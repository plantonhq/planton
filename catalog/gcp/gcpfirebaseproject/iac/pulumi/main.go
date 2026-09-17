package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcpfirebaseproject/iac/pulumi/module"
	gcpfirebaseprojectv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpfirebaseproject/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &gcpfirebaseprojectv1alpha1.GcpFirebaseProjectStackInput{}
		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return err
		}
		return module.Resources(ctx, stackInput)
	})
}
