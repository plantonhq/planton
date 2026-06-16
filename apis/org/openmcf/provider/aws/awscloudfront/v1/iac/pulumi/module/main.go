package module

import (
	"github.com/pkg/errors"
	awscloudfrontv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awscloudfront/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awscloudfrontv1.AwsCloudFrontStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	dist, err := createDistribution(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "create cloudfront distribution")
	}

	// Export outputs mapped to AwsCloudFrontStackOutputs
	ctx.Export(OpDistributionId, dist.ID())
	ctx.Export(OpDomainName, dist.DomainName)
	ctx.Export(OpHostedZoneId, pulumi.String("Z2FDTNDATAQYW2"))

	return nil
}
