package module

import (
	"github.com/pkg/errors"
	gcphierarchicalfirewallpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcphierarchicalfirewallpolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *gcphierarchicalfirewallpolicyv1alpha1.GcpHierarchicalFirewallPolicyStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	createdPolicy, err := firewallPolicy(ctx, locals, gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create hierarchical firewall policy")
	}

	if err := firewallPolicyRules(ctx, locals, gcpProvider, createdPolicy); err != nil {
		return errors.Wrap(err, "failed to create hierarchical firewall policy rules")
	}

	if err := firewallPolicyAssociations(ctx, locals, gcpProvider, createdPolicy); err != nil {
		return errors.Wrap(err, "failed to create hierarchical firewall policy associations")
	}

	return nil
}
