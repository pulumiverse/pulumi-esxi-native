import { VirtualMachine } from "@pulumiverse/esxi-native";
import { remote, types } from "@pulumi/command";
import * as random from "@pulumi/random";
import * as tls from "@pulumi/tls";
import {interpolate} from "@pulumi/pulumi";

export = async () => {
    // https://github.com/tsugliani/packer-alpine
    const ovfSource = "https://cloud.tsugliani.fr/ova/alpine-3.15.6.ova";

    const key = new tls.PrivateKey("ssh-key", {
        algorithm: "ECDSA",
        ecdsaCurve: "P384",
    });

    const password = new random.RandomPassword("password", {
        length: 16,
        special: false,
    });

    const vm = new VirtualMachine("vm-test", {
        diskStore: "nvme-ssd-datastore",
        os: "otherlinux",
        bootDiskSize: 8,
        numVCpus: 1,
        memSize: 512,
        shutdownTimeout: 5,
        //startupTimeout: 60,
        power: "on",
        networkInterfaces: [
            {
                virtualNetwork: "default"
            }
        ],
        // Specify ovf_properties specific to the source ovf/ova.
        // Use ovftool <filename>.ova to get details of which ovf_properties are available.
        ovfProperties: [
            {
                // using with templates generated name
                key: "guestinfo.hostname",
                value: "{{.Name}}"
            },
            {
                key: "guestinfo.gateway",
                value: "192.168.20.1"
            },
            // {
            //     key: "guestinfo.netprefix",
            //     value: "20"
            // },
            // {
            //     key: "guestinfo.ipaddress",
            //     value: "192.168.20.200"
            // },
            {
                key: "guestinfo.dns",
                value: "1.1.1.1"
            },
            {
                key: "guestinfo.domain",
                value: "alpine.local"
            },
            {
                key: "guestinfo.password",
                value: password.result
            },
            {
                key: "guestinfo.sshkey",
                value: key.publicKeyOpenssh.apply(v => v.trim())
            },
        ],
        // Specify an ovf file to use as a source.
        ovfSource: ovfSource,
        virtualHWVer: 14
    });

    const passwordConnection: types.input.remote.ConnectionArgs = {
        host: vm.ipAddress,
        user: "root",
        password: password.result,
    };

    new remote.Command("cloud-init-ssh-setup", {
        connection: passwordConnection,
        create: interpolate`mkdir -vp ~/.ssh; echo "${key.publicKeyOpenssh}" > ~/.ssh/authorized_keys`,
        delete: `rm ~/.ssh/authorized_keys`
    }, { deleteBeforeReplace: true });

    const connection: types.input.remote.ConnectionArgs = {
        host: vm.ipAddress,
        user: "root",
        privateKey: key.privateKeyOpenssh,
    };

    new remote.CopyFile("cloud-init", {
        connection,
        localPath: "./init.sh",
        remotePath: "init.sh",
    })

    new remote.Command("cloud-init-chmod", {
        connection,
        create: `chmod +x init.sh`,
        delete: `chmod -x init.sh`,
    }, { deleteBeforeReplace: true });

    new remote.Command("cloud-init-exec", {
        connection,
        create: `sh init.sh`,
    }, { deleteBeforeReplace: true });

    password.result.apply(console.log)

    return {
        "id": vm.id,
        "name": vm.name,
        "os": vm.os,
        "ip": vm.ipAddress,
    };
}
