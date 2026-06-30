package module

import (
	"github.com/pkg/errors"
	awsserverlesselasticachev1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsserverlesselasticache/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS ElastiCache Serverless resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsserverlesselasticachev1.AwsServerlessElasticacheStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := serverlessCache(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "serverless cache")
	}

	return nil
}
