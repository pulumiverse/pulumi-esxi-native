package esxi

import "github.com/pulumi/pulumi/sdk/v3/go/common/resource"

func VirtualDiskCreateParser(inputs resource.PropertyMap) VirtualMachine {
	vm := VirtualMachine{}

	return vm
}

func VirtualDiskCreate(vm VirtualMachine, esxi *Host) (resource.PropertyMap, error) {

	// create vm

	// read vm

	return vm.ToPropertyMap(), nil
}
