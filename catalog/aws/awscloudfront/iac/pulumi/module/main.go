package module

import (
	"github.com/pkg/errors"
	awscloudfrontv1alpha1 "github.com/plantonhq/planton/catalog/aws/awscloudfront/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *awscloudfrontv1alpha1.AwsCloudFrontIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsCloudFront.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	dist, err := createDistribution(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudfront distribution")
	}

	// Names match the Terraform module's outputs.tf key-for-key so both
	// engines present one contract to consumers.
	ctx.Export(OpDistributionId, dist.ID())
	ctx.Export(OpDistributionArn, dist.Arn)
	ctx.Export(OpDomainName, dist.DomainName)
	ctx.Export(OpHostedZoneId, dist.HostedZoneId)
	ctx.Export(OpStatus, dist.Status)

	return nil
}
