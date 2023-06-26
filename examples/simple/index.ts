import * as esxi from "@edmondshtogu/pulumi-esxi-native";

const vm = esxi.getVirtualMachineOutput({name: "vcsa"});

const vswitch = new esxi.VirtualSwitch("test", {name: "test"})

export const output = vm.name;
