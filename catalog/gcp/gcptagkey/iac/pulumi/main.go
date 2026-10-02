package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcptagkey/iac/pulumi/module"
	gcptagkeyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcptagkey/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &gcptagkeyv1alpha1.GcpTagKeyStackInput{}
		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return err
		}
		return module.Resources(ctx, stackInput)
	})
}
