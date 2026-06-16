package module

import (
	"github.com/pkg/errors"
	awsmemorydbclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsmemorydbcluster/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS MemoryDB cluster resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsmemorydbclusterv1.AwsMemorydbClusterStackInput) error {
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

	// MemoryDB cluster (always created)
	if err := cluster(ctx, locals, provider, createdSubnetGroup, createdParamGroup); err != nil {
		return errors.Wrap(err, "memorydb cluster")
	}

	return nil
}
