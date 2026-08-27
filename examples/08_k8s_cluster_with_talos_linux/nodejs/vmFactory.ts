import {VirtualMachine, DiskType, VirtualDisk} from "@pulumiverse/esxi-native";
import {VMVirtualDiskArgs} from "@pulumiverse/esxi-native/types/input";
import {Output} from "@pulumi/pulumi";

export interface VirtualMachineConfig {
    Type: string;
    Datastore: string;
    Network: string;
    OvfSource: string;
    StorageDisk?: {
        Size: number;
        Datastore: string;
    };
    Memory: number;
    Cpu: number;
    Disk: number;
}

export class VirtualMachineFactory {
    private readonly _talosCPConfig: Output<string>;
    private readonly _talosWorkerConfig: Output<string>;

    constructor(cpConfig: Output<string>, workerConfig: Output<string>) {
        this._talosCPConfig = cpConfig;
        this._talosWorkerConfig = workerConfig;
    }

    make(config: VirtualMachineConfig): VirtualMachine {
        const machinePowerState = "on";
        const machineOs = "other3xlinux-64";
        const name = `vm-${config.Type}`;
        const talosConfig = (config.Type == "control-plane" ?
            this._talosCPConfig : this._talosWorkerConfig).apply(config => `${config}{{ parsedTemplateOutput . | base64encode -}}`);

        let disks: VMVirtualDiskArgs[] = [];

        if (config.StorageDisk) {
            const storageDisk = new VirtualDisk(`${name}-k8s-vdisk`, {
                diskType: DiskType.ZeroedThick,
                diskStore: config.StorageDisk.Datastore,
                directory: `k8s-storage`,
                size: config.StorageDisk.Size
            })

            disks.push({
                virtualDiskId: storageDisk.id,
                slot: "0:1"
            });
        }

        return new VirtualMachine(name, {
            diskStore: config.Datastore,
            memSize: config.Memory,
            numVCpus: config.Cpu,
            bootDiskType: DiskType.Thin,
            bootDiskSize: config.Disk,
            power: machinePowerState,
            networkInterfaces: [
                {
                    virtualNetwork: config.Network,
                    nicType: "vmxnet3"
                }
            ],
            ovfSource: config.OvfSource,
            info: [
                {
                    key: "talos.config",
                    value: talosConfig
                },
            ],
            ovfProperties: [
                {
                    key: "talos.config",
                    value: talosConfig
                },
            ],
            os: machineOs,
            resourcePoolName: "/",
            // in this case we will set the startup timeout to 300 as the vm-tools need to be configured, up and running, so the IP can be fetched
            startupTimeout: 5,
            shutdownTimeout: 5,
            virtualDisks: disks,
            virtualHWVer: 13
        })
    }
}
