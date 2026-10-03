package module

import (
	"github.com/pkg/errors"
	awsserverlesselasticachev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsserverlesselasticache/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of the AWS ElastiCache Serverless cache.
// Subnets, security groups, and KMS keys attach by reference; this module
// provisions only the serverless cache resource and exports connection
// endpoints.
func Resources(ctx *pulumi.Context, iacInput *awsserverlesselasticachev1alpha1.AwsServerlessElasticacheIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := serverlessCache(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "serverless cache")
	}

	return nil
}
