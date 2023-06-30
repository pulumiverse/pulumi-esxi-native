package esxi

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

func VirtualMachineGetParser(inputs resource.PropertyMap) string {
	return inputs["name"].StringValue()
}

func VirtualMachineGet(name string, esxi *Host) (resource.PropertyMap, error) {
	id, _ := esxi.getVirtualMachineId(name)

	vm, err := esxi.readVirtualMachine(VirtualMachine{
		Id:             id,
		StartupTimeout: 600,
	})

	if err != nil || vm.Name == "" {
		return nil, err
	}

	result := vm.ToPropertyMap()
	result["id"] = resource.PropertyValue{
		V: id,
	}
	return result, err
}
