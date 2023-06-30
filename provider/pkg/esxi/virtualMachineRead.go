package esxi

import (
	"bufio"
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
	"regexp"
	"strconv"
	"strings"
)

func VirtualMachineReadParser(id string, inputs resource.PropertyMap) VirtualMachine {
	vm := VirtualMachine{
		Id:             id,
		StartupTimeout: int(inputs["startupTimeout"].NumberValue()),
	}

	return vm
}

func VirtualMachineRead(vm VirtualMachine, esxi *Host) (string, resource.PropertyMap, error) {
	// read vm
	vm, err := esxi.readVirtualMachine(vm)

	if err != nil || vm.Name == "" {
		return "", nil, err
	}

	result := vm.toMap()
	return vm.Id, resource.NewPropertyMapFromMap(result), err
}

func (esxi *Host) readVirtualMachine(vm VirtualMachine) (VirtualMachine, error) {
	command := fmt.Sprintf("vim-cmd  vmsvc/get.summary %s", vm.Id)
	stdout, err := esxi.Execute(command, "Get Guest summary")

	if strings.Contains(stdout, "Unable to find a VM corresponding") {
		vm.Name = ""
		vm.DiskStore = ""
		vm.BootDiskSize = ""
		vm.BootDiskType = ""
		vm.ResourcePoolName = ""
		vm.MemSize = ""
		vm.NumVCpus = ""
		vm.VirtualHWVer = ""
		vm.Os = ""
		vm.IpAddress = ""
		vm.VirtualHWVer = ""
		vm.NetworkInterfaces = make([]NetworkInterface, 0)
		vm.BootFirmware = ""
		vm.VirtualDisks = make([]VMVirtualDisk, 0)
		vm.Power = ""
		vm.Notes = ""
		vm.Info = nil

		return vm, nil
	}

	r, _ := regexp.Compile("")
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		switch {
		case strings.Contains(scanner.Text(), "name = "):
			r, _ = regexp.Compile("\".*\"")
			vm.Name = r.FindString(scanner.Text())
			nr := strings.NewReplacer(`"`, "", `"`, "")
			vm.Name = nr.Replace(vm.Name)
		case strings.Contains(scanner.Text(), "vmPathName = "):
			r, _ = regexp.Compile("\".*\"")
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
			vm.MemSize = nr.Replace(stdout)
			logging.V(9).Infof("readVirtualMachine: MemSize found => %s", vm.MemSize)

		case strings.Contains(scanner.Text(), "numvcpus = "):
			r, _ = regexp.Compile("\".*\"")
			stdout = r.FindString(scanner.Text())
			nr = strings.NewReplacer(`"`, "", `"`, "")
			vm.NumVCpus = nr.Replace(stdout)
			logging.V(9).Infof("readVirtualMachine: NumVCpus found => %s", vm.NumVCpus)

		case strings.Contains(scanner.Text(), "numa.autosize.vcpu."):
			r, _ = regexp.Compile("\".*\"")
			stdout = r.FindString(scanner.Text())
			nr = strings.NewReplacer(`"`, "", `"`, "")
			vm.NumVCpus = nr.Replace(stdout)
			logging.V(9).Infof("readVirtualMachine: numa.vcpu (NumVCpus) found => %s", vm.NumVCpus)

		case strings.Contains(scanner.Text(), "virtualHW.version = "):
			r, _ = regexp.Compile("\".*\"")
			stdout = r.FindString(scanner.Text())
			vm.VirtualHWVer = strings.Replace(stdout, `"`, "", -1)
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
					if strings.Contains(results[3], "fileName") == true {
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
			//case "generatedAddress":
			//	if isGeneratedMAC[index] == true {
			//		networkInterfaces[index].MacAddress = results[3]
			//	    logging.V(9).Infof("readVirtualMachine: %s => %s", results[0], results[3])
			//	}

			case "address":
				if isGeneratedMAC[index] == false {
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
	vd, err := esxi.readVirtualDisk(bootDiskPath)
	vm.BootDiskSize = strconv.Itoa(vd.Size)
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

	return vm, nil
}
