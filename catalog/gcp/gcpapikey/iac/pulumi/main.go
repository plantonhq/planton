package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcpapikey/iac/pulumi/module"
	gcpapikeyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpapikey/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &gcpapikeyv1alpha1.GcpApiKeyIacInput{}
		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return err
		}
		return module.Resources(ctx, iacInput)
	})
}
