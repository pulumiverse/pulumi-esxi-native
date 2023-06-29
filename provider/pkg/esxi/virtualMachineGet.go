package esxi

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

type VirtualMachineGetParams struct {
	name string
}

func VirtualMachineGetParser(inputs resource.PropertyMap) (VirtualMachineGetParams, error) {
	n := inputs["name"].StringValue()
	return VirtualMachineGetParams{
		name: n,
	}, nil
}

func VirtualMachineGet(params VirtualMachineGetParams, esxi *Host) (resource.PropertyMap, error) {
	outputs := resource.PropertyMap{
		"name": resource.PropertyValue{
			V: params.name,
		},
	}

	return outputs, nil
}
