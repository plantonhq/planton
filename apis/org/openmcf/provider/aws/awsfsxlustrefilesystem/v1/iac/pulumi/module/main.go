package module

import (
	"github.com/pkg/errors"
	awsfsxlustrefilesystemv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsfsxlustrefilesystem/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awsfsxlustrefilesystemv1.AwsFsxLustreFileSystemStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsFsxLustreFileSystem.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdFs, err := fileSystem(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create fsx lustre file system")
	}

	ctx.Export(OpFileSystemId, createdFs.ID())
	ctx.Export(OpFileSystemArn, createdFs.Arn)
	ctx.Export(OpDnsName, createdFs.DnsName)
	ctx.Export(OpMountName, createdFs.MountName)
	ctx.Export(OpNetworkInterfaceIds, createdFs.NetworkInterfaceIds)
	ctx.Export(OpVpcId, createdFs.VpcId)
	ctx.Export(OpFileSystemTypeVersion, createdFs.FileSystemTypeVersion)
	ctx.Export(OpOwnerId, createdFs.OwnerId)

	return nil
}
