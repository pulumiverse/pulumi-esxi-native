package esxi

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

func VirtualSwitchCreateParser(inputs resource.PropertyMap) VirtualSwitch {
	return VirtualSwitch{
		Name: inputs["name"].StringValue(),
	}
}

func VirtualSwitchCreate(vs VirtualSwitch, esxi *Host) (string, resource.PropertyMap, error) {
	//var uplinks []string
	//var remoteCmd string
	//var somthingWentWrong string
	//var err error
	//var i int

	//property := inputs["name"]
	//if property.HasValue() {
	//}
	//property.StringValue()
	//name := property.StringValue()
	//ports := inputs["ports"].NumberValue()
	//mtu := inputs["mtu"].NumberValue()
	//linkDiscoveryMode := inputs["linkDiscoveryMode"].StringValue()
	//promiscuousMode := inputs["promiscuousMode"].BoolValue()
	//macChanges := inputs["macChanges"].BoolValue()
	//forgedTransmits := inputs["forgedTransmits"].BoolValue()
	//somthingWentWrong = ""

	// Validate variables
	//if ports == 0 {
	//	ports = 128
	//}
	//
	//if mtu == 0 {
	//	mtu = 1500
	//}
	//
	//if len(linkDiscoveryMode) == 0 {
	//	linkDiscoveryMode = "listen"
	//}
	//
	//if linkDiscoveryMode != "down" && linkDiscoveryMode != "listen" &&
	//	linkDiscoveryMode != "advertise" && linkDiscoveryMode != "both" {
	//	return nil, fmt.Errorf("linkDiscoveryMode must be one of: down, listen, advertise or both")
	//}

	//uplinkCount, ok := inputs["uplink"].Mappable()
	//if !ok {
	//	uplinkCount = 0
	//	uplinks[0] = ""
	//}
	//if uplinkCount > 32 {
	//	uplinkCount = 32
	//}
	//for i = 0; i < uplinkCount; i++ {
	//	prefix := fmt.Sprintf("uplink.%d.", i)
	//
	//	if attr, ok := d.Get(prefix + "name").(string); ok && attr != "" {
	//		uplinks = append(uplinks, d.Get(prefix+"name").(string))
	//	}
	//}
	//
	////  Create vswitch
	//remoteCmd = fmt.Sprintf("esxcli network vswitch standard add -P %d -v \"%s\"",
	//	ports, name)

	//stdout, err := esxi.Execute(remoteCmd, "create virtual switch")
	//if strings.Contains(stdout, "this name already exists") {
	//	d.SetId("")
	//	return fmt.Errorf("Failed to add vswitch: %s, it already exists\n", name)
	//}
	//if err != nil {
	//	d.SetId("")
	//	return fmt.Errorf("Failed to add vswitch: %s\n%s\n", stdout, err)
	//}
	//
	////  Set id
	//d.SetId(name)
	//
	//err = vswitchUpdate(c, name, ports, mtu, uplinks, linkDiscoveryMode, promiscuousMode, macChanges, forgedTransmits)
	//if err != nil {
	//	somthingWentWrong = fmt.Sprintf("Failed to update vswitch: %s\n", err)
	//}
	//
	//// Refresh
	//ports, mtu, uplinks, linkDiscoveryMode, promiscuousMode, macChanges, forgedTransmits, err = vswitchRead(c, name)
	//if err != nil {
	//	d.SetId("")
	//	return nil
	//}
	//
	//// Change uplinks (list) to map
	//log.Printf("[resourceVSWITCHCreate] uplinks: %s\n", uplinks)
	//uplink := make([]map[string]interface{}, 0, 1)
	//
	//if len(uplinks) == 0 {
	//	uplink = nil
	//} else {
	//	for i, _ := range uplinks {
	//		out := make(map[string]interface{})
	//		out["name"] = uplinks[i]
	//		uplink = append(uplink, out)
	//	}
	//}
	//d.Set("uplink", uplink)
	//
	//d.Set("ports", ports)
	//d.Set("mtu", mtu)
	//d.Set("link_discovery_mode", linkDiscoveryMode)
	//d.Set("promiscuous_mode", promiscuousMode)
	//d.Set("mac_changes", macChanges)
	//d.Set("forged_transmits", forgedTransmits)
	//
	//if somthingWentWrong != "" {
	//	return fmt.Errorf(somthingWentWrong)
	//}

	result := vs.toMap()
	return "1", resource.NewPropertyMapFromMap(result), nil
}
