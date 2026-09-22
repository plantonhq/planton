package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcprediscluster/iac/pulumi/module"
	gcpredisclusterv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcprediscluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &gcpredisclusterv1alpha1.GcpRedisClusterStackInput{}
		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return err
		}
		return module.Resources(ctx, stackInput)
	})
}
