package esxi

import (
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

func VirtualMachineGetParser(inputs resource.PropertyMap) string {
	return inputs["name"].StringValue()
}

func VirtualMachineGet(name string, esxi *Host) (resource.PropertyMap, error) {
	id, err := esxi.getVirtualMachineId(name)
	if err != nil {
		return nil, fmt.Errorf("unable to find a virtual machine corresponding to the name '%s'", name)
	}

	return VirtualMachineGetById(id, esxi)
}

func VirtualMachineGetByIdParser(inputs resource.PropertyMap) string {
	return inputs["id"].StringValue()
}

func VirtualMachineGetById(id string, esxi *Host) (resource.PropertyMap, error) {
	vm := esxi.readVirtualMachine(VirtualMachine{
		Id:             id,
		StartupTimeout: 1,
	})

	if len(vm.Name) == 0 {
		return nil, fmt.Errorf("unable to find a virtual machine corresponding to the id '%s'", vm.Id)
	}

	result := vm.toMap(true)
	return resource.NewPropertyMapFromMap(result), nil
}
