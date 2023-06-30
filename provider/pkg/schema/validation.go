package schema

import (
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

func ValidatePortGroup(resourceToken string, inputs resource.PropertyMap) []*pulumirpc.CheckFailure {
	failures := make(map[string]string)

	if _, has := inputs["name"]; !has {
		failures["name"] = "The properly 'name' is required!"
	}

	return validateResource(resourceToken, failures)
}

func ValidateResourcePool(resourceToken string, inputs resource.PropertyMap) []*pulumirpc.CheckFailure {
	failures := make(map[string]string)

	if _, has := inputs["name"]; !has {
		failures["name"] = "The properly 'name' is required!"
	}

	return validateResource(resourceToken, failures)
}

func ValidateVirtualDisk(resourceToken string, inputs resource.PropertyMap) []*pulumirpc.CheckFailure {
	failures := make(map[string]string)

	if _, has := inputs["name"]; !has {
		failures["name"] = "The properly 'name' is required!"
	}

	if _, has := inputs["diskStore"]; !has {
		failures["diskStore"] = "The properly 'diskStore' is required!"
	}

	if _, has := inputs["directory"]; !has {
		failures["directory"] = "The properly 'directory' is required!"
	}

	if _, has := inputs["diskType"]; !has {
		failures["directory"] = "The properly 'diskType' is required!"
	}

	return validateResource(resourceToken, failures)
}

func ValidateVirtualMachine(resourceToken string, inputs resource.PropertyMap) []*pulumirpc.CheckFailure {
	failures := map[string]string{}
	invalidFormat := "The properly '%s' is invalid! The value %s"

	for propertyName, property := range inputs {
		key := string(propertyName)
		switch key {
		case "shutdownTimeout":
			if property.NumberValue() >= 0 && property.NumberValue() <= 600 {
				failures[key] = fmt.Sprintf(invalidFormat, key, "in beetween 0 and 600")
			}
		case "startupTimeout":
			if property.NumberValue() >= 0 && property.NumberValue() <= 600 {
				failures[key] = fmt.Sprintf(invalidFormat, key, "in beetween 0 and 600")
			}
		case "ovfPropertiesTimer":
			if property.NumberValue() >= 0 && property.NumberValue() <= 6000 {
				failures[key] = fmt.Sprintf(invalidFormat, key, "in beetween 0 and 6000")
			}
		case "info":
		case "ovfProperties":
			if items := property.ArrayValue(); len(items) > 0 {
				for i, ovfProperty := range items {
					itemErrorFormat := "The properly '%s' is required!"
					if _, has := ovfProperty.ObjectValue()["key"]; !has {
						itemKey := fmt.Sprintf("%s[%d]key", key, i)
						failures[itemKey] = fmt.Sprintf(itemErrorFormat, itemKey)
					}
					if _, has := ovfProperty.ObjectValue()["value"]; !has {
						itemKey := fmt.Sprintf("%s[%d]value", key, i)
						failures[itemKey] = fmt.Sprintf(itemErrorFormat, itemKey)
					}
				}
			}
		case "virtualDisks":
			if items := property.ArrayValue(); len(items) > 0 {
				for i, ovfProperty := range items {
					itemErrorFormat := "The properly '%s' is required!"
					if _, has := ovfProperty.ObjectValue()["virtualDiskId"]; !has {
						itemKey := fmt.Sprintf("%s[%d]virtualDiskId", key, i)
						failures[itemKey] = fmt.Sprintf(itemErrorFormat, itemKey)
					}
				}
			}
		}
	}

	if _, has := inputs["name"]; !has {
		failures["name"] = "The properly 'name' is required!"
	}

	if _, has := inputs["diskStore"]; !has {
		failures["diskStore"] = "The properly 'diskStore' is required!"
	}

	if _, has := inputs["resourcePoolName"]; !has {
		failures["resourcePoolName"] = "The properly 'resourcePoolName' is required!"
	}

	if _, has := inputs["memSize"]; !has {
		failures["memSize"] = "The properly 'memSize' is required!"
	}

	if _, has := inputs["numVCpus"]; !has {
		failures["numVCpus"] = "The properly 'numVCpus' is required!"
	}

	if _, has := inputs["os"]; !has {
		failures["os"] = "The properly 'os' is required!"
	}

	return validateResource(resourceToken, failures)
}

func ValidateVirtualSwitch(resourceToken string, inputs resource.PropertyMap) []*pulumirpc.CheckFailure {
	failures := make(map[string]string)

	if _, has := inputs["name"]; !has {
		failures["name"] = "The properly 'name' is required!"
	}

	for propertyName, property := range inputs {
		key := string(propertyName)
		switch key {
		case "upLinks":
			if upLinks := property.ArrayValue(); len(upLinks) > 0 {
				for i, upLink := range upLinks {
					if _, has := upLink.ObjectValue()["name"]; !has {
						itemKey := fmt.Sprintf("%s[%d]name", key, i)
						failures[itemKey] = fmt.Sprintf("The properly '%s' is required!", itemKey)
					}
				}
			}
		}
	}

	return validateResource(resourceToken, failures)
}

func validateResource(resourceToken string, failures map[string]string) []*pulumirpc.CheckFailure {
	var checkFailures []*pulumirpc.CheckFailure
	for property, reason := range failures {
		path := fmt.Sprintf("%s.%s", resourceToken, property)
		checkFailures = append(checkFailures, &pulumirpc.CheckFailure{Property: path, Reason: reason})
	}

	return checkFailures
}
