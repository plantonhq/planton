package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcpdeploytarget/iac/pulumi/module"
	gcpdeploytargetv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdeploytarget/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &gcpdeploytargetv1alpha1.GcpDeployTargetStackInput{}
		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return err
		}
		return module.Resources(ctx, stackInput)
	})
}
