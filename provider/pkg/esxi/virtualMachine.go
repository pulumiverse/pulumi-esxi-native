package esxi

import (
	"bufio"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
)

func VirtualMachineGet(inputs resource.PropertyMap, esxi *Host) (resource.PropertyMap, error) {
	var id string
	if nameProp, has := inputs["name"]; has {
		var err error
		id, err = esxi.getVirtualMachineId(nameProp.StringValue())
		if err != nil {
			return nil, err
		}
	} else if idProp, has := inputs["id"]; has {
		id = idProp.StringValue()
	}

	vm := esxi.readVirtualMachine(VirtualMachine{
		Id:             id,
		StartupTimeout: vmDefaultStartupTimeout,
	})

	if len(vm.Name) == 0 {
		return nil, fmt.Errorf("unable to find a virtual machine corresponding to the id '%s'", id)
	}

	result := vm.toMap(true)
	return resource.NewPropertyMapFromMap(result), nil
}

func VirtualMachineRead(id string, _ resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	// read vm
	vm := esxi.readVirtualMachine(VirtualMachine{
		Id:             id,
		StartupTimeout: vmDefaultStartupTimeout,
	})

	if len(vm.Name) == 0 {
		return "", nil, fmt.Errorf("unable to find a virtual machine corresponding to the id '%s'", id)
	}

	result := vm.toMap()
	return vm.Id, resource.NewPropertyMapFromMap(result), nil
}

func VirtualMachineCreate(inputs resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	vm := parseVirtualMachine("", inputs, esxi.Connection)

	powerOn := vm.Power == vmTurnedOn || vm.Power == ""
	vm, err := esxi.createVirtualMachine(vm)
	if err != nil {
		return "", nil, err
	}
	if powerOn {
		err = esxi.powerOnVirtualMachine(vm.Id)
		if err != nil {
			return "", nil, fmt.Errorf("failed to power on the virtual machine")
		}
		vm.Power = vmTurnedOn
	}

	// read vm
	vm = esxi.readVirtualMachine(vm)

	result := vm.toMap()
	return vm.Id, resource.NewPropertyMapFromMap(result), nil
}

func VirtualMachineUpdate(id string, inputs resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	vm := parseVirtualMachine(id, inputs, esxi.Connection)

	currentPowerState := esxi.getVirtualMachinePowerState(vm.Id)
	if currentPowerState == vmTurnedOn || currentPowerState == vmTurnedSuspended {
		esxi.powerOffVirtualMachine(vm.Id, vm.ShutdownTimeout)
	}

	// make updates to vmx file
	err := esxi.updateVmxContents(true, vm)
	if err != nil {
		return id, nil, fmt.Errorf("failed to update vmx contents: %w", err)
	}

	// Grow boot disk
	bootDiskVmdkPath, _ := esxi.getBootDiskPath(vm.Id)

	didGrow, err := esxi.growVirtualDisk(bootDiskVmdkPath, vm.BootDiskSize)
	if err != nil {
		return id, nil, fmt.Errorf("failed to grow boot disk: %w", err)
	}
	if didGrow {
		_ = esxi.reloadVirtualMachine(id)
	}
	//  power on
	if vm.Power == vmTurnedOn {
		err = esxi.powerOnVirtualMachine(id)
		if err != nil {
			return id, nil, fmt.Errorf("failed to power on: %w", err)
		}
	}

	result := vm.toMap()
	return vm.Id, resource.NewPropertyMapFromMap(result), nil
}

func VirtualMachineDelete(id string, esxi *Host) error {
	var command, stdout string
	var err error

	esxi.powerOffVirtualMachine(id, vmDefaultShutdownTimeout)

	// remove storage from vmx so it doesn't get deleted by the vim-cmd destroy
	err = esxi.cleanStorageFromVmx(id)
	if err != nil {
		logging.V(logLevel).Infof("VirtualMachineDelete: failed clean storage from id: %s (to be deleted)", id)
	}

	const waitTime = 5
	time.Sleep(waitTime * time.Second)
	command = fmt.Sprintf("vim-cmd vmsvc/destroy %s", id)
	stdout, err = esxi.Execute(command, "vmsvc/destroy")
	if err != nil {
		logging.V(logLevel).Infof("VirtualMachineDelete: failed to destroy vm: %s", stdout)
		return fmt.Errorf("failed to destroy vm: %w", err)
	}

	return nil
}

