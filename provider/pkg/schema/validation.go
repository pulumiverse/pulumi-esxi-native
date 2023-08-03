package schema

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

const (
	invalidFormat    = "The property '%s' is invalid! The value %s"
	propertyRequired = "The property '%s' is required!"

	// Maximum values for timeout properties.
	maxShutdownTimeout = 600
	maxStartupTimeout  = 600
	maxOvfProperties   = 6000

	maxVlanId   = 4095
	maxDiskSize = 62000

	// Maximum values for VirtualMachine properties.
	maxNetworkInterfaces = 10
	maxVirtualDisks      = 59
	maxUplinks           = 32
)

// ValidatePortGroup validates a port group resource.
func ValidatePortGroup(resourceToken string, inputs resource.PropertyMap) []*pulumirpc.CheckFailure {
	failures := make(map[string]string)

	// Validate required properties.
	checkRequiredProperty("name", inputs, &failures)
	checkRequiredProperty("vSwitch", inputs, &failures)

	// Validate boolean properties.
	booleanProps := []string{"forgedTransmits", "promiscuousMode", "macChanges"}
	for _, key := range booleanProps {
		if value, has := inputs[resource.PropertyKey(key)]; has {
			strVal := value.StringValue()
			if strVal != "true" && strVal != "false" && strVal != "" {
				failures[key] = fmt.Sprintf(invalidFormat, key, "must be true, false, or empty to inherit")
			}
		}
	}

	validatePropertyValueInBetween("vlan", 0, maxVlanId, inputs, &failures)

	return validateResource(resourceToken, failures)
}

// ValidateResourcePool validates a resource pool resource.
func ValidateResourcePool(resourceToken string, inputs resource.PropertyMap) []*pulumirpc.CheckFailure {
	failures := make(map[string]string)

	// Validate required property "name".
	if property, has := inputs["name"]; !has || property.StringValue() == "/" {
		failures["name"] = fmt.Sprintf(propertyRequired, "name")
	} else if property.StringValue()[0] == '/' {
		failures["name"] = "The property 'name' cannot start with '/'!"
	}

	// Validate "cpuShares" and "memShares".
	validateCPUShares("cpuShares", inputs, &failures)
	validateCPUShares("memShares", inputs, &failures)

	// Validate boolean properties.
	booleanProps := []string{"cpuMinExpandable", "memMinExpandable"}
	for _, key := range booleanProps {
		if value, has := inputs[resource.PropertyKey(key)]; has {
			strVal := value.StringValue()
			if strVal != "true" && strVal != "false" {
				failures[key] = fmt.Sprintf(invalidFormat, key, "must be true or false")
			}
		}
	}

	return validateResource(resourceToken, failures)
}

func ValidateVirtualDisk(resourceToken string, inputs resource.PropertyMap) []*pulumirpc.CheckFailure {
	failures := map[string]string{}

	requiredProps := []string{"name", "diskStore", "directory", "diskType"}
	for _, key := range requiredProps {
		checkRequiredProperty(key, inputs, &failures)
	}

	validateDiskType("diskType", inputs, &failures)

	return validateResource(resourceToken, failures)
}

// ValidateVirtualMachine validates a virtual machine resource.
func ValidateVirtualMachine(resourceToken string, inputs resource.PropertyMap) []*pulumirpc.CheckFailure {
	failures := map[string]string{}

	// Validate required properties.
	requiredProps := []string{"name", "diskStore", "resourcePoolName", "memSize", "numVCpus", "os"}
	for _, key := range requiredProps {
		checkRequiredProperty(key, inputs, &failures)
	}

	validateDiskType("bootDiskType", inputs, &failures)
	validatePropertyValueInBetween("bootDiskSize", 0, maxDiskSize, inputs, &failures)
	validatePropertyValueInBetween("shutdownTimeout", 0, maxShutdownTimeout, inputs, &failures)
	validatePropertyValueInBetween("startupTimeout", 0, maxStartupTimeout, inputs, &failures)
	validatePropertyValueInBetween("ovfPropertiesTimer", 0, maxOvfProperties, inputs, &failures)
	validateKeyValuePairsProperty("ovfProperties", inputs, &failures)
	validateKeyValuePairsProperty("info", inputs, &failures)
	validateVirtualMachineOs(inputs, &failures)
	validateNetworkInterfaces(inputs, &failures)
	validateVirtualDisks(inputs, &failures)

	// TODO: recheck if it is okay
	// "virtualHWVer":
	//	if contains([]string{"0","4","7","8","9","10","11","12","13","14"}, property.StringValue()) {
	//		failures[key] = fmt.Sprintf(invalidFormat, key, "4,7,8,9,10,11,12,13 or 14")
	//	}

	return validateResource(resourceToken, failures)
}

func ValidateVirtualSwitch(resourceToken string, inputs resource.PropertyMap) []*pulumirpc.CheckFailure {
	failures := map[string]string{}

	checkRequiredProperty("name", inputs, &failures)
	validateLinkDiscoveryMode(inputs, &failures)
	validateUplinks(inputs, &failures)

	return validateResource(resourceToken, failures)
}

