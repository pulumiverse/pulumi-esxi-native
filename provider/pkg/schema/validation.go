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
)

func ValidatePortGroup(resourceToken string, inputs resource.PropertyMap) []*pulumirpc.CheckFailure {
	failures := make(map[string]string)

	if _, has := inputs["name"]; !has {
		failures["name"] = fmt.Sprintf(propertyRequired, "name")
	}

	if _, has := inputs["vSwitch"]; !has {
		failures["vSwitch"] = fmt.Sprintf(propertyRequired, "vSwitch")
	}

	for propertyName, property := range inputs {
		key := string(propertyName)
		switch key {
		case "forgedTransmits":
		case "promiscuousMode":
		case "macChanges":
			value := property.StringValue()
			if value != "true" && value != "false" && value != "" {
				failures[key] = fmt.Sprintf(invalidFormat, key, "must be true, false or empty to inherit")
			}
		}
	}

	return validateResource(resourceToken, failures)
}

func ValidateResourcePool(resourceToken string, inputs resource.PropertyMap) []*pulumirpc.CheckFailure {
	failures := make(map[string]string)

	if property, has := inputs["name"]; !has {
		failures["name"] = fmt.Sprintf(propertyRequired, "name")
	} else if value := property.StringValue(); has && value == "/" {
		failures["name"] = fmt.Sprintf(propertyRequired, "name")
	} else if has && value[0] == '/' {
		failures["name"] = "The property 'name' cannot start with '/'!"
	}

	for propertyName, property := range inputs {
		key := string(propertyName)
		switch key {
		case "cpuMinExpandable":
		case "memMinExpandable":
			value := property.StringValue()
			if value != "true" && value != "false" {
				failures[key] = fmt.Sprintf(invalidFormat, key, "must be true or false")
			}
		case "cpuShares":
		case "memShares":
			value := property.StringValue()
			if _, err := strconv.Atoi(value); !contains([]string{"low", "normal", "high"}, value) && err != nil {
				failures[key] = fmt.Sprintf(invalidFormat, key, fmt.Sprintf("must be low/normal/high/<custom> (%s)", err))
			}
		}
	}

	return validateResource(resourceToken, failures)
}

func ValidateVirtualDisk(resourceToken string, inputs resource.PropertyMap) []*pulumirpc.CheckFailure {
	failures := make(map[string]string)

	if _, has := inputs["name"]; !has {
		failures["name"] = fmt.Sprintf(propertyRequired, "name")
	}

	if _, has := inputs["diskStore"]; !has {
		failures["diskStore"] = fmt.Sprintf(propertyRequired, "diskStore")
	}

	if _, has := inputs["directory"]; !has {
		failures["directory"] = fmt.Sprintf(propertyRequired, "directory")
	}

	if _, has := inputs["diskType"]; !has {
		failures["diskType"] = fmt.Sprintf(propertyRequired, "diskType")
	}

	for propertyName, property := range inputs {
		key := string(propertyName)
		if key == "diskType" {
			value := property.StringValue()
			if _, err := strconv.Atoi(value); !contains([]string{"thin", "zeroedthick", "eagerzeroedthick"}, value) && err != nil {
				failures[key] = fmt.Sprintf(invalidFormat, key, fmt.Sprintf("must be one of the thin, zeroedthick or eagerzeroedthick (%s)", err))
			}
		}
	}

	return validateResource(resourceToken, failures)
}

