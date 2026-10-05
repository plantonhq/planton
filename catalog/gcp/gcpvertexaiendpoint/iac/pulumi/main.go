package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcpvertexaiendpoint/iac/pulumi/module"
	gcpvertexaiendpointv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaiendpoint/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &gcpvertexaiendpointv1alpha1.GcpVertexAiEndpointIacInput{}
		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return err
		}
		return module.Resources(ctx, iacInput)
	})
}
