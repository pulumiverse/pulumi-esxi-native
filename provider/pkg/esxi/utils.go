package esxi

import "reflect"

// Contains checks if an item is present in a collection
func Contains[T comparable](collection []T, value T) bool {
	for _, item := range collection {
		if item == value {
			return true
		}
	}
	return false
}

// ContainsValue checks if an item property value is present in a collection
func ContainsValue[V comparable, T any](collection []T, selector func(T) V, value V) bool {
	for _, item := range collection {
		if selector(item) == value {
			return true
		}
	}
	return false
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
					slice[j] = structToMap(field.Index(j).Interface())
				}
				result[key] = slice
			}
		case reflect.Invalid:
			// Handle reflect.Invalid case
			result[key] = nil
		case reflect.Bool:
			result[key] = field.Bool()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			result[key] = field.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			result[key] = field.Uint()
		case reflect.Float32, reflect.Float64:
			result[key] = field.Float()
		case reflect.Complex64, reflect.Complex128:
			result[key] = field.Complex()
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.String, reflect.UnsafePointer:
			// Handle other common cases
			result[key] = field.Interface()
		default:
			// Handle the rest of the cases (unlikely to occur in practice)
			result[key] = field.Interface()
		}
	}

	return result
}
