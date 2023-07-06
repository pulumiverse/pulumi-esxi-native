import * as esxi from "@edmondshtogu/pulumi-esxi-native";

const vmw = esxi.getVirtualMachineOutput({name: "vm-worker-1-on-prem-metal-b887abc"});

//const vSwitch = new esxi.VirtualSwitch("test", {})

export const output = vmw.id;
//export const vSwitchNameOutput = vSwitch.name;
