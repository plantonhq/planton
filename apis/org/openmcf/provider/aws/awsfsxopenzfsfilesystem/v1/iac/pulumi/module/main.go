package module

import (
	"github.com/pkg/errors"
	awsfsxopenzfsfilesystemv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsfsxopenzfsfilesystem/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awsfsxopenzfsfilesystemv1.AwsFsxOpenzfsFileSystemStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsFsxOpenzfsFileSystem.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdFs, err := fileSystem(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create fsx openzfs file system")
	}

	ctx.Export(OpFileSystemId, createdFs.ID())
	ctx.Export(OpFileSystemArn, createdFs.Arn)
	ctx.Export(OpDnsName, createdFs.DnsName)
	ctx.Export(OpEndpointIpAddress, createdFs.EndpointIpAddress)
	ctx.Export(OpRootVolumeId, createdFs.RootVolumeId)
	ctx.Export(OpNetworkInterfaceIds, createdFs.NetworkInterfaceIds)
	ctx.Export(OpVpcId, createdFs.VpcId)
	ctx.Export(OpOwnerId, createdFs.OwnerId)

	return nil
}
