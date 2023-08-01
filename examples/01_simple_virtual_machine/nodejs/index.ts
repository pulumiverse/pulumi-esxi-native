import * as esxi from "@pulumiverse/pulumi-esxi-native";

let vm: esxi.VirtualMachine;
vm = new esxi.VirtualMachine("vm-test", {
    diskStore: "nvme-ssd-datastore",
    networkInterfaces: [
        {
            virtualNetwork: "default"
        }
    ]
});

export const id = vm.id;
export const name = vm.name;
export const os = vm.os;
