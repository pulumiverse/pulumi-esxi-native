package esxi

import (
	"fmt"
	"github.com/jszwec/csvutil"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"regexp"
	"strconv"
	"strings"
)

func PortGroupCreate(inputs resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	var pg PortGroup
	if parsed, err := parsePortGroup("", inputs); err == nil {
		pg = parsed
	} else {
		return "", nil, err
	}

	command := fmt.Sprintf("esxcli network vswitch standard portgroup add -v \"%s\" -p \"%s\"",
		pg.VSwitch, pg.Name)

	stdout, err := esxi.Execute(command, "create port group")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create port group: %s err:%s", stdout, err)
	}

	err = esxi.updatePortGroup(pg)
	if err != nil {
		return "", nil, err
	}

	return esxi.readPortGroup(pg)
}

func PortGroupUpdate(id string, inputs resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	var pg PortGroup
	if parsed, err := parsePortGroup(id, inputs); err == nil {
		pg = parsed
	} else {
		return "", nil, err
	}

	err := esxi.updatePortGroup(pg)
	if err != nil {
		return "", nil, err
	}

	return esxi.readPortGroup(pg)
}

func PortGroupDelete(id string, esxi *Host) error {
	var command string

	if name, vSwitch, err := extractId(id); err == nil {
		command = fmt.Sprintf("esxcli network vswitch standard portgroup remove -v \"%s\" -p \"%s\"",
			vSwitch, name)
	} else {
		return err
	}

	stdout, err := esxi.Execute(command, "delete port group")
	if err != nil {
		return fmt.Errorf("failed to delete port group: %s err:%s", stdout, err)
	}

	return nil
}

func PortGroupRead(id string, inputs resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	var pg PortGroup
	if parsed, err := parsePortGroup(id, inputs); err == nil {
		pg = parsed
	} else {
		return "", nil, err
	}

	return esxi.readPortGroup(pg)
}

func extractId(id string) (name, vSwitch string, err error) {
	if idParts := strings.Split(id, "/"); len(id) > 0 && len(idParts) == 2 {
		name = idParts[1]
		vSwitch = idParts[0]
		err = nil
	} else {
		name = ""
		vSwitch = ""
		err = fmt.Errorf("port group id is invalid %s", id)
	}

	return name, vSwitch, err
}

func parsePortGroup(id string, inputs resource.PropertyMap) (PortGroup, error) {
	pg := PortGroup{}

	if len(id) > 0 {
		if name, vSwitch, err := extractId(id); err == nil {
			pg.Name = name
			pg.VSwitch = vSwitch
		} else if _, hasName := inputs["name"]; !hasName {
			// it isn't called rather from create or update
			return pg, err
		}
	} else {
		// already controlled for their existence during the validation
		pg.Name = inputs["name"].StringValue()
		pg.VSwitch = inputs["vSwitch"].StringValue()
	}

	pg.Id = fmt.Sprintf("%s/%s", pg.VSwitch, pg.Name)

	pg.Vlan = int(inputs["vlan"].NumberValue())

	if property, has := inputs["promiscuousMode"]; has {
		pg.PromiscuousMode = property.StringValue()
	} else {
		pg.PromiscuousMode = ""
	}

	if property, has := inputs["macChanges"]; has {
		pg.MacChanges = property.StringValue()
	} else {
		pg.MacChanges = ""
	}

	if property, has := inputs["forgedTransmits"]; has {
		pg.ForgedTransmits = property.StringValue()
	} else {
		pg.ForgedTransmits = ""
	}

	return pg, nil
}

func (esxi *Host) updatePortGroup(pg PortGroup) error {
	command := fmt.Sprintf("esxcli network vswitch standard portgroup set -v \"%s\" -p \"%s\"",
		pg.Vlan, pg.Name)

	stdout, err := esxi.Execute(command, "port group set vlan")
	if err != nil {
		return fmt.Errorf("failed to set port group vlan: %s err:%s", stdout, err)
	}

	// set the security policies.
	if len(pg.PromiscuousMode) > 0 {
		command = fmt.Sprintf("--allow-promiscuous=%s ", pg.PromiscuousMode)
	}
	if len(pg.ForgedTransmits) > 0 {
		command = fmt.Sprintf("%s--allow-forged-transmits=%s ", command, pg.ForgedTransmits)
	}
	if len(pg.MacChanges) > 0 {
		command = fmt.Sprintf("%s--allow-mac-change=%s ", command, pg.MacChanges)
	}
	command = fmt.Sprintf("esxcli network vswitch standard portgroup policy security set -p \"%s\" -u %s", pg.Name, command)
	stdout, err = esxi.Execute(command, "port group set security policy")
	if err != nil {
		return fmt.Errorf("failed to set port group security policy: %s err:%s", stdout, err)
	}

	return nil
}

func (esxi *Host) readPortGroup(pg PortGroup) (string, resource.PropertyMap, error) {
	//  get port group info
	command := fmt.Sprintf("esxcli network vswitch standard portgroup list | grep -m 1 \"^%s  \"", pg.Name)

	stdout, err := esxi.Execute(command, "port group list")
	if stdout == "" {
		return "", nil, fmt.Errorf("failed to list port group: %s err: %s", stdout, err)
	}

	re, _ := regexp.Compile("( {2}.* {2})  +[0-9]+  +[0-9]+$")
	if len(re.FindStringSubmatch(stdout)) > 0 {
		pg.VSwitch = strings.Trim(re.FindStringSubmatch(stdout)[1], " ")
	} else {
		pg.VSwitch = ""
	}

	re, _ = regexp.Compile("  +([0-9]+)$")
	if len(re.FindStringSubmatch(stdout)) > 0 {
		pg.Vlan, _ = strconv.Atoi(re.FindStringSubmatch(stdout)[1])
	} else {
		pg.Vlan = 0
	}

	policy, err := esxi.readPortGroupSecurityPolicy(pg.Name)
	if err != nil {
		return "", nil, err
	}

	pg.MacChanges = strconv.FormatBool(policy.AllowMACAddressChange)
	pg.ForgedTransmits = strconv.FormatBool(policy.AllowForgedTransmits)
	pg.PromiscuousMode = strconv.FormatBool(policy.AllowPromiscuous)

	result := pg.toMap()
	return pg.Id, resource.NewPropertyMapFromMap(result), nil
}

func (esxi *Host) readPortGroupSecurityPolicy(name string) (*PortGroupSecurityPolicy, error) {
	command := fmt.Sprintf("esxcli --formatter=csv network vswitch standard portgroup policy security get -p \"%s\"", name)
	stdout, err := esxi.Execute(command, "port group security policy")
	if stdout == "" {
		return nil, fmt.Errorf("failed to get the port group security policy: %s\n%s\n", stdout, err)
	}

	var policies []PortGroupSecurityPolicy
	if err = csvutil.Unmarshal([]byte(stdout), &policies); err != nil || len(policies) != 1 {
		return nil, fmt.Errorf("failed to parse the port group security policy: %s\n%s\n", stdout, err)
	}

	return &policies[0], nil
}

func (pg *PortGroup) toMap(keepId ...bool) map[string]interface{} {
	outputs := structToMap(pg)
	if len(keepId) != 0 && !keepId[0] {
		delete(outputs, "id")
	}
	return outputs
}
