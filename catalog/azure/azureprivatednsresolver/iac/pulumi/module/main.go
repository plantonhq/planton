package module

import (
	"fmt"

	"github.com/pkg/errors"
	azureprivatednsresolverv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureprivatednsresolver/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/privatedns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureprivatednsresolverv1alpha1.AzurePrivateDnsResolverIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzurePrivateDnsResolver.Spec

	// The resolver anchors to ONE virtual network (Azure allows at most
	// one resolver per network -- enforced at deploy time). Everything
	// except tags is create-only.
	createdResolver, err := privatedns.NewResolver(ctx,
		spec.Name,
		&privatedns.ResolverArgs{
			Name:              pulumi.String(spec.Name),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Location:          pulumi.String(spec.Region),
			VirtualNetworkId:  pulumi.String(spec.VirtualNetworkId.GetValue()),
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create dns private resolver %s", spec.Name)
	}

	// Inbound endpoints -- the private IPs on-premises DNS forwarders
	// send queries to. Each occupies its own dedicated subnet delegated
	// to "Microsoft.Network/dnsResolvers" (ARM validates the delegation
	// and network membership at deploy time). The FIRST endpoint
	// declared in the spec is the primary: its IP rides the singular
	// inbound_endpoint_ip output.
	inboundEndpointIps := pulumi.StringMap{}
	var firstInboundEndpointIp pulumi.StringInput = pulumi.String("")
	for i, endpoint := range spec.InboundEndpoints {
		// Unspecified applies "Dynamic", the provider's own default,
		// kept explicit so both engines send identical wire shapes.
		allocationMethod := "Dynamic"
		if wire, ok := allocationMethodWire[endpoint.PrivateIpAllocationMethod]; ok {
			allocationMethod = wire
		}

		ipConfigurationArgs := &privatedns.ResolverInboundEndpointIpConfigurationsArgs{
			SubnetId:                  pulumi.String(endpoint.SubnetId.GetValue()),
			PrivateIpAllocationMethod: pulumi.String(allocationMethod),
		}
		// Only STATIC allocation carries an address (spec-validated); the
		// provider rejects an address on Dynamic before touching ARM.
		if endpoint.PrivateIpAddress != "" {
			ipConfigurationArgs.PrivateIpAddress = pulumi.String(endpoint.PrivateIpAddress)
		}

		createdInboundEndpoint, err := privatedns.NewResolverInboundEndpoint(ctx,
			fmt.Sprintf("%s-inbound-%s", spec.Name, endpoint.Name),
			&privatedns.ResolverInboundEndpointArgs{
				Name:                 pulumi.String(endpoint.Name),
				PrivateDnsResolverId: createdResolver.ID(),
				// Endpoints deploy into the resolver's region (their
				// subnets belong to the resolver's own network).
				Location:         pulumi.String(spec.Region),
				IpConfigurations: ipConfigurationArgs,
				Tags:             pulumi.ToStringMap(locals.AzureTags),
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create inbound endpoint %s", endpoint.Name)
		}

		// The effective address -- static or dynamically assigned --
		// read back from ARM after create.
		endpointIp := createdInboundEndpoint.IpConfigurations.PrivateIpAddress().Elem()
		inboundEndpointIps[endpoint.Name] = endpointIp
		if i == 0 {
			firstInboundEndpointIp = endpointIp
		}
	}

	// Outbound endpoints -- the egress points queries leave Azure
	// through, steered by the forwarding rulesets that bind them. The
	// FIRST endpoint declared in the spec is the primary: its ARM id
	// rides the singular outbound_endpoint_id output that
	// AzurePrivateDnsResolverForwardingRuleset references by default.
	outboundEndpointIds := pulumi.StringMap{}
	var firstOutboundEndpointId pulumi.StringInput = pulumi.String("")
	for i, endpoint := range spec.OutboundEndpoints {
		createdOutboundEndpoint, err := privatedns.NewResolverOutboundEndpoint(ctx,
			fmt.Sprintf("%s-outbound-%s", spec.Name, endpoint.Name),
			&privatedns.ResolverOutboundEndpointArgs{
				Name:                 pulumi.String(endpoint.Name),
				PrivateDnsResolverId: createdResolver.ID(),
				Location:             pulumi.String(spec.Region),
				SubnetId:             pulumi.String(endpoint.SubnetId.GetValue()),
				Tags:                 pulumi.ToStringMap(locals.AzureTags),
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create outbound endpoint %s", endpoint.Name)
		}

		outboundEndpointIds[endpoint.Name] = createdOutboundEndpoint.ID()
		if i == 0 {
			firstOutboundEndpointId = createdOutboundEndpoint.ID()
		}
	}

	ctx.Export(OpDnsResolverId, createdResolver.ID())
	ctx.Export(OpDnsResolverName, createdResolver.Name)
	ctx.Export(OpInboundEndpointIp, firstInboundEndpointIp)
	ctx.Export(OpInboundEndpointIps, inboundEndpointIps)
	ctx.Export(OpOutboundEndpointId, firstOutboundEndpointId)
	ctx.Export(OpOutboundEndpointIds, outboundEndpointIds)

	return nil
}
