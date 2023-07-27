package esxi

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
)

func (esxi *Host) createPlainVirtualMachine(vm VirtualMachine) (VirtualMachine, error) {
	// check if path already exists.
	fullPATH := fmt.Sprintf("/vmfs/volumes/%s/%s", vm.DiskStore, vm.Name)
	bootDiskVmdkPath := fmt.Sprintf("\"/vmfs/volumes/%s/%s/%s.vmdk\"", vm.DiskStore, vm.Name, vm.Name)
	command := fmt.Sprintf("ls -d %s", bootDiskVmdkPath)
	stdout, _ := esxi.Execute(command, "check if guest path already exists.")
	if !strings.Contains(stdout, "No such file or directory") {
		return VirtualMachine{}, fmt.Errorf("virtual machine may already exists. vmdkPATH:%s", bootDiskVmdkPath)
	}

	command = fmt.Sprintf("ls -d \"%s\"", fullPATH)
	stdout, _ = esxi.Execute(command, "check if guest path already exists.")
	if strings.Contains(stdout, "No such file or directory") {
		command = fmt.Sprintf("mkdir \"%s\"", fullPATH)
		_, err := esxi.Execute(command, "create guest path")
		if err != nil {
			return VirtualMachine{}, fmt.Errorf("Failed to create guest path. fullPATH:%s\n", fullPATH)
		}
	}

	hasISO := false
	isoFileName := ""
	// Build VM by default/black config
	vmxContents := "config.version = \"8\"\n" +
		fmt.Sprintf("virtualHW.version = \"%d\"\n", vm.VirtualHWVer) +
		fmt.Sprintf("displayName = \"%s\"\n", vm.Name) +
		fmt.Sprintf("numvcpus = \"%d\"\n", vm.NumVCpus) +
		fmt.Sprintf("memSize = \"%d\"\n", vm.MemSize) +
		fmt.Sprintf("guestOS = \"%s\"\n", vm.Os) +
		fmt.Sprintf("annotation = \"%s\"\n", vm.Notes) +
		fmt.Sprintf("floppy0.present = \"%s\"\n", "FALSE") +
		fmt.Sprintf("scsi0.present = \"TRUE\"\n") +
		fmt.Sprintf("scsi0.sharedBus = \"none\"\n") +
		fmt.Sprintf("scsi0.virtualDev = \"lsilogic\"\n") +
		fmt.Sprintf("disk.EnableUUID = \"TRUE\"\n") +
		fmt.Sprintf("pciBridge0.present = \"TRUE\"\n") +
		fmt.Sprintf("pciBridge4.present = \"TRUE\"\n") +
		fmt.Sprintf("pciBridge4.virtualDev = \"pcieRootPort\"\n") +
		fmt.Sprintf("pciBridge4.functions = \"8\"\n") +
		fmt.Sprintf("pciBridge5.present = \"TRUE\"\n") +
		fmt.Sprintf("pciBridge5.virtualDev = \"pcieRootPort\"\n") +
		fmt.Sprintf("pciBridge5.functions = \"8\"\n") +
		fmt.Sprintf("pciBridge6.present = \"TRUE\"\n") +
		fmt.Sprintf("pciBridge6.virtualDev = \"pcieRootPort\"\n") +
		fmt.Sprintf("pciBridge6.functions = \"8\"\n") +
		fmt.Sprintf("pciBridge7.present = \"TRUE\"\n") +
		fmt.Sprintf("pciBridge7.virtualDev = \"pcieRootPort\"\n") +
		fmt.Sprintf("pciBridge7.functions = \"8\"\n") +
		fmt.Sprintf("scsi0:0.present = \"TRUE\"\n") +
		fmt.Sprintf("scsi0:0.fileName = \"%s.vmdk\"\n", vm.Name) +
		fmt.Sprintf("scsi0:0.deviceType = \"%s\"\n", "scsi-hardDisk") +
		fmt.Sprintf("nvram = \"%s.nvram\"\n", vm.Name)
	if vm.BootFirmware == "efi" {
		vmxContents = vmxContents + "firmware = \\\"efi\\\"\n"
	} else if vm.BootFirmware == "bios" {
		vmxContents = vmxContents + "firmware = \\\"bios\\\"\n"
	}
	// TODO: to be checked how we can set the ISO file
	if hasISO {
		vmxContents = vmxContents +
			fmt.Sprintf("ide1:0.present = \"TRUE\"\n") +
			fmt.Sprintf("ide1:0.fileName = \\\"emptyBackingString\\\"\n") +
			fmt.Sprintf("ide1:0.deviceType = \\\"atapi-cdrom\\\"\n") +
			fmt.Sprintf("ide1:0.startConnected = \"FALSE\"\n") +
			fmt.Sprintf("ide1:0.clientDevice = \"TRUE\"\n")
	} else {
		vmxContents = vmxContents +
			fmt.Sprintf("ide1:0.present = \"TRUE\"\n") +
			fmt.Sprintf("ide1:0.fileName = \"%s\"\n", isoFileName) +
			fmt.Sprintf("ide1:0.deviceType = \"cdrom-raw\"\n")
	}

	// Write vmx file to esxi host
	dstVmxFile := fmt.Sprintf("%s/%s.vmx", fullPATH, vm.Name)

	command = fmt.Sprintf("echo \"%s\" >\"%s\"", vmxContents, dstVmxFile)
	vmxContents, err := esxi.WriteFile(vmxContents, dstVmxFile, "write vmx file")

	// Create boot disk (vmdk)
	command = fmt.Sprintf("vmkfstools -c %dG -d %s \"%s/%s.vmdk\"", vm.BootDiskSize, vm.BootDiskType, fullPATH, vm.Name)
	_, err = esxi.Execute(command, "vmkfstools (make boot disk)")
	if err != nil {
		command = fmt.Sprintf("rm -fr \"%s\"", fullPATH)
		stdout, _ = esxi.Execute(command, "cleanup guest path because of failed events")
		return VirtualMachine{}, fmt.Errorf("Failed to vmkfstools (make boot disk):%s\n", err)
	}

	poolID, err := esxi.getResourcePoolId(vm.ResourcePoolName)
	if err != nil {
		return VirtualMachine{}, fmt.Errorf("failed to use Resource Pool ID:%s", poolID)
	}
	command = fmt.Sprintf("vim-cmd solo/registervm \"%s\" %s %s", dstVmxFile, vm.Name, poolID)
	_, err = esxi.Execute(command, "solo/registervm")
	if err != nil {
		command = fmt.Sprintf("rm -fr \"%s\"", fullPATH)
		stdout, _ = esxi.Execute(command, "cleanup guest path because of failed events")
		return VirtualMachine{}, fmt.Errorf("failed to register guest:%s", err)
	}

	return vm, nil
}

