import * as esxi from "@pulumiverse/pulumi-esxi-native";
import {DiskType, ResourcePool, VirtualDisk} from "@pulumiverse/pulumi-esxi-native";
import {concat} from "@pulumi/pulumi";

// Resource Pools
let pool1, pool2 : ResourcePool;
pool1 = new ResourcePool("pool1", {
    name: "Pulumi",
    cpuMin: 100,
    cpuMinExpandable: "true",
    cpuMax: 8000,
    cpuShares: "normal",
    memMin: 200,
    memMinExpandable: "false",
    memMax: 8192,
    memShares: "normal"
})

pool2 = new ResourcePool("pool2", {
    name: concat(pool1.name, "/pool2"),
})

// Virtual Disks
let vdisk1, vdisk2 : VirtualDisk
vdisk1 = new VirtualDisk("vdisk1", {
    diskType: DiskType.ZeroedThick,
    diskStore: "nvme-ssd-datastore",
    directory: "Pulumi",
    name: "vdisk_1.vmdk",
    size: 10
})

vdisk2 = new VirtualDisk("vdisk2", {
    diskType: DiskType.EagerZeroedThick,
    diskStore: "nvme-ssd-datastore",
    directory: "Pulumi",
    size: 15
})

// ESXI Guest resource
// This Guest VM is a clone of an existing Guest VM named "centos7" (must exist and
// be powered off), located in the "Templates" resource pool.  vmtest03 will be powered
// on by default by terraform.  The virtual network "VM Network", must already exist on
// your esxi host!

let vm: esxi.VirtualMachine;
vm = new esxi.VirtualMachine("vm-test", {
    diskStore: "nvme-ssd-datastore",
    os: "centos-64",
    bootDiskType: DiskType.Thin,
    bootDiskSize: 35,
    memSize: 2048,
    numVCpus: 2,
    resourcePoolName: pool2.name,
    power: "on",
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
    shutdownTimeout: 30,
    virtualDisks: [
        {
            virtualDiskId: vdisk1.id,
            slot: "0:1"
        },
        {
            virtualDiskId: vdisk2.id,
            slot: "0:2"
        }
    ]
})