func checkRequiredProperty(property string, inputs resource.PropertyMap, failures *map[string]string) {
	if _, has := inputs[resource.PropertyKey(property)]; !has {
		(*failures)[property] = fmt.Sprintf(propertyRequired, property)
	}
}

func validatePropertyValueInBetween(key string, min, max float64, inputs resource.PropertyMap, failures *map[string]string) {
	if val, has := inputs[resource.PropertyKey(key)]; has && val.NumberValue() < min || val.NumberValue() > max {
		(*failures)[key] = fmt.Sprintf(invalidFormat, key, fmt.Sprintf("expected to be in the range (%f - %f), got %f", min, max, val.NumberValue()))
	}
}

func validateCPUShares(key string, inputs resource.PropertyMap, failures *map[string]string) {
	if value, has := inputs[resource.PropertyKey(key)]; has {
		strVal := value.StringValue()
		if _, err := strconv.Atoi(strVal); !contains([]string{"low", "normal", "high"}, strVal) || err != nil {
			(*failures)[key] = fmt.Sprintf(invalidFormat, key, fmt.Sprintf("must be low/normal/high/<custom> (%s)", err))
		}
	}
}

func validateDiskType(key string, inputs resource.PropertyMap, failures *map[string]string) {
	if prop, has := inputs[resource.PropertyKey(key)]; has {
		value := prop.StringValue()
		if _, err := strconv.Atoi(value); !contains([]string{"thin", "zeroedthick", "eagerzeroedthick"}, value) && err != nil {
			(*failures)[key] = fmt.Sprintf(invalidFormat, key, fmt.Sprintf("must be one of the thin, zeroedthick, or eagerzeroedthick (%s)", err))
		}
	}
}

func validateVirtualMachineOs(inputs resource.PropertyMap, failures *map[string]string) {
	key := "os"
	if !validateVirtualMachineOsType(inputs[resource.PropertyKey(key)].StringValue()) {
		(*failures)[key] = fmt.Sprintf(invalidFormat, key, "should be from here: https://github.com/josenk/vagrant-vmware-esxi/wiki/VMware-ESXi-6.5-guestOS-types")
	}
}

func validateNetworkInterfaces(inputs resource.PropertyMap, failures *map[string]string) {
	key := "networkInterfaces"
	property, hasProperty := inputs[resource.PropertyKey(key)]
	if !hasProperty {
		return
	}
	items := property.ArrayValue()
	if len(items) > maxNetworkInterfaces {
		(*failures)[key] = fmt.Sprintf(invalidFormat, key, fmt.Sprintf("must contain max %d network interfaces, currently '%d'", maxNetworkInterfaces, len(items)))
	}
	if len(items) > 0 {
		for i, item := range items {
			if _, has := item.ObjectValue()["virtualNetwork"]; !has {
				itemKey := fmt.Sprintf("%s[%d].key", key, i)
				(*failures)[itemKey] = fmt.Sprintf(propertyRequired, itemKey)
			}
			if nicType, has := item.ObjectValue()["nicType"]; has && !validateNicType(nicType.StringValue()) {
				itemKey := fmt.Sprintf("%s[%d].nicType", key, i)
				(*failures)[itemKey] = fmt.Sprintf("The property '%s' must be vlance, flexible, e1000, e1000e, vmxnet, vmxnet2, or vmxnet3!", itemKey)
			}
		}
	}
}

func validateKeyValuePairsProperty(key string, inputs resource.PropertyMap, failures *map[string]string) {
	property, hasProperty := inputs[resource.PropertyKey(key)]
	if hasProperty && len(property.ArrayValue()) > 0 {
		for i, item := range property.ArrayValue() {
			if _, has := item.ObjectValue()["key"]; !has {
				itemKey := fmt.Sprintf("%s[%d].key", key, i)
				(*failures)[itemKey] = fmt.Sprintf(propertyRequired, itemKey)
			}
			if _, has := item.ObjectValue()["value"]; !has {
				itemKey := fmt.Sprintf("%s[%d].value", key, i)
				(*failures)[itemKey] = fmt.Sprintf(propertyRequired, itemKey)
			}
		}
	}
}

func validateVirtualDisks(inputs resource.PropertyMap, failures *map[string]string) {
	key := "virtualDisks"
	property, hasProperty := inputs[resource.PropertyKey(key)]
	if !hasProperty {
		return
	}
	if length := len(property.ArrayValue()); length > maxVirtualDisks {
		(*failures)[key] = fmt.Sprintf(invalidFormat, key, fmt.Sprintf("must contain max %d virtual disks, currently '%d'", maxVirtualDisks, length))
	}
	if len(property.ArrayValue()) > 0 {
		for i, ovfProperty := range property.ArrayValue() {
			if _, has := ovfProperty.ObjectValue()["virtualDiskId"]; !has {
				itemKey := fmt.Sprintf("%s[%d].virtualDiskId", key, i)
				(*failures)[itemKey] = fmt.Sprintf(propertyRequired, itemKey)
			}
			if slot, has := ovfProperty.ObjectValue()["slot"]; has {
				check := validateVirtualDiskSlot(slot.StringValue())
				if check != "ok" {
					itemKey := fmt.Sprintf("%s[%d].slot", key, i)
					(*failures)[itemKey] = fmt.Sprintf("The property '%s' is not valid: %s!", itemKey, check)
				}
			}
		}
	}
}

