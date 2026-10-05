package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcpdeploytarget/iac/pulumi/module"
	gcpdeploytargetv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdeploytarget/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &gcpdeploytargetv1alpha1.GcpDeployTargetIacInput{}
		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return err
		}
		return module.Resources(ctx, iacInput)
	})
}
