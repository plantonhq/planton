package module

import (
	"github.com/pkg/errors"
	azureprivatelinkservicev1alpha1 "github.com/plantonhq/planton/catalog/azure/azureprivatelinkservice/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	// The classic SDK registers the Private Link Service resource under
	// the legacy "privatedns" module (token
	// azure:privatedns/linkService:LinkService) -- a historical placement
	// in the upstream provider, not a DNS resource. The wire surface is
	// azurerm_private_link_service.
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/privatedns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureprivatelinkservicev1alpha1.AzurePrivateLinkServiceIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzurePrivateLinkService.Spec

	// The NAT configurations consumer traffic is source-NATed through.
	// Setting a static private address flips ARM's allocation method to
	// Static -- the provider derives the method from the address's
	// presence, so the module only carries the address itself.
	natIpConfigurations := privatedns.LinkServiceNatIpConfigurationArray{}
	for _, natConfig := range spec.NatIpConfigurations {
		natArgs := &privatedns.LinkServiceNatIpConfigurationArgs{
			Name:     pulumi.String(natConfig.Name),
			SubnetId: pulumi.String(natConfig.SubnetId.GetValue()),
			Primary:  pulumi.Bool(natConfig.Primary),
		}
		if natConfig.PrivateIpAddress != "" {
			natArgs.PrivateIpAddress = pulumi.String(natConfig.PrivateIpAddress)
		}
		if natConfig.GetPrivateIpAddressVersion() != "" {
			natArgs.PrivateIpAddressVersion = pulumi.String(natConfig.GetPrivateIpAddressVersion())
		}
		natIpConfigurations = append(natIpConfigurations, natArgs)
	}

	serviceArgs := &privatedns.LinkServiceArgs{
		Name:                 pulumi.String(spec.Name),
		Location:             pulumi.String(spec.Region),
		ResourceGroupName:    pulumi.String(locals.ResourceGroupName),
		NatIpConfigurations:  natIpConfigurations,
		ProxyProtocolEnabled: pulumi.Bool(spec.ProxyProtocolEnabled),
		Tags:                 pulumi.ToStringMap(locals.AzureTags),
	}

	// Exactly one traffic destination (spec-validated): the Standard load
	// balancer frontends the service fronts, or one fixed private IP.
	if len(spec.LoadBalancerFrontendIpConfigurationIds) > 0 {
		frontendIds := pulumi.StringArray{}
		for _, frontendId := range spec.LoadBalancerFrontendIpConfigurationIds {
			frontendIds = append(frontendIds, pulumi.String(frontendId.GetValue()))
		}
		serviceArgs.LoadBalancerFrontendIpConfigurationIds = frontendIds
	}
	if spec.DestinationIpAddress != "" {
		serviceArgs.DestinationIpAddress = pulumi.String(spec.DestinationIpAddress)
	}

	if len(spec.AutoApprovalSubscriptionIds) > 0 {
		serviceArgs.AutoApprovalSubscriptionIds = pulumi.ToStringArray(spec.AutoApprovalSubscriptionIds)
	}
	if len(spec.VisibilitySubscriptionIds) > 0 {
		serviceArgs.VisibilitySubscriptionIds = pulumi.ToStringArray(spec.VisibilitySubscriptionIds)
	}
	if len(spec.Fqdns) > 0 {
		serviceArgs.Fqdns = pulumi.ToStringArray(spec.Fqdns)
	}

	createdService, err := privatedns.NewLinkService(ctx,
		spec.Name,
		serviceArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create private link service %s", spec.Name)
	}

	ctx.Export(OpPrivateLinkServiceId, createdService.ID())
	ctx.Export(OpPrivateLinkServiceName, createdService.Name)
	// The alias is the consumer-facing handle -- what another tenant uses
	// to request a private-endpoint connection without any RBAC here.
	ctx.Export(OpAlias, createdService.Alias)

	return nil
}
