package esxi

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
)

const (
	rootPool       = "ha-root-pool"
	basePool       = "Resources"
	digitsPattern  = "-?[0-9]+"
	uDigitsPattern = "[0-9]+"
)

func ResourcePoolCreate(inputs resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	var command string
	rp := parseResourcePool("", inputs)
	parentPool := basePool
	i := strings.LastIndex(rp.Name, "/")
	const lastSlashIndex = 2
	if i > lastSlashIndex {
		parentPool = rp.Name[:i]
		rp.Name = rp.Name[i+1:]
	}

	//  Check if already exists
	stdout, _ := esxi.getResourcePoolId(rp.Name)
	if stdout != "" {
		rp.Id = stdout
		return esxi.readResourcePool(rp)
	}

	command = fmt.Sprintf("--cpu-min=%d", rp.CpuMin)
	command = fmt.Sprintf("%s --cpu-min-expandable=%s", command, rp.CpuMinExpandable)
	if rp.CpuMax > 0 {
		command = fmt.Sprintf("%s --cpu-max=%d", command, rp.CpuMax)
	}
	if Contains([]string{"low", "normal", "high"}, rp.CpuShares) {
		command = fmt.Sprintf("%s --cpu-shares=%s", command, rp.CpuShares)
	} else {
		shares, _ := strconv.Atoi(rp.CpuShares)
		command = fmt.Sprintf("%s --cpu-shares=%d", command, shares)
	}
	command = fmt.Sprintf("%s --mem-min=%d", command, rp.MemMin)
	command = fmt.Sprintf("%s --mem-min-expandable=%s", command, rp.MemMinExpandable)
	if rp.MemMax > 0 {
		command = fmt.Sprintf("%s --mem-max=%d", command, rp.MemMax)
	}
	if Contains([]string{"low", "normal", "high"}, rp.MemShares) {
		command = fmt.Sprintf("%s --mem-shares=%s", command, rp.MemShares)
	} else {
		shares, _ := strconv.Atoi(rp.MemShares)
		command = fmt.Sprintf("%s --mem-shares=%d", command, shares)
	}

	parentPoolId, err := esxi.getResourcePoolId(parentPool)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get parent pool id: %w", err)
	}

	command = fmt.Sprintf("%s %s %s", command, parentPoolId, rp.Name)
	command = fmt.Sprintf("vim-cmd hostsvc/rsrc/create %s", command)

	stdout, err = esxi.Execute(command, "create resource pool")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create resource pool %s: %w", stdout, err)
	}

	id, err := esxi.getResourcePoolId(rp.Name)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get resource pool %s: %w", id, err)
	}

	rp.Id = id
	return esxi.readResourcePool(rp)
}

func ResourcePoolUpdate(id string, inputs resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	var command string
	rp := parseResourcePool(id, inputs)

	stdout, err := esxi.getResourcePoolName(rp.Id)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get resource pool name: %w", err)
	}
	if stdout != rp.Name {
		command = fmt.Sprintf("vim-cmd hostsvc/rsrc/rename %s %s", rp.Id, rp.Name)
		_, err = esxi.Execute(command, "update resource pool")
		if err != nil {
			return "", nil, fmt.Errorf("failed to update resource pool: %w", err)
		}
	}

	command = ""
	if rp.CpuMin > 0 {
		command = fmt.Sprintf("--cpu-min=%d", rp.CpuMin)
	}
	command = fmt.Sprintf("%s --cpu-min-expandable=%s", command, rp.CpuMinExpandable)
	if rp.CpuMax > 0 {
		command = fmt.Sprintf("%s --cpu-max=%d", command, rp.CpuMax)
	}
	if Contains([]string{"low", "normal", "high"}, rp.CpuShares) {
		command = fmt.Sprintf("%s --cpu-shares=%s", command, rp.CpuShares)
	} else {
		shares, _ := strconv.Atoi(rp.CpuShares)
		command = fmt.Sprintf("%s --cpu-shares=%d", command, shares)
	}
	if rp.MemMin > 0 {
		command = fmt.Sprintf("%s --mem-min=%d", command, rp.MemMin)
	}
	command = fmt.Sprintf("%s --mem-min-expandable=%s", command, rp.MemMinExpandable)
	if rp.MemMax > 0 {
		command = fmt.Sprintf("%s --mem-max=%d", command, rp.MemMax)
	}
	if Contains([]string{"low", "normal", "high"}, rp.MemShares) {
		command = fmt.Sprintf("%s --mem-shares=%s", command, rp.MemShares)
	} else {
		shares, _ := strconv.Atoi(rp.MemShares)
		command = fmt.Sprintf("%s --mem-shares=%d", command, shares)
	}

	command = fmt.Sprintf("%s %s", command, rp.Id)
	command = fmt.Sprintf("vim-cmd hostsvc/rsrc/pool_config_set %s", command)

	stdout, err = esxi.Execute(command, "update resource pool")
	r := strings.NewReplacer("'vim.ResourcePool:", "", "'", "")
	stdout = r.Replace(stdout)
	if err != nil {
		return "", nil, fmt.Errorf("failed to update resource pool %s: %w", stdout, err)
	}

	return esxi.readResourcePool(rp)
}

