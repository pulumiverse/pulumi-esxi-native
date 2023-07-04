package esxi

import (
	"reflect"
)

func (pg *PortGroup) toMap(keepId ...bool) map[string]interface{} {
	outputs := structToMap(pg)
	if len(keepId) != 0 && !keepId[0] {
		delete(outputs, "id")
	}

	return outputs
}

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

	if vm.BootDiskType == "Unknown" || len(vm.BootDiskType) == 0 {
		delete(outputs, "bootDiskType")
	}

	if len(vm.Info) == 0 {
		delete(outputs, "info")
	}

	// Do network interfaces
	if len(vm.NetworkInterfaces) == 0 || len(vm.NetworkInterfaces[0].VirtualNetwork) == 0 {
		delete(outputs, "networkInterfaces")
	}

	// Do virtual disks
	if len(vm.VirtualDisks) == 0 || len(vm.VirtualDisks[0].VirtualDiskId) == 0 {
		delete(outputs, "virtualDisks")
	}

	return outputs
}

func (vs *VirtualSwitch) toMap(keepId ...bool) map[string]interface{} {
	outputs := structToMap(vs)
	if len(keepId) != 0 && !keepId[0] {
		delete(outputs, "id")
	}

	// Do up links
	if len(vs.Uplinks) == 0 || len(vs.Uplinks[0].Name) == 0 {
		delete(outputs, "uplinks")
	}

	return outputs
}

func structToMap(dataStruct interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	value := reflect.ValueOf(dataStruct)
	typeOfStruct := value.Type()

	if typeOfStruct.Kind() == reflect.Ptr {
		value = value.Elem()
		typeOfStruct = typeOfStruct.Elem()
	}

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := typeOfStruct.Field(i)

		// Convert the first letter of the field name to lowercase
		key := string(fieldType.Name[0]+32) + fieldType.Name[1:]

		switch field.Kind() {
		case reflect.Struct:
			result[key] = structToMap(field.Interface())
		case reflect.Array, reflect.Slice:
			if field.Len() > 0 {
				slice := make([]interface{}, field.Len())
				for j := 0; j < field.Len(); j++ {
					switch field.Index(j).Kind() {
					case reflect.Struct:
						slice[j] = structToMap(field.Index(j).Interface())
					default:
						slice[j] = field.Index(j).Interface()
					}
				}
				result[key] = slice
			}
		default:
			result[key] = field.Interface()
		}
	}

	return result
}
