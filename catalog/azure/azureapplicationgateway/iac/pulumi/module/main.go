package module

import (
	"github.com/pkg/errors"
	azureapplicationgatewayv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureapplicationgateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureapplicationgatewayv1alpha1.AzureApplicationGatewayStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureApplicationGateway.Spec

	// SKU name and tier carry the same value on the v2 platform (and
	// Basic); capacity is fixed sizing, autoscale replaces it (spec
	// validation guarantees exactly one of the two).
	skuArgs := &network.ApplicationGatewaySkuArgs{
		Name: pulumi.String(skuStrings[spec.Sku]),
		Tier: pulumi.String(skuStrings[spec.Sku]),
	}
	if spec.Capacity != nil {
		skuArgs.Capacity = pulumi.Int(int(spec.GetCapacity()))
	}

	// One atomic ARM resource: every sub-object below wires to the others
	// BY NAME within this args tree. Applies routinely run 15-25 minutes.
	gatewayArgs := &network.ApplicationGatewayArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
		Sku:               skuArgs,

		// The gateway's dedicated-subnet anchor -- pure ARM plumbing
		// derived from the spec's subnet_id; users never name it.
		GatewayIpConfigurations: network.ApplicationGatewayGatewayIpConfigurationArray{
			&network.ApplicationGatewayGatewayIpConfigurationArgs{
				Name:     pulumi.String(locals.GatewayIpConfigName),
				SubnetId: pulumi.String(spec.SubnetId.GetValue()),
			},
		},

		FrontendIpConfigurations: buildFrontendIpConfigurations(spec.FrontendIpConfigurations),
		FrontendPorts:            buildFrontendPorts(spec.FrontendPorts),
		BackendAddressPools:      buildBackendAddressPools(spec.BackendAddressPools),
	}
	if spec.Autoscale != nil {
		autoscaleArgs := &network.ApplicationGatewayAutoscaleConfigurationArgs{
			MinCapacity: pulumi.Int(int(spec.Autoscale.GetMinCapacity())),
		}
		if spec.Autoscale.MaxCapacity != nil {
			autoscaleArgs.MaxCapacity = pulumi.Int(int(spec.Autoscale.GetMaxCapacity()))
		}
		gatewayArgs.AutoscaleConfiguration = autoscaleArgs
	}

	if len(spec.Zones) > 0 {
		gatewayArgs.Zones = pulumi.ToStringArray(spec.Zones)
	}

	if spec.Identity != nil {
		identityIds := make([]string, 0, len(spec.Identity.IdentityIds))
		for _, identityId := range spec.Identity.IdentityIds {
			identityIds = append(identityIds, identityId.GetValue())
		}
		identityArgs := &network.ApplicationGatewayIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(identityIds) > 0 {
			identityArgs.IdentityIds = pulumi.ToStringArray(identityIds)
		}
		gatewayArgs.Identity = identityArgs
	}

	if len(spec.BackendHttpSettings) > 0 {
		gatewayArgs.BackendHttpSettings = buildBackendHttpSettings(spec.BackendHttpSettings)
	}
	if len(spec.HttpListeners) > 0 {
		gatewayArgs.HttpListeners = buildHttpListeners(spec.HttpListeners)
	}
	if len(spec.RequestRoutingRules) > 0 {
		gatewayArgs.RequestRoutingRules = buildRequestRoutingRules(spec.RequestRoutingRules)
	}
	if len(spec.UrlPathMaps) > 0 {
		gatewayArgs.UrlPathMaps = buildUrlPathMaps(spec.UrlPathMaps)
	}
	if len(spec.Probes) > 0 {
		gatewayArgs.Probes = buildProbes(spec.Probes)
	}
	if len(spec.SslCertificates) > 0 {
		gatewayArgs.SslCertificates = buildSslCertificates(spec.SslCertificates)
	}
	if len(spec.TrustedRootCertificates) > 0 {
		gatewayArgs.TrustedRootCertificates = buildTrustedRootCertificates(spec.TrustedRootCertificates)
	}
	if len(spec.TrustedClientCertificates) > 0 {
		gatewayArgs.TrustedClientCertificates = buildTrustedClientCertificates(spec.TrustedClientCertificates)
	}
	if len(spec.SslProfiles) > 0 {
		gatewayArgs.SslProfiles = buildSslProfiles(spec.SslProfiles)
	}
	if spec.SslPolicy != nil {
		gatewayArgs.SslPolicy = buildGlobalSslPolicy(spec.SslPolicy)
	}
	if len(spec.RedirectConfigurations) > 0 {
		gatewayArgs.RedirectConfigurations = buildRedirectConfigurations(spec.RedirectConfigurations)
	}
	if len(spec.RewriteRuleSets) > 0 {
		gatewayArgs.RewriteRuleSets = buildRewriteRuleSets(spec.RewriteRuleSets)
	}
	if len(spec.Listeners) > 0 {
		gatewayArgs.Listeners = buildLayer4Listeners(spec.Listeners)
	}
	if len(spec.Backends) > 0 {
		gatewayArgs.Backends = buildLayer4Backends(spec.Backends)
	}
	if len(spec.RoutingRules) > 0 {
		gatewayArgs.RoutingRules = buildLayer4RoutingRules(spec.RoutingRules)
	}
	if len(spec.CustomErrorConfigurations) > 0 {
		gatewayArgs.CustomErrorConfigurations = buildCustomErrorConfigurations(spec.CustomErrorConfigurations)
	}
	if len(spec.PrivateLinkConfigurations) > 0 {
		gatewayArgs.PrivateLinkConfigurations = buildPrivateLinkConfigurations(spec.PrivateLinkConfigurations)
	}

	// The WAF policy attachment (WAF_v2 only; per-listener and per-path
	// overrides live on their blocks).
	if spec.FirewallPolicyId.GetValue() != "" {
		gatewayArgs.FirewallPolicyId = pulumi.String(spec.FirewallPolicyId.GetValue())
	}
	gatewayArgs.ForceFirewallPolicyAssociation = pulumi.Bool(spec.ForceFirewallPolicyAssociation)

	// Presence-guarded false-default: unset falls back to Azure's default
	// (HTTP/2 off) -- stack inputs built from a manifest do NOT
	// materialize proto defaults.
	if spec.Http2Enabled != nil {
		gatewayArgs.Http2Enabled = pulumi.Bool(spec.GetHttp2Enabled())
	} else {
		gatewayArgs.Http2Enabled = pulumi.Bool(false)
	}
	gatewayArgs.FipsEnabled = pulumi.Bool(spec.FipsEnabled)

	// Request/response buffering (Azure defaults both to true when the
	// block is absent; spec validation requires both when declared).
	if spec.GlobalConfiguration != nil {
		gatewayArgs.Global = &network.ApplicationGatewayGlobalArgs{
			RequestBufferingEnabled:  pulumi.Bool(spec.GlobalConfiguration.GetRequestBufferingEnabled()),
			ResponseBufferingEnabled: pulumi.Bool(spec.GlobalConfiguration.GetResponseBufferingEnabled()),
		}
	}

	createdGateway, err := network.NewApplicationGateway(ctx,
		spec.Name,
		gatewayArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create application gateway %s", spec.Name)
	}

	// Export stack outputs. The name-keyed maps are the composition
	// seams: NICs and scale sets join pools through
	// backend_address_pool_ids; frontends chain through
	// frontend_ip_configuration_ids.
	ctx.Export(OpApplicationGatewayId, createdGateway.ID())
	ctx.Export(OpApplicationGatewayName, createdGateway.Name)

	ctx.Export(OpBackendAddressPoolIds, createdGateway.BackendAddressPools.ApplyT(func(pools []network.ApplicationGatewayBackendAddressPool) map[string]string {
		ids := make(map[string]string, len(pools))
		for _, pool := range pools {
			if pool.Id != nil {
				ids[pool.Name] = *pool.Id
			}
		}
		return ids
	}))

	ctx.Export(OpFrontendIpConfigurationIds, createdGateway.FrontendIpConfigurations.ApplyT(func(frontends []network.ApplicationGatewayFrontendIpConfiguration) map[string]string {
		ids := make(map[string]string, len(frontends))
		for _, frontend := range frontends {
			if frontend.Id != nil {
				ids[frontend.Name] = *frontend.Id
			}
		}
		return ids
	}))

	// The private frontends' addresses (a public frontend's address
	// lives on its referenced AzurePublicIp resource).
	ctx.Export(OpPrivateIpAddress, createdGateway.FrontendIpConfigurations.ApplyT(func(frontends []network.ApplicationGatewayFrontendIpConfiguration) string {
		for _, frontend := range frontends {
			if frontend.PrivateIpAddress != nil && *frontend.PrivateIpAddress != "" {
				return *frontend.PrivateIpAddress
			}
		}
		return ""
	}))
	ctx.Export(OpPrivateIpAddresses, createdGateway.FrontendIpConfigurations.ApplyT(func(frontends []network.ApplicationGatewayFrontendIpConfiguration) []string {
		addresses := make([]string, 0, len(frontends))
		for _, frontend := range frontends {
			if frontend.PrivateIpAddress != nil && *frontend.PrivateIpAddress != "" {
				addresses = append(addresses, *frontend.PrivateIpAddress)
			}
		}
		return addresses
	}))

	return nil
}
