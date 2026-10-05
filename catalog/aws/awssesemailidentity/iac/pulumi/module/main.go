package module

import (
	"github.com/pkg/errors"
	awssesemailidentityv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssesemailidentity/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates the SES email identity, its satellites, and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awssesemailidentityv1alpha1.AwsSesEmailIdentityIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsSesEmailIdentity.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	result, err := emailIdentity(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create ses email identity")
	}

	ctx.Export(OpIdentityArn, result.Arn)
	ctx.Export(OpEmailIdentity, result.EmailIdentity)
	ctx.Export(OpIdentityType, result.IdentityType)
	ctx.Export(OpVerificationStatus, result.VerificationStatus)
	ctx.Export(OpDkimTokens, result.DkimTokens)

	return nil
}
