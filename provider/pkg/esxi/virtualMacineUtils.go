package esxi

import (
	"bytes"
	"fmt"
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
			return VirtualMachine{}, fmt.Errorf("failed to create guest path. fullPATH: %s", fullPATH)
		}
	}

	hasISO := false
	isoFileName := ""
	// Build VMX file content
	vmxContents := fmt.Sprintf(`config.version = "8"
virtualHW.version = "%d"
displayName = "%s"
numvcpus = "%d"
memSize = "%d"
guestOS = "%s"
annotation = "%s"
floppy0.present = "FALSE"
scsi0.present = "TRUE"
scsi0.sharedBus = "none"
scsi0.virtualDev = "lsilogic"
disk.EnableUUID = "TRUE"
pciBridge0.present = "TRUE"
pciBridge4.present = "TRUE"
pciBridge4.virtualDev = "pcieRootPort"
pciBridge4.functions = "8"
pciBridge5.present = "TRUE"
pciBridge5.virtualDev = "pcieRootPort"
pciBridge5.functions = "8"
pciBridge6.present = "TRUE"
pciBridge6.virtualDev = "pcieRootPort"
pciBridge6.functions = "8"
pciBridge7.present = "TRUE"
pciBridge7.virtualDev = "pcieRootPort"
pciBridge7.functions = "8"
scsi0:0.present = "TRUE"
scsi0:0.fileName = "%s.vmdk"
scsi0:0.deviceType = "scsi-hardDisk"
nvram = "%s.nvram"`, vm.VirtualHWVer, vm.Name, vm.NumVCpus, vm.MemSize, vm.Os, vm.Notes, vm.Name, vm.Name)

	if vm.BootFirmware == "efi" {
		vmxContents += "\nfirmware = \"efi\""
	} else if vm.BootFirmware == "bios" {
		vmxContents += "\nfirmware = \"bios\""
	}

	// Check and set ISO file
	if hasISO {
		vmxContents += fmt.Sprintf(`
ide1:0.present = "TRUE"
ide1:0.fileName = "%s"
ide1:0.deviceType = "cdrom-raw"`, isoFileName)
	} else {
		vmxContents += `
ide1:0.present = "TRUE"
ide1:0.fileName = "emptyBackingString"
ide1:0.deviceType = "atapi-cdrom"
ide1:0.startConnected = "FALSE"
ide1:0.clientDevice = "TRUE"`
	}

	// Write vmx file to esxi host
	dstVmxFile := fmt.Sprintf("%s/%s.vmx", fullPATH, vm.Name)

	_, err := esxi.WriteFile(vmxContents, dstVmxFile, "write vmx file")
	if err != nil {
		return VirtualMachine{}, fmt.Errorf("failed to write vmx file %w", err)
	}

	// Create boot disk (vmdk)
	command = fmt.Sprintf("vmkfstools -c %dG -d %s \"%s/%s.vmdk\"", vm.BootDiskSize, vm.BootDiskType, fullPATH, vm.Name)
	_, err = esxi.Execute(command, "vmkfstools (make boot disk)")
	if err != nil {
		command = fmt.Sprintf("rm -fr \"%s\"", fullPATH)
		_, _ = esxi.Execute(command, "cleanup guest path because of failed events")
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
		_, _ = esxi.Execute(command, "cleanup guest path because of failed events")
		return VirtualMachine{}, fmt.Errorf("failed to register guest:%s", err)
	}

	return vm, nil
}