func (esxi *Host) createVirtualMachine(vm VirtualMachine) (VirtualMachine, error) {
	hasOvfProperties := false
	// Check if Disk Store already exists
	err := esxi.validateDiskStore(vm.DiskStore)
	if err != nil {
		return VirtualMachine{}, fmt.Errorf("failed to validate disk store: %s", err)
	}

	// Check if guest already exists
	// get VM ID (by name)
	id, err := esxi.getVirtualMachineId(vm.Name)

	if id != "" {
		// We don't need to create the VM. It already exists.
		// Power off guest if it's powered on.
		currentPowerState := esxi.getVirtualMachinePowerState(id)
		if currentPowerState == "on" || currentPowerState == "suspended" {
			_, err = esxi.powerOffVirtualMachine(id, vm.ShutdownTimeout)
			if err != nil {
				return VirtualMachine{}, fmt.Errorf("failed to power off: %s", err)
			}
		}
	} else if vm.SourcePath == "none" {
		vm, err = esxi.createPlainVirtualMachine(vm)
		if err != nil {
			return vm, err
		}
	} else {
		// Build VM with ovftool
		// Check if source file exist.
		const (
			httpSchema  = "http://"
			httpsSchema = "https://"
		)
		if strings.HasPrefix(vm.SourcePath, httpSchema) || strings.HasPrefix(vm.SourcePath, httpsSchema) {
			resp, err := http.Get(vm.SourcePath)
			if (err != nil) || (resp.StatusCode != 200) {
				logging.V(9).Infof("URL not accessible: %s", vm.SourcePath)
				logging.V(9).Infof("URL StatusCode: %d", resp.StatusCode)
				logging.V(9).Infof("URL Error: %s", err)
				defer func(Body io.ReadCloser) {
					_ = Body.Close()
				}(resp.Body)
				return VirtualMachine{}, fmt.Errorf("URL not accessible: %s. err:%s", vm.SourcePath, err)
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
		} else if strings.HasPrefix(vm.SourcePath, "vi://") {
			logging.V(9).Infof("Source is Guest VM (vi).\n")
		} else {
			logging.V(9).Infof("Source is local.\n")
			if _, err := os.Stat(vm.SourcePath); os.IsNotExist(err) {
				logging.V(9).Infof("File not found, Error: %s\n", err)
				return VirtualMachine{}, fmt.Errorf("file not found locally: %s", vm.SourcePath)
			}
		}

		// Set params for packer
		if vm.BootDiskType == "zeroedthick" {
			vm.BootDiskType = "thick"
		}

		username := url.QueryEscape(esxi.Connection.UserName)
		password := url.QueryEscape(esxi.Connection.Password)
		dstPath := fmt.Sprintf("vi://%s:%s@%s:%s/%s", username, password, esxi.Connection.Host, esxi.Connection.SslPort, vm.ResourcePoolName)

		netParam := ""
		if (strings.HasSuffix(vm.SourcePath, ".ova") || strings.HasSuffix(vm.SourcePath, ".ovf")) && len(vm.NetworkInterfaces) > 0 && vm.NetworkInterfaces[0].VirtualNetwork != "" {
			netParam = fmt.Sprintf(" --network='%s'", vm.NetworkInterfaces[0].VirtualNetwork)
		}

		extraParams := ""
		if (len(vm.OvfProperties) > 0) && (strings.HasSuffix(vm.SourcePath, ".ova") || strings.HasSuffix(vm.SourcePath, ".ovf")) {
			hasOvfProperties = true
			// in order to process any OVF params, guest should be immediately powered on
			// This is because the ESXi host doesn't have a cache to store the OVF parameters, like the vCenter Server does.
			// Therefore, you MUST use the ‘--X:injectOvfEnv’ option with the ‘--poweron’ option
			extraParams = "--X:injectOvfEnv --allowExtraConfig --powerOn "

			for _, prop := range vm.OvfProperties {
				extraParams = fmt.Sprintf("%s --prop:%s='%s' ", extraParams, prop.Key, prop.Value)
			}
			log.Println("ovf_properties extra_params: " + extraParams)
		}

		ovfCmd := fmt.Sprintf("ovftool --acceptAllEulas --noSSLVerify --X:useMacNaming=false %s -dm=%d --name='%s' --overwrite -ds='%s'%s '%s' '%s'",
			extraParams, vm.BootDiskSize, vm.Name, vm.DiskStore, netParam, vm.SourcePath, dstPath)
		re := regexp.MustCompile(`vi://.*?@`)

		osShellCmd := "/bin/bash"
		osShellCmdOpt := "-c"

		var ovfBat *os.File
		if runtime.GOOS == "windows" {
			osShellCmd = "cmd.exe"
			osShellCmdOpt = "/c"

			ovfCmd = strings.Replace(ovfCmd, "'", "\"", -1)

			ovfBat, _ = ioutil.TempFile("", "ovfCmd*.bat")

			_, err = os.Stat(ovfBat.Name())
			// delete file if exists
			if os.IsExist(err) {
				err = os.Remove(ovfBat.Name())
				if err != nil {
					return VirtualMachine{}, fmt.Errorf("unable to delete existing %s: %w", ovfBat.Name(), err)
				}
			}

			//  create new batch file
			file, err := os.Create(ovfBat.Name())
			if err != nil {
				defer file.Close()
				return VirtualMachine{}, fmt.Errorf("unable to create %s: %w", ovfBat.Name(), err)
			}

			_, err = file.WriteString(strings.Replace(ovfCmd, "%", "%%", -1))
			if err != nil {
				defer file.Close()
				return VirtualMachine{}, fmt.Errorf("unable to write to %s: %w", ovfBat.Name(), err)
			}

			err = file.Close()
			if err != nil {
				defer file.Close()
				return VirtualMachine{}, fmt.Errorf("unable to close %s: %w", ovfBat.Name(), err)
			}
			ovfCmd = ovfBat.Name()
		}

		//  Execute ovftool script (or batch) here.
		cmd := exec.Command(osShellCmd, osShellCmdOpt, ovfCmd)
		var out bytes.Buffer
		cmd.Stdout = &out
		err = cmd.Run()

		//  Attempt to delete tmp batch file.
		if ovfBat != nil {
			_ = cmd.Wait()
			_ = os.Remove(ovfBat.Name())
		}

		if err != nil {
			return VirtualMachine{}, fmt.Errorf("there was an ovftool error: cmd<%s>; stdout<%s>; err<%w>",
				re.ReplaceAllString(ovfCmd, "vi://****:******@"), out.String(), err)
		}
	}

	// get id (by name)
	vm.Id, err = esxi.getVirtualMachineId(vm.Name)
	if err != nil {
		return VirtualMachine{}, fmt.Errorf("failed to get vm id: %s", err)
	}

	// ovfProperties require packer to power on the VM to inject the properties.
	// Unfortunately, there is no way to know when cloud-init is finished?!?!?  Just need
	// to wait for ovfPropertiesTimer seconds, then shutdown/power-off to continue...
	if hasOvfProperties {
		currentPowerState := esxi.getVirtualMachinePowerState(vm.Id)
		if currentPowerState != "on" {
			return vm, fmt.Errorf("failed to poweron after ovfProperties injection")
		}
		// allow cloud-init to process.
		duration := time.Duration(vm.OvfPropertiesTimer) * time.Second

		time.Sleep(duration)
		_, err = esxi.powerOffVirtualMachine(vm.Id, vm.ShutdownTimeout)
		if err != nil {
			return vm, fmt.Errorf("failed to shutdown after ovfProperties injection")
		}
	}

	// Grow boot disk to boot_disk_size
	bootDiskVmdkPath, _ := esxi.getBootDiskPath(vm.Id)

	_, err = esxi.growVirtualDisk(bootDiskVmdkPath, vm.BootDiskSize)
	if err != nil {
		return vm, fmt.Errorf("failed to grow boot disk: %s", err)
	}

	// make updates to vmx file
	err = esxi.updateVmxContents(true, vm)
	if err != nil {
		return vm, fmt.Errorf("failed to update vmx contents: %s", err)
	}

	return vm, nil
}

func (esxi *Host) getVirtualMachineId(name string) (string, error) {
	var command, id string
	var err error

	command = fmt.Sprintf("vim-cmd vmsvc/getallvms 2>/dev/null |sort -n | "+
		"grep -m 1 \"[0-9] * %s .*%s\" |awk '{print $1}' ", name, name)

	id, err = esxi.Execute(command, "get vm Id")
	logging.V(9).Infof("getVirtualMachineId: result => %s", id)
	if err != nil {
		logging.V(9).Infof("getVirtualMachineId: Failed get vm id => %s", err)
		return "", fmt.Errorf("unable to find a virtual machine corresponding to the name '%s'", name)
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

	// Get location of vmx file on esxi host
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

	if vm.MemSize != 0 {
		re := regexp.MustCompile("memSize = \".*\"")
		regexReplacement = fmt.Sprintf("memSize = \"%d\"", vm.MemSize)
		vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)
	}

	if vm.NumVCpus != 0 {
		if strings.Contains(vmxContents, "numvcpus = ") {
			re := regexp.MustCompile("numvcpus = \".*\"")
			regexReplacement = fmt.Sprintf("numvcpus = \"%d\"", vm.NumVCpus)
			vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)
		} else {
			logging.V(9).Infof("updateVmxContents: Add numVCpu => %d", vm.NumVCpus)
			vmxContents += fmt.Sprintf("\nnumvcpus = \"%d\"", vm.NumVCpus)
		}
	}

	if vm.VirtualHWVer != 0 {
		re := regexp.MustCompile("virtualHW.version = \".*\"")
		regexReplacement = fmt.Sprintf("virtualHW.version = \"%d\"", vm.VirtualHWVer)
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

	// add/modify virtual disks
	var tmpvar string
	var vmxContentsNew string
	var i, j int

	// Remove all disks
	regexReplacement = fmt.Sprintf("")
	for i = 0; i < 4; i++ {
		for j = 0; j < 16; j++ {
			if (i != 0 || j != 0) && j != 7 {
				re := regexp.MustCompile(fmt.Sprintf("scsi%d:%d.*\n", i, j))
				vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)
			}
		}
	}

	// Add disks that are managed by pulumi
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

	// Create/update networks network_interfaces
	// Define default nic type.
	var defaultNetworkType, networkType string
	if vm.NetworkInterfaces[0].NicType != "" {
		defaultNetworkType = vm.NetworkInterfaces[0].NicType
	} else {
		defaultNetworkType = "e1000"
	}

	// If this is first time provisioning, delete all the old ethernet configuration.
	if isNew {
		logging.V(9).Infof("updateVmxContents:Delete old ethernet configuration => %d", i)
		regexReplacement = fmt.Sprintf("")
		for i = 0; i < 9; i++ {
			re := regexp.MustCompile(fmt.Sprintf("ethernet%d.*\n", i))
			vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)
		}
	}

	// Add/Modify virtual networks.
	networkType = ""
	for i, ni := range vm.NetworkInterfaces {
		logging.V(9).Infof("updateVmxContents: ethernet%d", i)

		if len(ni.VirtualNetwork) == 0 && strings.Contains(vmxContents, "ethernet"+strconv.Itoa(i)) {
			// This is Modify (Delete existing network configuration)
			logging.V(9).Infof("updateVmxContents: Modify ethernet%d - Delete existing.", i)
			regexReplacement = fmt.Sprintf("")
			re := regexp.MustCompile(fmt.Sprintf("ethernet%d.*\n", i))
			vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)
		}

		if ni.VirtualNetwork != "" && strings.Contains(vmxContents, "ethernet"+strconv.Itoa(i)) {
			// This is Modify
			logging.V(9).Infof("updateVmxContents: Modify ethernet%d - Modify existing.", i)

			// Modify Network Name
			re := regexp.MustCompile("ethernet" + strconv.Itoa(i) + ".networkName = \".*\"")
			regexReplacement = fmt.Sprintf("ethernet"+strconv.Itoa(i)+".networkName = \"%s\"", ni.VirtualNetwork)
			vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)

			// Modify virtual Device
			re = regexp.MustCompile("ethernet" + strconv.Itoa(i) + ".virtualDev = \".*\"")
			regexReplacement = fmt.Sprintf("ethernet"+strconv.Itoa(i)+".virtualDev = \"%s\"", ni.NicType)
			vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)

			// Modify MAC (dynamic to static only. static to dynamic is not implemented)
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

		if ni.VirtualNetwork != "" && !strings.Contains(vmxContents, "ethernet"+strconv.Itoa(i)) {
			// This is created
			// Set virtual_network name
			logging.V(9).Infof("updateVmxContents: ethernet%d Create New: %s", i, ni.VirtualNetwork)
			tmpvar = fmt.Sprintf("\nethernet%d.networkName = \"%s\"\n", i, ni.VirtualNetwork)
			vmxContentsNew = tmpvar

			// Set mac address
			if ni.MacAddress != "" {
				tmpvar = fmt.Sprintf("ethernet%d.addressType = \"static\"\n", i)
				vmxContentsNew = vmxContentsNew + tmpvar

				tmpvar = fmt.Sprintf("ethernet%d.address = \"%s\"\n", i, ni.MacAddress)
				vmxContentsNew = vmxContentsNew + tmpvar
			}

			// Set network type
			if len(ni.NicType) == 0 {
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

	// Add disk UUID
	if !strings.Contains(vmxContents, "disk.EnableUUID") {
		vmxContents = vmxContents + "\ndisk.EnableUUID = \"TRUE\""
	}

	// Write vmx file to esxi host
	logging.V(9).Infof("updateVmxContents: New vm_name.vmx => %s", vmxContents)

	dstVmxFile, err := esxi.getDstVmxFile(vm.Id)

	vmxContents, err = esxi.CopyFile(strings.Replace(vmxContents, "\\\"", "\"", -1), dstVmxFile, "write vmx file")

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

	// Write vmx file to esxi host
	dstVmxFile, err := esxi.getDstVmxFile(id)
	vmxContents, err = esxi.CopyFile(strings.Replace(vmxContents, "\\\"", "\"", -1), dstVmxFile, "write vmx file")

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

	if strings.Contains(stdout, "Powered off") {
		return "off"
	} else if strings.Contains(stdout, "Powered on") {
		return "on"
	} else if strings.Contains(stdout, "Suspended") {
		return "suspended"
	} else {
		return "Unknown"
	}
}

func (esxi *Host) getVirtualMachineIpAddress(id string, startupTimeout int) string {
	var command, stdout, ipAddress, ipAddress2 string
	var uptime int

	// Check if powered off
	if esxi.getVirtualMachinePowerState(id) != "on" {
		return ""
	}

	// Check uptime of guest.
	uptime = 0
	for uptime < startupTimeout {
		// Primary method to get IP
		command = fmt.Sprintf("vim-cmd vmsvc/get.guest %s 2>/dev/null |sed '1!G;h;$!d' |awk '/deviceConfigId = 4000/,/ipAddress/' |grep -m 1 -oE '((1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])\\.){3}(1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])'", id)
		stdout, _ = esxi.Execute(command, "get ip_address method 1")
		ipAddress = stdout
		if ipAddress != "" {
			return ipAddress
		}

		time.Sleep(3 * time.Second)

		// Get uptime if above failed.
		command = fmt.Sprintf("vim-cmd vmsvc/get.summary %s 2>/dev/null | grep 'uptimeSeconds ='|sed 's/^.*= //g'|sed s/,//g", id)
		stdout, err := esxi.Execute(command, "get uptime")
		if err != nil {
			return ""
		}
		uptime, _ = strconv.Atoi(stdout)
	}

	// Alternate method to get IP
	command = fmt.Sprintf("vim-cmd vmsvc/get.guest %s 2>/dev/null | grep -m 1 '^   ipAddress = ' | grep -oE '((1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])\\.){3}(1?[0-9][0-9]?|2[0-4][0-9]|25[0-5])'", id)
	stdout, _ = esxi.Execute(command, "get ip_address method 2")
	ipAddress2 = stdout
	if ipAddress2 != "" {
		return ipAddress2
	}

	return ""
}

func (vm *VirtualMachine) toMap(keepId ...bool) map[string]interface{} {
	outputs := structToMap(vm)
	if len(keepId) != 0 && !keepId[0] {
		delete(outputs, "id")
	}
	delete(outputs, "sourcePath")
	delete(outputs, "ovfProperties")
	delete(outputs, "ovfPropertiesTimer")

	if vm.BootDiskType == "Unknown" || len(vm.BootDiskType) == 0 {
		delete(outputs, "bootDiskType")
	}

	if len(vm.Info) == 0 {
		delete(outputs, "info")
	}

	// Do network interfaces
	if len(vm.NetworkInterfaces) == 0 || len(vm.NetworkInterfaces[0].VirtualNetwork) == 0 {
		delete(outputs, "networkInterfaces")
	}

	// Do virtual disks
	if len(vm.VirtualDisks) == 0 || len(vm.VirtualDisks[0].VirtualDiskId) == 0 {
		delete(outputs, "virtualDisks")
	}

	return outputs
}
