package module

import (
	"github.com/pkg/errors"
	awsfsxlustrefilesystemv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsfsxlustrefilesystem/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *awsfsxlustrefilesystemv1alpha1.AwsFsxLustreFileSystemIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsFsxLustreFileSystem.Spec.Region)
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
