import {base64gzip, userdata, ovfSource} from './utils';
import {VirtualMachine} from "@pulumiverse/pulumi-esxi-native";

base64gzip(userdata)
    .then((base64Compressed) => {
        // This VM is "bare-metal". It will be powered on by the
        // Pulumi, but it will not boot to any OS. It will however attempt
        // to network boot on the port group configured above.
        new VirtualMachine("vm-test", {
            diskStore: "nvme-ssd-datastore",
            networkInterfaces: [
                {
                    virtualNetwork: "default"
                }
            ],
            info: [
                {
                    key: "userdata.encoding",
                    value: "gzip+base64"
                },
                {
                    key: "userdata",
                    value: base64Compressed
                },
            ],
            // Specify an ovf file to use as a source.
            ovfSource: ovfSource,
        });
    })
    .catch((error) => {
        console.error('Error compressing:', error);
    });
