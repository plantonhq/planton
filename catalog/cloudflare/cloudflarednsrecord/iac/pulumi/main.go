package main

import (
	"github.com/plantonhq/planton/catalog/cloudflare/cloudflarednsrecord/iac/pulumi/module"
	cloudflarednsrecordv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarednsrecord/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &cloudflarednsrecordv1alpha1.CloudflareDnsRecordIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return err
		}

		return module.Resources(ctx, iacInput)
	})
}
