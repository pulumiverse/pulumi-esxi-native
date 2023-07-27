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
		StartupTimeout: 1,
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
		StartupTimeout: 1,
	})

	if len(vm.Name) == 0 {
		return "", nil, fmt.Errorf("unable to find a virtual machine corresponding to the id '%s'", id)
	}

	result := vm.toMap()
	return vm.Id, resource.NewPropertyMapFromMap(result), nil
}

func VirtualMachineCreate(inputs resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	var vm VirtualMachine
	if parsed, err := parseVirtualMachine("", inputs, esxi.Connection); err == nil {
		vm = parsed
	} else {
		return "", nil, err
	}
	powerOn := vm.Power == "on" || vm.Power == ""
	vm, err := esxi.createVirtualMachine(vm)
	if err != nil {
		return "", nil, err
	}
	if powerOn {
		_, err = esxi.powerOnVirtualMachine(vm.Id)
		if err != nil {
			return "", nil, fmt.Errorf("failed to power on the virtual machine")
		}
		vm.Power = "on"
	}

	result := vm.toMap()
	return vm.Id, resource.NewPropertyMapFromMap(result), nil
}

func VirtualMachineUpdate(id string, inputs resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	var vm VirtualMachine
	if parsed, err := parseVirtualMachine(id, inputs, esxi.Connection); err == nil {
		vm = parsed
	} else {
		return id, nil, err
	}

	currentPowerState := esxi.getVirtualMachinePowerState(vm.Id)
	if currentPowerState == "on" || currentPowerState == "suspended" {
		_, err := esxi.powerOffVirtualMachine(vm.Id, vm.ShutdownTimeout)
		if err != nil {
			return id, nil, fmt.Errorf("failed to shutdown %s", err)
		}
	}

	// make updates to vmx file
	err := esxi.updateVmxContents(true, vm)
	if err != nil {
		return id, nil, fmt.Errorf("failed to update vmx contents: %s", err)
	}

	// Grow boot disk
	bootDiskVmdkPath, _ := esxi.getBootDiskPath(vm.Id)

	didGrow, err := esxi.growVirtualDisk(bootDiskVmdkPath, vm.BootDiskSize)
	if err != nil {
		return id, nil, fmt.Errorf("failed to grow boot disk: %s", err)
	}
	if didGrow {
		err = esxi.reloadVirtualMachine(id)
	}
	//  power on
	if vm.Power == "on" {
		_, err = esxi.powerOnVirtualMachine(id)
		if err != nil {
			return id, nil, fmt.Errorf("failed to power on: %s", err)
		}
	}

	result := vm.toMap()
	return vm.Id, resource.NewPropertyMapFromMap(result), nil
}

func VirtualMachineDelete(id string, esxi *Host) error {
	var command, stdout string
	var err error

	_, err = esxi.powerOffVirtualMachine(id, 30)
	if err != nil {
		return fmt.Errorf("failed to power off: %s", err)
	}

	// remove storage from vmx so it doesn't get deleted by the vim-cmd destroy
	err = esxi.cleanStorageFromVmx(id)
	if err != nil {
		logging.V(9).Infof("VirtualMachineDelete: failed clean storage from id: %s (to be deleted)", id)
	}

	time.Sleep(5 * time.Second)
	command = fmt.Sprintf("vim-cmd vmsvc/destroy %s", id)
	stdout, err = esxi.Execute(command, "vmsvc/destroy")
	if err != nil {
		logging.V(9).Infof("VirtualMachineDelete: failed to destroy vm: %s", stdout)
		return fmt.Errorf("failed to destroy vm: %s", err)
	}

	return nil
}

