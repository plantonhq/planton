package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcpcertmanagertrustconfig/iac/pulumi/module"
	gcpcertmanagertrustconfigv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcertmanagertrustconfig/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &gcpcertmanagertrustconfigv1alpha1.GcpCertManagerTrustConfigStackInput{}
		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return err
		}
		return module.Resources(ctx, stackInput)
	})
}
