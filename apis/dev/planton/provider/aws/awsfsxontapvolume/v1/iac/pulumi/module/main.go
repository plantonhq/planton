package module

import (
	"github.com/pkg/errors"
	awsfsxontapvolumev1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsfsxontapvolume/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awsfsxontapvolumev1.AwsFsxOntapVolumeStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsFsxOntapVolume.Spec.Region)
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
