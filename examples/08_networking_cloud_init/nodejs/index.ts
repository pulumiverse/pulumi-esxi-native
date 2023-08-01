import { gzip } from 'zlib';
import {Output} from "@pulumi/pulumi";
import {PortGroup, VirtualSwitch, VirtualMachine} from "@pulumiverse/pulumi-esxi-native";

function base64gzip(input: string): Promise<string> {
    return new Promise((resolve, reject) => {
        gzip(input, (error, compressedBuffer) => {
            if (error) {
                reject(error);
            } else {
                const base64Compressed = compressedBuffer.toString('base64');
                resolve(base64Compressed);
            }
        });
    });
}

let output: Output<string>;

const ipAddress = "10.10.10.10/24";
const gateway = "10.10.10.1";
const nameserver = "8.8.8.8";

const metadata = `
network:
    version: 2
    ethernets:
        ens192:
            dhcp4: false
            addresses:
                - ${ipAddress}
            gateway4: ${gateway}
            nameservers:
                addresses:
                    - ${nameserver}
`;


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
    // vSwitch: "vSwitch0", // using default vlan
});

base64gzip(metadata)
    .then((base64Compressed) => {
        // This VM is "bare-metal". It will be powered on by the
        // Pulumi, but it will not boot to any OS. It will however attempt
        // to network boot on the port group configured above.

        const vm = new VirtualMachine("vm-test", {
            diskStore: "nvme-ssd-datastore",
            networkInterfaces: [
                {
                    virtualNetwork: portGroup.name
                }
            ],
            info: [
                {
                    key: "metadata.encoding",
                    value: "gzip+base64"
                },
                {
                    key: "metadata",
                    value: base64Compressed
                },
            ],
            power: "on"
        });

        output = vm.ipAddress;
    })
    .catch((error) => {
        console.error('Error compressing:', error);
    });


