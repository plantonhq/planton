package module

import (
	"github.com/pkg/errors"
	awsfsxontapfilesystemv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsfsxontapfilesystem/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/fsx"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *awsfsxontapfilesystemv1alpha1.AwsFsxOntapFileSystemIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsFsxOntapFileSystem.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdFs, err := fileSystem(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create fsx ontap file system")
	}

	ctx.Export(OpFileSystemId, createdFs.ID())
	ctx.Export(OpFileSystemArn, createdFs.Arn)
	ctx.Export(OpNetworkInterfaceIds, createdFs.NetworkInterfaceIds)
	ctx.Export(OpVpcId, createdFs.VpcId)
	ctx.Export(OpOwnerId, createdFs.OwnerId)

	// ONTAP endpoints: management and intercluster.
	ctx.Export(OpManagementDnsName, createdFs.Endpoints.ApplyT(func(endpoints []fsx.OntapFileSystemEndpoint) string {
		if len(endpoints) > 0 && len(endpoints[0].Managements) > 0 {
			if endpoints[0].Managements[0].DnsName != nil {
				return *endpoints[0].Managements[0].DnsName
			}
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpManagementIpAddresses, createdFs.Endpoints.ApplyT(func(endpoints []fsx.OntapFileSystemEndpoint) []string {
		if len(endpoints) > 0 && len(endpoints[0].Managements) > 0 {
			return endpoints[0].Managements[0].IpAddresses
		}
		return nil
	}).(pulumi.StringArrayOutput))

	ctx.Export(OpInterclusterDnsName, createdFs.Endpoints.ApplyT(func(endpoints []fsx.OntapFileSystemEndpoint) string {
		if len(endpoints) > 0 && len(endpoints[0].Interclusters) > 0 {
			if endpoints[0].Interclusters[0].DnsName != nil {
				return *endpoints[0].Interclusters[0].DnsName
			}
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpInterclusterIpAddresses, createdFs.Endpoints.ApplyT(func(endpoints []fsx.OntapFileSystemEndpoint) []string {
		if len(endpoints) > 0 && len(endpoints[0].Interclusters) > 0 {
			return endpoints[0].Interclusters[0].IpAddresses
		}
		return nil
	}).(pulumi.StringArrayOutput))

	return nil
}
