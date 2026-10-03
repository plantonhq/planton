package module

import (
	"github.com/pkg/errors"
	awsfsxontapvolumev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsfsxontapvolume/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *awsfsxontapvolumev1alpha1.AwsFsxOntapVolumeIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsFsxOntapVolume.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdVolume, err := volume(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create fsx ontap volume")
	}

	ctx.Export(OpVolumeId, createdVolume.ID())
	ctx.Export(OpArn, createdVolume.Arn)
	ctx.Export(OpUuid, createdVolume.Uuid)
	ctx.Export(OpFileSystemId, createdVolume.FileSystemId)
	ctx.Export(OpFlexcacheEndpointType, createdVolume.FlexcacheEndpointType)
	ctx.Export(OpOntapVolumeType, createdVolume.OntapVolumeType)

	return nil
}