func (esxi *Host) createVirtualMachine(vm VirtualMachine) (VirtualMachine, error) {
	// Step 1: Check if Disk Store already exists
	err := esxi.validateDiskStore(vm.DiskStore)
	if err != nil {
		return VirtualMachine{}, fmt.Errorf("failed to validate disk store: %s", err)
	}

	// Step 2: Check if guest already exists
	id, err := esxi.getOrCreateVirtualMachine(vm)
	if err != nil {
		return VirtualMachine{}, err
	}

	// Step 3: Handle OVF properties, if present
	err = esxi.handleOvfProperties(id, vm)
	if err != nil {
		return VirtualMachine{}, err
	}

	// Step 4: Grow boot disk to boot_disk_size
	err = esxi.growBootDisk(id, vm.BootDiskSize)
	if err != nil {
		return VirtualMachine{}, err
	}

	// Step 5: Make updates to the vmx file
	err = esxi.updateVmxContents(true, vm)
	if err != nil {
		return VirtualMachine{}, fmt.Errorf("failed to update vmx contents: %s", err)
	}

	return vm, nil
}

// getOrCreateVirtualMachine checks if the virtual machine already exists or creates it if not.
func (esxi *Host) getOrCreateVirtualMachine(vm VirtualMachine) (string, error) {
	id, err := esxi.getVirtualMachineId(vm.Name)
	if err != nil {
		return "", fmt.Errorf("failed to get VM ID: %s", err)
	}

	switch {
	case id != "":
		// VM already exists, power off guest if it's powered on or suspended
		currentPowerState := esxi.getVirtualMachinePowerState(id)
		if currentPowerState == vmTurnedOn || currentPowerState == vmTurnedSuspended {
			esxi.powerOffVirtualMachine(id, vm.ShutdownTimeout)
		}
	case vm.SourcePath == "none":
		// Create a plain virtual machine
		vm, err = esxi.createPlainVirtualMachine(vm)
		if err != nil {
			return "", err
		}
	default:
		// Build VM with ovftool or copy from local source
		err = esxi.buildVirtualMachineFromSource(vm)
		if err != nil {
			return "", err
		}

		// Retrieve the VM ID after building the virtual machine
		id, err = esxi.getVirtualMachineId(vm.Name)
		if err != nil {
			return "", fmt.Errorf("failed to get VM ID: %s", err)
		}
	}

	return id, nil
}

// handleOvfProperties handles OVF properties injection and power off if necessary.
func (esxi *Host) handleOvfProperties(id string, vm VirtualMachine) error {
	if len(vm.OvfProperties) > 0 {
		currentPowerState := esxi.getVirtualMachinePowerState(id)
		if currentPowerState != vmTurnedOn {
			return fmt.Errorf("failed to power on after ovfProperties injection")
		}

		// Allow cloud-init to process.
		duration := time.Duration(vm.OvfPropertiesTimer) * time.Second
		time.Sleep(duration)
		esxi.powerOffVirtualMachine(id, vm.ShutdownTimeout)
	}
	return nil
}

// growBootDisk grows the boot disk to the specified size.
func (esxi *Host) growBootDisk(id string, bootDiskSize int) error {
	bootDiskVmdkPath, _ := esxi.getBootDiskPath(id)
	_, err := esxi.growVirtualDisk(bootDiskVmdkPath, bootDiskSize)
	if err != nil {
		return fmt.Errorf("failed to grow boot disk: %s", err)
	}
	return nil
}

