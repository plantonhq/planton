package module

import (
	"github.com/pkg/errors"
	awscloudmapnamespacev1alpha1 "github.com/plantonhq/planton/catalog/aws/awscloudmapnamespace/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of the namespace (whichever type
// arm), its services and statically registered instances, and exports
// outputs.
func Resources(ctx *pulumi.Context, iacInput *awscloudmapnamespacev1alpha1.AwsCloudMapNamespaceIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := namespace(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "namespace")
	}

	return nil
}
