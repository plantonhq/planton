package module

import (
	"github.com/pkg/errors"
	gcpfirewallrulev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpfirewallrule/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi program entry point invoked by the Planton CLI.
// It wires provider credentials, initializes locals, and creates the firewall rule.
func Resources(ctx *pulumi.Context, iacInput *gcpfirewallrulev1alpha1.GcpFirewallRuleIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := firewall(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create firewall rule")
	}

	return nil
}