func validateLinkDiscoveryMode(inputs resource.PropertyMap, failures *map[string]string) {
	key := "linkDiscoveryMode"
	if prop, has := inputs[resource.PropertyKey(key)]; has {
		value := prop.StringValue()
		if !contains([]string{"down", "listen", "advertise", "both"}, value) {
			(*failures)[key] = fmt.Sprintf(invalidFormat, key, "must be one of down, listen, advertise, or both")
		}
	}
}

func validateUplinks(inputs resource.PropertyMap, failures *map[string]string) {
	key := "upLinks"
	if prop, has := inputs[resource.PropertyKey(key)]; has {
		upLinks := prop.ArrayValue()
		if len(upLinks) > maxUplinks {
			(*failures)[key] = fmt.Sprintf(invalidFormat, key, fmt.Sprintf("must contain a maximum of %d up links, currently '%d'", maxUplinks, len(upLinks)))
		}
		for i, upLink := range upLinks {
			if _, has := upLink.ObjectValue()["name"]; !has {
				itemKey := fmt.Sprintf("%s[%d].name", key, i)
				(*failures)[itemKey] = fmt.Sprintf("The property '%s' is required!", itemKey)
			}
		}
	}
}

// validateVirtualMachineOsType checks if the OS type is valid.
func validateVirtualMachineOsType(os string) bool {
	//  All valid Guest OS's
	allGuestOSs := [...]string{
		"amazonlinux",
		"asianux",
		"centos",
		"coreos",
		"darwin",
		"debian",
		"dos",
		"ecomstation",
		"fedora",
		"freebsd",
		"genericlinux",
		"mandrake",
		"mandriva",
		"netware",
		"nld9",
		"oes",
		"openserver",
		"opensuse",
		"oraclelinux",
		"os2",
		"other24xlinux",
		"other26xlinux",
		"other3xlinux",
		"other3xlinux-64",
		"other",
		"otherguest",
		"otherlinux",
		"redhat",
		"rhel",
		"sjds",
		"sles",
		"solaris",
		"suse",
		"turbolinux",
		"ubuntu",
		"unixware",
		"vmkernel",
		"vmwarephoton",
		"win31",
		"win95",
		"win98",
		"windows",
		"windowshyperv",
		"winlonghorn",
		"winme",
		"winnetbusiness",
		"winnetdatacenter",
		"winnetenterprise",
		"winnetstandard",
		"winnetweb",
		"winnt",
		"winvista",
		"winxphome",
		"winxppro",
	}

	os = fmt.Sprintf("%s\n", strings.ToLower(os))
	for i := 0; i < len(allGuestOSs); i++ {
		if strings.Contains(os, allGuestOSs[i]) {
			return true
		}
	}
	return false
}

// validateNicType checks if the NIC type is valid.
func validateNicType(nicType string) bool {
	if nicType == "" {
		return true
	}

	allNICtypes := `
    vlance
    flexible
    e1000
    e1000e
    vmxnet
    vmxnet2
    vmxnet3
	  `
	nicType = fmt.Sprintf(" %s\n", nicType)
	return strings.Contains(allNICtypes, nicType)
}

// validateVirtualDiskSlot checks if the virtual disk slot is valid.
func validateVirtualDiskSlot(slot string) string {
	var result string
	const invalidSciId = 7

	// Split on comma.
	fields := strings.Split(slot+":UnSet", ":")

	// if using simple format
	if fields[1] == "UnSet" {
		fields[1] = fields[0]
		fields[0] = "0"
	}

	field0i, _ := strconv.Atoi(fields[0])
	field1i, _ := strconv.Atoi(fields[1])
	result = "ok"

	if field0i < 0 || field0i > 3 {
		result = "scsi controller id out of range, should be between 0 and 3"
	}
	if field1i < 0 || field1i > 15 {
		result = "scsi id out of range, should be between 0 and 15"
	}
	if field0i == 0 && field1i == 0 {
		result = "scsi id 0 used by boot disk"
	}
	if field1i == invalidSciId {
		result = fmt.Sprintf("scsi id %d not allowed", invalidSciId)
	}

	return result
}

// validateResource checks for failures and generates CheckFailure messages.
func validateResource(resourceToken string, failures map[string]string) []*pulumirpc.CheckFailure {
	checkFailures := make([]*pulumirpc.CheckFailure, 0, len(failures))
	for property, reason := range failures {
		path := fmt.Sprintf("%s.%s", resourceToken, property)
		checkFailures = append(checkFailures, &pulumirpc.CheckFailure{Property: path, Reason: reason})
	}
	return checkFailures
}

// contains checks if an item is present in a collection.
func contains[T comparable](collection []T, value T) bool {
	for _, item := range collection {
		if item == value {
			return true
		}
	}
	return false
}
