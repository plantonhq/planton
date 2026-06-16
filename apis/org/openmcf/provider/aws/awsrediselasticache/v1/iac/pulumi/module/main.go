package module

import (
	"github.com/pkg/errors"
	awsrediselasticachev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsrediselasticache/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS ElastiCache Redis/Valkey resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsrediselasticachev1.AwsRedisElasticacheStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// Subnet group (only when subnet_ids provided)
	createdSubnetGroup, err := subnetGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "subnet group")
	}

	// Parameter group (when parameters provided with family)
	createdParamGroup, err := parameterGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "parameter group")
	}

	// Replication group (always created)
	if err := replicationGroup(ctx, locals, provider, createdSubnetGroup, createdParamGroup); err != nil {
		return errors.Wrap(err, "replication group")
	}

	return nil
}
