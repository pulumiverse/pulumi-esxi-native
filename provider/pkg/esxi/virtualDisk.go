package esxi

import "github.com/pulumi/pulumi/sdk/v3/go/common/resource"

func VirtualDiskCreate(vd VirtualDisk, esxi *Host) (resource.PropertyMap, error) {

	// create vd

	// read vd

	return nil, nil
}

func VirtualDiskUpdate(vd VirtualDisk, esxi *Host) (string, resource.PropertyMap, error) {

	return "", nil, nil
}

func VirtualDiskDelete(id string, esxi *Host) error {

	return nil
}

func VirtualDiskRead(vd VirtualDisk, esxi *Host) (string, resource.PropertyMap, error) {

	return "", nil, nil
}

func parseVirtualDisk(id string, inputs resource.PropertyMap) VirtualDisk {

	return VirtualDisk{
		Name: id,
	}
}
