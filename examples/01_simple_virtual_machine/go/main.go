package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-esxi-native/sdk/go/esxi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		vm, err := esxi.VirtualMachine
		if err != nil {
			return err
		}

		ctx.Export("arn", logGroup.Arn)
		return nil
	})
}
