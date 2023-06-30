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

func (receiver *ResourceService) Create(token string, inputs resource.PropertyMap, esxi *Host) (rId string, result resource.PropertyMap, err error) {
	token = fmt.Sprintf("%s:Create", token)
	return receiver.call(token, "", inputs, esxi)
}

func (receiver *ResourceService) Update(token string, id string, inputs resource.PropertyMap, esxi *Host) (rId string, result resource.PropertyMap, err error) {
	token = fmt.Sprintf("%s:Update", token)
	return receiver.call(token, id, inputs, esxi)
}

func (receiver *ResourceService) Delete(token string, id string, inputs resource.PropertyMap, esxi *Host) (rId string, result resource.PropertyMap, err error) {
	token = fmt.Sprintf("%s:Delete", token)
	return receiver.call(token, id, inputs, esxi)
}

func (receiver *ResourceService) Read(token string, id string, inputs resource.PropertyMap, esxi *Host) (rId string, result resource.PropertyMap, err error) {
	token = fmt.Sprintf("%s:Read", token)
	return receiver.call(token, id, inputs, esxi)
}

func (receiver *ResourceService) Invoke(token string, inputs resource.PropertyMap, esxi *Host) (result resource.PropertyMap, err error) {
	mapper := receiver.functions[token]
	params := mapper.getParams("", inputs, esxi)
	functionHandler := reflect.ValueOf(mapper.handler)
	var res []reflect.Value
	res = functionHandler.Call(params)
	result = res[0].Interface().(resource.PropertyMap)
	return
}

func (receiver *ResourceService) call(token string, id string, inputs resource.PropertyMap, esxi *Host) (rId string, result resource.PropertyMap, err error) {
	mapper := receiver.functions[token]
	params := mapper.getParams(id, inputs, esxi)
	functionHandler := reflect.ValueOf(mapper.handler)
	var res []reflect.Value
	res = functionHandler.Call(params)
	rId = res[0].Interface().(string)
	result = res[1].Interface().(resource.PropertyMap)
	return
}

func (m *functionMapper) getParams(id string, inputs resource.PropertyMap, esxi *Host) []reflect.Value {
	var parsedParams []reflect.Value
	parser := reflect.ValueOf(m.parser)
	if len(id) > 0 {
		parsedParams = parser.Call([]reflect.Value{reflect.ValueOf(id), reflect.ValueOf(inputs)})
	} else {
		parsedParams = parser.Call([]reflect.Value{reflect.ValueOf(inputs)})
	}

	params := make([]reflect.Value, len(parsedParams)+1)
	esxiIndex := 0
	for i, parsedParam := range parsedParams {
		esxiIndex = i + 1
		params[i] = parsedParam
	}
	params[esxiIndex] = reflect.ValueOf(esxi)

	return params
}
