package esxi

import (
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

func CreateVirtualMachine() {

}

func UpdateVirtualMachine() {

}

func DeleteVirtualMachine() {

}

func ReadVirtualMachine() {

}

func GetVirtualMachine(inputs resource.PropertyMap, esxi *Host) (resource.PropertyMap, error) {
	// Replace the below random number implementation with logic specific to your provider
	if !inputs["name"].IsString() {
		return nil, fmt.Errorf("expected input property 'name' of type 'string' but got '%s", inputs["name"].TypeString())
	}

	n := inputs["name"].StringValue()

	outputs := resource.PropertyMap{
		"name": resource.PropertyValue{
			V: n,
		},
	}

	return outputs, nil
}
