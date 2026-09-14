package module

import (
	"github.com/pkg/errors"
	azurevpnsitev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevpnsite/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurevpnsitev1alpha1.AzureVpnSiteStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureVpnSite.Spec

	// Create the VPN site -- the Virtual WAN address-book entry for one
	// branch location. The object is free and provisions in seconds; it
	// deploys nothing at the branch. Deleting a site requires the
	// connections pointing at it to be gone first.
	siteArgs := &network.VpnSiteArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		VirtualWanId:      pulumi.String(spec.VirtualWanId.GetValue()),
		// The prefixes Azure routes into the tunnels. Empty is legal when
		// every link speaks BGP (the spec documents the convention).
		AddressCidrs: pulumi.ToStringArray(spec.AddressCidrs),
		Tags:         pulumi.ToStringMap(locals.AzureTags),
	}

	// The provider validates these as non-empty when configured -- omit
	// instead of sending "" so an unset spec field stays unset on the
	// wire (mirroring the Terraform module's null handling).
	if spec.DeviceVendor != "" {
		siteArgs.DeviceVendor = pulumi.String(spec.DeviceVendor)
	}
	if spec.DeviceModel != "" {
		siteArgs.DeviceModel = pulumi.String(spec.DeviceModel)
	}

	// The branch's internet links -- ARM returns each link's ID, which
	// the link_ids output republishes keyed by name for connections to
	// pin to.
	if len(spec.Links) > 0 {
		links := network.VpnSiteLinkArray{}
		for _, link := range spec.Links {
			linkArgs := &network.VpnSiteLinkArgs{
				Name:        pulumi.String(link.Name),
				SpeedInMbps: pulumi.Int(int(link.SpeedInMbps)),
			}
			// The spec guarantees at least one endpoint per link; omission
			// (not "") keeps the provider's validations off the unset one.
			if link.IpAddress != "" {
				linkArgs.IpAddress = pulumi.String(link.IpAddress)
			}
			if link.Fqdn != "" {
				linkArgs.Fqdn = pulumi.String(link.Fqdn)
			}
			if link.ProviderName != "" {
				linkArgs.ProviderName = pulumi.String(link.ProviderName)
			}
			if link.Bgp != nil {
				linkArgs.Bgp = &network.VpnSiteLinkBgpArgs{
					Asn:            pulumi.Int(int(link.Bgp.Asn)),
					PeeringAddress: pulumi.String(link.Bgp.PeeringAddress),
				}
			}
			links = append(links, linkArgs)
		}
		siteArgs.Links = links
	}

	// O365 breakout categories for SD-WAN partners. Unset sends nothing
	// (ARM's no-breakout default).
	if spec.O365Policy != nil {
		o365PolicyArgs := &network.VpnSiteO365PolicyArgs{}
		if spec.O365Policy.TrafficCategory != nil {
			o365PolicyArgs.TrafficCategory = &network.VpnSiteO365PolicyTrafficCategoryArgs{
				AllowEndpointEnabled:    pulumi.Bool(spec.O365Policy.TrafficCategory.GetAllowEndpointEnabled()),
				DefaultEndpointEnabled:  pulumi.Bool(spec.O365Policy.TrafficCategory.GetDefaultEndpointEnabled()),
				OptimizeEndpointEnabled: pulumi.Bool(spec.O365Policy.TrafficCategory.GetOptimizeEndpointEnabled()),
			}
		}
		siteArgs.O365Policy = o365PolicyArgs
	}

	createdSite, err := network.NewVpnSite(ctx,
		spec.Name,
		siteArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create vpn site %s", spec.Name)
	}

	// ARM generates each link's ID; republish them keyed by the link's
	// name so connections can reference
	// status.outputs.link_ids.<link-name>.
	linkIds := createdSite.Links.ApplyT(func(links []network.VpnSiteLink) map[string]string {
		ids := map[string]string{}
		for _, link := range links {
			if link.Id != nil {
				ids[link.Name] = *link.Id
			}
		}
		return ids
	}).(pulumi.StringMapOutput)

	ctx.Export(OpVpnSiteId, createdSite.ID())
	ctx.Export(OpVpnSiteName, createdSite.Name)
	ctx.Export(OpLinkIds, linkIds)

	return nil
}
