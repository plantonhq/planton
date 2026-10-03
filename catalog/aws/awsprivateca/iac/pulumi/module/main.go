package module

import (
	"github.com/pkg/errors"
	awsprivatecav1alpha1 "github.com/plantonhq/planton/catalog/aws/awsprivateca/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of the CA with its composed
// activation, issued certificates, permission, and policy, and
// exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awsprivatecav1alpha1.AwsPrivateCaIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := certificateAuthority(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "certificate authority")
	}

	return nil
}
