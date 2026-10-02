package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcpvertexaifeatureonlinestore/iac/pulumi/module"
	gcpvertexaifeatureonlinestorev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaifeatureonlinestore/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &gcpvertexaifeatureonlinestorev1alpha1.GcpVertexAiFeatureOnlineStoreStackInput{}
		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return err
		}
		return module.Resources(ctx, stackInput)
	})
}
