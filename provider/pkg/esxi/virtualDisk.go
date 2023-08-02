package esxi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
)

func VirtualDiskCreate(inputs resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	vd := parseVirtualDisk("", inputs)
	// create vd
	var id, command string
	var err error

	err = esxi.validateDiskStore(vd.DiskStore)
	if err != nil {
		return "", nil, fmt.Errorf("failed to validate disk store: %w", err)
	}

	// Create dir if required
	command = fmt.Sprintf("mkdir -p \"/vmfs/volumes/%s/%s\"", vd.DiskStore, vd.Directory)
	_, _ = esxi.Execute(command, "create virtual disk dir")

	command = fmt.Sprintf("ls -d \"/vmfs/volumes/%s/%s\"", vd.DiskStore, vd.Directory)
	_, err = esxi.Execute(command, "validate dir exists")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create virtual disk directory: %w", err)
	}

	// id is just the full path name
	id = fmt.Sprintf("/vmfs/volumes/%s/%s/%s", vd.DiskStore, vd.Directory, vd.Name)

	// validate if it exists already
	command = fmt.Sprintf("ls -l \"%s\"", id)
	_, err = esxi.Execute(command, "validate disk store exists")
	if err == nil {
		return "", nil, err
	}

	command = fmt.Sprintf("/bin/vmkfstools -c %dG -d %s \"%s\"", vd.Size, vd.DiskType, id)
	_, err = esxi.Execute(command, "Create virtual disk")
	if err != nil {
		return "", nil, fmt.Errorf("unable to create virtual disk")
	}

	// read vd
	if err == nil {
		return esxi.readVirtualDisk(id)
	} else {
		return "", nil, fmt.Errorf("failed to create virtual disk: %s err: %w", vd.Name, err)
	}
}

func VirtualDiskUpdate(id string, inputs resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	vd := parseVirtualDisk(id, inputs)

	changed, err := esxi.growVirtualDisk(vd.Id, vd.Size)
	if err != nil && !changed {
		return "", nil, fmt.Errorf("failed to grow virtual disk: %w", err)
	}

	return id, inputs, nil
}

func VirtualDiskDelete(id string, esxi *Host) error {
	vd, err := esxi.getVirtualDisk(id)
	if err != nil && strings.Contains(err.Error(), "invalid virtual disk id") {
		return err
	}

	//  Destroy virtual disk.
	command := fmt.Sprintf("/bin/vmkfstools -U \"%s\"", id)
	stdout, err := esxi.Execute(command, "destroy virtual disk")
	if err != nil {
		if strings.Contains(err.Error(), "Process exited with status 255") {
			logging.V(logLevel).Infof("already deleted:%s", id)
		} else {
			logging.V(logLevel).Infof("failed destroy virtual disk id: %s", stdout)
			return fmt.Errorf("failed to destroy virtual disk: %w", err)
		}
	}

	command = fmt.Sprintf("ls -al \"/vmfs/volumes/%s/%s/\" |wc -l", vd.DiskStore, vd.Directory)

	stdout, _ = esxi.Execute(command, "check if storage dir is empty")
	if stdout == "3" {
		{
			//  Delete empty dir.  Ignore stdout and errors.
			command = fmt.Sprintf("rmdir \"/vmfs/volumes/%s/%s\"", vd.DiskStore, vd.Directory)
			_, _ = esxi.Execute(command, "rmdir empty storage dir")
		}
	}

	return nil
}

func VirtualDiskRead(id string, _ resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	return esxi.readVirtualDisk(id)
}

func parseVirtualDisk(id string, inputs resource.PropertyMap) VirtualDisk {
	vd := VirtualDisk{}
	if len(id) > 0 {
		vd.Id = id
	}

	vd.Name = inputs["name"].StringValue()
	if !strings.HasSuffix(vd.Name, ".vmdk") {
		vd.Name = fmt.Sprintf("%s.vmdk", vd.Name)
	}

	vd.DiskStore = inputs["diskStore"].StringValue()
	vd.Directory = inputs["directory"].StringValue()
	vd.DiskType = inputs["diskType"].StringValue()

	if property, has := inputs["size"]; has && property.NumberValue() != 0 {
		vd.Size = int(property.NumberValue())
	} else {
		vd.Size = 1
	}

	return vd
}

func (esxi *Host) readVirtualDisk(id string) (string, resource.PropertyMap, error) {
	vd, err := esxi.getVirtualDisk(id)
	if err != nil && strings.Contains(err.Error(), "invalid virtual disk id") {
		return "", nil, err
	}

	result := vd.toMap()
	return vd.Id, resource.NewPropertyMapFromMap(result), nil
}

