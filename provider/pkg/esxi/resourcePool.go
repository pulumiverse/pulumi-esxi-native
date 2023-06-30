package esxi

import "github.com/pulumi/pulumi/sdk/v3/go/common/resource"

func ResourcePoolCreate(rp ResourcePool, esxi *Host) (string, resource.PropertyMap, error) {

	return "", nil, nil
}

func ResourcePoolUpdate(rp ResourcePool, esxi *Host) (string, resource.PropertyMap, error) {

	return "", nil, nil
}

func ResourcePoolDelete(id string, esxi *Host) error {

	return nil
}

func ResourcePoolRead(rp ResourcePool, esxi *Host) (string, resource.PropertyMap, error) {

	return "", nil, nil
}

func parseResourcePool(id string, inputs resource.PropertyMap) ResourcePool {

	return ResourcePool{
		Name: id,
	}
}
