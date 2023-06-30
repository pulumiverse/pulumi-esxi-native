package esxi

import (
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (esxi *Host) getVirtualMachineId(name string) (string, error) {
	var command, id string
	var err error

	command = fmt.Sprintf("vim-cmd vmsvc/getallvms 2>/dev/null |sort -n | "+
		"grep -m 1 \"[0-9] * %s .*%s\" |awk '{print $1}' ", name, name)

	id, err = esxi.Execute(command, "get vm Id")
	logging.V(9).Infof("getVirtualMachineId: result => %s", id)
	if err != nil {
		logging.V(9).Infof("getVirtualMachineId: Failed get vm id => %s", err)
		return "", fmt.Errorf("Failed get vm id: %s\n", err)
	}

	return id, nil
}

func (esxi *Host) validateVirtualMachineId(id string) (string, error) {
	var command string
	var err error

	command = fmt.Sprintf("vim-cmd vmsvc/getallvms 2>/dev/null | awk '{print $1}' | "+
		"grep '^%s$'", id)

	id, err = esxi.Execute(command, "validate vm id exists")
	logging.V(9).Infof("validateVirtualMachineId: result => %s", id)
	if err != nil {
		logging.V(9).Infof("validateVirtualMachineId: Failed get vm by id => %s", err)
		return "", fmt.Errorf("Failed get vm id: %s\n", err)
	}

	return id, nil
}

func (esxi *Host) getBootDiskPath(id string) (string, error) {
	var command, stdout string
	var err error

	command = fmt.Sprintf("vim-cmd vmsvc/device.getdevices %s | grep -A10 -e 'key = 2000' -e 'key = 3000' -e 'key = 16000'|grep -m 1 fileName", id)
	stdout, err = esxi.Execute(command, "get boot disk")
	if err != nil {
		logging.V(9).Infof("getBootDiskPath: Failed get boot disk path => %s", stdout)
		return "Failed get boot disk path:", err
	}
	r := strings.NewReplacer("fileName = \"[", "/vmfs/volumes/", "] ", "/", "\",", "")
	return r.Replace(stdout), err
}

func (esxi *Host) getDstVmxFile(id string) (string, error) {
	var dstVmxDs, dstVmx, dstVmxFile string

	//      -Get location of vmx file on esxi host
	command := fmt.Sprintf("vim-cmd vmsvc/get.config %s | grep vmPathName|grep -oE \"\\[.*\\]\"", id)
	stdout, err := esxi.Execute(command, "get dstVmxDs")
	dstVmxDs = stdout
	dstVmxDs = strings.Trim(dstVmxDs, "[")
	dstVmxDs = strings.Trim(dstVmxDs, "]")

	command = fmt.Sprintf("vim-cmd vmsvc/get.config %s | grep vmPathName|awk '{print $NF}'|sed 's/[\"|,]//g'", id)
	stdout, err = esxi.Execute(command, "get dstVmx")
	dstVmx = stdout

	dstVmxFile = "/vmfs/volumes/" + dstVmxDs + "/" + dstVmx
	return dstVmxFile, err
}

func (esxi *Host) readVmxContents(id string) (string, error) {
	var command, vmxContents string

	dstVmxFile, err := esxi.getDstVmxFile(id)
	command = fmt.Sprintf("cat \"%s\"", dstVmxFile)
	vmxContents, err = esxi.Execute(command, "read vmx file")

	return vmxContents, err
}

func (esxi *Host) updateVmxContents(isNew bool, vm VirtualMachine) error {

	var regexReplacement string

	vmxContents, err := esxi.readVmxContents(vm.Id)
	if err != nil {
		logging.V(9).Infof("updateVmxContents: Failed get vmx contents => %s", err)
		return fmt.Errorf("Failed to get vmx contents: %s\n", err)
	}
	if strings.Contains(vmxContents, "Unable to find a VM corresponding") {
		return nil
	}

	var memSize, numVCpus, virtualHWVer int
	memSize, _ = strconv.Atoi(vm.MemSize)
	numVCpus, _ = strconv.Atoi(vm.NumVCpus)
	virtualHWVer, _ = strconv.Atoi(vm.VirtualHWVer)

	if memSize != 0 {
		re := regexp.MustCompile("memSize = \".*\"")
		regexReplacement = fmt.Sprintf("memSize = \"%d\"", memSize)
		vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)
	}

	if numVCpus != 0 {
		if strings.Contains(vmxContents, "numvcpus = ") {
			re := regexp.MustCompile("numvcpus = \".*\"")
			regexReplacement = fmt.Sprintf("numvcpus = \"%d\"", numVCpus)
			vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)
		} else {
			logging.V(9).Infof("updateVmxContents: Add numVCpu => %d", numVCpus)
			vmxContents += fmt.Sprintf("\nnumvcpus = \"%d\"", numVCpus)
		}
	}

	if virtualHWVer != 0 {
		re := regexp.MustCompile("virtualHW.version = \".*\"")
		regexReplacement = fmt.Sprintf("virtualHW.version = \"%d\"", virtualHWVer)
		vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)
	}

	if vm.Os != "" {
		re := regexp.MustCompile("guestOS = \".*\"")
		regexReplacement = fmt.Sprintf("guestOS = \"%s\"", vm.Os)
		vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)
	}

	re := regexp.MustCompile("firmware = \".*\"")
	regexReplacement = fmt.Sprintf("firmware = \"%s\"", vm.BootFirmware)
	vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)

	// modify annotation
	if vm.Notes != "" {
		vm.Notes = strings.Replace(vm.Notes, "\"", "|22", -1)
		if strings.Contains(vmxContents, "annotation") {
			re := regexp.MustCompile("annotation = \".*\"")
			regexReplacement = fmt.Sprintf("annotation = \"%s\"", vm.Notes)
			vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)
		} else {
			regexReplacement = fmt.Sprintf("\nannotation = \"%s\"", vm.Notes)
			vmxContents += regexReplacement
		}
	}

	if len(vm.Info) > 0 {
		parsedVmx := ParseVMX(vmxContents)
		for _, config := range vm.Info {
			logging.V(9).Infof("SAVING %s => %s", config.Key, config.Value)
			parsedVmx["guestinfo."+config.Key] = config.Value
		}
		vmxContents = EncodeVMX(parsedVmx)
	}

	//
	//  add/modify virtual disks
	//
	var tmpvar string
	var vmxContentsNew string
	var i, j int

	//
	//  Remove all disks
	//
	regexReplacement = fmt.Sprintf("")
	for i = 0; i < 4; i++ {
		for j = 0; j < 16; j++ {

			if (i != 0 || j != 0) && j != 7 {
				re := regexp.MustCompile(fmt.Sprintf("scsi%d:%d.*\n", i, j))
				vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)
			}
		}
	}

	//
	//  Add disks that are managed by pulumi
	//
	for _, vd := range vm.VirtualDisks {
		if vd.VirtualDiskId != "" {
			logging.V(9).Infof("updateVmxContents: Adding => %s", vd.Slot)
			tmpvar = fmt.Sprintf("scsi%s.deviceType = \"scsi-hardDisk\"\n", vd.Slot)
			if !strings.Contains(vmxContents, tmpvar) {
				vmxContents += "\n" + tmpvar
			}

			tmpvar = fmt.Sprintf("scsi%s.fileName", vd.Slot)
			if strings.Contains(vmxContents, tmpvar) {
				re := regexp.MustCompile(tmpvar + " = \".*\"")
				regexReplacement = fmt.Sprintf(tmpvar+" = \"%s\"", vd.VirtualDiskId)
				vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)
			} else {
				regexReplacement = fmt.Sprintf("\n"+tmpvar+" = \"%s\"", vd.VirtualDiskId)
				vmxContents += "\n" + regexReplacement
			}

			tmpvar = fmt.Sprintf("scsi%s.present = \"true\"\n", vd.Slot)
			if !strings.Contains(vmxContents, tmpvar) {
				vmxContents += "\n" + tmpvar
			}

		}
	}

	//
	//  Create/update networks network_interfaces
	//

	//  Define default nic type.
	var defaultNetworkType, networkType string
	if vm.NetworkInterfaces[0].NicType != "" {
		defaultNetworkType = vm.NetworkInterfaces[0].NicType
	} else {
		defaultNetworkType = "e1000"
	}

	//  If this is first time provisioning, delete all the old ethernet configuration.
	if isNew == true {
		logging.V(9).Infof("updateVmxContents:Delete old ethernet configuration => %d", i)
		regexReplacement = fmt.Sprintf("")
		for i = 0; i < 9; i++ {
			re := regexp.MustCompile(fmt.Sprintf("ethernet%d.*\n", i))
			vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)
		}
	}

	//  Add/Modify virtual networks.
	networkType = ""
	for i, ni := range vm.NetworkInterfaces {
		logging.V(9).Infof("updateVmxContents: ethernet%d", i)

		if ni.VirtualNetwork == "" && strings.Contains(vmxContents, "ethernet"+strconv.Itoa(i)) == true {
			//  This is Modify (Delete existing network configuration)
			logging.V(9).Infof("updateVmxContents: Modify ethernet%d - Delete existing.", i)
			regexReplacement = fmt.Sprintf("")
			re := regexp.MustCompile(fmt.Sprintf("ethernet%d.*\n", i))
			vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)
		}

		if ni.VirtualNetwork != "" && strings.Contains(vmxContents, "ethernet"+strconv.Itoa(i)) == true {
			//  This is Modify
			logging.V(9).Infof("updateVmxContents: Modify ethernet%d - Modify existing.", i)

			//  Modify Network Name
			re := regexp.MustCompile("ethernet" + strconv.Itoa(i) + ".networkName = \".*\"")
			regexReplacement = fmt.Sprintf("ethernet"+strconv.Itoa(i)+".networkName = \"%s\"", ni.VirtualNetwork)
			vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)

			//  Modify virtual Device
			re = regexp.MustCompile("ethernet" + strconv.Itoa(i) + ".virtualDev = \".*\"")
			regexReplacement = fmt.Sprintf("ethernet"+strconv.Itoa(i)+".virtualDev = \"%s\"", ni.NicType)
			vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)

			//  Modify MAC (dynamic to static only. static to dynamic is not implemented)
			if ni.MacAddress != "" {
				logging.V(9).Infof("updateVmxContents: ethernet%d Modify MAC: %s", i, ni.MacAddress)

				re = regexp.MustCompile("ethernet" + strconv.Itoa(i) + ".[a-zA-Z]*ddress = \".*\"")
				regexReplacement = fmt.Sprintf("ethernet"+strconv.Itoa(i)+".address = \"%s\"", ni.MacAddress)
				vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)

				re = regexp.MustCompile("ethernet" + strconv.Itoa(i) + ".addressType = \".*\"")
				regexReplacement = fmt.Sprintf("ethernet" + strconv.Itoa(i) + ".addressType = \"static\"")
				vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)

				re = regexp.MustCompile("ethernet" + strconv.Itoa(i) + ".generatedAddressOffset = \".*\"")
				regexReplacement = fmt.Sprintf("")
				vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)
			}
		}

		if ni.VirtualNetwork != "" && strings.Contains(vmxContents, "ethernet"+strconv.Itoa(i)) == false {
			//  This is created

			//  Set virtual_network name
			logging.V(9).Infof("updateVmxContents: ethernet%d Create New: %s", i, ni.VirtualNetwork)
			tmpvar = fmt.Sprintf("\nethernet%d.networkName = \"%s\"\n", i, ni.VirtualNetwork)
			vmxContentsNew = tmpvar

			//  Set mac address
			if ni.MacAddress != "" {
				tmpvar = fmt.Sprintf("ethernet%d.addressType = \"static\"\n", i)
				vmxContentsNew = vmxContentsNew + tmpvar

				tmpvar = fmt.Sprintf("ethernet%d.address = \"%s\"\n", i, ni.MacAddress)
				vmxContentsNew = vmxContentsNew + tmpvar
			}

			//  Set network type
			if ni.NicType == "" {
				networkType = defaultNetworkType
			} else {
				networkType = ni.NicType
			}

			tmpvar = fmt.Sprintf("ethernet%d.virtualDev = \"%s\"\n", i, networkType)
			vmxContentsNew = vmxContentsNew + tmpvar

			tmpvar = fmt.Sprintf("ethernet%d.present = \"TRUE\"\n", i)

			vmxContents = vmxContents + vmxContentsNew + tmpvar
		}
	}

	//  Add disk UUID
	if !strings.Contains(vmxContents, "disk.EnableUUID") {
		vmxContents = vmxContents + "\ndisk.EnableUUID = \"TRUE\""
	}

	//
	//  Write vmx file to esxi host
	//
	logging.V(9).Infof("updateVmxContents: New vm_name.vmx => %s", vmxContents)

	dstVmxFile, err := esxi.getDstVmxFile(vm.Id)

	vmxContents, err = esxi.CopyFile(strings.Replace(vmxContents, "\\\"", "\"", -1), dstVmxFile, "write guest_name.vmx file")

	err = esxi.reloadVirtualMachine(vm.Id)
	return err
}

