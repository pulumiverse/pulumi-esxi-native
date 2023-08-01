import {VirtualMachine, DiskType, VirtualDisk} from "@pulumiverse/pulumi-esxi-native";
import {VMVirtualDiskArgs} from "@pulumiverse/pulumi-esxi-native/types/input";

export interface VirtualMachineConfig {
    Index: number
    Type: string;
    Datastore: string;
    Network: string;
    OvaRemoteUrl: string;
    StorageDisk?: {
        Size: number;
        Datastore: string;
    };
    Memory: number;
    Cpu: number;
    Disk: number;
}

export class VirtualMachineFactory {
    private readonly _talosCPConfig: string;
    private readonly _talosWorkerConfig: string;

    constructor(cpConfig: string, workerConfig: string) {
        this._talosCPConfig = cpConfig;
        this._talosWorkerConfig = workerConfig;
    }

    make(config: VirtualMachineConfig): VirtualMachine {
        const machinePowerState = "on";
        const machineOs = "other3xlinux-64";
        const name = `${config.Type}-${config.Index}`;
        const talosConfig = config.Type == "control-plane" ?
            this._talosCPConfig : this._talosWorkerConfig;

        const disks: VMVirtualDiskArgs[] = [];

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
            info: [
                {
                    key: "guestinfo.talos.config",
                    value: talosConfig
                },
            ],
            memSize: config.Memory,
            numVCpus: config.Cpu,
            bootDiskType: DiskType.Thin,
            bootDiskSize: config.Disk,
            power: machinePowerState,
            // ovfNetworkMap: [{
            //   "Network 1": config.Network,
            // }],
            networkInterfaces: [
                {
                    virtualNetwork: config.Network,
                    nicType: "vmxnet3"
                }
            ],
            ovfSource: config.OvaRemoteUrl,
            os: machineOs,
            resourcePoolName: "/",
            startupTimeout: 35,
            shutdownTimeout: 30,
            virtualDisks: disks
        })
    }
}
