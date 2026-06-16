package module

import (
	"github.com/pkg/errors"
	awsfsxontapstoragevirtualmachinev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsfsxontapstoragevirtualmachine/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/fsx"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awsfsxontapstoragevirtualmachinev1.AwsFsxOntapStorageVirtualMachineStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsFsxOntapStorageVirtualMachine.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdSvm, err := svm(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create fsx ontap storage virtual machine")
	}

	ctx.Export(OpSvmId, createdSvm.ID())
	ctx.Export(OpArn, createdSvm.Arn)
	ctx.Export(OpUuid, createdSvm.Uuid)
	ctx.Export(OpSubtype, createdSvm.Subtype)

	// SVM endpoints: iSCSI, management, NFS, SMB.
	// Each endpoint type has dns_name and ip_addresses.

	ctx.Export(OpIscsiDnsName, createdSvm.Endpoints.ApplyT(func(endpoints []fsx.OntapStorageVirtualMachineEndpoint) string {
		if len(endpoints) > 0 && len(endpoints[0].Iscsis) > 0 {
			if endpoints[0].Iscsis[0].DnsName != nil {
				return *endpoints[0].Iscsis[0].DnsName
			}
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpIscsiIpAddresses, createdSvm.Endpoints.ApplyT(func(endpoints []fsx.OntapStorageVirtualMachineEndpoint) []string {
		if len(endpoints) > 0 && len(endpoints[0].Iscsis) > 0 {
			return endpoints[0].Iscsis[0].IpAddresses
		}
		return nil
	}).(pulumi.StringArrayOutput))

	ctx.Export(OpManagementDnsName, createdSvm.Endpoints.ApplyT(func(endpoints []fsx.OntapStorageVirtualMachineEndpoint) string {
		if len(endpoints) > 0 && len(endpoints[0].Managements) > 0 {
			if endpoints[0].Managements[0].DnsName != nil {
				return *endpoints[0].Managements[0].DnsName
			}
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpManagementIpAddresses, createdSvm.Endpoints.ApplyT(func(endpoints []fsx.OntapStorageVirtualMachineEndpoint) []string {
		if len(endpoints) > 0 && len(endpoints[0].Managements) > 0 {
			return endpoints[0].Managements[0].IpAddresses
		}
		return nil
	}).(pulumi.StringArrayOutput))

	ctx.Export(OpNfsDnsName, createdSvm.Endpoints.ApplyT(func(endpoints []fsx.OntapStorageVirtualMachineEndpoint) string {
		if len(endpoints) > 0 && len(endpoints[0].Nfs) > 0 {
			if endpoints[0].Nfs[0].DnsName != nil {
				return *endpoints[0].Nfs[0].DnsName
			}
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpNfsIpAddresses, createdSvm.Endpoints.ApplyT(func(endpoints []fsx.OntapStorageVirtualMachineEndpoint) []string {
		if len(endpoints) > 0 && len(endpoints[0].Nfs) > 0 {
			return endpoints[0].Nfs[0].IpAddresses
		}
		return nil
	}).(pulumi.StringArrayOutput))

	ctx.Export(OpSmbDnsName, createdSvm.Endpoints.ApplyT(func(endpoints []fsx.OntapStorageVirtualMachineEndpoint) string {
		if len(endpoints) > 0 && len(endpoints[0].Smbs) > 0 {
			if endpoints[0].Smbs[0].DnsName != nil {
				return *endpoints[0].Smbs[0].DnsName
			}
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpSmbIpAddresses, createdSvm.Endpoints.ApplyT(func(endpoints []fsx.OntapStorageVirtualMachineEndpoint) []string {
		if len(endpoints) > 0 && len(endpoints[0].Smbs) > 0 {
			return endpoints[0].Smbs[0].IpAddresses
		}
		return nil
	}).(pulumi.StringArrayOutput))

	return nil
}
