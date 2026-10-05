package module

import (
	"github.com/pkg/errors"
	awssesaccountsettingsv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssesaccountsettings/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources manages the region's SES account settings and exports
// outputs.
func Resources(ctx *pulumi.Context, iacInput *awssesaccountsettingsv1alpha1.AwsSesAccountSettingsIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := accountSettings(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "ses account settings")
	}

	return nil
}
