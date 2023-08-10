import {setupTalos, getTalosCpConfig, getTalosWorkerConfig} from './talosSetup';
import {VirtualMachineFactory, VirtualMachineConfig} from './vmFactory';


const ovfUrl = "https://github.com/siderolabs/talos/releases/download/v1.4.7/vmware-amd64.ova";
const vmConfigs: VirtualMachineConfig[] = [
    {
        Index: 1,
        Datastore: "sata-evo-ssd-datastore",
        Network: "default",
        OvaRemoteUrl: ovfUrl,
        Type: "control-plane",
        Disk: 40,
        Memory: 4096,
        Cpu: 4
    },
    {
        Index: 2,
        Datastore: "sata-evo-ssd-datastore",
        Network: "default",
        OvaRemoteUrl: ovfUrl,
        Type: "worker",
        Disk: 50,
        Memory: 8192,
        Cpu: 4,
        StorageDisk: {
            Datastore: "nvme-ssd-datastore",
            Size: 200
        }
    }
]

// Talos Linux Setup
//   See: https://www.talos.dev/v1.4/talos-guides/install/virtualized-platforms/vmware/
setupTalos().then(_ => {
    const factory = new VirtualMachineFactory(getTalosCpConfig(), getTalosWorkerConfig())
    vmConfigs.forEach(config => factory.make(config))
});
