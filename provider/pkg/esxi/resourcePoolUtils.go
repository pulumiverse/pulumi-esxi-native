package esxi

import (
	"bufio"
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
	"regexp"
	"strconv"
	"strings"
)

func (esxi *Host) getResourcePoolId(name string) (string, error) {
	if name == "/" || name == "Resources" {
		return "ha-root-pool", nil
	}

	result := strings.Split(name, "/")
	name = result[len(result)-1]

	r := strings.NewReplacer("objID>", "", "</objID", "")
	command := fmt.Sprintf("grep -A1 '<name>%s</name>' /etc/vmware/hostd/pools.xml | grep -m 1 -o objID.*objID", name)
	stdout, err := esxi.Execute(command, "get existing resource pool id")
	if err != nil {
		logging.V(9).Infof("getResourcePoolName: Failed get existing resource pool id => %s", stdout)
		return "", fmt.Errorf("failed to get existing resource pool id: %s", err)
	} else {
		stdout = r.Replace(stdout)
		return stdout, nil
	}
}

func (esxi *Host) getResourcePoolName(id string) (string, error) {
	var resourcePoolName, fullResourcePoolName string

	fullResourcePoolName = ""

	if id == "ha-root-pool" {
		return "/", nil
	}

	// Get full Resource Pool Path
	command := fmt.Sprintf("grep -A1 '<objID>%s</objID>' /etc/vmware/hostd/pools.xml | grep '<path>'", id)
	stdout, err := esxi.Execute(command, "get resource pool path")
	if err != nil {
		logging.V(9).Infof("getResourcePoolName: Failed get resource pool PATH => %s", stdout)
		return "", fmt.Errorf("Failed to get pool path: %s\n", err)
	}

	re := regexp.MustCompile(`[/<>\n]`)
	result := re.Split(stdout, -1)

	for i := range result {

		resourcePoolName = ""
		if result[i] != "path" && result[i] != "host" && result[i] != "user" && result[i] != "" {

			r := strings.NewReplacer("name>", "", "</name", "")
			command = fmt.Sprintf("grep -B1 '<objID>%s</objID>' /etc/vmware/hostd/pools.xml | grep -o name.*name", result[i])
			stdout, _ = esxi.Execute(command, "get resource pool name")
			resourcePoolName = r.Replace(stdout)

			if resourcePoolName != "" {
				if result[i] == id {
					fullResourcePoolName = fullResourcePoolName + resourcePoolName
				} else {
					fullResourcePoolName = fullResourcePoolName + resourcePoolName + "/"
				}
			}
		}
	}

	return fullResourcePoolName, nil
}

func (esxi *Host) readResourcePool(rp ResourcePool) (ResourcePool, error) {
	// Get full Resource Pool Path
	command := fmt.Sprintf("vim-cmd hostsvc/rsrc/pool_config_get %s", rp.Id)
	stdout, err := esxi.Execute(command, "get resource pool config")
	if strings.Contains(stdout, "deleted") == true {
		return rp, err
	}
	if err != nil {
		return rp, fmt.Errorf("failed to get resource pool config: %s", err)
	}

	isCpuFlag := true

	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		switch {
		case strings.Contains(scanner.Text(), "memoryAllocation = "):
			isCpuFlag = false

		case strings.Contains(scanner.Text(), "reservation = "):
			r, _ := regexp.Compile("[0-9]+")
			if isCpuFlag == true {
				rp.CpuMin, _ = strconv.Atoi(r.FindString(scanner.Text()))
			} else {
				rp.MemMin, _ = strconv.Atoi(r.FindString(scanner.Text()))
			}

		case strings.Contains(scanner.Text(), "expandableReservation = "):
			r, _ := regexp.Compile("(true|false)")
			if isCpuFlag == true {
				rp.CpuMinExpandable = r.FindString(scanner.Text())
			} else {
				rp.MemMinExpandable = r.FindString(scanner.Text())
			}

		case strings.Contains(scanner.Text(), "limit = "):
			r, _ := regexp.Compile("-?[0-9]+")
			tmpvar, _ := strconv.Atoi(r.FindString(scanner.Text()))
			if tmpvar < 0 {
				tmpvar = 0
			}
			if isCpuFlag == true {
				rp.CpuMax = tmpvar
			} else {
				rp.MemMax = tmpvar
			}

		case strings.Contains(scanner.Text(), "shares = "):
			r, _ := regexp.Compile("[0-9]+")
			if isCpuFlag == true {
				rp.CpuShares = r.FindString(scanner.Text())
			} else {
				rp.MemShares = r.FindString(scanner.Text())
			}

		case strings.Contains(scanner.Text(), "level = "):
			r, _ := regexp.Compile("(low|high|normal)")
			if r.FindString(scanner.Text()) != "" {
				if isCpuFlag == true {
					rp.CpuShares = r.FindString(scanner.Text())
				} else {
					rp.MemShares = r.FindString(scanner.Text())
				}
			}
		}
	}

	rp.Name, err = esxi.getResourcePoolName(rp.Id)
	if err != nil {
		return rp, fmt.Errorf("failed to get pool name: %s", err)
	}

	return rp, nil
}
