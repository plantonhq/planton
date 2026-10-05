package main

import (
	"github.com/plantonhq/planton/catalog/cloudflare/cloudflareruleset/iac/pulumi/module"
	cloudflarerulesetv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareruleset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &cloudflarerulesetv1alpha1.CloudflareRulesetIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return err
		}

		return module.Resources(ctx, iacInput)
	})
}
