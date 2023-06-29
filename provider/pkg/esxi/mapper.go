package esxi

import (
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"reflect"
)

type functionMapper struct {
	parser  interface{}
	handler interface{}
}
type functionsMapper map[string]functionMapper

type ResourceService struct {
	functions functionsMapper
}

func NewResourceService() *ResourceService {
	return &ResourceService{
		functionsMapper{
			"esxi-native:index:PortGroup:Create":      functionMapper{PortGroupCreateParser, PortGroupCreate},
			"esxi-native:index:PortGroup:Update":      functionMapper{PortGroupUpdateParser, PortGroupUpdate},
			"esxi-native:index:PortGroup:Delete":      functionMapper{PortGroupDeleteParser, PortGroupDelete},
			"esxi-native:index:PortGroup:Read":        functionMapper{PortGroupReadParser, PortGroupRead},
			"esxi-native:index:ResourcePool:Create":   functionMapper{ResourcePoolCreateParser, ResourcePoolCreate},
			"esxi-native:index:ResourcePool:Update":   functionMapper{ResourcePoolUpdateParser, ResourcePoolUpdate},
			"esxi-native:index:ResourcePool:Delete":   functionMapper{ResourcePoolDeleteParser, ResourcePoolDelete},
			"esxi-native:index:ResourcePool:Read":     functionMapper{ResourcePoolReadParser, ResourcePoolRead},
			"esxi-native:index:VirtualDisk:Create":    functionMapper{VirtualDiskCreateParser, VirtualDiskCreate},
			"esxi-native:index:VirtualDisk:Update":    functionMapper{VirtualDiskUpdateParser, VirtualDiskUpdate},
			"esxi-native:index:VirtualDisk:Delete":    functionMapper{VirtualDiskDeleteParser, VirtualDiskDelete},
			"esxi-native:index:VirtualDisk:Read":      functionMapper{VirtualDiskReadParser, VirtualDiskRead},
			"esxi-native:index:VirtualMachine:Create": functionMapper{VirtualMachineCreateParser, VirtualMachineCreate},
			"esxi-native:index:VirtualMachine:Update": functionMapper{VirtualMachineUpdateParser, VirtualMachineUpdate},
			"esxi-native:index:VirtualMachine:Delete": functionMapper{VirtualMachineDeleteParser, VirtualMachineDelete},
			"esxi-native:index:VirtualMachine:Read":   functionMapper{VirtualMachineReadParser, VirtualMachineRead},
			"esxi-native:index:getVirtualMachine":     functionMapper{VirtualMachineGetParser, VirtualMachineGet},
			"esxi-native:index:VirtualSwitch:Create":  functionMapper{VirtualSwitchCreateParser, VirtualSwitchCreate},
			"esxi-native:index:VirtualSwitch:Update":  functionMapper{VirtualSwitchUpdateParser, VirtualSwitchUpdate},
			"esxi-native:index:VirtualSwitch:Delete":  functionMapper{VirtualSwitchDeleteParser, VirtualSwitchDelete},
			"esxi-native:index:VirtualSwitch:Read":    functionMapper{VirtualSwitchReadParser, VirtualSwitchRead},
		},
	}
}

func (receiver *ResourceService) Invoke(token string, inputs resource.PropertyMap, esxi *Host) (result interface{}, err error) {
	return receiver.call(token, inputs, esxi)
}

func (receiver *ResourceService) Create(token string, inputs resource.PropertyMap, esxi *Host) (result interface{}, err error) {
	token = fmt.Sprintf("%s:Create", token)
	return receiver.call(token, inputs, esxi)
}

func (receiver *ResourceService) Update(token string, inputs resource.PropertyMap, esxi *Host) (result interface{}, err error) {
	token = fmt.Sprintf("%s:Update", token)
	return receiver.call(token, inputs, esxi)
}

func (receiver *ResourceService) Delete(token string, inputs resource.PropertyMap, esxi *Host) (result interface{}, err error) {
	token = fmt.Sprintf("%s:Delete", token)
	return receiver.call(token, inputs, esxi)
}

func (receiver *ResourceService) Read(token string, inputs resource.PropertyMap, esxi *Host) (result interface{}, err error) {
	token = fmt.Sprintf("%s:Read", token)
	return receiver.call(token, inputs, esxi)
}

func (receiver *ResourceService) call(token string, inputs resource.PropertyMap, esxi *Host) (result interface{}, err error) {
	paramsParser := reflect.ValueOf(receiver.functions[token].parser)
	paramsParserResult := paramsParser.Call([]reflect.Value{reflect.ValueOf(inputs)})

	functionHandler := reflect.ValueOf(receiver.functions[token].handler)
	params := []reflect.Value{reflect.ValueOf(paramsParserResult), reflect.ValueOf(esxi)}

	var res []reflect.Value
	res = functionHandler.Call(params)
	result = res[0].Interface()
	return
}
