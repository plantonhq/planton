package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/fsx"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func svm(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*fsx.OntapStorageVirtualMachine, error) {
	spec := locals.AwsFsxOntapStorageVirtualMachine.Spec
	name := locals.AwsFsxOntapStorageVirtualMachine.Metadata.Name

	args := &fsx.OntapStorageVirtualMachineArgs{
		FileSystemId: pulumi.String(spec.FileSystemId.GetValue()),
		Name:         pulumi.StringPtr(spec.Name),
		Tags:         pulumi.ToStringMap(locals.AwsTags),
	}

	// Root volume security style (ForceNew, default UNIX via OpenMCF middleware).
	if spec.GetRootVolumeSecurityStyle() != "" {
		args.RootVolumeSecurityStyle = pulumi.StringPtr(spec.GetRootVolumeSecurityStyle())
	}

	// SVM admin password (sensitive, optional).
	if spec.SvmAdminPassword != "" {
		args.SvmAdminPassword = pulumi.StringPtr(spec.SvmAdminPassword)
	}

	// Active Directory configuration (optional — for SMB access).
	if spec.ActiveDirectoryConfiguration != nil {
		ad := spec.ActiveDirectoryConfiguration

		dnsIps := make(pulumi.StringArray, 0, len(ad.DnsIps))
		for _, ip := range ad.DnsIps {
			dnsIps = append(dnsIps, pulumi.String(ip))
		}

		smadArgs := &fsx.OntapStorageVirtualMachineActiveDirectoryConfigurationSelfManagedActiveDirectoryConfigurationArgs{
			DomainName: pulumi.String(ad.DomainName),
			DnsIps:     dnsIps,
			Username:   pulumi.String(ad.Username),
			Password:   pulumi.String(ad.Password),
		}

		if ad.GetFileSystemAdministratorsGroup() != "" {
			smadArgs.FileSystemAdministratorsGroup = pulumi.StringPtr(ad.GetFileSystemAdministratorsGroup())
		}

		if ad.OrganizationalUnitDistinguishedName != "" {
			smadArgs.OrganizationalUnitDistinguishedName = pulumi.StringPtr(ad.OrganizationalUnitDistinguishedName)
		}

		adArgs := &fsx.OntapStorageVirtualMachineActiveDirectoryConfigurationArgs{
			SelfManagedActiveDirectoryConfiguration: smadArgs,
		}

		if ad.NetbiosName != "" {
			adArgs.NetbiosName = pulumi.StringPtr(ad.NetbiosName)
		}

		args.ActiveDirectoryConfiguration = adArgs
	}

	createdSvm, err := fsx.NewOntapStorageVirtualMachine(ctx, name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create fsx ontap storage virtual machine")
	}

	return createdSvm, nil
}
