package module

import (
	"github.com/pkg/errors"
	azurebastionhostv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurebastionhost/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	// The classic SDK's bridge parks BastionHost in the COMPUTE package
	// (token azure:compute/bastionHost:BastionHost), not network where
	// the ARM resource lives -- do not "fix" this import into a break.
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurebastionhostv1alpha1.AzureBastionHostIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureBastionHost.Spec

	// Fixed at 2 on Developer/Basic (the provider default, always sent);
	// 2-50 on Standard/Premium (spec-validated).
	scaleUnits := 2
	if spec.ScaleUnits != nil {
		scaleUnits = int(*spec.ScaleUnits)
	}

	// Available on every SKU; the provider defaults it true.
	copyPasteEnabled := true
	if spec.CopyPasteEnabled != nil {
		copyPasteEnabled = *spec.CopyPasteEnabled
	}

	hostArgs := &compute.BastionHostArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Sku:               pulumi.String(locals.SkuWireValue),
		ScaleUnits:        pulumi.Int(scaleUnits),
		CopyPasteEnabled:  pulumi.Bool(copyPasteEnabled),
		// Standard/Premium feature knobs (spec-validated against the
		// SKU). kerberos_enabled is applied at CREATE only -- the
		// provider has no update path for it and silently ignores later
		// changes.
		FileCopyEnabled:         pulumi.Bool(spec.FileCopyEnabled),
		IpConnectEnabled:        pulumi.Bool(spec.IpConnectEnabled),
		KerberosEnabled:         pulumi.Bool(spec.KerberosEnabled),
		ShareableLinkEnabled:    pulumi.Bool(spec.ShareableLinkEnabled),
		TunnelingEnabled:        pulumi.Bool(spec.TunnelingEnabled),
		SessionRecordingEnabled: pulumi.Bool(spec.SessionRecordingEnabled),
		Tags:                    pulumi.ToStringMap(locals.AzureTags),
	}

	// Developer SKU only (spec-validated): the shared-infrastructure
	// host attaches to a virtual network directly -- no subnet, no
	// public IP.
	if spec.VirtualNetworkId.GetValue() != "" {
		hostArgs.VirtualNetworkId = pulumi.String(spec.VirtualNetworkId.GetValue())
	}

	// The dedicated-infrastructure binding: the "AzureBastionSubnet"
	// (the exact ARM name -- validated at deploy time) plus the public
	// IP the host binds exclusively. Premium may omit the public IP to
	// deploy private-only (surfaced in the private_only_enabled output).
	if spec.IpConfiguration != nil {
		ipConfigurationArgs := &compute.BastionHostIpConfigurationArgs{
			Name:     pulumi.String(spec.IpConfiguration.Name),
			SubnetId: pulumi.String(spec.IpConfiguration.SubnetId.GetValue()),
		}
		if spec.IpConfiguration.PublicIpAddressId.GetValue() != "" {
			ipConfigurationArgs.PublicIpAddressId = pulumi.String(spec.IpConfiguration.PublicIpAddressId.GetValue())
		}
		hostArgs.IpConfiguration = ipConfigurationArgs
	}

	// Availability zone pinning; empty deploys regionally. Fixed at
	// creation.
	if len(spec.Zones) > 0 {
		hostArgs.Zones = pulumi.ToStringArray(spec.Zones)
	}

	createdHost, err := compute.NewBastionHost(ctx,
		spec.Name,
		hostArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create bastion host %s", spec.Name)
	}

	ctx.Export(OpBastionHostId, createdHost.ID())
	ctx.Export(OpBastionHostName, createdHost.Name)
	ctx.Export(OpDnsName, createdHost.DnsName)
	ctx.Export(OpPrivateOnlyEnabled, createdHost.PrivateOnlyEnabled)

	return nil
}
