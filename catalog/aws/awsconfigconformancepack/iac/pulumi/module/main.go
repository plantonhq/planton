package module

import (
	"github.com/pkg/errors"
	awsconfigconformancepackv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsconfigconformancepack/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of the conformance pack (account or
// organization scope) and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awsconfigconformancepackv1alpha1.AwsConfigConformancePackIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := conformancePack(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "conformance pack")
	}

	return nil
}
