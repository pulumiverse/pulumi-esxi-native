package esxi

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

func VirtualSwitchCreate(inputs resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	var vs VirtualSwitch
	if parsed, err := parseVirtualSwitch("", inputs); err == nil {
		vs = parsed
	} else {
		return "", nil, err
	}

	//  Create vswitch
	command := fmt.Sprintf("esxcli network vswitch standard add -P %d -v \"%s\"", vs.Ports, vs.Name)
	stdout, err := esxi.Execute(command, "create vswitch")
	if strings.Contains(stdout, "this name already exists") {
		return "", nil, fmt.Errorf("failed to create vswitch: %s, it already exists", vs.Name)
	}
	if err != nil {
		return "", nil, fmt.Errorf("failed to create vswitch: %s err: %s", stdout, err)
	}

	var somethingWentWrong string
	err = esxi.updateVirtualSwitch(vs)
	if err != nil {
		somethingWentWrong = fmt.Sprintf("failed to update vswitch: %s", err)
	}

	// Refresh
	id, result, err := esxi.readVirtualSwitch(vs.Name)
	if err != nil {
		return "", nil, err
	}

	if somethingWentWrong != "" {
		return "", nil, fmt.Errorf(somethingWentWrong)
	}

	return id, result, nil
}

func VirtualSwitchUpdate(id string, inputs resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	var vs VirtualSwitch
	if parsed, err := parseVirtualSwitch(id, inputs); err == nil {
		vs = parsed
	} else {
		return "", nil, err
	}

	err := esxi.updateVirtualSwitch(vs)
	if err != nil {
		return "", nil, fmt.Errorf("failed to update vswitch: %s", err)
	}

	return esxi.readVirtualSwitch(vs.Name)
}

func VirtualSwitchDelete(id string, esxi *Host) error {
	command := fmt.Sprintf("esxcli network vswitch standard remove -v \"%s\"", id)

	stdout, err := esxi.Execute(command, "delete vswitch")
	if err != nil {
		return fmt.Errorf("failed to delete vswitch: %s err: %s", stdout, err)
	}

	return nil
}

func VirtualSwitchRead(id string, _ resource.PropertyMap, esxi *Host) (string, resource.PropertyMap, error) {
	return esxi.readVirtualSwitch(id)
}

func parseVirtualSwitch(id string, inputs resource.PropertyMap) (VirtualSwitch, error) {
	vs := VirtualSwitch{}

	if len(id) > 0 {
		vs.Id = id
		vs.Name = id
	} else {
		vs.Name = inputs["name"].StringValue()
		vs.Id = vs.Name
	}

	if property, has := inputs["ports"]; has && property.NumberValue() != 0 {
		vs.Ports = int(property.NumberValue())
	} else {
		vs.Ports = 128
	}
	if property, has := inputs["mtu"]; has && property.NumberValue() != 0 {
		vs.Mtu = int(property.NumberValue())
	} else {
		vs.Mtu = 1500
	}
	if property, has := inputs["linkDiscoveryMode"]; has {
		vs.LinkDiscoveryMode = property.StringValue()
	} else {
		vs.LinkDiscoveryMode = "listen"
	}
	if property, has := inputs["promiscuousMode"]; has {
		vs.PromiscuousMode = property.BoolValue()
	} else {
		vs.PromiscuousMode = false
	}
	if property, has := inputs["macChanges"]; has {
		vs.MacChanges = property.BoolValue()
	} else {
		vs.MacChanges = false
	}
	if property, has := inputs["forgedTransmits"]; has {
		vs.ForgedTransmits = property.BoolValue()
	} else {
		vs.ForgedTransmits = false
	}
	if property, has := inputs["upLinks"]; has {
		if upLinks := property.ArrayValue(); len(upLinks) > 0 {
			vs.Uplinks = make([]Uplink, len(upLinks))
			for i, upLink := range upLinks {
				vs.Uplinks[i] = Uplink{
					Name: upLink.ObjectValue()["name"].StringValue(),
				}
			}
		}
	} else {
		vs.Uplinks = make([]Uplink, 0)
	}

	return vs, nil
}

func (esxi *Host) readVirtualSwitch(name string) (string, resource.PropertyMap, error) {
	vs, err := esxi.getVirtualSwitch(name)
	if err != nil {
		return "", nil, err
	}

	result := vs.toMap()
	return vs.Id, resource.NewPropertyMapFromMap(result), nil
}

