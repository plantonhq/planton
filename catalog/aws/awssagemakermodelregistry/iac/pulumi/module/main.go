package module

import (
	"github.com/pkg/errors"
	awssagemakermodelregistryv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssagemakermodelregistry/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of the model package group and its
// folded resource policy and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awssagemakermodelregistryv1alpha1.AwsSagemakerModelRegistryIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := modelRegistry(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "sagemaker model registry")
	}

	return nil
}
