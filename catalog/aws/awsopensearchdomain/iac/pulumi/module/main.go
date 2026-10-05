package module

import (
	"github.com/pkg/errors"
	awsopensearchdomainv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsopensearchdomain/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS OpenSearch Service domain resources and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awsopensearchdomainv1alpha1.AwsOpenSearchDomainIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	osDomain, err := domain(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "opensearch domain")
	}

	if err := satellites(ctx, locals, provider, osDomain); err != nil {
		return errors.Wrap(err, "opensearch domain satellites")
	}

	return nil
}
