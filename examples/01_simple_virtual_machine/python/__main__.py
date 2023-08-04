import pulumi
from typing import Sequence
from pulumiverse_esxi_native import VirtualMachine, NetworkInterfaceArgs

vm = VirtualMachine("vm-test",
                    disk_store="nvme-ssd-datastore",
                    network_interfaces=Sequence[NetworkInterfaceArgs(
                        virtual_network="default"
                    )])

pulumi.export("id", vm.id)
pulumi.export("name", vm.name)
pulumi.export("os", vm.os)
