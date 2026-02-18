package module

import (
	"github.com/pkg/errors"
	awsfsxwindowsfilesystemv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsfsxwindowsfilesystem/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awsfsxwindowsfilesystemv1.AwsFsxWindowsFileSystemStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.AwsFsxWindowsFileSystem.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.AwsFsxWindowsFileSystem.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	createdFs, err := fileSystem(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create fsx windows file system")
	}

	ctx.Export(OpFileSystemId, createdFs.ID())
	ctx.Export(OpFileSystemArn, createdFs.Arn)
	ctx.Export(OpDnsName, createdFs.DnsName)
	ctx.Export(OpPreferredFileServerIp, createdFs.PreferredFileServerIp)
	ctx.Export(OpRemoteAdministrationEndpoint, createdFs.RemoteAdministrationEndpoint)
	ctx.Export(OpNetworkInterfaceIds, createdFs.NetworkInterfaceIds)
	ctx.Export(OpVpcId, createdFs.VpcId)
	ctx.Export(OpOwnerId, createdFs.OwnerId)

	return nil
}