func ResourcePoolDelete(id string, esxi *Host) error {
	command := fmt.Sprintf("vim-cmd hostsvc/rsrc/destroy %s", id)

	stdout, err := esxi.Execute(command, "delete resource pool")
	if err != nil {
		return fmt.Errorf("failed to delete resource pool: %s err: %w", stdout, err)
	}

	return nil
}

func ResourcePoolRead(id string, inputs resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	rp := parseResourcePool(id, inputs)
	return esxi.readResourcePool(rp)
}

func parseResourcePool(id string, inputs resource.PropertyMap) ResourcePool {
	rp := ResourcePool{}

	if len(id) > 0 {
		rp.Id = id
	}

	rp.Name = inputs["name"].StringValue()
	if rp.Name == string('/') {
		rp.Name = basePool
	}
	if rp.Name[0] == '/' {
		rp.Name = rp.Name[1:]
	}

	if property, has := inputs["cpuMin"]; has {
		rp.CpuMin = int(property.NumberValue())
	} else {
		rp.CpuMin = 100
	}
	if property, has := inputs["cpuMinExpandable"]; has {
		rp.CpuMinExpandable = property.StringValue()
	} else {
		rp.CpuMinExpandable = trueValue
	}
	if property, has := inputs["cpuMax"]; has {
		rp.CpuMax = int(property.NumberValue())
	} else {
		rp.CpuMax = 0
	}
	if property, has := inputs["cpuShares"]; has {
		rp.CpuShares = strings.ToLower(property.StringValue())
	} else {
		rp.CpuShares = "normal"
	}
	if property, has := inputs["memMin"]; has {
		rp.MemMin = int(property.NumberValue())
	} else {
		rp.MemMin = 200
	}
	if property, has := inputs["memMinExpandable"]; has {
		rp.MemMinExpandable = property.StringValue()
	} else {
		rp.MemMinExpandable = trueValue
	}
	if property, has := inputs["memMax"]; has {
		rp.MemMax = int(property.NumberValue())
	} else {
		rp.MemMax = 0
	}
	if property, has := inputs["memShares"]; has {
		rp.MemShares = strings.ToLower(property.StringValue())
	} else {
		rp.MemShares = "normal"
	}

	return rp
}

func (esxi *Host) readResourcePool(rp ResourcePool) (string, resource.PropertyMap, error) {
	rp, err := esxi.getResourcePoolDetails(rp)
	if err != nil {
		return "", nil, err
	}

	result := rp.toMap()
	return rp.Id, resource.NewPropertyMapFromMap(result), nil
}

func (esxi *Host) getResourcePoolId(name string) (string, error) {
	if name == "/" || name == basePool {
		return rootPool, nil
	}

	result := strings.Split(name, "/")
	name = result[len(result)-1]

	r := strings.NewReplacer("objID>", "", "</objID", "")
	command := fmt.Sprintf("grep -A1 '<name>%s</name>' /etc/vmware/hostd/pools.xml | grep -m 1 -o objID.*objID", name)
	stdout, err := esxi.Execute(command, "get existing resource pool id")
	if err != nil {
		logging.V(logLevel).Infof("getResourcePoolName: Failed get existing resource pool id => %s", stdout)
		return "", fmt.Errorf("failed to get existing resource pool id: %w", err)
	} else {
		stdout = r.Replace(stdout)
		return stdout, nil
	}
}

