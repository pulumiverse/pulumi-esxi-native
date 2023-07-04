package esxi

import (
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"strings"
)

func ResourcePoolCreate(rp ResourcePool, esxi *Host) (string, resource.PropertyMap, error) {

	return "", nil, nil
}

func ResourcePoolUpdate(rp ResourcePool, esxi *Host) (string, resource.PropertyMap, error) {

	return "", nil, nil
}

func ResourcePoolDelete(id string, esxi *Host) error {
	command := fmt.Sprintf("vim-cmd hostsvc/rsrc/destroy %s", id)

	stdout, err := esxi.Execute(command, "delete resource pool")
	if err != nil {
		return fmt.Errorf("failed to delete resource pool: %s err:%s", stdout, err)
	}

	return nil
}

func ResourcePoolRead(rp ResourcePool, esxi *Host) (string, resource.PropertyMap, error) {

	return "", nil, nil
}

func parseResourcePool(id string, inputs resource.PropertyMap) ResourcePool {
	rp := ResourcePool{}

	if len(id) > 0 {
		rp.Id = id
	}

	rp.Name = inputs["name"].StringValue()

	if property, has := inputs["cpuMin"]; has {
		rp.CpuMin = int(property.NumberValue())
	}
	if property, has := inputs["cpuMinExpandable"]; has {
		rp.CpuMinExpandable = property.StringValue()
	}
	if property, has := inputs["cpuMax"]; has {
		rp.CpuMax = int(property.NumberValue())
	}
	if property, has := inputs["cpuShares"]; has {
		rp.CpuShares = strings.ToLower(property.StringValue())
	}
	if property, has := inputs["memMin"]; has {
		rp.MemMin = int(property.NumberValue())
	}
	if property, has := inputs["memMinExpandable"]; has {
		rp.MemMinExpandable = property.StringValue()
	}
	if property, has := inputs["memMax"]; has {
		rp.MemMax = int(property.NumberValue())
	}
	if property, has := inputs["memShares"]; has {
		rp.MemShares = strings.ToLower(property.StringValue())
	}

	return rp
}
