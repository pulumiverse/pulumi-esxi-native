package esxi

import (
	"reflect"
)

func (vm *VirtualMachine) toMap(keepId ...bool) map[string]interface{} {
	outputs := structToMap(vm)
	if len(keepId) != 0 && !keepId[0] {
		delete(outputs, "id")
	}
	delete(outputs, "cloneFromVirtualMachine")
	delete(outputs, "ovfHostPathSource")
	delete(outputs, "ovfSourceLocalPath")
	delete(outputs, "ovfProperties")
	delete(outputs, "ovfPropertiesTimer")

	if vm.BootDiskType == "Unknown" || vm.BootDiskType == "" {
		delete(outputs, "bootDiskType")
	}

	if len(vm.Info) == 0 {
		delete(outputs, "info")
	}

	// Do network interfaces
	if len(vm.NetworkInterfaces) == 0 || vm.NetworkInterfaces[0].VirtualNetwork == "" {
		delete(outputs, "networkInterfaces")
	}

	// Do virtual disks
	if len(vm.VirtualDisks) == 0 || vm.VirtualDisks[0].VirtualDiskId == "" {
		delete(outputs, "virtualDisks")
	}

	return outputs
}

func structToMap(data interface{}) map[string]interface{} {
	value := reflect.ValueOf(data)
	typ := reflect.TypeOf(data)

	if typ.Kind() == reflect.Ptr {
		value = value.Elem()
		typ = typ.Elem()
	}

	result := make(map[string]interface{})

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := value.Field(i).Interface()
		// Convert the first letter of the field name to lowercase
		key := string(field.Name[0]+32) + field.Name[1:]
		result[key] = fieldValue
	}

	return result
}
