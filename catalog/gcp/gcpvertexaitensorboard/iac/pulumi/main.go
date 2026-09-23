package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcpvertexaitensorboard/iac/pulumi/module"
	gcpvertexaitensorboardv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaitensorboard/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &gcpvertexaitensorboardv1alpha1.GcpVertexAiTensorboardStackInput{}
		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return err
		}
		return module.Resources(ctx, stackInput)
	})
}
