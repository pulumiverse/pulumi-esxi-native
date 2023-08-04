import * as esxi from "@pulumiverse/esxi-native";

const vm = new esxi.VirtualMachine("vm-test", {
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