func ValidateVirtualMachine(resourceToken string, inputs resource.PropertyMap) []*pulumirpc.CheckFailure {
	failures := map[string]string{}

	for propertyName, property := range inputs {
		key := string(propertyName)
		switch key {
		case "bootDiskType":
			value := property.StringValue()
			if _, err := strconv.Atoi(value); !contains([]string{"thin", "zeroedthick", "eagerzeroedthick"}, value) && err != nil {
				failures[key] = fmt.Sprintf(invalidFormat, key, fmt.Sprintf("must be one of the thin, zeroedthick or eagerzeroedthick (%s)", err))
			}
		case "bootDiskSize":
			if property.NumberValue() < 1 || property.NumberValue() > 62000 {
				failures[key] = fmt.Sprintf(invalidFormat, key, "should be in beetween 1 and 62000")
			}
		case "shutdownTimeout":
			if property.NumberValue() < 0 || property.NumberValue() > 600 {
				failures[key] = fmt.Sprintf(invalidFormat, key, "should be in beetween 0 and 600")
			}
		case "startupTimeout":
			if property.NumberValue() < 0 || property.NumberValue() > 600 {
				failures[key] = fmt.Sprintf(invalidFormat, key, "should be in beetween 0 and 600")
			}
		case "ovfPropertiesTimer":
			if property.NumberValue() < 0 || property.NumberValue() > 6000 {
				failures[key] = fmt.Sprintf(invalidFormat, key, "should be in beetween 0 and 6000")
			}
		case "os":
			if !validateVirtualMachineOsType(property.StringValue()) {
				failures[key] = fmt.Sprintf(invalidFormat, key, "should be from here: https://github.com/josenk/vagrant-vmware-esxi/wiki/VMware-ESXi-6.5-guestOS-types")
			}
		case "networkInterfaces":
			items := property.ArrayValue()
			if len(items) > 10 {
				failures[key] = fmt.Sprintf(invalidFormat, key, fmt.Sprintf("must contain max 10 network interfaces, currently '%d'", len(items)))
			}
			if len(items) > 0 {
				for i, item := range items {
					if nicType, has := item.ObjectValue()["nicType"]; has && !validateNicType(nicType.StringValue()) {
						itemKey := fmt.Sprintf("%s[%d].nicType", key, i)
						failures[itemKey] = fmt.Sprintf("The property '%s' must be vlance, flexible, e1000, e1000e, vmxnet, vmxnet2 or vmxnet3!", itemKey)
					}
				}
			}
		case "info":
		case "ovfProperties":
			items := property.ArrayValue()
			if len(items) > 0 {
				for i, ovfProperty := range items {
					if _, has := ovfProperty.ObjectValue()["key"]; !has {
						itemKey := fmt.Sprintf("%s[%d].key", key, i)
						failures[itemKey] = fmt.Sprintf(propertyRequired, itemKey)
					}
					if _, has := ovfProperty.ObjectValue()["value"]; !has {
						itemKey := fmt.Sprintf("%s[%d].value", key, i)
						failures[itemKey] = fmt.Sprintf(propertyRequired, itemKey)
					}
				}
			}
		case "virtualDisks":
			items := property.ArrayValue()
			if length := len(items); length > 59 {
				failures[key] = fmt.Sprintf(invalidFormat, key, fmt.Sprintf("must contain max 59 virtual disks, currently '%d'", length))
			}
			if len(items) > 0 {
				for i, ovfProperty := range items {
					itemErrorFormat := "The property '%s' is required!"
					if _, has := ovfProperty.ObjectValue()["virtualDiskId"]; !has {
						itemKey := fmt.Sprintf("%s[%d].virtualDiskId", key, i)
						failures[itemKey] = fmt.Sprintf(itemErrorFormat, itemKey)
					}
					if slot, has := ovfProperty.ObjectValue()["slot"]; has {
						check := validateVirtualDiskSlot(slot.StringValue())
						if check != "ok" {
							itemKey := fmt.Sprintf("%s[%d].slot", key, i)
							failures[itemKey] = fmt.Sprintf("The property '%s' is not valid: %s!", itemKey, check)
						}
					}
				}
			}
			// TODO: recheck if it is okay
			//  case "virtualHWVer":
			//	  if contains([]string{"0","4","7","8","9","10","11","12","13","14"}, property.StringValue()) {
			//		  failures[key] = fmt.Sprintf(invalidFormat, key, "4,7,8,9,10,11,12,13 or 14")
			//	  }
		}
	}

	if _, has := inputs["name"]; !has {
		failures["name"] = fmt.Sprintf(propertyRequired, "name")
	}

	if _, has := inputs["diskStore"]; !has {
		failures["diskStore"] = fmt.Sprintf(propertyRequired, "diskStore")
	}

	return validateResource(resourceToken, failures)
}

func ValidateVirtualSwitch(resourceToken string, inputs resource.PropertyMap) []*pulumirpc.CheckFailure {
	failures := make(map[string]string)

	if _, has := inputs["name"]; !has {
		failures["name"] = fmt.Sprintf(propertyRequired, "name")
	}

	for propertyName, property := range inputs {
		key := string(propertyName)
		switch key {
		case "linkDiscoveryMode":
			value := property.StringValue()
			if !contains([]string{"down", "listen", "advertise", "both"}, value) {
				failures[key] = fmt.Sprintf(invalidFormat, key, "must be one of down, listen, advertise or both")
			}
		case "upLinks":
			upLinks := property.ArrayValue()
			if len(upLinks) > 0 {
				for i, upLink := range upLinks {
					if _, has := upLink.ObjectValue()["name"]; !has {
						itemKey := fmt.Sprintf("%s[%d].name", key, i)
						failures[itemKey] = fmt.Sprintf("The property '%s' is required!", itemKey)
					}
				}
			}
			if length := len(upLinks); length > 32 {
				failures[key] = fmt.Sprintf(invalidFormat, key, fmt.Sprintf("must contain max 32 up links, currently '%d'", length))
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

// Contains checks if an item is present in a collection
func contains[T comparable](collection []T, value T) bool {
	for _, item := range collection {
		if item == value {
			return true
		}
	}
	return false
}

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

func validateVirtualDiskSlot(slot string) string {
	var result string

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
	if field1i == 7 {
		result = "scsi id 7 not allowed"
	}

	return result
}

func validateSCSIType(scsiType string) bool {
	if scsiType == "" {
		return true
	}

	allSCSItypes := `
    lsilogic
    pvscsi
    lsisas1068
	  `
	scsiType = fmt.Sprintf(" %s\n", scsiType)
	return strings.Contains(allSCSItypes, scsiType)
}