func (esxi *Host) validateDiskStore(diskStore string) error {
	var command, stdout string
	var err error

	command = "esxcli storage filesystem list | grep '/vmfs/volumes/.*[VMFS|NFS]' |awk '{for(i=2;i<=NF-5;++i)printf $i\" \" ; printf \"\\n\"}'"
	stdout, err = esxi.Execute(command, "get list of disk stores")
	if err != nil {
		return fmt.Errorf("unable to get list of disk stores: %w", err)
	}

	if !strings.Contains(stdout, diskStore) {
		command = "esxcli storage filesystem rescan"
		_, _ = esxi.Execute(command, "refresh filesystems")

		command = "esxcli storage filesystem list | grep '/vmfs/volumes/.*[VMFS|NFS]' |awk '{for(i=2;i<=NF-5;++i)printf $i\" \" ; printf \"\\n\"}'"
		stdout, err = esxi.Execute(command, "get list of disk stores")
		if err != nil {
			return fmt.Errorf("unable to get list of disk stores: %w", err)
		}
		if !strings.Contains(stdout, diskStore) {
			return fmt.Errorf("disk store %s does not exist; available disk stores: %s", diskStore, stdout)
		}
	}
	return nil
}

func (esxi *Host) growVirtualDisk(id string, size int) (bool, error) {
	var didGrowDisk bool

	current, err := esxi.getVirtualDisk(id)

	if current.Size == size {
		return true, nil
	}

	if current.Size > size {
		return false, fmt.Errorf("not able to shrink virtual disk: %s", id)
	}

	if current.Size < size {
		command := fmt.Sprintf("/bin/vmkfstools -X %dG \"%s\"", size, id)
		stdout, err := esxi.Execute(command, "grow disk")
		if err != nil {
			return false, fmt.Errorf("%s err: %w", stdout, err)
		}
		didGrowDisk = true
	}

	return didGrowDisk, err
}

func (esxi *Host) getVirtualDisk(id string) (VirtualDisk, error) {
	var diskStore, diskDir, diskName, diskType, flatSize string
	var diskSize int
	var flatSizeI64 int64
	var s []string

	const pathParts = 3
	path := strings.TrimPrefix(id, "/vmfs/volumes/")
	// Extract the values from the id string
	parts := strings.Split(path, "/")
	if len(parts) < pathParts {
		return VirtualDisk{}, fmt.Errorf("invalid virtual disk id: '%s'", id)
	}

	// Access the individual parts
	diskStore = parts[0]
	diskName = parts[len(parts)-1]
	if len(parts) == pathParts {
		diskDir = parts[1]
	} else {
		diskDir = strings.TrimLeft(path, fmt.Sprintf("%s/", diskStore))
		diskDir = strings.TrimRight(diskDir, fmt.Sprintf("/%s", diskName))
	}

	// Test if virtual disk exists
	command := fmt.Sprintf("test -s \"%s\"", id)
	stdout, err := esxi.Execute(command, "test if virtual disk exists")
	if err != nil {
		return VirtualDisk{}, fmt.Errorf("virtual disk %s doesn't exist, err: %s %w", id, stdout, err)
	}

	//  Get virtual disk flat size
	const diskNameParts = 2
	s = strings.Split(diskName, ".")
	if len(s) < diskNameParts {
		return VirtualDisk{}, fmt.Errorf("virtual disk name %s is not valid", diskName)
	}
	diskNameFlat := fmt.Sprintf("%s-flat.%s", s[0], s[1])

	command = fmt.Sprintf("ls -l \"/vmfs/volumes/%s/%s/%s\" | awk '{print $5}'",
		diskStore, diskDir, diskNameFlat)
	flatSize, err = esxi.Execute(command, "Get size")
	if err != nil {
		return VirtualDisk{}, fmt.Errorf("failed to read virtual disk %s size, err: %s %w", id, flatSize, err)
	}
	flatSizeI64, _ = strconv.ParseInt(flatSize, 10, 64)
	const bytesSize = 1024
	diskSize = int(flatSizeI64 / bytesSize / bytesSize / bytesSize)

	// Determine virtual disk type  (only works if Guest is powered off)
	command = fmt.Sprintf("vmkfstools -t0 \"%s\" |grep -q 'VMFS Z- LVID:' && echo true", id)
	isZeroedThick, _ := esxi.Execute(command, "Get disk type.  Is zeroedthick.")

	command = fmt.Sprintf("vmkfstools -t0 \"%s\" |grep -q 'VMFS -- LVID:' && echo true", id)
	isEagerZeroedThick, _ := esxi.Execute(command, "Get disk type.  Is eagerzeroedthick.")

	command = fmt.Sprintf("vmkfstools -t0 \"%s\" |grep -q 'NOMP -- :' && echo true", id)
	isThin, _ := esxi.Execute(command, "Get disk type.  Is thin.")

	switch {
	case isThin == trueValue:
		diskType = vdThin
	case isZeroedThick == trueValue:
		diskType = vdZeroedThick
	case isEagerZeroedThick == trueValue:
		diskType = vdEagerZeroedThick
	default:
		diskType = esxiUnknown
	}

	return VirtualDisk{
		diskDir, diskStore, diskType, id, diskName, diskSize,
	}, nil
}

func (vd *VirtualDisk) toMap(keepId ...bool) map[string]interface{} {
	outputs := structToMap(vd)
	if len(keepId) != 0 && !keepId[0] {
		delete(outputs, "id")
	}
	if vd.DiskType == esxiUnknown {
		delete(outputs, "diskType")
	}
	return outputs
}
