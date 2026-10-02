package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcpvertexaiagentengine/iac/pulumi/module"
	gcpvertexaiagentenginev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaiagentengine/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngineStackInput{}
		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return err
		}
		return module.Resources(ctx, stackInput)
	})
}
