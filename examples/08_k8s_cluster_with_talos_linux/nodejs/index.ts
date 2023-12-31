import {VirtualMachineFactory, VirtualMachineConfig} from './vmFactory';
import {VirtualMachine} from "@pulumiverse/esxi-native";
import {local} from "@pulumi/command";
import {interpolate} from "@pulumi/pulumi";

export = async () => {
    const outputs = {
        nodes: []
    }

    // See this repo for more details: https://github.com/tsugliani/packer-alpine
    const ovfSource = "https://github.com/siderolabs/talos/releases/download/v1.4.8/vmware-amd64.ova";

    const controlPlaneIp = '192.168.20.20'

    const vmConfigs: VirtualMachineConfig[] = [
        {
            Datastore: "sata-evo-ssd-datastore",
            Network: "default",
            OvfSource: ovfSource,
            Type: "control-plane",
            Disk: 40,
            Memory: 4096,
            Cpu: 4
        },
        {
            Datastore: "sata-evo-ssd-datastore",
            Network: "default",
            OvfSource: ovfSource,
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
    const cpConfigPatch = JSON.stringify([
        {
            op: "add",
            path: "/machine/network",
            value: {
                hostname: "{{.Name}}", // adding name as template
                interfaces: [
                    {
                        interface: "eth0", dhcp: true, vip: {ip: controlPlaneIp}
                    }
                ]
            }
        },
        {
            op: "replace",
            path: "/cluster/inlineManifests",
            value: [
                {
                    name: "talos-vmtoolsd-config",
                    contents: "apiVersion: v1\ndata:\n  talosconfig: <TALOS_CONFIG>\nkind: Secret\nmetadata:\n  name: talos-vmtoolsd-config\n  namespace: kube-system\n"
                }
            ]
        },
        {
            op: "replace",
            path: "/cluster/extraManifests",
            value: [ "https://raw.githubusercontent.com/mologie/talos-vmtoolsd/master/deploy/unstable.yaml" ]
        }
    ])

    const workerConfigPatch = JSON.stringify([
        {
            op: "add",
            path: "/machine/network/hostname",
            value: "{{.Name}}" // adding name as template
        }
    ]).toString()

    const talosCtl = local.runOutput({
        command: `curl -sL https://talos.dev/install | sh`,
        dir: process.cwd(),
        assetPaths: ["/usr/local/bin/talosctl"]
    }).assetPaths[0];

    const configs = new local.Command("generate-configs", {
        create: talosCtl.apply(cmd => `${cmd} gen config pulumi-esxi https://${controlPlaneIp}:6443 --config-patch-control-plane '${cpConfigPatch}' --config-patch-worker '${workerConfigPatch}' --force`),
        dir: process.cwd(),
    })

    const validateConfigs = new local.Command("validate-configs", {
        create: `talosctl validate --config controlplane.yaml --mode cloud && talosctl validate --config worker.yaml --mode cloud`,
        dir: process.cwd(),
    }, {dependsOn: configs})

    const talosConfig = new local.Command("talos-config", {
        create: `cat talosconfig`,
        dir: process.cwd(),
    }, {dependsOn: configs})

    Object.assign(outputs, outputs, {
        talosConfig: talosConfig.stdout,
    });

    const controlPlaneConfig = new local.Command("control-plane-config", {
        create: "config=$(cat controlplane.yaml) && echo ${config//<TALOS_CONFIG>/$(base64 talosconfig)}",
        dir: process.cwd(),
    }, {dependsOn: talosConfig})

    const workerConfig = new local.Command("worker-config", {
        create: `cat worker.yaml`,
        dir: process.cwd(),
    }, {dependsOn: validateConfigs})

    let vms: VirtualMachine[] = []
    let cp: VirtualMachine
    const factory = new VirtualMachineFactory(controlPlaneConfig.stdout, workerConfig.stdout)

    vmConfigs.forEach(config => {
            const vm = factory.make(config);
            if (config.Type ==  "control-plane" && cp == null)
                cp = vm
            vms.push(vm)
            outputs.nodes.push({
                    id: vm.id,
                    name: vm.name,
                    type: config.Type,
                })
        })

    const bootstrap = new local.Command("cluster-bootstrap", {
        create: `talosctl --talosconfig talosconfig bootstrap -e ${cp.ipAddress} -n ${cp.ipAddress}`,
        dir: process.cwd(),
    }, {dependsOn: vms})

    const k8sConfig = new local.Command("cluster-bootstrap", {
        create: `talosctl --talosconfig talosconfig bootstrap -e ${controlPlaneIp} -n ${controlPlaneIp}`,
        dir: process.cwd(),
    }, {dependsOn: bootstrap})

    Object.assign(outputs, outputs, {
        k8sConfig: k8sConfig.stdout,
    });

    return outputs;
}