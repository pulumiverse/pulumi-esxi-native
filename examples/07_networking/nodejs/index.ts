import {PortGroup, VirtualMachine, VirtualSwitch} from "@pulumiverse/pulumi-esxi-native";

// ESXI vSwitch resource
// Example vSwitch with defaults.
// Uncomment the uplink block to connect this vSwitch to your nic.
const vSwitch = new VirtualSwitch("my-v-switch", {
    name: "My vSwitch",
    // uplinks: [
    //     {
    //         name: "vmnic0"
    //     }
    // ]
});

// ESXI Port Group resource
// Example port group with default, connecting to the above vSwitch.
const portGroup = new PortGroup("my-port-group", {
    name: "My Port Group",
    vSwitch: vSwitch.name,
});

new VirtualMachine("vm-test", {
    diskStore: "nvme-ssd-datastore",
    networkInterfaces: [
        {
            virtualNetwork: portGroup.name
        }
    ]
});