func (esxi *Host) getResourcePoolName(id string) (string, error) {
	var resourcePoolName, fullResourcePoolName string

	fullResourcePoolName = ""

	if id == rootPool {
		return "/", nil
	}

	// Get full Resource Pool Path
	command := fmt.Sprintf("grep -A1 '<objID>%s</objID>' /etc/vmware/hostd/pools.xml | grep '<path>'", id)
	stdout, err := esxi.Execute(command, "get resource pool path")
	if err != nil {
		logging.V(logLevel).Infof("getResourcePoolName: Failed get resource pool PATH => %s", stdout)
		return "", fmt.Errorf("failed to get pool path: %w", err)
	}

	re := regexp.MustCompile(`[/<>\n]`)
	result := re.Split(stdout, -1)

	for i := range result {
		if result[i] != "path" && result[i] != "host" && result[i] != "user" && result[i] != "" {
			r := strings.NewReplacer("name>", "", "</name", "")
			command = fmt.Sprintf("grep -B1 '<objID>%s</objID>' /etc/vmware/hostd/pools.xml | grep -o name.*name", result[i])
			stdout, _ = esxi.Execute(command, "get resource pool name")
			resourcePoolName = r.Replace(stdout)

			if resourcePoolName != "" {
				if result[i] == id {
					fullResourcePoolName += resourcePoolName
				} else {
					fullResourcePoolName = fullResourcePoolName + resourcePoolName + "/"
				}
			}
		}
	}

	return fullResourcePoolName, nil
}

func (esxi *Host) getResourcePoolDetails(rp ResourcePool) (ResourcePool, error) {
	// Get full Resource Pool Path
	command := fmt.Sprintf("vim-cmd hostsvc/rsrc/pool_config_get %s", rp.Id)
	stdout, err := esxi.Execute(command, "get resource pool config")
	if strings.Contains(stdout, "deleted") {
		return rp, err
	}
	if err != nil {
		return rp, fmt.Errorf("failed to get resource pool config: %w", err)
	}

	isCpuFlag := true

	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		switch {
		case strings.Contains(scanner.Text(), "memoryAllocation = "):
			isCpuFlag = false

		case strings.Contains(scanner.Text(), "reservation = "):
			r := regexp.MustCompile(uDigitsPattern)
			if isCpuFlag {
				rp.CpuMin, _ = strconv.Atoi(r.FindString(scanner.Text()))
			} else {
				rp.MemMin, _ = strconv.Atoi(r.FindString(scanner.Text()))
			}

		case strings.Contains(scanner.Text(), "expandableReservation = "):
			r := regexp.MustCompile("(true|false)")
			if isCpuFlag {
				rp.CpuMinExpandable = r.FindString(scanner.Text())
			} else {
				rp.MemMinExpandable = r.FindString(scanner.Text())
			}

		case strings.Contains(scanner.Text(), "limit = "):
			r := regexp.MustCompile(digitsPattern)
			tmpVar, _ := strconv.Atoi(r.FindString(scanner.Text()))
			if tmpVar < 0 {
				tmpVar = 0
			}
			if isCpuFlag {
				rp.CpuMax = tmpVar
			} else {
				rp.MemMax = tmpVar
			}

		case strings.Contains(scanner.Text(), "shares = "):
			r := regexp.MustCompile(uDigitsPattern)
			if isCpuFlag {
				rp.CpuShares = r.FindString(scanner.Text())
			} else {
				rp.MemShares = r.FindString(scanner.Text())
			}

		case strings.Contains(scanner.Text(), "level = "):
			r := regexp.MustCompile("(low|high|normal)")
			match := r.FindString(scanner.Text())
			if match != "" {
				if isCpuFlag {
					rp.CpuShares = match
				} else {
					rp.MemShares = match
				}
			}
		}
	}

	rp.Name, err = esxi.getResourcePoolName(rp.Id)
	if err != nil {
		return rp, fmt.Errorf("failed to get pool name: %w", err)
	}

	return rp, nil
}

func (rp *ResourcePool) toMap(keepId ...bool) map[string]interface{} {
	outputs := structToMap(rp)
	if len(keepId) != 0 && !keepId[0] {
		delete(outputs, "id")
	}
	return outputs
}
