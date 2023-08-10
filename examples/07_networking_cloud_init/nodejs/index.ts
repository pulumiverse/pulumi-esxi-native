import {PortGroup, VirtualSwitch, VirtualMachine} from "@pulumiverse/esxi-native";

export = async () => {
    // https://github.com/tsugliani/packer-alpine
    const ovfSource = "https://cloud.tsugliani.fr/ova/alpine-3.15.6.ova";
    const ipAddress = "10.10.10.10/24";
    const gateway = "10.10.10.1";
    const nameserver = "8.8.8.8";


    // ESXI vSwitch resource
    // Example vSwitch with defaults.
    // Uncomment the uplink block to connect this vSwitch to your nic.
    const vSwitch = new VirtualSwitch("my-v-switch", {
        // uplinks: [
        //     {
        //         name: "vmnic0"
        //     }
        // ]
    });

    // ESXI Port Group resource
    // Example port group with default, connecting to the above vSwitch.
    const portGroup = new PortGroup("my-port-group", {
        vSwitch: vSwitch.name,
        // vSwitch: "vSwitch0", // using default vlan
    });

    const vm = new VirtualMachine("vm-test", {
        diskStore: "nvme-ssd-datastore",
        os: "otherlinux",
        bootDiskSize: 8,
        numVCpus: 1,
        memSize: 512,
        shutdownTimeout: 5,
        power: "on",
        networkInterfaces: [
            {
                virtualNetwork: portGroup.name
            }
        ],
        // Specify ovf_properties specific to the source ovf/ova.
        // Use ovftool <filename>.ova to get details of which ovf_properties are available.
        ovfProperties: [
            {
                key: "guestinfo.hostname",
                value: "pulumi"
            },
            {
                key: "guestinfo.gateway",
                value: gateway
            },
            // {
            //     key: "guestinfo.netprefix",
            //     value: "20"
            // },
            {
                key: "guestinfo.ipaddress",
                value: ipAddress
            },
            {
                key: "guestinfo.dns",
                value: nameserver
            },
            {
                key: "guestinfo.domain",
                value: "alpine.local"
            },
            {
                key: "guestinfo.password",
                value: "secret"
            },
        ],
        // Specify an ovf file to use as a source.
        ovfSource: ovfSource,
        virtualHWVer: 14
    });

    return {
        "id": vm.id,
        "ip": vm.ipAddress,
        "name": vm.name,
        "os": vm.os,
        "vlan": portGroup.vlan
    };
}
