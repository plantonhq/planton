package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcpvpcpeering/iac/pulumi/module"
	gcpvpcpeeringv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvpcpeering/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &gcpvpcpeeringv1alpha1.GcpVpcPeeringIacInput{}
		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return err
		}
		return module.Resources(ctx, iacInput)
	})
}