// buildVirtualMachineFromSource builds the virtual machine using ovftool or copies from a local source.
func (esxi *Host) buildVirtualMachineFromSource(vm VirtualMachine) error {
	switch {
	case strings.HasPrefix(vm.SourcePath, "http://") || strings.HasPrefix(vm.SourcePath, "https://"):
		// If the source is a remote URL, check its accessibility
		resp, err := http.Get(vm.SourcePath)
		if err != nil || resp.StatusCode != http.StatusOK {
			return fmt.Errorf("URL not accessible: %s", vm.SourcePath)
		}
		defer resp.Body.Close()
	case strings.HasPrefix(vm.SourcePath, "vi://"):
		logging.V(logLevel).Infof("Source is Guest VM (vi).\n")
	default:
		// If the source is a local file, check if it exists
		if _, err := os.Stat(vm.SourcePath); os.IsNotExist(err) {
			return fmt.Errorf("file not found locally: %s", vm.SourcePath)
		}
	}

	// Set params for packer
	if vm.BootDiskType == "zeroedthick" {
		vm.BootDiskType = "thick"
	}

	username := url.QueryEscape(esxi.Connection.UserName)
	password := url.QueryEscape(esxi.Connection.Password)
	dstPath := fmt.Sprintf("vi://%s:%s@%s:%s/", username, password, esxi.Connection.Host, esxi.Connection.SslPort)
	if vm.ResourcePoolName != "/" {
		dstPath = fmt.Sprintf("%s/%s", dstPath, vm.ResourcePoolName)
	}

	netParam := ""
	if (strings.HasSuffix(vm.SourcePath, ".ova") || strings.HasSuffix(vm.SourcePath, ".ovf")) && len(vm.NetworkInterfaces) > 0 && vm.NetworkInterfaces[0].VirtualNetwork != "" {
		netParam = fmt.Sprintf(" --network='%s'", vm.NetworkInterfaces[0].VirtualNetwork)
	}

	extraParams := "--X:logToConsole --X:logLevel=info"
	if len(vm.OvfProperties) > 0 && (strings.HasSuffix(vm.SourcePath, ".ova") || strings.HasSuffix(vm.SourcePath, ".ovf")) {
		// Inject OVF properties if available
		extraParams = fmt.Sprintf("%s --X:injectOvfEnv --allowExtraConfig --powerOn", extraParams)

		for _, prop := range vm.OvfProperties {
			extraParams = fmt.Sprintf("%s --prop:%s='%s'", extraParams, prop.Key, prop.Value)
		}
	}

	ovfCmd := fmt.Sprintf("ovftool --acceptAllEulas --noSSLVerify --X:useMacNaming=false %s -dm=%s --name='%s' --overwrite -ds='%s'%s '%s' '%s'",
		extraParams, vm.BootDiskType, vm.Name, vm.DiskStore, netParam, vm.SourcePath, dstPath)

	osShellCmd := "/bin/bash"
	osShellCmdOpt := "-c"

	var ovfBat *os.File
	if runtime.GOOS == "windows" {
		// For Windows, create a batch file and execute it
		ovfCmd = strings.ReplaceAll(ovfCmd, "'", "\"")

		var err error
		ovfBat, err = os.CreateTemp("", "ovfCmd*.bat")
		if err != nil {
			return fmt.Errorf("unable to create temporary batch file: %w", err)
		}
		defer os.Remove(ovfBat.Name())

		// Write the ovftool command to the batch file
		file, err := os.Create(ovfBat.Name())
		if err != nil {
			return fmt.Errorf("unable to create batch file: %w", err)
		}
		defer file.Close()

		_, err = file.WriteString(strings.ReplaceAll(ovfCmd, "%", "%%"))
		if err != nil {
			return fmt.Errorf("unable to write to batch file: %w", err)
		}

		err = file.Close()
		if err != nil {
			return fmt.Errorf("unable to close batch file: %w", err)
		}

		ovfCmd = ovfBat.Name()
		osShellCmd = "cmd.exe"
		osShellCmdOpt = "/c"
	}

	// Execute ovftool command
	cmd := exec.Command(osShellCmd, osShellCmdOpt, ovfCmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	// Clean up temporary batch file for Windows
	if runtime.GOOS == "windows" {
		_ = cmd.Wait()
		_ = os.Remove(ovfBat.Name())
	}

	// Check for errors during ovftool execution
	if err != nil {
		return fmt.Errorf("ovftool error: %w; command: %s; stdout: %s", err, ovfCmd, out.String())
	}

	return nil
}

func (esxi *Host) getVirtualMachineId(name string) (string, error) {
	var command, id string
	var err error

	command = fmt.Sprintf("vim-cmd vmsvc/getallvms 2>/dev/null |sort -n | "+
		"grep -m 1 \"[0-9] * %s .*%s\" |awk '{print $1}' ", name, name)

	id, err = esxi.Execute(command, "get vm Id")
	logging.V(logLevel).Infof("getVirtualMachineId: result => %s", id)
	if err != nil {
		logging.V(logLevel).Infof("getVirtualMachineId: Failed get vm id => %s", err)
		return "", fmt.Errorf("unable to find a virtual machine corresponding to the name '%s'", name)
	}

	return id, nil
}

/*
func (esxi *Host) validateVirtualMachineId(id string) (string, error) {
	var command string
	var err error

	command = fmt.Sprintf("vim-cmd vmsvc/getallvms 2>/dev/null | awk '{print $1}' | "+
		"grep '^%s$'", id)

	id, err = esxi.Execute(command, "validate vm id exists")
	logging.V(logLevel).Infof("validateVirtualMachineId: result => %s", id)
	if err != nil {
		logging.V(logLevel).Infof("validateVirtualMachineId: Failed get vm by id => %s", err)
		return "", fmt.Errorf("Failed get vm id: %s\n", err)
	}

	return id, nil
}
*/

func (esxi *Host) getBootDiskPath(id string) (string, error) {
	var command, stdout string
	var err error

	command = fmt.Sprintf("vim-cmd vmsvc/device.getdevices %s | grep -A10 -e 'key = 2000' -e 'key = 3000' -e 'key = 16000'|grep -m 1 fileName", id)
	stdout, err = esxi.Execute(command, "get boot disk")
	if err != nil {
		logging.V(logLevel).Infof("getBootDiskPath: Failed get boot disk path => %s", stdout)
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
	// Read existing vmxContents
	vmxContents, err := esxi.readVmxContents(vm.Id)
	if err != nil {
		return fmt.Errorf("Failed to get vmx contents: %s\n", err)
	}
	if strings.Contains(vmxContents, "Unable to find a VM corresponding") {
		// VM is not found, return without any updates.
		return nil
	}

	// Update VM settings based on provided VirtualMachine struct fields.
	if vm.MemSize != 0 {
		vmxContents = replaceVMXSetting("memSize", vm.MemSize, vmxContents)
	}

	if vm.NumVCpus != 0 {
		vmxContents = replaceVMXSetting("numvcpus", vm.NumVCpus, vmxContents)
	}

	if vm.VirtualHWVer != 0 {
		vmxContents = replaceVMXSetting("virtualHW.version", vm.VirtualHWVer, vmxContents)
	}

	if vm.Os != "" {
		vmxContents = replaceVMXSetting("guestOS", vm.Os, vmxContents)
	}

	vmxContents = replaceVMXSetting("firmware", vm.BootFirmware, vmxContents)

	// Modify annotation
	if vm.Notes != "" {
		vm.Notes = strings.ReplaceAll(vm.Notes, "\"", "|22")
		vmxContents = replaceVMXSetting("annotation", vm.Notes, vmxContents)
	}

	if len(vm.Info) > 0 {
		parsedVmx := ParseVMX(vmxContents)
		for _, config := range vm.Info {
			logging.V(logLevel).Infof("SAVING %s => %s", config.Key, config.Value)
			parsedVmx["guestinfo."+config.Key] = config.Value
		}
		vmxContents = EncodeVMX(parsedVmx)
	}

	// Add/Modify virtual disks
	vmxContents = removeAllDisks(vmxContents)
	vmxContents = addVirtualDisks(vm.VirtualDisks, vmxContents)

	// Create/Update network interfaces
	vmxContents = manageNetworkInterfaces(isNew, vm.NetworkInterfaces, vmxContents)

	// Add disk UUID
	if !strings.Contains(vmxContents, "disk.EnableUUID") {
		vmxContents += "\ndisk.EnableUUID = \"TRUE\""
	}

	// Write updated vmxContents back to ESXi host
	dstVmxFile, err := esxi.getDstVmxFile(vm.Id)
	if err != nil {
		return fmt.Errorf("failed to get destination vmx file: %w", err)
	}

	_, err = esxi.CopyFile(strings.ReplaceAll(vmxContents, "\\\"", "\""), dstVmxFile, "write vmx file")
	if err != nil {
		return fmt.Errorf("failed to write vmx file: %w", err)
	}

	err = esxi.reloadVirtualMachine(vm.Id)
	return err
}

// replaceVMXSetting replaces or adds the given VMX setting in the vmxContents.
func replaceVMXSetting(settingName string, value interface{}, vmxContents string) string {
	re := regexp.MustCompile(settingName + ` = ".*"`)
	regexReplacement := fmt.Sprintf(settingName+` = "%v"`, value)
	return re.ReplaceAllString(vmxContents, regexReplacement)
}

// removeAllDisks removes all disk settings from vmxContents.
func removeAllDisks(vmxContents string) string {
	regexReplacement := ""
	for i := 0; i < 4; i++ {
		for j := 0; j < 16; j++ {
			if (i != 0 || j != 0) && j != 7 {
				re := regexp.MustCompile(fmt.Sprintf("scsi%d:%d.*\n", i, j))
				vmxContents = re.ReplaceAllString(vmxContents, regexReplacement)
			}
		}
	}
	return vmxContents
}

// addVirtualDisks adds the given virtual disks to the vmxContents.
func addVirtualDisks(virtualDisks []VMVirtualDisk, vmxContents string) string {
	for _, vd := range virtualDisks {
		if vd.VirtualDiskId != "" {
			slot := vd.Slot
			vmxContents += fmt.Sprintf(`
scsi%s.deviceType = "scsi-hardDisk"
scsi%s.fileName = "%s"
scsi%s.present = "true"
`, slot, slot, vd.VirtualDiskId, slot)
		}
	}
	return vmxContents
}

// manageNetworkInterfaces creates/updates network interfaces in the vmxContents.
func manageNetworkInterfaces(isNew bool, networkInterfaces []NetworkInterface, vmxContents string) string {
	defaultNetworkType := "e1000"

	if isNew {
		// This is a new VM, delete all old ethernet configurations.
		for i := 0; i < 9; i++ {
			re := regexp.MustCompile(fmt.Sprintf("ethernet%d.*\n", i))
			vmxContents = re.ReplaceAllString(vmxContents, "")
		}
	}

	for i, ni := range networkInterfaces {
		if len(ni.VirtualNetwork) == 0 {
			// No virtual network specified, skip this interface.
			continue
		}

		if !isNew {
			// Check if the ethernet configuration exists for an existing VM.
			if !strings.Contains(vmxContents, "ethernet"+strconv.Itoa(i)) {
				// This is a newly created interface, add its configuration.
				vmxContents += fmt.Sprintf(`
ethernet%d.networkName = "%s"
`, i, ni.VirtualNetwork)
			}
		} else {
			// For a new VM, add its ethernet configuration.
			networkType := defaultNetworkType
			if ni.NicType != "" {
				networkType = ni.NicType
			}
			macAddressSetting := ""
			if ni.MacAddress != "" {
				macAddressSetting = fmt.Sprintf(`
ethernet%d.addressType = "static"
ethernet%d.address = "%s"
`, i, i, ni.MacAddress)
			}

			vmxContents += fmt.Sprintf(`
ethernet%d.networkName = "%s"
ethernet%d.virtualDev = "%s"%s
ethernet%d.present = "TRUE"
`, i, ni.VirtualNetwork, i, networkType, macAddressSetting, i)
		}
	}

	return vmxContents
}

func (esxi *Host) cleanStorageFromVmx(id string) error {
	vmxContents, err := esxi.readVmxContents(id)
	if err != nil {
		logging.V(logLevel).Infof("cleanStorageFromVmx: Failed get vmx contents => %s", err)
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
	_, err = esxi.CopyFile(strings.ReplaceAll(vmxContents, "\\\"", "\""), dstVmxFile, "write vmx file")
	if err != nil {
		return fmt.Errorf("failed to write vmx file %w", err)
	}

	err = esxi.reloadVirtualMachine(id)
	return err
}

func (esxi *Host) reloadVirtualMachine(id string) error {
	command := fmt.Sprintf("vim-cmd vmsvc/reload %s", id)
	_, err := esxi.Execute(command, "vmsvc/reload")

	return err
}

func (esxi *Host) powerOnVirtualMachine(id string) error {
	if esxi.getVirtualMachinePowerState(id) == vmTurnedOn {
		return nil
	}

	command := fmt.Sprintf("vim-cmd vmsvc/power.on %s", id)
	_, err := esxi.Execute(command, "vmsvc/power.on")

	time.Sleep(vmSleepBetweenPowerStateChecks * time.Second)

	if esxi.getVirtualMachinePowerState(id) == vmTurnedOn {
		return nil
	}

	return err
}

// powerOffVirtualMachine powers off a virtual machine on the ESXi host with the given ID.
// If the virtual machine is already turned off, it returns immediately.
// If the virtual machine is turned on, it tries to gracefully shut it down before powering off.
// The shutdownTimeout parameter specifies the maximum time (in seconds) to wait for the VM to shut down.
// If shutdownTimeout is 0, the VM will be powered off immediately without attempting a graceful shutdown.
func (esxi *Host) powerOffVirtualMachine(id string, shutdownTimeout int) {
	savedPowerState := esxi.getVirtualMachinePowerState(id)

	if savedPowerState == vmTurnedOff {
		// VM is already turned off, no need to do anything.
		return
	}

	if savedPowerState == vmTurnedOn {
		if shutdownTimeout > 0 {
			// Try to gracefully shut down the VM first.
			command := fmt.Sprintf("vim-cmd vmsvc/power.shutdown %s", id)
			_, _ = esxi.Execute(command, "vmsvc/power.shutdown")
			time.Sleep(vmSleepBetweenPowerStateChecks * time.Second)

			for i := 0; i < (shutdownTimeout / vmSleepBetweenPowerStateChecks); i++ {
				if esxi.getVirtualMachinePowerState(id) == vmTurnedOff {
					// VM is successfully shut down.
					return
				}
				time.Sleep(vmSleepBetweenPowerStateChecks * time.Second)
			}
		}

		// VM is either still running after the timeout or no graceful shutdown attempted.
		// Power off the VM forcefully.
		command := fmt.Sprintf("vim-cmd vmsvc/power.off %s", id)
		_, _ = esxi.Execute(command, "vmsvc/power.off")
		time.Sleep(1 * time.Second)

		return
	}

	// VM power state is unknown, just power it off forcefully.
	command := fmt.Sprintf("vim-cmd vmsvc/power.off %s", id)
	_, _ = esxi.Execute(command, "vmsvc/power.off")
}

func (esxi *Host) getVirtualMachinePowerState(id string) string {
	command := fmt.Sprintf("vim-cmd vmsvc/power.getstate %s", id)
	stdout, _ := esxi.Execute(command, "vmsvc/power.getstate")
	if strings.Contains(stdout, "Unable to find a VM corresponding") {
		return "Unknown"
	}

	switch {
	case strings.Contains(stdout, "Powered off"):
		return vmTurnedOff
	case strings.Contains(stdout, "Powered on"):
		return vmTurnedOn
	case strings.Contains(stdout, "Suspended"):
		return vmTurnedSuspended
	default:
		return "Unknown"
	}
}

func (esxi *Host) getVirtualMachineIpAddress(id string, startupTimeout int) string {
	var command, stdout, ipAddress, ipAddress2 string
	var uptime int

	// Check if powered off
	if esxi.getVirtualMachinePowerState(id) != vmTurnedOn {
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

		time.Sleep(vmSleepBetweenPowerStateChecks * time.Second)

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
