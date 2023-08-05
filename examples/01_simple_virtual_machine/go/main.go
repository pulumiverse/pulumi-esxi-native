package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-esxi-native/sdk/go/esxi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		vm, err := esxi.NewVirtualMachine(ctx, "vm-test", &esxi.VirtualMachineArgs{
			DiskStore: pulumi.String("nvme-ssd-datastore"),
			NetworkInterfaces: esxi.NetworkInterfaceArray{
				esxi.NetworkInterfaceArgs{
					VirtualNetwork: pulumi.String("default"),
				},
			},
		})
		if err != nil {
			return err
		}

		ctx.Export("id", vm.ID())
		ctx.Export("name", vm.Name)
		ctx.Export("os", vm.Os)
		return nil
	})
}
