package module

import (
	"github.com/pkg/errors"
	awsfsxontapvolumev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsfsxontapvolume/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awsfsxontapvolumev1.AwsFsxOntapVolumeStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(awsProviderConfig.GetRegion()),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
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
