package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcpmonitoringslo/iac/pulumi/module"
	gcpmonitoringslov1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmonitoringslo/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &gcpmonitoringslov1alpha1.GcpMonitoringSloIacInput{}
		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return err
		}
		return module.Resources(ctx, iacInput)
	})
}
