package esxi

import "github.com/pulumi/pulumi/sdk/v3/go/common/resource"

func PortGroupCreate(pg PortGroup, esxi *Host) (string, resource.PropertyMap, error) {

	return "", nil, nil

}

func PortGroupDelete(id string, esxi *Host) error {

	return nil
}

func PortGroupUpdate(pg PortGroup, esxi *Host) (string, resource.PropertyMap, error) {

	return "", nil, nil
}

func PortGroupRead(pg PortGroup, esxi *Host) (string, resource.PropertyMap, error) {

	return "", nil, nil

}

func parsePortGroup(id string, inputs resource.PropertyMap) PortGroup {

	return PortGroup{
		Name: id,
	}
}