func (esxi *Host) cleanStorageFromVmx(id string) error {
	vmxContents, err := esxi.readVmxContents(id)
	if err != nil {
		logging.V(9).Infof("cleanStorageFromVmx: Failed get vmx contents => %s", err)
		return fmt.Errorf("Failed to get vmx contents: %s\n", err)
	}

	for x := 0; x < 4; x++ {
		for y := 0; y < 16; y++ {
			if !(x == 0 && y == 0) {
				regexReplacement := fmt.Sprintf("scsi%d:%d.*", x, y)
				re := regexp.MustCompile(regexReplacement)
				vmxContents = re.ReplaceAllString(vmxContents, "")
			}
		}
	}

	//
	//  Write vmx file to esxi host
	//

	dstVmxFile, err := esxi.getDstVmxFile(id)
	vmxContents, err = esxi.CopyFile(strings.Replace(vmxContents, "\\\"", "\"", -1), dstVmxFile, "write guest_name.vmx file")

	err = esxi.reloadVirtualMachine(id)
	return err
}

func (esxi *Host) reloadVirtualMachine(id string) error {
	command := fmt.Sprintf("vim-cmd vmsvc/reload %s", id)
	_, err := esxi.Execute(command, "vmsvc/reload")

	return err
}

func (esxi *Host) powerOnVirtualMachine(id string) (string, error) {
	if esxi.getVirtualMachinePowerState(id) == "on" {
		return "", nil
	}

	command := fmt.Sprintf("vim-cmd vmsvc/power.on %s", id)
	stdout, err := esxi.Execute(command, "vmsvc/power.on")
	time.Sleep(3 * time.Second)

	if esxi.getVirtualMachinePowerState(id) == "on" {
		return stdout, nil
	}

	return stdout, err
}

