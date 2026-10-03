package module

import (
	"github.com/pkg/errors"
	awsecsclusterv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsecscluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the ECS cluster and its capacity: the cluster
// itself, one capacity provider per folded EC2 entry, and a single
// association that puts the Fargate built-ins and the EC2 providers onto
// the cluster together with the default strategy. The cluster composes
// onto its neighbors instead of embedding them: the auto-scaling groups
// that provide EC2 capacity, the KMS keys that encrypt exec sessions and
// Fargate storage, and the Cloud Map namespace Service Connect uses all
// attach by reference.
func Resources(ctx *pulumi.Context, iacInput *awsecsclusterv1alpha1.AwsEcsClusterIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static keys, keyless
	// web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsEcsCluster.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdCluster, err := cluster(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create ECS cluster")
	}

	createdProviders, err := capacityProviders(ctx, locals, provider, createdCluster)
	if err != nil {
		return errors.Wrap(err, "failed to create ECS cluster capacity providers")
	}

	ctx.Export(OpClusterName, createdCluster.Name)
	ctx.Export(OpClusterArn, createdCluster.Arn)

	// The full capacity vocabulary services can use in a strategy: the
	// associated built-ins plus every folded provider name (EC2 then
	// managed-instances), in spec order (matching the Terraform module
	// element-for-element).
	spec := locals.AwsEcsCluster.Spec
	capacityProviderNames := pulumi.StringArray{}
	for _, builtin := range spec.CapacityProviders {
		capacityProviderNames = append(capacityProviderNames, pulumi.String(builtin))
	}
	capacityProviderArns := pulumi.StringArray{}
	for _, createdProvider := range createdProviders {
		capacityProviderNames = append(capacityProviderNames, createdProvider.Name)
		capacityProviderArns = append(capacityProviderArns, createdProvider.Arn)
	}
	ctx.Export(OpCapacityProviderNames, capacityProviderNames)
	ctx.Export(OpCapacityProviderArns, capacityProviderArns)

	return nil
}
