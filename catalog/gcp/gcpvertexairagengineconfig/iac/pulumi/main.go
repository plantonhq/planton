package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcpvertexairagengineconfig/iac/pulumi/module"
	gcpvertexairagengineconfigv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexairagengineconfig/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &gcpvertexairagengineconfigv1alpha1.GcpVertexAiRagEngineConfigStackInput{}
		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return err
		}
		return module.Resources(ctx, stackInput)
	})
}