func parseVirtualMachine(id string, inputs resource.PropertyMap, connection *ConnectionInfo) (VirtualMachine, error) {
	vm := VirtualMachine{}

	if len(id) > 0 {
		vm.Id = id
	}

	vm.Name = inputs["name"].StringValue()

	if property, has := inputs["cloneFromVirtualMachine"]; has {
		password := url.QueryEscape(connection.Password)
		vm.SourcePath = fmt.Sprintf("vi://%s:%s@%s:%s/%s", connection.UserName, password, connection.Host, connection.SslPort, property.StringValue())
	} else if property, has = inputs["ovfLocalSource"]; has {
		vm.SourcePath = fmt.Sprintf("local://%s", property.StringValue())
	} else if property, has = inputs["ovfSource"]; has {
		vm.SourcePath = property.StringValue()
	} else {
		vm.SourcePath = "none"
	}
	if property, has := inputs["bootFirmware"]; has {
		vm.BootFirmware = property.StringValue()
	} else {
		vm.BootFirmware = ""
	}
	vm.DiskStore = inputs["diskStore"].StringValue()
	if property, has := inputs["resourcePoolName"]; has {
		vm.ResourcePoolName = property.StringValue()
		if vm.ResourcePoolName == "ha-root-pool" {
			vm.ResourcePoolName = "/"
		}
	} else {
		vm.ResourcePoolName = "/"
	}
	if property, has := inputs["bootDiskSize"]; has {
		vm.BootDiskSize = int(property.NumberValue())
	} else {
		vm.BootDiskSize = 16
	}
	if property, has := inputs["bootDiskType"]; has {
		vm.BootDiskType = property.StringValue()
	} else {
		vm.BootDiskType = "thin"
	}
	if property, has := inputs["memSize"]; has {
		vm.MemSize = int(property.NumberValue())
	} else {
		vm.MemSize = 512
	}
	if property, has := inputs["numVCpus"]; has {
		vm.NumVCpus = int(property.NumberValue())
	} else {
		vm.NumVCpus = 1
	}
	if property, has := inputs["virtualHWVer"]; has {
		vm.VirtualHWVer = int(property.NumberValue())
	} else {
		vm.VirtualHWVer = 13
	}
	if property, has := inputs["networkInterfaces"]; has {
		if items := property.ArrayValue(); len(items) > 0 {
			vm.NetworkInterfaces = make([]NetworkInterface, len(items))
			for i, item := range items {
				vm.NetworkInterfaces[i] = NetworkInterface{
					VirtualNetwork: item.ObjectValue()["virtualNetwork"].StringValue(),
					MacAddress:     "",
					NicType:        "",
				}
				if macAddress, hasProp := item.ObjectValue()["macAddress"]; hasProp {
					vm.NetworkInterfaces[i].MacAddress = macAddress.StringValue()
				}
				if nicType, hasProp := item.ObjectValue()["nicType"]; hasProp {
					vm.NetworkInterfaces[i].MacAddress = nicType.StringValue()
				}
			}
		}
	} else {
		vm.NetworkInterfaces = make([]NetworkInterface, 0)
	}
	if property, has := inputs["os"]; has {
		vm.Os = property.StringValue()
	} else {
		vm.Os = "centos"
	}
	if property, has := inputs["power"]; has {
		vm.Power = property.StringValue()
	} else {
		vm.Power = ""
	}
	if property, has := inputs["startupTimeout"]; has && property.NumberValue() > 0 {
		vm.StartupTimeout = int(property.NumberValue())
	} else {
		vm.StartupTimeout = 120
	}
	if property, has := inputs["shutdownTimeout"]; has && property.NumberValue() > 0 {
		vm.ShutdownTimeout = int(property.NumberValue())
	} else {
		vm.ShutdownTimeout = 20
	}
	if property, has := inputs["virtualDisks"]; has {
		if items := property.ArrayValue(); len(items) > 0 {
			vm.VirtualDisks = make([]VMVirtualDisk, len(items))
			for i, item := range items {
				vm.VirtualDisks[i] = VMVirtualDisk{
					VirtualDiskId: item.ObjectValue()["virtualDiskId"].StringValue(),
					Slot:          "",
				}
				if slot, hasSlot := item.ObjectValue()["slot"]; hasSlot {
					vm.VirtualDisks[i].Slot = slot.StringValue()
				}
			}
		}
	} else {
		vm.VirtualDisks = make([]VMVirtualDisk, 0)
	}
	if property, has := inputs["ovfProperties"]; has {
		if items := property.ArrayValue(); len(items) > 0 {
			vm.OvfProperties = make([]KeyValuePair, len(items))
			for i, item := range items {
				vm.OvfProperties[i] = KeyValuePair{
					Key:   item.ObjectValue()["key"].StringValue(),
					Value: item.ObjectValue()["value"].StringValue(),
				}
			}
		}
	} else {
		vm.OvfProperties = make([]KeyValuePair, 0)
	}
	if property, has := inputs["ovfPropertiesTimer"]; has && property.NumberValue() > 0 {
		vm.ShutdownTimeout = int(property.NumberValue())
	} else {
		vm.ShutdownTimeout = 90
	}
	if property, has := inputs["notes"]; has {
		vm.Notes = strings.Replace(property.StringValue(), "\"", "|22", -1)
	} else {
		vm.Notes = ""
	}
	if property, has := inputs["info"]; has {
		if items := property.ArrayValue(); len(items) > 0 {
			vm.Info = make([]KeyValuePair, len(items))
			for i, item := range items {
				vm.Info[i] = KeyValuePair{
					Key:   item.ObjectValue()["key"].StringValue(),
					Value: item.ObjectValue()["value"].StringValue(),
				}
			}
		}
	} else {
		vm.Info = make([]KeyValuePair, 0)
	}

	return vm, nil
}