func (esxi *Host) powerOffVirtualMachine(id string, shutdownTimeout int) (string, error) {
	var command, stdout string

	savedPowerState := esxi.getVirtualMachinePowerState(id)
	if savedPowerState == "off" {
		return "", nil

	} else if savedPowerState == "on" {

		if shutdownTimeout != 0 {
			command = fmt.Sprintf("vim-cmd vmsvc/power.shutdown %s", id)
			stdout, _ = esxi.Execute(command, "vmsvc/power.shutdown")
			time.Sleep(3 * time.Second)

			for i := 0; i < (shutdownTimeout / 3); i++ {
				if esxi.getVirtualMachinePowerState(id) == "off" {
					return stdout, nil
				}
				time.Sleep(3 * time.Second)
			}
		}

		command = fmt.Sprintf("vim-cmd vmsvc/power.off %s", id)
		stdout, _ = esxi.Execute(command, "vmsvc/power.off")
		time.Sleep(1 * time.Second)

		return stdout, nil

	} else {
		command = fmt.Sprintf("vim-cmd vmsvc/power.off %s", id)
		stdout, _ = esxi.Execute(command, "vmsvc/power.off")
		return stdout, nil
	}
}

func (esxi *Host) getVirtualMachinePowerState(id string) string {
	command := fmt.Sprintf("vim-cmd vmsvc/power.getstate %s", id)
	stdout, _ := esxi.Execute(command, "vmsvc/power.getstate")
	if strings.Contains(stdout, "Unable to find a VM corresponding") {
		return "Unknown"
	}

	if strings.Contains(stdout, "Powered off") == true {
		return "off"
	} else if strings.Contains(stdout, "Powered on") == true {
		return "on"
	} else if strings.Contains(stdout, "Suspended") == true {
		return "suspended"
	} else {
		return "Unknown"
	}
}

