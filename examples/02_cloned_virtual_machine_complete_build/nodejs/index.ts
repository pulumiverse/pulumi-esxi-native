import {ResourcePool, VirtualMachine} from "@edmondshtogu/pulumi-esxi-native";
import {concat} from "@pulumi/pulumi";

// This Guest VM is a clone of an existing Guest VM named "centos7" (must exist and
// be powered off), located in the "Templates" resource pool.  vmtest02 will be powered
// on by default by terraform.  The virtual network "VM Network", must already exist on
// your esxi host!

let pool : ResourcePool;
pool = new ResourcePool("templates", {
    name: "Templates"
})

let template: VirtualMachine;
template = new VirtualMachine("vm-template", {
    name: "centos7",
    diskStore: "nvme-ssd-datastore",
    networkInterfaces: [
        {
            virtualNetwork: "default"
        }
    ],
    resourcePoolName: pool.name,
    os: "centos-64",
    power: "off"
});

let clone: VirtualMachine;
clone = new VirtualMachine("vm-clone",{
    diskStore: "nvme-ssd-datastore",
    os: "centos-64",
    bootDiskType: "thin",
    bootDiskSize: 35,
    memSize: 2048,
    numVCpus: 2,
    resourcePoolName: "/",
    power: "on",
    //  cloneFromVirtualMachine uses ovftool to clone an existing VM on your esxi host.
    //  This example will clone a VM named "centos7", we created above, located in the "Templates" resource pool.
    //  ovfSource uses ovftool to produce a clone from an ovf or vmx image. (typically produced using the ovf_tool).
    //    Basically clone_from_vm clones from sources on the esxi host and ovf_source clones from sources on your local hard disk or a URL.
    //    These two options are mutually exclusive.
    cloneFromVirtualMachine:  concat(pool.name, "/", template.name),
    networkInterfaces: [
        {
            virtualNetwork: "default",
            macAddress: "00:50:56:a1:b1:c2",
            nicType: "e1000"
        },
        {
            virtualNetwork: "Management Network"
        }
    ],
    startupTimeout: 45,
    shutdownTimeout: 30
})

export const id = clone.id;
export const name = clone.name;
export const os = clone.os;