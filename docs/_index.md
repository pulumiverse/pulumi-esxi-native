---
title: ESXi Native
meta_desc: Provides an overview of the ESXi Native Provider for Pulumi.
layout: package
---

The ESXi Native provider is used to provision VMs directly on an ESXi hypervisor without a need for vCenter or vSphere.

## Example

{{< chooser language "typescript,python,go,csharp,yaml" >}}


{{% choosable language typescript %}}
```typescript
import * as esxi from "@pulumiverse/esxi-native";


export = async () => {
    const vm = new esxi.VirtualMachine("vm-test", {
        diskStore: "nvme-ssd-datastore",
        networkInterfaces: [
            {
                virtualNetwork: "default"
            }
        ]
    });

    return {
        "id": vm.id,
        "name": vm.name,
        "os": vm.os,
    };
}
```
{{% /choosable %}}

{{% choosable language python %}}
```python
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
```
{{% /choosable %}}

{{% choosable language go %}}
```go
package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-esxi-native/sdk/go/esxi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		vm, err := esxi.VirtualMachine(ctx, "rotating", &time.RotatingArgs{
			RotationDays: pulumi.Int(30),
			Triggers: pulumi.StringMap{
				"trigger1": pulumi.String(timeTrigger),
			},
		})
		if err != nil {
			return err
		}
		offset, err := time.NewOffset(ctx, "offset", &time.OffsetArgs{
			OffsetDays: pulumi.Int(7),
		})
		if err != nil {
			return err
		}
		ctx.Export("rotating-output-unix", rotating.Unix)
		ctx.Export("rotating-output-rfc3339", rotating.Rfc3339)
		ctx.Export("offset-date", pulumi.All(offset.Day, offset.Month, offset.Year).ApplyT(func(_args []interface{}) (string, error) {
			day := _args[0].(int)
			month := _args[1].(int)
			year := _args[2].(int)
			return fmt.Sprintf("%v-%v-%v", day, month, year), nil
		}).(pulumi.StringOutput))
		return nil
	})
}
```
{{% /choosable %}}

{{% choosable language csharp %}}
```csharp
using System.Collections.Generic;
using System.Linq;
using Pulumi;
using Pulumi.EsxiNative;
using Pulumi.EsxiNative.Inputs;

return await Deployment.RunAsync(() =>
{
    var vm = new VirtualMachine("vm-test", new VirtualMachineArgs
    {
        DiskStore = "nvme-ssd-datastore",
        NetworkInterfaces = new NetworkInterfaceArgs[]
        {
            {
                VirtualNetwork = "default"
            }
        }
    });

    return new Dictionary<string, object?>
    {
        ["id"] = vm.Id,
        ["name"] = vm.Name,
        ["os"] = vm.Os,
    };
});
```
{{% /choosable %}}

{{< /chooser >}}