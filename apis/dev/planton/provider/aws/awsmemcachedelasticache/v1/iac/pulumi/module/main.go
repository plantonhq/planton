package module

import (
	"github.com/pkg/errors"
	awsmemcachedelasticachev1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsmemcachedelasticache/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS ElastiCache Memcached resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsmemcachedelasticachev1.AwsMemcachedElasticacheStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.Spec.Region)
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

	// Cluster (always created)
	if err := cluster(ctx, locals, provider, createdSubnetGroup, createdParamGroup); err != nil {
		return errors.Wrap(err, "cluster")
	}

	return nil
}
