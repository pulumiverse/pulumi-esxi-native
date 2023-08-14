import {VirtualMachine} from "@pulumiverse/esxi-native";

export = async () => {
    // https://github.com/tsugliani/packer-alpine
    const ovfSource = "https://cloud.tsugliani.fr/ova/alpine-3.15.6.ova";

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
                virtualNetwork: "default"
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
                value: "192.168.20.1"
            },
            {
                key: "guestinfo.dns",
                value: "1.1.1.1"
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
        "name": vm.name,
        "os": vm.os,
    };
}