func (esxi *Host) readVirtualMachine(vm VirtualMachine) VirtualMachine {
	command := fmt.Sprintf("vim-cmd  vmsvc/get.summary %s", vm.Id)
	stdout, err := esxi.Execute(command, "Get Guest summary")

	if strings.Contains(stdout, "Unable to find a VM corresponding") {
		vm = VirtualMachine{Name: ""}
		return vm
	}

	r, _ := regexp.Compile("")
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		switch {
		case strings.Contains(scanner.Text(), "name = "):
			r, _ = regexp.Compile("\".*\"")
			vm.Name = r.FindString(scanner.Text())
			nr := strings.NewReplacer("\"", "", "\"", "")
			vm.Name = nr.Replace(vm.Name)
		case strings.Contains(scanner.Text(), "vmPathName = "):
			r, _ = regexp.Compile("\\[.*]")
			vm.DiskStore = r.FindString(scanner.Text())
			nr := strings.NewReplacer("[", "", "]", "")
			vm.DiskStore = nr.Replace(vm.DiskStore)
		}
	}

	//  Get resource pool that this VM is located
	command = fmt.Sprintf(`grep -A2 'objID>%s</objID' /etc/vmware/hostd/pools.xml | grep -o resourcePool.*resourcePool`, vm.Id)
	stdout, err = esxi.Execute(command, "check if guest is in resource pool")
	nr := strings.NewReplacer("resourcePool>", "", "</resourcePool", "")
	vmResourcePoolId := nr.Replace(stdout)
	logging.V(9).Infof("readVirtualMachine: resource_pool_name|%s| scanner.Text() => |%s|", vmResourcePoolId, stdout)
	vm.ResourcePoolName, err = esxi.getResourcePoolName(vmResourcePoolId)
	logging.V(9).Infof("readVirtualMachine: resource_pool_name|%s| scanner.Text() => |%s|", vmResourcePoolId, err)

	//
	//  Read vmx file into memory to read settings
	//
	//      -Get location of vmx file on esxi host
	command = fmt.Sprintf("vim-cmd vmsvc/get.config %s | grep vmPathName|grep -oE \"\\[.*\\]\"", vm.Id)
	stdout, err = esxi.Execute(command, "get dst_vmx_ds")
	dstVmxDs := stdout
	dstVmxDs = strings.Trim(dstVmxDs, "[")
	dstVmxDs = strings.Trim(dstVmxDs, "]")

	command = fmt.Sprintf("vim-cmd vmsvc/get.config %s | grep vmPathName|awk '{print $NF}'|sed 's/[\"|,]//g'", vm.Id)
	stdout, err = esxi.Execute(command, "get dst_vmx")
	dstVmx := stdout

	dstVmxFile := "/vmfs/volumes/" + dstVmxDs + "/" + dstVmx

	logging.V(9).Infof("readVirtualMachine: dstVmxFile => %s", dstVmxFile)
	logging.V(9).Infof("readVirtualMachine: vm.DiskStore => %s  dstVmxDs => %s", vm.DiskStore, dstVmxDs)

	command = fmt.Sprintf("cat \"%s\"", dstVmxFile)
	vmxContents, err := esxi.Execute(command, "read guest_name.vmx file")

	vm.VirtualDisks = []VMVirtualDisk{}

	// Used to keep track if a network interface is using static or generated macs.
	var isGeneratedMAC [10]bool
	networkInterfaces := make([]NetworkInterface, 10)

	//  Read vmxContents line-by-line to get current settings.
	scanner = bufio.NewScanner(strings.NewReader(vmxContents))
	for scanner.Scan() {
		switch {
		case strings.Contains(scanner.Text(), "memSize = "):
			r, _ = regexp.Compile("\".*\"")
			stdout = r.FindString(scanner.Text())
			nr = strings.NewReplacer(`"`, "", `"`, "")
			vm.MemSize, _ = strconv.Atoi(nr.Replace(stdout))
			logging.V(9).Infof("readVirtualMachine: MemSize found => %s", vm.MemSize)

		case strings.Contains(scanner.Text(), "numvcpus = "):
			r, _ = regexp.Compile("\".*\"")
			stdout = r.FindString(scanner.Text())
			nr = strings.NewReplacer(`"`, "", `"`, "")
			vm.NumVCpus, _ = strconv.Atoi(nr.Replace(stdout))
			logging.V(9).Infof("readVirtualMachine: NumVCpus found => %s", vm.NumVCpus)

		case strings.Contains(scanner.Text(), "numa.autosize.vcpu."):
			r, _ = regexp.Compile("\".*\"")
			stdout = r.FindString(scanner.Text())
			nr = strings.NewReplacer(`"`, "", `"`, "")
			vm.NumVCpus, _ = strconv.Atoi(nr.Replace(stdout))
			logging.V(9).Infof("readVirtualMachine: numa.vcpu (NumVCpus) found => %s", vm.NumVCpus)

		case strings.Contains(scanner.Text(), "virtualHW.version = "):
			r, _ = regexp.Compile("\".*\"")
			stdout = r.FindString(scanner.Text())
			vm.VirtualHWVer, _ = strconv.Atoi(strings.Replace(stdout, `"`, "", -1))
			logging.V(9).Infof("readVirtualMachine: VirtualHWVer found => %s", vm.VirtualHWVer)

		case strings.Contains(scanner.Text(), "guestOS = "):
			r, _ = regexp.Compile("\".*\"")
			stdout = r.FindString(scanner.Text())
			vm.Os = strings.Replace(stdout, `"`, "", -1)
			logging.V(9).Infof("readVirtualMachine: Os found => %s", vm.Os)

		case strings.Contains(scanner.Text(), "scsi"):
			re := regexp.MustCompile("scsi([0-3]):([0-9]{1,2}).(.*) = \"(.*)\"")
			results := re.FindStringSubmatch(scanner.Text())
			if len(results) > 4 {
				logging.V(9).Infof("readVirtualMachine: %s : %s . %s = %s", results[1], results[2], results[3], results[4])

				if (results[1] == "0") && (results[2] == "0") {
					// Skip boot disk
				} else {
					if strings.Contains(results[3], "fileName") {
						logging.V(9).Infof("readVirtualMachine: %s => %s", results[0], results[4])

						vm.VirtualDisks = append(vm.VirtualDisks, VMVirtualDisk{
							fmt.Sprintf("%s:%s", results[1], results[2]),
							results[4],
						})
					}
				}
			}

		case strings.Contains(scanner.Text(), "ethernet"):
			re := regexp.MustCompile("ethernet(.).(.*) = \"(.*)\"")
			results := re.FindStringSubmatch(scanner.Text())
			index, _ := strconv.Atoi(results[1])

			switch results[2] {
			case "networkName":
				networkInterfaces[index].VirtualNetwork = results[3]
				logging.V(9).Infof("readVirtualMachine: %s => %s", results[0], results[3])

			case "addressType":
				if results[3] == "generated" {
					isGeneratedMAC[index] = true
				}

			//  Done't save generatedAddress...   It should not be saved because it
			//  should be considered dynamic & is breaks the update MAC address code.
			//  case "generatedAddress":
			//	  if isGeneratedMAC[index] == true {
			//		networkInterfaces[index].MacAddress = results[3]
			//	    logging.V(9).Infof("readVirtualMachine: %s => %s", results[0], results[3])
			//	  }

			case "address":
				if !isGeneratedMAC[index] {
					networkInterfaces[index].MacAddress = results[3]
					logging.V(9).Infof("readVirtualMachine: %s => %s", results[0], results[3])
				}

			case "virtualDev":
				networkInterfaces[index].NicType = results[3]
				logging.V(9).Infof("readVirtualMachine: %s => %s", results[0], results[3])
			}

		case strings.Contains(scanner.Text(), "firmware = "):
			r, _ = regexp.Compile("\".*\"")
			stdout = r.FindString(scanner.Text())
			vm.BootFirmware = strings.Replace(stdout, `"`, "", -1)
			logging.V(9).Infof("readVirtualMachine: BootFirmware found => %s", vm.BootFirmware)

		case strings.Contains(scanner.Text(), "annotation = "):
			r, _ = regexp.Compile("\".*\"")
			stdout = r.FindString(scanner.Text())
			vm.Notes = strings.Replace(stdout, `"`, "", -1)
			vm.Notes = strings.Replace(vm.Notes, "|22", "\"", -1)
			logging.V(9).Infof("readVirtualMachine: Notes found => %s", vm.Notes)
		}
	}

	for _, ni := range networkInterfaces {
		if len(ni.VirtualNetwork) > 0 {
			vm.NetworkInterfaces = append(vm.NetworkInterfaces, ni)
		}
	}

	parsedVmx := ParseVMX(vmxContents)

	//  Get power state
	vm.Power = esxi.getVirtualMachinePowerState(vm.Id)
	logging.V(9).Infof("readVirtualMachine: Power => %s", vm.Power)

	//
	// Get IP address (need vmware tools installed)
	//
	if vm.Power == "on" {
		vm.IpAddress = esxi.getVirtualMachineIpAddress(vm.Id, vm.StartupTimeout)
		logging.V(9).Infof("readVirtualMachine: IpAddress found => %s", vm.IpAddress)
	} else {
		vm.IpAddress = ""
	}

	// Get boot disk size
	bootDiskPath, _ := esxi.getBootDiskPath(vm.Id)
	vd, err := esxi.getVirtualDisk(bootDiskPath)
	vm.BootDiskSize = vd.Size
	vm.BootDiskType = vd.DiskType

	// Get guestinfo value
	info := make(map[string]string)
	for key, value := range parsedVmx {
		if strings.Contains(key, "guestinfo") {
			shortKey := strings.Replace(key, "guestinfo.", "", -1)
			info[shortKey] = value
		}
	}
	i := 0
	vm.Info = make([]KeyValuePair, len(info))
	for key, value := range info {
		vm.Info[i] = KeyValuePair{key, value}
		i++
	}

	return vm
}
