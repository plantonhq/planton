package module

import (
	"fmt"

	"github.com/pkg/errors"
	azureapplicationgatewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azureapplicationgateway/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureapplicationgatewayv1.AzureApplicationGatewayStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	azureProviderConfig := stackInput.ProviderConfig

	// Create azure provider using the credentials from the input
	azureProvider, err := azure.NewProvider(ctx,
		"azure",
		&azure.ProviderArgs{
			ClientId:       pulumi.String(azureProviderConfig.ClientId),
			ClientSecret:   pulumi.String(azureProviderConfig.ClientSecret),
			SubscriptionId: pulumi.String(azureProviderConfig.SubscriptionId),
			TenantId:       pulumi.String(azureProviderConfig.TenantId),
		})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureApplicationGateway.Spec

	// Build SKU block.
	// Name and Tier use the same value (e.g. "Standard_v2" or "WAF_v2").
	// Capacity is set only when autoscale is not configured.
	skuArgs := network.ApplicationGatewaySkuArgs{
		Name: pulumi.String(spec.Sku),
		Tier: pulumi.String(spec.Sku),
	}
	if spec.Autoscale == nil {
		// Use fixed capacity (defaults to 2 via proto default)
		capacity := int(spec.GetCapacity())
		if capacity == 0 {
			capacity = 2
		}
		skuArgs.Capacity = pulumi.IntPtr(capacity)
	}

	// Build gateway IP configuration (single entry, auto-derived name).
	gatewayIpConfigs := network.ApplicationGatewayGatewayIpConfigurationArray{
		&network.ApplicationGatewayGatewayIpConfigurationArgs{
			Name:     pulumi.String(locals.GatewayIpConfigName),
			SubnetId: pulumi.String(spec.SubnetId.GetValue()),
		},
	}

	// Build frontend IP configuration (single entry, auto-derived name).
	frontendIpConfigs := network.ApplicationGatewayFrontendIpConfigurationArray{
		&network.ApplicationGatewayFrontendIpConfigurationArgs{
			Name:              pulumi.String(locals.FrontendIpConfigName),
			PublicIpAddressId: pulumi.StringPtr(spec.PublicIpId.GetValue()),
		},
	}

	// Build frontend ports (auto-derived from listeners as "{listener_name}-port").
	frontendPorts := network.ApplicationGatewayFrontendPortArray{}
	for _, listener := range spec.HttpListeners {
		frontendPorts = append(frontendPorts, &network.ApplicationGatewayFrontendPortArgs{
			Name: pulumi.String(fmt.Sprintf("%s-port", listener.Name)),
			Port: pulumi.Int(int(listener.Port)),
		})
	}

	// Build backend address pools from spec.
	backendPools := network.ApplicationGatewayBackendAddressPoolArray{}
	for _, pool := range spec.BackendAddressPools {
		poolArgs := &network.ApplicationGatewayBackendAddressPoolArgs{
			Name: pulumi.String(pool.Name),
		}
		if len(pool.Fqdns) > 0 {
			poolArgs.Fqdns = pulumi.ToStringArray(pool.Fqdns)
		}
		if len(pool.IpAddresses) > 0 {
			poolArgs.IpAddresses = pulumi.ToStringArray(pool.IpAddresses)
		}
		backendPools = append(backendPools, poolArgs)
	}

	// Build backend HTTP settings from spec.
	backendHttpSettings := network.ApplicationGatewayBackendHttpSettingArray{}
	for _, settings := range spec.BackendHttpSettings {
		settingsArgs := &network.ApplicationGatewayBackendHttpSettingArgs{
			Name:     pulumi.String(settings.Name),
			Port:     pulumi.Int(int(settings.Port)),
			Protocol: pulumi.String(settings.Protocol),
		}

		// Cookie-based affinity (defaults to "Disabled")
		cookieAffinity := settings.GetCookieBasedAffinity()
		if cookieAffinity == "" {
			cookieAffinity = "Disabled"
		}
		settingsArgs.CookieBasedAffinity = pulumi.String(cookieAffinity)

		// Request timeout (defaults to 30)
		requestTimeout := int(settings.GetRequestTimeout())
		if requestTimeout == 0 {
			requestTimeout = 30
		}
		settingsArgs.RequestTimeout = pulumi.IntPtr(requestTimeout)

		// Optional probe name reference
		if settings.ProbeName != "" {
			settingsArgs.ProbeName = pulumi.StringPtr(settings.ProbeName)
		}

		// Optional host name override
		if settings.HostName != "" {
			settingsArgs.HostName = pulumi.StringPtr(settings.HostName)
		}

		// Pick host name from backend address
		if settings.GetPickHostNameFromBackendAddress() {
			settingsArgs.PickHostNameFromBackendAddress = pulumi.BoolPtr(true)
		}

		backendHttpSettings = append(backendHttpSettings, settingsArgs)
	}

	// Build HTTP listeners from spec.
	httpListeners := network.ApplicationGatewayHttpListenerArray{}
	for _, listener := range spec.HttpListeners {
		listenerArgs := &network.ApplicationGatewayHttpListenerArgs{
			Name:                        pulumi.String(listener.Name),
			FrontendIpConfigurationName: pulumi.String(locals.FrontendIpConfigName),
			FrontendPortName:            pulumi.String(fmt.Sprintf("%s-port", listener.Name)),
			Protocol:                    pulumi.String(listener.Protocol),
		}

		// Optional host name for host-based routing
		if listener.HostName != "" {
			listenerArgs.HostName = pulumi.StringPtr(listener.HostName)
		}

		// Optional SSL certificate reference for HTTPS listeners
		if listener.SslCertificateName != "" {
			listenerArgs.SslCertificateName = pulumi.StringPtr(listener.SslCertificateName)
		}

		httpListeners = append(httpListeners, listenerArgs)
	}

	// Build request routing rules from spec.
	// Only Basic rule type is supported.
	routingRules := network.ApplicationGatewayRequestRoutingRuleArray{}
	for _, rule := range spec.RequestRoutingRules {
		routingRules = append(routingRules, &network.ApplicationGatewayRequestRoutingRuleArgs{
			Name:                    pulumi.String(rule.Name),
			RuleType:                pulumi.String("Basic"),
			HttpListenerName:        pulumi.String(rule.HttpListenerName),
			BackendAddressPoolName:  pulumi.StringPtr(rule.BackendAddressPoolName),
			BackendHttpSettingsName: pulumi.StringPtr(rule.BackendHttpSettingsName),
			Priority:                pulumi.IntPtr(int(rule.Priority)),
		})
	}

	// Build Application Gateway args
	appGwArgs := &network.ApplicationGatewayArgs{
		Name:                     pulumi.String(spec.Name),
		Location:                 pulumi.String(spec.Region),
		ResourceGroupName:        pulumi.String(locals.ResourceGroupName),
		Sku:                      skuArgs,
		GatewayIpConfigurations:  gatewayIpConfigs,
		FrontendIpConfigurations: frontendIpConfigs,
		FrontendPorts:            frontendPorts,
		BackendAddressPools:      backendPools,
		BackendHttpSettings:      backendHttpSettings,
		HttpListeners:            httpListeners,
		RequestRoutingRules:      routingRules,
		Tags:                     pulumi.ToStringMap(locals.AzureTags),
		EnableHttp2:              pulumi.BoolPtr(spec.GetEnableHttp2()),
	}

	// Build health probes (optional).
	if len(spec.Probes) > 0 {
		probes := network.ApplicationGatewayProbeArray{}
		for _, probe := range spec.Probes {
			probeArgs := &network.ApplicationGatewayProbeArgs{
				Name:     pulumi.String(probe.Name),
				Protocol: pulumi.String(probe.Protocol),
				Path:     pulumi.String(probe.Path),
			}

			// Optional host header
			if probe.Host != "" {
				probeArgs.Host = pulumi.StringPtr(probe.Host)
			}

			// Interval (defaults to 30)
			interval := int(probe.GetInterval())
			if interval == 0 {
				interval = 30
			}
			probeArgs.Interval = pulumi.Int(interval)

			// Timeout (defaults to 30)
			timeout := int(probe.GetTimeout())
			if timeout == 0 {
				timeout = 30
			}
			probeArgs.Timeout = pulumi.Int(timeout)

			// Unhealthy threshold (defaults to 3)
			unhealthyThreshold := int(probe.GetUnhealthyThreshold())
			if unhealthyThreshold == 0 {
				unhealthyThreshold = 3
			}
			probeArgs.UnhealthyThreshold = pulumi.Int(unhealthyThreshold)

			probes = append(probes, probeArgs)
		}
		appGwArgs.Probes = probes
	}

	// Build SSL certificates (optional).
	if len(spec.SslCertificates) > 0 {
		sslCerts := network.ApplicationGatewaySslCertificateArray{}
		for _, cert := range spec.SslCertificates {
			sslCerts = append(sslCerts, &network.ApplicationGatewaySslCertificateArgs{
				Name:             pulumi.String(cert.Name),
				KeyVaultSecretId: pulumi.StringPtr(cert.KeyVaultSecretId),
			})
		}
		appGwArgs.SslCertificates = sslCerts
	}

	// Build identity block (optional, only if identity_ids is non-empty).
	// Required when SSL certificates reference Key Vault secrets.
	if len(spec.IdentityIds) > 0 {
		identityIdStrs := pulumi.StringArray{}
		for _, idRef := range spec.IdentityIds {
			identityIdStrs = append(identityIdStrs, pulumi.String(idRef.GetValue()))
		}
		appGwArgs.Identity = network.ApplicationGatewayIdentityArgs{
			Type:        pulumi.String("UserAssigned"),
			IdentityIds: identityIdStrs,
		}
	}

	// Build WAF configuration (optional, only if waf_enabled is true).
	if spec.GetWafEnabled() {
		wafMode := spec.GetWafMode()
		if wafMode == "" {
			wafMode = "Prevention"
		}
		appGwArgs.WafConfiguration = network.ApplicationGatewayWafConfigurationArgs{
			Enabled:        pulumi.Bool(true),
			FirewallMode:   pulumi.String(wafMode),
			RuleSetType:    pulumi.StringPtr("OWASP"),
			RuleSetVersion: pulumi.String("3.2"),
		}
	}

	// Build autoscale configuration (optional, mutually exclusive with SKU capacity).
	if spec.Autoscale != nil {
		autoscaleArgs := network.ApplicationGatewayAutoscaleConfigurationArgs{
			MinCapacity: pulumi.Int(int(spec.Autoscale.MinCapacity)),
		}
		if spec.Autoscale.MaxCapacity != nil {
			autoscaleArgs.MaxCapacity = pulumi.IntPtr(int(spec.Autoscale.GetMaxCapacity()))
		}
		appGwArgs.AutoscaleConfiguration = autoscaleArgs
	}

	// Create the Application Gateway.
	appGateway, err := network.NewApplicationGateway(ctx,
		spec.Name,
		appGwArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Application Gateway %s", spec.Name)
	}

	// Export stack outputs.
	ctx.Export(OpAppGatewayId, appGateway.ID())
	ctx.Export(OpAppGatewayName, appGateway.Name)

	return nil
}
