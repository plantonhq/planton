package module

import (
	"github.com/pkg/errors"
	gcpglobalforwardingrulev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpglobalforwardingrule/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *gcpglobalforwardingrulev1alpha1.GcpGlobalForwardingRuleIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	gcpProvider, err := pulumigoogleprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup google provider")
	}

	if err := forwardingRule(ctx, locals, gcpProvider); err != nil {
		return errors.Wrap(err, "failed to create forwarding rule")
	}

	return nil
}
