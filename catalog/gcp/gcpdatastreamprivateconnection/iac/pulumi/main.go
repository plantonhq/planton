package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcpdatastreamprivateconnection/iac/pulumi/module"
	gcpdatastreamprivateconnectionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdatastreamprivateconnection/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &gcpdatastreamprivateconnectionv1alpha1.GcpDatastreamPrivateConnectionStackInput{}
		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return err
		}
		return module.Resources(ctx, stackInput)
	})
}
