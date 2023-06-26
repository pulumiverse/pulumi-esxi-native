package esxi

import (
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"reflect"
)

type FunctionsMapper map[string]interface{}

type Mapper struct {
	Functions FunctionsMapper
}

func NewMapper() *Mapper {
	return &Mapper{
		FunctionsMapper{
			"esxi-native:index:VirtualMachine:Create": CreateVirtualMachine,
			"esxi-native:index:VirtualMachine:Update": UpdateVirtualMachine,
			"esxi-native:index:VirtualMachine:Delete": DeleteVirtualMachine,
			"esxi-native:index:VirtualMachine:Read":   ReadVirtualMachine,
			"esxi-native:index:ResourcePool:Create":   CreateResourcePool,
			"esxi-native:index:ResourcePool:Update":   UpdateResourcePool,
			"esxi-native:index:ResourcePool:Delete":   DeleteResourcePool,
			"esxi-native:index:ResourcePool:Read":     ReadResourcePool,
			"esxi-native:index:VirtualDisk:Create":    CreateVirtualDisk,
			"esxi-native:index:VirtualDisk:Update":    UpdateVirtualDisk,
			"esxi-native:index:VirtualDisk:Delete":    DeleteVirtualDisk,
			"esxi-native:index:VirtualDisk:Read":      ReadVirtualDisk,
			"esxi-native:index:VirtualSwitch:Create":  CreateVirtualSwitch,
			"esxi-native:index:VirtualSwitch:Update":  UpdateVirtualSwitch,
			"esxi-native:index:VirtualSwitch:Delete":  DeleteVirtualSwitch,
			"esxi-native:index:VirtualSwitch:Read":    ReadVirtualSwitch,
			"esxi-native:index:getVirtualMachine":     GetVirtualMachine,
		},
	}
}

func (receiver *Mapper) Invoke(token string, inputs resource.PropertyMap, esxi *Host) (result interface{}, err error) {
	return receiver.call(token, inputs, esxi)
}

func (receiver *Mapper) Create(token string, inputs resource.PropertyMap, esxi *Host) (result interface{}, err error) {
	token = fmt.Sprintf("%s:Create", token)
	return receiver.call(token, inputs, esxi)
}

func (receiver *Mapper) Update(token string, inputs resource.PropertyMap, esxi *Host) (result interface{}, err error) {
	token = fmt.Sprintf("%s:Update", token)
	return receiver.call(token, inputs, esxi)
}

func (receiver *Mapper) Delete(token string, inputs resource.PropertyMap, esxi *Host) (result interface{}, err error) {
	token = fmt.Sprintf("%s:Delete", token)
	return receiver.call(token, inputs, esxi)
}

func (receiver *Mapper) Read(token string, inputs resource.PropertyMap, esxi *Host) (result interface{}, err error) {
	token = fmt.Sprintf("%s:Read", token)
	return receiver.call(token, inputs, esxi)
}

func (receiver *Mapper) call(token string, inputs resource.PropertyMap, esxi *Host) (result interface{}, err error) {
	f := reflect.ValueOf(receiver.Functions[token])
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(inputs)
	in[1] = reflect.ValueOf(esxi)

	var res []reflect.Value
	res = f.Call(in)
	result = res[0].Interface()
	return
}
