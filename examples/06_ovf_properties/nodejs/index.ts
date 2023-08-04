import {base64gzip, userdata, ovfSource} from './utils';
import {VirtualMachine} from "@pulumiverse/esxi-native";

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
            // Specify an ovf file to use as a source.
            ovfSource: ovfSource,
            // Specify ovf_properties specific to the source ovf/ova.
            // Use ovftool <filename>.ova to get details of which ovf_properties are available.
            ovfProperties: [
                {
                    key: "hostname",
                    value: "pulumi"
                },
                {
                    key: "user-data",
                    value: base64Compressed
                }
            ],
            // Default ovfPropertiesTimer is 90 seconds. ovfProperties are injected on first boot.
            // This value should be high enough to allow the virtual machine to fully boot to
            // a linux prompt. The second boot is needed to configure the virtual machine as
            // specified. (cpus, memory, adding or expanding disks, etc...)
            ovfPropertiesTimer: 90
        });
    })
    .catch((error) => {
        console.error('Error compressing:', error);
    });
