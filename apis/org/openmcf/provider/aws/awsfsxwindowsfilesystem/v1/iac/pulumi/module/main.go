package module

import (
	"github.com/pkg/errors"
	awsfsxwindowsfilesystemv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsfsxwindowsfilesystem/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awsfsxwindowsfilesystemv1.AwsFsxWindowsFileSystemStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsFsxWindowsFileSystem.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
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