func parseVirtualMachine(id string, inputs resource.PropertyMap, connection *ConnectionInfo) VirtualMachine {
	vm := VirtualMachine{}

	if len(id) > 0 {
		vm.Id = id
	}

	vm.Name = inputs["name"].StringValue()
	vm.SourcePath = parseSourcePath(inputs, connection)
	vm.BootFirmware = parseStringProperty(inputs, "bootFirmware", "bios")
	vm.DiskStore = inputs["diskStore"].StringValue()
	vm.ResourcePoolName = parseStringProperty(inputs, "resourcePoolName", "/")
	if vm.ResourcePoolName == rootPool {
		vm.ResourcePoolName = "/"
	}

	vm.BootDiskSize = parseIntProperty(inputs, "bootDiskSize", vmDefaultBootDiskSize)
	vm.BootDiskType = parseStringProperty(inputs, "bootDiskType", vdThin)
	vm.MemSize = parseIntProperty(inputs, "memSize", vmDefaultMemSize)
	vm.NumVCpus = parseIntProperty(inputs, "numVCpus", vmDefaultNumVCpus)
	vm.VirtualHWVer = parseIntProperty(inputs, "virtualHWVer", vmDefaultVirtualHWVer)
	vm.NetworkInterfaces = parseNetworkInterfaces(inputs)
	vm.Os = parseStringProperty(inputs, "os", vmDefaultOs)
	vm.Power = parseStringProperty(inputs, "power", vmTurnedOn)
	vm.StartupTimeout = parseIntProperty(inputs, "startupTimeout", vmDefaultStartupTimeout)
	vm.ShutdownTimeout = parseIntProperty(inputs, "shutdownTimeout", vmDefaultShutdownTimeout)
	vm.VirtualDisks = parseVirtualDisks(inputs)
	vm.OvfProperties = parseKeyValuePairsProperty(inputs, "ovfProperties")
	vm.Notes = parseStringProperty(inputs, "notes", "")
	vm.Info = parseKeyValuePairsProperty(inputs, "info")

	return vm
}

func parseSourcePath(inputs resource.PropertyMap, connection *ConnectionInfo) string {
	if property, has := inputs["cloneFromVirtualMachine"]; has {
		password := url.QueryEscape(connection.Password)
		return fmt.Sprintf("vi://%s:%s@%s:%s/%s", connection.UserName, password, connection.Host, connection.SslPort, property.StringValue())
	}
	if property, has := inputs["ovfSource"]; has {
		return property.StringValue()
	}
	return "none"
}

func parseStringProperty(inputs resource.PropertyMap, key string, defaultValue string) string {
	if property, has := inputs[resource.PropertyKey(key)]; has {
		if property.IsSecret() {
			return property.SecretValue().Element.StringValue()
		}
		return property.StringValue()
	}
	return defaultValue
}

func parseIntProperty(inputs resource.PropertyMap, key string, defaultValue int) int {
	if property, has := inputs[resource.PropertyKey(key)]; has {
		return int(property.NumberValue())
	}
	return defaultValue
}

func parseNetworkInterfaces(inputs resource.PropertyMap) []NetworkInterface {
	if property, has := inputs["networkInterfaces"]; has {
		if items := property.ArrayValue(); len(items) > 0 {
			interfaces := make([]NetworkInterface, len(items))
			for i, item := range items {
				interfaces[i] = NetworkInterface{
					VirtualNetwork: parseStringProperty(item.ObjectValue(), "virtualNetwork", ""),
					MacAddress:     parseStringProperty(item.ObjectValue(), "macAddress", ""),
					NicType:        parseStringProperty(item.ObjectValue(), "nicType", ""),
				}
			}
			return interfaces
		}
	}
	return []NetworkInterface{}
}

