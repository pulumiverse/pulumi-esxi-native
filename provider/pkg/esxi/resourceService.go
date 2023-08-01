package esxi

import (
	"fmt"
	"reflect"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/pulumiverse/pulumi-esxi-native/provider/pkg/schema"
)

type functionsMapper map[string]interface{}

type ResourceService struct {
	functions functionsMapper
}

func NewResourceService() *ResourceService {
	return &ResourceService{
		functionsMapper{
			"esxi-native:index:PortGroup:Create":        PortGroupCreate,
			"esxi-native:index:PortGroup:Update":        PortGroupUpdate,
			"esxi-native:index:PortGroup:Delete":        PortGroupDelete,
			"esxi-native:index:PortGroup:Read":          PortGroupRead,
			"esxi-native:index:ResourcePool:Create":     ResourcePoolCreate,
			"esxi-native:index:ResourcePool:Update":     ResourcePoolUpdate,
			"esxi-native:index:ResourcePool:Delete":     ResourcePoolDelete,
			"esxi-native:index:ResourcePool:Read":       ResourcePoolRead,
			"esxi-native:index:VirtualDisk:Create":      VirtualDiskCreate,
			"esxi-native:index:VirtualDisk:Update":      VirtualDiskUpdate,
			"esxi-native:index:VirtualDisk:Delete":      VirtualDiskDelete,
			"esxi-native:index:VirtualDisk:Read":        VirtualDiskRead,
			"esxi-native:index:VirtualMachine:Create":   VirtualMachineCreate,
			"esxi-native:index:VirtualMachine:Update":   VirtualMachineUpdate,
			"esxi-native:index:VirtualMachine:Delete":   VirtualMachineDelete,
			"esxi-native:index:VirtualMachine:Read":     VirtualMachineRead,
			"esxi-native:index:getVirtualMachine":       VirtualMachineGet,
			"esxi-native:index:getVirtualMachineById":   VirtualMachineGet,
			"esxi-native:index:VirtualSwitch:Create":    VirtualSwitchCreate,
			"esxi-native:index:VirtualSwitch:Update":    VirtualSwitchUpdate,
			"esxi-native:index:VirtualSwitch:Delete":    VirtualSwitchDelete,
			"esxi-native:index:VirtualSwitch:Read":      VirtualSwitchRead,
			"esxi-native:index:PortGroup:Validate":      schema.ValidatePortGroup,
			"esxi-native:index:ResourcePool:Validate":   schema.ValidateResourcePool,
			"esxi-native:index:VirtualDisk:Validate":    schema.ValidateVirtualDisk,
			"esxi-native:index:VirtualMachine:Validate": schema.ValidateVirtualMachine,
			"esxi-native:index:VirtualSwitch:Validate":  schema.ValidateVirtualSwitch,
		},
	}
}

func (receiver *ResourceService) Validate(token string, inputs resource.PropertyMap) ([]*pulumirpc.CheckFailure, error) {
	handler, ok := receiver.functions[fmt.Sprintf("%s:Validate", token)]
	if !ok {
		return nil, fmt.Errorf("unknown operation '%s'", token)
	}
	params := []reflect.Value{reflect.ValueOf(token), reflect.ValueOf(inputs)}

	functionHandler := reflect.ValueOf(handler)
	functionResult := functionHandler.Call(params)
	result := functionResult[0].Interface().([]*pulumirpc.CheckFailure)
	return result, nil
}

func (receiver *ResourceService) Invoke(token string, inputs resource.PropertyMap, esxi *Host) (resource.PropertyMap, error) {
	handler, ok := receiver.functions[token]
	if !ok {
		return nil, fmt.Errorf("unknown function '%s'", token)
	}
	params := []reflect.Value{
		reflect.ValueOf(inputs),
		reflect.ValueOf(esxi),
	}

	functionHandler := reflect.ValueOf(handler)
	functionResult := functionHandler.Call(params)
	result := functionResult[0].Interface().(resource.PropertyMap)
	err := functionResult[1].Interface()
	if err != nil {
		return result, err.(error)
	}
	return result, nil
}

func (receiver *ResourceService) Create(token string, inputs resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	token = fmt.Sprintf("%s:Create", token)
	return receiver.call(token, "", inputs, esxi)
}

func (receiver *ResourceService) Update(token string, id string, inputs resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	token = fmt.Sprintf("%s:Update", token)
	return receiver.call(token, id, inputs, esxi)
}

func (receiver *ResourceService) Read(token string, id string, inputs resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	token = fmt.Sprintf("%s:Read", token)
	return receiver.call(token, id, inputs, esxi)
}

func (receiver *ResourceService) Delete(token string, id string, esxi *Host) error {
	token = fmt.Sprintf("%s:Delete", token)
	handler, ok := receiver.functions[token]
	if !ok {
		return fmt.Errorf("unknown operation '%s'", token)
	}

	params := []reflect.Value{
		reflect.ValueOf(id), reflect.ValueOf(esxi),
	}

	functionHandler := reflect.ValueOf(handler)
	functionResult := functionHandler.Call(params)
	err := functionResult[0].Interface()
	if err != nil {
		return err.(error)
	}

	return nil
}

func (receiver *ResourceService) call(token string, id string, inputs resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	handler, ok := receiver.functions[token]
	if !ok {
		return "", nil, fmt.Errorf("unknown operation '%s'", token)
	}
	var params []reflect.Value
	if len(id) > 0 {
		params = []reflect.Value{reflect.ValueOf(id), reflect.ValueOf(inputs), reflect.ValueOf(esxi)}
	} else {
		params = []reflect.Value{reflect.ValueOf(inputs), reflect.ValueOf(esxi)}
	}

	functionHandler := reflect.ValueOf(handler)
	var functionResult []reflect.Value
	functionResult = functionHandler.Call(params)
	resourceId := functionResult[0].Interface().(string)
	resourceData := functionResult[1].Interface().(resource.PropertyMap)
	err := functionResult[2].Interface()
	if err != nil {
		return resourceId, resourceData, err.(error)
	}

	return resourceId, resourceData, nil
}
