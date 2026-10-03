package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcpnetworkendpointgroup/iac/pulumi/module"
	gcpnetworkendpointgroupv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpnetworkendpointgroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &gcpnetworkendpointgroupv1alpha1.GcpNetworkEndpointGroupIacInput{}
		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return err
		}
		return module.Resources(ctx, iacInput)
	})
}
