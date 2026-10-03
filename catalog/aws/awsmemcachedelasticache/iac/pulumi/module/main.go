package module

import (
	"github.com/pkg/errors"
	awsmemcachedelasticachev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsmemcachedelasticache/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS ElastiCache Memcached resources.
// Subnet and parameter groups are managed inline only when the spec brings
// raw subnets or inline parameters; existing group names short-circuit
// creation (CEL-enforced mutual exclusion). The cluster itself is always
// provisioned and exports connection endpoints.
func Resources(ctx *pulumi.Context, iacInput *awsmemcachedelasticachev1alpha1.AwsMemcachedElasticacheIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdSubnetGroup, err := subnetGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "subnet group")
	}

	createdParamGroup, err := parameterGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "parameter group")
	}

	if err := cluster(ctx, locals, provider, createdSubnetGroup, createdParamGroup); err != nil {
		return errors.Wrap(err, "cluster")
	}

	return nil
}