func (esxi *Host) updateVirtualSwitch(vs VirtualSwitch) error {
	var foundUplinks []string
	var command, stdout string
	var err error

	//  Set mtu and cdp
	command = fmt.Sprintf("esxcli network vswitch standard set -m %d -c \"%s\" -v \"%s\"",
		vs.Mtu, vs.LinkDiscoveryMode, vs.Name)

	stdout, err = esxi.Execute(command, "set vswitch mtu, link_discovery_mode")
	if err != nil {
		return fmt.Errorf("failed to set vswitch mtu: %s err: %s", stdout, err)
	}

	//  Set security
	command = fmt.Sprintf("esxcli network vswitch standard policy security set -f %t -m %t -p %t -v \"%s\"",
		vs.PromiscuousMode, vs.MacChanges, vs.ForgedTransmits, vs.Name)

	stdout, err = esxi.Execute(command, "set vswitch security")
	if err != nil {
		return fmt.Errorf("failed to set vswitch security: %s err: %s", stdout, err)
	}

	//  Update uplinks
	command = fmt.Sprintf("esxcli network vswitch standard list -v \"%s\"", vs.Name)
	stdout, err = esxi.Execute(command, "vswitch list")

	if err != nil {
		return fmt.Errorf("failed to list vswitch: %s err: %s", stdout, err)
	}

	re := regexp.MustCompile(`Uplinks: (.*)`)
	foundUplinksRaw := strings.Fields(re.FindStringSubmatch(stdout)[1])
	for _, s := range foundUplinksRaw {
		foundUplinks = append(foundUplinks, strings.Replace(s, ",", "", -1))
	}

	//  Add uplink if needed
	for i := range vs.Uplinks {
		if !Contains(foundUplinks, vs.Uplinks[i].Name) {
			command = fmt.Sprintf("esxcli network vswitch standard uplink add -u \"%s\" -v \"%s\"",
				vs.Uplinks[i].Name, vs.Name)

			stdout, err = esxi.Execute(command, "vswitch add uplink")
			if strings.Contains(stdout, "Not a valid pnic") {
				return fmt.Errorf("uplink not found: %s", vs.Uplinks[i].Name)
			}
			if err != nil {
				return fmt.Errorf("failed to add vswitch uplink: %s err: %s", stdout, err)
			}
		}
	}

	//  Remove uplink if needed
	selector := func(upLink Uplink) string { return upLink.Name }
	for _, item := range foundUplinks {
		if !ContainsValue(vs.Uplinks, selector, item) {
			log.Printf("[vswitchUpdate] delete uplink (%s)\n", item)
			command = fmt.Sprintf("esxcli network vswitch standard uplink remove -u \"%s\" -v \"%s\"",
				item, vs.Name)

			stdout, err = esxi.Execute(command, "vswitch remove uplink")
			if err != nil {
				return fmt.Errorf("failed to remove vswitch uplink: %s err: %s", stdout, err)
			}
		}
	}

	return nil
}

func (esxi *Host) getVirtualSwitch(name string) (VirtualSwitch, error) {
	vs := VirtualSwitch{
		Id:   name,
		Name: name,
	}
	var command, stdout string
	var err error

	command = fmt.Sprintf("esxcli network vswitch standard list -v \"%s\"", name)
	stdout, _ = esxi.Execute(command, "vswitch list")

	if stdout == "" {
		return VirtualSwitch{}, fmt.Errorf(stdout)
	}

	re, _ := regexp.Compile(`Configured Ports: ([0-9]*)`)
	if len(re.FindStringSubmatch(stdout)) > 0 {
		vs.Ports, _ = strconv.Atoi(re.FindStringSubmatch(stdout)[1])
	} else {
		vs.Ports = 128
	}

	re, _ = regexp.Compile(`MTU: ([0-9]*)`)
	if len(re.FindStringSubmatch(stdout)) > 0 {
		vs.Mtu, _ = strconv.Atoi(re.FindStringSubmatch(stdout)[1])
	} else {
		vs.Mtu = 1500
	}

	re, _ = regexp.Compile(`CDP Status: ([a-z]*)`)
	if len(re.FindStringSubmatch(stdout)) > 0 {
		vs.LinkDiscoveryMode = re.FindStringSubmatch(stdout)[1]
	} else {
		vs.LinkDiscoveryMode = "listen"
	}

	re, _ = regexp.Compile(`Uplinks: (.*)`)
	if len(re.FindStringSubmatch(stdout)) > 0 {
		foundUplinks := strings.Fields(re.FindStringSubmatch(stdout)[1])
		for _, s := range foundUplinks {
			vs.Uplinks = append(vs.Uplinks, Uplink{Name: strings.Replace(s, ",", "", -1)})
		}
	} else {
		vs.Uplinks = vs.Uplinks[:0]
	}

	command = fmt.Sprintf("esxcli network vswitch standard policy security get -v \"%s\"", name)
	stdout, _ = esxi.Execute(command, "vswitch policy security get")

	if stdout == "" {
		log.Printf("[vswitchRead] Failed to run %s: %s\n", "vswitch policy security get", err)
		return VirtualSwitch{}, fmt.Errorf(stdout)
	}

	re, _ = regexp.Compile(`Allow Promiscuous: (.*)`)
	if len(re.FindStringSubmatch(stdout)) > 0 {
		vs.PromiscuousMode, _ = strconv.ParseBool(re.FindStringSubmatch(stdout)[1])
	} else {
		vs.PromiscuousMode = false
	}

	re, _ = regexp.Compile(`Allow MAC Address Change: (.*)`)
	if len(re.FindStringSubmatch(stdout)) > 0 {
		vs.MacChanges, _ = strconv.ParseBool(re.FindStringSubmatch(stdout)[1])
	} else {
		vs.MacChanges = false
	}

	re, _ = regexp.Compile(`Allow Forged Transmits: (.*)`)
	if len(re.FindStringSubmatch(stdout)) > 0 {
		vs.ForgedTransmits, _ = strconv.ParseBool(re.FindStringSubmatch(stdout)[1])
	} else {
		vs.ForgedTransmits = false
	}

	return vs, nil
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
