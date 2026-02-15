package module

import (
	"github.com/pkg/errors"
	gcpfirewallrulev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpfirewallrule/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the Pulumi program entry point invoked by the OpenMCF CLI.
// It wires provider credentials, initializes locals, and creates the firewall rule.
func Resources(ctx *pulumi.Context, stackInput *gcpfirewallrulev1.GcpFirewallRuleStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := firewall(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create firewall rule")
	}

	return nil
}
