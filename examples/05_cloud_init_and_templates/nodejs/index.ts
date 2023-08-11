import { VirtualMachine } from "@pulumiverse/esxi-native";
import { remote, types } from "@pulumi/command";
import * as random from "@pulumi/random";
import * as tls from "@pulumi/tls";
import {interpolate} from "@pulumi/pulumi";

export = async () => {
    // See this repo for more details: https://github.com/tsugliani/packer-alpine
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
                value: key.publicKeyOpenssh
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

    const sshKeyConnection: types.input.remote.ConnectionArgs = {
        host: vm.ipAddress,
        user: "root",
        privateKey: key.privateKeyOpenssh,
    };

    // We poll the server until it responds.
    //
    // Because other commands depend on this command, other commands are guaranteed
    // to hit an already booted server.
    const passPoll = new remote.Command("poll-password", {
        connection: { ...passwordConnection, dialErrorLimit: -1 },
        create: "echo 'Connection established using password connection'",
    }, { customTimeouts: { create: "10m" } })

    // Seems that the ova file is not setting up the SSH key, we add the ssh public key to the server connecting with password.
    const sshSetup = new remote.Command("cloud-init-ssh-setup", {
        connection: passwordConnection,
        create: interpolate`mkdir -vp /root/.ssh; echo "${key.publicKeyOpenssh}" > /root/.ssh/authorized_keys`,
        delete: `rm /root/.ssh/authorized_keys`,
    }, { deleteBeforeReplace: true, dependsOn: passPoll });

    const keyPoll = new remote.Command("poll-ssh-key", {
        connection: { ...sshKeyConnection, dialErrorLimit: -1 },
        create: "echo 'Connection established using ssh key'",
    }, { customTimeouts: { create: "10m" }, dependsOn: sshSetup })

    const copyFile = new remote.CopyFile("cloud-init-copy-script", {
        connection: sshKeyConnection,
        localPath: "./init.sh",
        remotePath: "/root/init.sh",
    }, { dependsOn: keyPoll })

    const chmod = new remote.Command("cloud-init-chmod-script", {
        connection: sshKeyConnection,
        create: `chmod +x /root/init.sh`,
    }, { deleteBeforeReplace: true, dependsOn: copyFile });

    const init = new remote.Command("cloud-init-exec-script", {
        connection: sshKeyConnection,
        create: `sh /root/init.sh`,
    }, { deleteBeforeReplace: true, dependsOn: chmod });

    return {
        "id": vm.id,
        "name": vm.name,
        "os": vm.os,
        "ip": vm.ipAddress,
        "init": init.stdout
    };
}