func (esxi *Host) getVirtualMachineIpAddress(id string, startupTimeout int) string {
	var command, stdout, ipAddress, ipAddress2 string
	var uptime int

	//  Check if powered off
	if esxi.getVirtualMachinePowerState(id) != "on" {
		return ""
	}

	//
	//  Check uptime of guest.
	//
	uptime = 0
	for uptime < startupTimeout {
		//  Primary method to get IP
		command = fmt.Sprintf("vim-cmd vmsvc/get.guest %s 2>/dev/null |sed '1!G;h;$!d' |awk '/deviceConfigId = 4000/,/ipAddress/' |grep -m 1 -oE '((1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])\\.){3}(1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])'", id)
		stdout, _ = esxi.Execute(command, "get ip_address method 1")
		ipAddress = stdout
		if ipAddress != "" {
			return ipAddress
		}

		time.Sleep(3 * time.Second)

		//  Get uptime if above failed.
		command = fmt.Sprintf("vim-cmd vmsvc/get.summary %s 2>/dev/null | grep 'uptimeSeconds ='|sed 's/^.*= //g'|sed s/,//g", id)
		stdout, err := esxi.Execute(command, "get uptime")
		if err != nil {
			return ""
		}
		uptime, _ = strconv.Atoi(stdout)
	}

	//
	// Alternate method to get IP
	//
	command = fmt.Sprintf("vim-cmd vmsvc/get.guest %s 2>/dev/null | grep -m 1 '^   ipAddress = ' | grep -oE '((1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])\\.){3}(1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])'", id)
	stdout, _ = esxi.Execute(command, "get ip_address method 2")
	ipAddress2 = stdout
	if ipAddress2 != "" {
		return ipAddress2
	}

	return ""
}
