package main

import (
	"github.com/plantonhq/planton/catalog/gcp/gcpfirewallrule/iac/pulumi/module"
	gcpfirewallrulev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpfirewallrule/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		iacInput := &gcpfirewallrulev1alpha1.GcpFirewallRuleIacInput{}

		if err := iacinput.LoadIacInput(ctx, iacInput); err != nil {
			return err
		}

		return module.Resources(ctx, iacInput)
	})
}
