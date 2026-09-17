package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcpfirebasewebapp/iac/pulumi/module"
	gcpfirebasewebappv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpfirebasewebapp/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &gcpfirebasewebappv1alpha1.GcpFirebaseWebAppStackInput{}
		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return err
		}
		return module.Resources(ctx, stackInput)
	})
}
