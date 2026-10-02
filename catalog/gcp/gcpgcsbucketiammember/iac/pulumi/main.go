package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcpgcsbucketiammember/iac/pulumi/module"
	gcpgcsbucketiammemberv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpgcsbucketiammember/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &gcpgcsbucketiammemberv1alpha1.GcpGcsBucketIamMemberStackInput{}
		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return err
		}
		return module.Resources(ctx, stackInput)
	})
}