func parseVirtualDisks(inputs resource.PropertyMap) []VMVirtualDisk {
	if property, has := inputs["virtualDisks"]; has {
		if items := property.ArrayValue(); len(items) > 0 {
			virtualDisks := make([]VMVirtualDisk, len(items))
			for i, item := range items {
				virtualDisks[i] = VMVirtualDisk{
					VirtualDiskId: parseStringProperty(item.ObjectValue(), "virtualDiskId", ""),
					Slot:          parseStringProperty(item.ObjectValue(), "slot", ""),
				}
			}
			return virtualDisks
		}
	}
	return []VMVirtualDisk{}
}

func parseKeyValuePairsProperty(inputs resource.PropertyMap, key string) []KeyValuePair {
	if property, has := inputs[resource.PropertyKey(key)]; has {
		if items := property.ArrayValue(); len(items) > 0 {
			properties := make([]KeyValuePair, len(items))
			for i, item := range items {
				properties[i] = KeyValuePair{
					Key:   parseStringProperty(item.ObjectValue(), "key", ""),
					Value: parseStringProperty(item.ObjectValue(), "value", ""),
				}
			}
			return properties
		}
	}
	return []KeyValuePair{}
}

func (esxi *Host) readVirtualMachine(vm VirtualMachine) VirtualMachine {
	var err error
	command := fmt.Sprintf("vim-cmd  vmsvc/get.summary %s", vm.Id)
	stdout, _ := esxi.Execute(command, "Get Guest summary")

	if strings.Contains(stdout, "Unable to find a VM corresponding") {
		vm = VirtualMachine{Name: ""}
		return vm
	}

	vm.patchWithSummary(stdout)

	//  Get resource pool that this VM is located
	vmResourcePoolId := esxi.getVMResourcePoolId(vm)
	vm.ResourcePoolName, err = esxi.getResourcePoolName(vmResourcePoolId)
	logging.V(logLevel).Infof("readVirtualMachine: resource_pool_name|%s| scanner.Text() => |%s|", vmResourcePoolId, err)

	vmxContents := esxi.readVMXContents(vm)

	vm.patchWithVMXContents(vmxContents)

	//  Get power state
	vm.Power = esxi.getVirtualMachinePowerState(vm.Id)
	logging.V(logLevel).Infof("readVirtualMachine: Power => %s", vm.Power)

	// Get IP address (need vmware tools installed)
	if vm.Power == vmTurnedOn {
		vm.IpAddress = esxi.getVirtualMachineIpAddress(vm.Id, vm.StartupTimeout)
		logging.V(logLevel).Infof("readVirtualMachine: IpAddress found => %s", vm.IpAddress)
	} else {
		vm.IpAddress = ""
	}

	// Get boot disk size
	bootDiskPath, _ := esxi.getBootDiskPath(vm.Id)
	vd, _ := esxi.getVirtualDisk(bootDiskPath)
	vm.BootDiskSize = vd.Size
	vm.BootDiskType = vd.DiskType

	// Get Info
	vm.Info = extractGuestInfo(vmxContents)

	return vm
}

func (esxi *Host) getVMResourcePoolId(vm VirtualMachine) string {
	command := fmt.Sprintf(`grep -A2 'objID>%s</objID' /etc/vmware/hostd/pools.xml | grep -o resourcePool.*resourcePool`, vm.Id)
	stdout, _ := esxi.Execute(command, "check if guest is in resource pool")
	nr := strings.NewReplacer("resourcePool>", "", "</resourcePool", "")
	vmResourcePoolId := nr.Replace(stdout)
	logging.V(logLevel).Infof("readVirtualMachine: resource_pool_name|%s| scanner.Text() => |%s|", vmResourcePoolId, stdout)
	return vmResourcePoolId
}

