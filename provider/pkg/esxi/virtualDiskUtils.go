package esxi

import (
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
	"strconv"
	"strings"
)

// Read virtual Disk details
func (esxi *Host) readVirtualDisk(id string) (VirtualDisk, error) {
	var diskStore, diskDir, diskName, diskType, flatSize string
	var diskSize int
	var flatSizeI64 int64
	var s []string

	//  Split id into it's variables
	s = strings.Split(id, "/")
	logging.V(9).Infof("readVirtualDisk: len=%d cap=%d %v", len(s), cap(s), s)
	if len(s) < 6 {
		return VirtualDisk{
			"", "", "", "", 0,
		}, nil
	} else if len(s) > 6 {
		diskDir = strings.Join(s[4:len(s)-1], "/")
	} else {
		diskDir = s[4]
	}
	diskStore = s[3]
	diskName = s[len(s)-1]

	// Test if virtual disk exists
	command := fmt.Sprintf("test -s \"%s\"", id)
	_, err := esxi.Execute(command, "test if virtual disk exists")
	if err != nil {
		return VirtualDisk{
			"", "", "", "", 0,
		}, err
	}

	//  Get virtual disk flat size
	s = strings.Split(diskName, ".")
	if len(s) < 2 {
		return VirtualDisk{
			"", "", "", "", 0,
		}, err
	}
	diskNameFlat := fmt.Sprintf("%s-flat.%s", s[0], s[1])

	command = fmt.Sprintf("ls -l \"/vmfs/volumes/%s/%s/%s\" | awk '{print $5}'",
		diskStore, diskDir, diskNameFlat)
	flatSize, err = esxi.Execute(command, "Get size")
	if err != nil {
		return VirtualDisk{
			"", "", "", "", 0,
		}, err
	}
	flatSizeI64, _ = strconv.ParseInt(flatSize, 10, 64)
	diskSize = int(flatSizeI64 / 1024 / 1024 / 1024)

	// Determine virtual disk type  (only works if Guest is powered off)
	command = fmt.Sprintf("vmkfstools -t0 \"%s\" |grep -q 'VMFS Z- LVID:' && echo true", id)
	isZeroedThick, _ := esxi.Execute(command, "Get disk type.  Is zeroedthick.")

	command = fmt.Sprintf("vmkfstools -t0 \"%s\" |grep -q 'VMFS -- LVID:' && echo true", id)
	isEagerZeroedThick, _ := esxi.Execute(command, "Get disk type.  Is eagerzeroedthick.")

	command = fmt.Sprintf("vmkfstools -t0 \"%s\" |grep -q 'NOMP -- :' && echo true", id)
	isThin, _ := esxi.Execute(command, "Get disk type.  Is thin.")

	if isThin == "true" {
		diskType = "thin"
	} else if isZeroedThick == "true" {
		diskType = "zeroedthick"
	} else if isEagerZeroedThick == "true" {
		diskType = "eagerzeroedthick"
	} else {
		diskType = "Unknown"
	}

	return VirtualDisk{
		diskDir, diskStore, diskType, diskName, diskSize,
	}, err
}
