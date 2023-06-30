package esxi

import (
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
	"regexp"
	"strings"
)

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
			command := fmt.Sprintf("grep -B1 '<objID>%s</objID>' /etc/vmware/hostd/pools.xml | grep -o name.*name", result[i])
			stdout, _ := esxi.Execute(command, "get resource pool name")
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
