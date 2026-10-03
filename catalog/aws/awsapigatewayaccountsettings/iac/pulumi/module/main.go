package module

import (
	"github.com/pkg/errors"
	awsapigatewayaccountsettingsv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsapigatewayaccountsettings/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources manages the region's API Gateway account settings and
// exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awsapigatewayaccountsettingsv1alpha1.AwsApiGatewayAccountSettingsIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := accountSettings(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "api gateway account settings")
	}

	return nil
}
