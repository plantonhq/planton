package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcpbigtabletable/iac/pulumi/module"
	gcpbigtabletablev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbigtabletable/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &gcpbigtabletablev1alpha1.GcpBigtableTableIacInput{}
		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return err
		}
		return module.Resources(ctx, iacInput)
	})
}