func (esxi *Host) readVMXContents(vm VirtualMachine) string {
	// Implement reading VMX contents from the ESXi host
	command := fmt.Sprintf("vim-cmd vmsvc/get.config %s | grep vmPathName|grep -oE \"\\[.*\\]\"", vm.Id)
	stdout, _ := esxi.Execute(command, "get dst_vmx_ds")
	dstVmxDs := stdout
	dstVmxDs = strings.Trim(dstVmxDs, "[")
	dstVmxDs = strings.Trim(dstVmxDs, "]")

	command = fmt.Sprintf("vim-cmd vmsvc/get.config %s | grep vmPathName|awk '{print $NF}'|sed 's/[\"|,]//g'", vm.Id)
	stdout, _ = esxi.Execute(command, "get dst_vmx")
	dstVmx := stdout

	dstVmxFile := fmt.Sprintf("/vmfs/volumes/%s/%s", dstVmxDs, dstVmx)

	logging.V(logLevel).Infof("readVirtualMachine: dstVmxFile => %s", dstVmxFile)
	logging.V(logLevel).Infof("readVirtualMachine: vm.DiskStore => %s  dstVmxDs => %s", vm.DiskStore, dstVmxDs)

	command = fmt.Sprintf("cat \"%s\"", dstVmxFile)
	vmxContents, _ := esxi.Execute(command, "read guest_name.vmx file")
	return vmxContents
}

func (vm *VirtualMachine) patchWithSummary(summary string) {
	// Implement parsing VM info from summary
	scanner := bufio.NewScanner(strings.NewReader(summary))
	for scanner.Scan() {
		switch {
		case strings.Contains(scanner.Text(), "name = "):
			r := regexp.MustCompile("\".*\"")
			vm.Name = r.FindString(scanner.Text())
			nr := strings.NewReplacer("\"", "", "\"", "")
			vm.Name = nr.Replace(vm.Name)
		case strings.Contains(scanner.Text(), "vmPathName = "):
			r := regexp.MustCompile(`\[.*]`)
			vm.DiskStore = r.FindString(scanner.Text())
			nr := strings.NewReplacer("[", "", "]", "")
			vm.DiskStore = nr.Replace(vm.DiskStore)
		}
	}
}

