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