func (vm *VirtualMachine) patchWithVMXContents(vmxContents string) {
	vm.VirtualDisks = []VMVirtualDisk{}

	// Used to keep track if a network interface is using static or generated macs.
	const interfacesCount = 10
	var isGeneratedMAC [interfacesCount]bool
	networkInterfaces := make([]NetworkInterface, interfacesCount)

	r := regexp.MustCompile("\".*\"")
	//  Read vmxContents line-by-line to get current settings.
	scanner := bufio.NewScanner(strings.NewReader(vmxContents))
	for scanner.Scan() {
		switch {
		case strings.Contains(scanner.Text(), "memSize = "):
			stdout := r.FindString(scanner.Text())
			nr := strings.NewReplacer(`"`, "", `"`, "")
			vm.MemSize, _ = strconv.Atoi(nr.Replace(stdout))
			logging.V(logLevel).Infof("readVirtualMachine: MemSize found => %d", vm.MemSize)

		case strings.Contains(scanner.Text(), "numvcpus = "):
			stdout := r.FindString(scanner.Text())
			nr := strings.NewReplacer(`"`, "", `"`, "")
			vm.NumVCpus, _ = strconv.Atoi(nr.Replace(stdout))
			logging.V(logLevel).Infof("readVirtualMachine: NumVCpus found => %d", vm.NumVCpus)

		case strings.Contains(scanner.Text(), "numa.autosize.vcpu."):
			stdout := r.FindString(scanner.Text())
			nr := strings.NewReplacer(`"`, "", `"`, "")
			vm.NumVCpus, _ = strconv.Atoi(nr.Replace(stdout))
			logging.V(logLevel).Infof("readVirtualMachine: numa.vcpu (NumVCpus) found => %d", vm.NumVCpus)

		case strings.Contains(scanner.Text(), "virtualHW.version = "):
			stdout := r.FindString(scanner.Text())
			vm.VirtualHWVer, _ = strconv.Atoi(strings.ReplaceAll(stdout, `"`, ""))
			logging.V(logLevel).Infof("readVirtualMachine: VirtualHWVer found => %d", vm.VirtualHWVer)

		case strings.Contains(scanner.Text(), "guestOS = "):
			stdout := r.FindString(scanner.Text())
			vm.Os = strings.ReplaceAll(stdout, `"`, "")
			logging.V(logLevel).Infof("readVirtualMachine: Os found => %s", vm.Os)

		case strings.Contains(scanner.Text(), "scsi"):
			re := regexp.MustCompile("scsi([0-3]):([0-9]{1,2}).(.*) = \"(.*)\"")
			results := re.FindStringSubmatch(scanner.Text())
			const scsiParts = 4
			if len(results) > scsiParts {
				logging.V(logLevel).Infof("readVirtualMachine: %s : %s . %s = %s", results[1], results[2], results[3], results[4])

				if (results[1] == "0") && (results[2] == "0") {
					// Skip boot disk
				} else if strings.Contains(results[3], "fileName") {
					logging.V(logLevel).Infof("readVirtualMachine: %s => %s", results[0], results[4])

					vm.VirtualDisks = append(vm.VirtualDisks, VMVirtualDisk{
						fmt.Sprintf("%s:%s", results[1], results[2]),
						results[4],
					})
				}
			}

		case strings.Contains(scanner.Text(), "ethernet"):
			re := regexp.MustCompile("ethernet(.).(.*) = \"(.*)\"")
			results := re.FindStringSubmatch(scanner.Text())
			index, _ := strconv.Atoi(results[1])

			switch results[2] {
			case "networkName":
				networkInterfaces[index].VirtualNetwork = results[3]
				logging.V(logLevel).Infof("readVirtualMachine: %s => %s", results[0], results[3])

			case "addressType":
				if results[3] == "generated" {
					isGeneratedMAC[index] = true
				}

			//  Done't save generatedAddress...   It should not be saved because it
			//  should be considered dynamic & is breaks the update MAC address code.
			//  case "generatedAddress":
			//	  if isGeneratedMAC[index] == true {
			//		networkInterfaces[index].MacAddress = results[3]
			//	    logging.V(logLevel).Infof("readVirtualMachine: %s => %s", results[0], results[3])
			//	  }

			case "address":
				if !isGeneratedMAC[index] {
					networkInterfaces[index].MacAddress = results[3]
					logging.V(logLevel).Infof("readVirtualMachine: %s => %s", results[0], results[3])
				}

			case "virtualDev":
				networkInterfaces[index].NicType = results[3]
				logging.V(logLevel).Infof("readVirtualMachine: %s => %s", results[0], results[3])
			}

		case strings.Contains(scanner.Text(), "firmware = "):
			stdout := r.FindString(scanner.Text())
			vm.BootFirmware = strings.ReplaceAll(stdout, `"`, "")
			logging.V(logLevel).Infof("readVirtualMachine: BootFirmware found => %s", vm.BootFirmware)

		case strings.Contains(scanner.Text(), "annotation = "):
			stdout := r.FindString(scanner.Text())
			vm.Notes = strings.ReplaceAll(stdout, `"`, "")
			vm.Notes = strings.ReplaceAll(vm.Notes, "|22", "\"")
			logging.V(logLevel).Infof("readVirtualMachine: Notes found => %s", vm.Notes)
		}
	}

	for _, ni := range networkInterfaces {
		if len(ni.VirtualNetwork) > 0 {
			vm.NetworkInterfaces = append(vm.NetworkInterfaces, ni)
		}
	}
}

func extractGuestInfo(vmxContents string) []KeyValuePair {
	parsedVmx := ParseVMX(vmxContents)
	// Get guestinfo value
	info := make(map[string]string)
	for key, value := range parsedVmx {
		if strings.Contains(key, "guestinfo") {
			shortKey := strings.ReplaceAll(key, "guestinfo.", "")
			info[shortKey] = value
		}
	}
	i := 0
	infoProperties := make([]KeyValuePair, len(info))
	for key, value := range info {
		infoProperties[i] = KeyValuePair{key, value}
		i++
	}
	return infoProperties
}
