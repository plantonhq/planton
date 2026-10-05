package module

import (
	azureapplicationgatewayv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureapplicationgateway/v1alpha1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The builders below translate each spec block into its SDK args,
// preserving the by-name wiring between sub-objects. Optional strings are
// only forwarded when non-empty so the ARM payload stays free of empty
// no-op properties.

func buildFrontendIpConfigurations(frontends []*azureapplicationgatewayv1alpha1.AzureApplicationGatewayFrontendIpConfiguration) network.ApplicationGatewayFrontendIpConfigurationArray {
	out := make(network.ApplicationGatewayFrontendIpConfigurationArray, 0, len(frontends))
	for _, frontend := range frontends {
		args := &network.ApplicationGatewayFrontendIpConfigurationArgs{
			Name: pulumi.String(frontend.Name),
		}
		if frontend.PublicIpAddressId.GetValue() != "" {
			args.PublicIpAddressId = pulumi.String(frontend.PublicIpAddressId.GetValue())
		}
		if frontend.SubnetId.GetValue() != "" {
			args.SubnetId = pulumi.String(frontend.SubnetId.GetValue())
		}
		if frontend.PrivateIpAddress != "" {
			args.PrivateIpAddress = pulumi.String(frontend.PrivateIpAddress)
		}
		if frontend.PrivateIpAddressAllocation != azureapplicationgatewayv1alpha1.AzureApplicationGatewayIpAllocation_azure_application_gateway_ip_allocation_unspecified {
			args.PrivateIpAddressAllocation = pulumi.String(ipAllocationStrings[frontend.PrivateIpAddressAllocation])
		}
		if frontend.PrivateLinkConfigurationName != "" {
			args.PrivateLinkConfigurationName = pulumi.String(frontend.PrivateLinkConfigurationName)
		}
		out = append(out, args)
	}
	return out
}

func buildFrontendPorts(ports []*azureapplicationgatewayv1alpha1.AzureApplicationGatewayFrontendPort) network.ApplicationGatewayFrontendPortArray {
	out := make(network.ApplicationGatewayFrontendPortArray, 0, len(ports))
	for _, port := range ports {
		out = append(out, &network.ApplicationGatewayFrontendPortArgs{
			Name: pulumi.String(port.Name),
			Port: pulumi.Int(int(port.Port)),
		})
	}
	return out
}

func buildBackendAddressPools(pools []*azureapplicationgatewayv1alpha1.AzureApplicationGatewayBackendAddressPool) network.ApplicationGatewayBackendAddressPoolArray {
	out := make(network.ApplicationGatewayBackendAddressPoolArray, 0, len(pools))
	for _, pool := range pools {
		args := &network.ApplicationGatewayBackendAddressPoolArgs{
			Name: pulumi.String(pool.Name),
		}
		if len(pool.Fqdns) > 0 {
			args.Fqdns = pulumi.ToStringArray(pool.Fqdns)
		}
		if len(pool.IpAddresses) > 0 {
			args.IpAddresses = pulumi.ToStringArray(pool.IpAddresses)
		}
		out = append(out, args)
	}
	return out
}

func buildBackendHttpSettings(settings []*azureapplicationgatewayv1alpha1.AzureApplicationGatewayBackendHttpSettings) network.ApplicationGatewayBackendHttpSettingArray {
	out := make(network.ApplicationGatewayBackendHttpSettingArray, 0, len(settings))
	for _, setting := range settings {
		// cookie_based_affinity is ARM's Enabled/Disabled string behind
		// the spec's boolean.
		affinity := "Disabled"
		if setting.CookieBasedAffinityEnabled {
			affinity = "Enabled"
		}
		args := &network.ApplicationGatewayBackendHttpSettingArgs{
			Name:                              pulumi.String(setting.Name),
			Port:                              pulumi.Int(int(setting.Port)),
			Protocol:                          pulumi.String(protocolStrings[setting.Protocol]),
			CookieBasedAffinity:               pulumi.String(affinity),
			PickHostNameFromBackendAddress:    pulumi.Bool(setting.PickHostNameFromBackendAddress),
			DedicatedBackendConnectionEnabled: pulumi.Bool(setting.DedicatedBackendConnectionEnabled),
		}
		if setting.AffinityCookieName != "" {
			args.AffinityCookieName = pulumi.String(setting.AffinityCookieName)
		}
		if setting.Path != "" {
			args.Path = pulumi.String(setting.Path)
		}
		// Presence-guarded: unset falls back to Azure's default (30s) --
		// IaC inputs built from a manifest do NOT materialize proto
		// defaults.
		if setting.RequestTimeout != nil {
			args.RequestTimeout = pulumi.Int(int(setting.GetRequestTimeout()))
		} else {
			args.RequestTimeout = pulumi.Int(30)
		}
		if setting.ProbeName != "" {
			args.ProbeName = pulumi.String(setting.ProbeName)
		}
		if setting.HostName != "" {
			args.HostName = pulumi.String(setting.HostName)
		}
		if len(setting.TrustedRootCertificateNames) > 0 {
			args.TrustedRootCertificateNames = pulumi.ToStringArray(setting.TrustedRootCertificateNames)
		}
		if setting.ConnectionDraining != nil {
			args.ConnectionDraining = &network.ApplicationGatewayBackendHttpSettingConnectionDrainingArgs{
				Enabled:         pulumi.Bool(setting.ConnectionDraining.Enabled),
				DrainTimeoutSec: pulumi.Int(int(setting.ConnectionDraining.DrainTimeoutSec)),
			}
		}
		out = append(out, args)
	}
	return out
}

func buildHttpListeners(listeners []*azureapplicationgatewayv1alpha1.AzureApplicationGatewayHttpListener) network.ApplicationGatewayHttpListenerArray {
	out := make(network.ApplicationGatewayHttpListenerArray, 0, len(listeners))
	for _, listener := range listeners {
		args := &network.ApplicationGatewayHttpListenerArgs{
			Name:                        pulumi.String(listener.Name),
			FrontendIpConfigurationName: pulumi.String(listener.FrontendIpConfigurationName),
			FrontendPortName:            pulumi.String(listener.FrontendPortName),
			Protocol:                    pulumi.String(protocolStrings[listener.Protocol]),
			RequireSni:                  pulumi.Bool(listener.RequireSni),
		}
		if len(listener.HostNames) > 0 {
			args.HostNames = pulumi.ToStringArray(listener.HostNames)
		}
		if listener.SslCertificateName != "" {
			args.SslCertificateName = pulumi.String(listener.SslCertificateName)
		}
		if listener.SslProfileName != "" {
			args.SslProfileName = pulumi.String(listener.SslProfileName)
		}
		if listener.FirewallPolicyId.GetValue() != "" {
			args.FirewallPolicyId = pulumi.String(listener.FirewallPolicyId.GetValue())
		}
		if len(listener.CustomErrorConfigurations) > 0 {
			errorConfigs := make(network.ApplicationGatewayHttpListenerCustomErrorConfigurationArray, 0, len(listener.CustomErrorConfigurations))
			for _, errorConfig := range listener.CustomErrorConfigurations {
				errorConfigs = append(errorConfigs, &network.ApplicationGatewayHttpListenerCustomErrorConfigurationArgs{
					StatusCode:         pulumi.String(statusCodeStrings[errorConfig.StatusCode]),
					CustomErrorPageUrl: pulumi.String(errorConfig.CustomErrorPageUrl),
				})
			}
			args.CustomErrorConfigurations = errorConfigs
		}
		out = append(out, args)
	}
	return out
}

func buildRequestRoutingRules(rules []*azureapplicationgatewayv1alpha1.AzureApplicationGatewayRequestRoutingRule) network.ApplicationGatewayRequestRoutingRuleArray {
	out := make(network.ApplicationGatewayRequestRoutingRuleArray, 0, len(rules))
	for _, rule := range rules {
		args := &network.ApplicationGatewayRequestRoutingRuleArgs{
			Name:             pulumi.String(rule.Name),
			RuleType:         pulumi.String(ruleTypeStrings[rule.RuleType]),
			HttpListenerName: pulumi.String(rule.HttpListenerName),
			Priority:         pulumi.Int(int(rule.Priority)),
		}
		if rule.BackendAddressPoolName != "" {
			args.BackendAddressPoolName = pulumi.String(rule.BackendAddressPoolName)
		}
		if rule.BackendHttpSettingsName != "" {
			args.BackendHttpSettingsName = pulumi.String(rule.BackendHttpSettingsName)
		}
		if rule.UrlPathMapName != "" {
			args.UrlPathMapName = pulumi.String(rule.UrlPathMapName)
		}
		if rule.RedirectConfigurationName != "" {
			args.RedirectConfigurationName = pulumi.String(rule.RedirectConfigurationName)
		}
		if rule.RewriteRuleSetName != "" {
			args.RewriteRuleSetName = pulumi.String(rule.RewriteRuleSetName)
		}
		out = append(out, args)
	}
	return out
}

func buildUrlPathMaps(pathMaps []*azureapplicationgatewayv1alpha1.AzureApplicationGatewayUrlPathMap) network.ApplicationGatewayUrlPathMapArray {
	out := make(network.ApplicationGatewayUrlPathMapArray, 0, len(pathMaps))
	for _, pathMap := range pathMaps {
		args := &network.ApplicationGatewayUrlPathMapArgs{
			Name: pulumi.String(pathMap.Name),
		}
		if pathMap.DefaultBackendAddressPoolName != "" {
			args.DefaultBackendAddressPoolName = pulumi.String(pathMap.DefaultBackendAddressPoolName)
		}
		if pathMap.DefaultBackendHttpSettingsName != "" {
			args.DefaultBackendHttpSettingsName = pulumi.String(pathMap.DefaultBackendHttpSettingsName)
		}
		if pathMap.DefaultRedirectConfigurationName != "" {
			args.DefaultRedirectConfigurationName = pulumi.String(pathMap.DefaultRedirectConfigurationName)
		}
		if pathMap.DefaultRewriteRuleSetName != "" {
			args.DefaultRewriteRuleSetName = pulumi.String(pathMap.DefaultRewriteRuleSetName)
		}
		pathRules := make(network.ApplicationGatewayUrlPathMapPathRuleArray, 0, len(pathMap.PathRules))
		for _, pathRule := range pathMap.PathRules {
			ruleArgs := &network.ApplicationGatewayUrlPathMapPathRuleArgs{
				Name:  pulumi.String(pathRule.Name),
				Paths: pulumi.ToStringArray(pathRule.Paths),
			}
			if pathRule.BackendAddressPoolName != "" {
				ruleArgs.BackendAddressPoolName = pulumi.String(pathRule.BackendAddressPoolName)
			}
			if pathRule.BackendHttpSettingsName != "" {
				ruleArgs.BackendHttpSettingsName = pulumi.String(pathRule.BackendHttpSettingsName)
			}
			if pathRule.RedirectConfigurationName != "" {
				ruleArgs.RedirectConfigurationName = pulumi.String(pathRule.RedirectConfigurationName)
			}
			if pathRule.RewriteRuleSetName != "" {
				ruleArgs.RewriteRuleSetName = pulumi.String(pathRule.RewriteRuleSetName)
			}
			if pathRule.FirewallPolicyId.GetValue() != "" {
				ruleArgs.FirewallPolicyId = pulumi.String(pathRule.FirewallPolicyId.GetValue())
			}
			pathRules = append(pathRules, ruleArgs)
		}
		args.PathRules = pathRules
		out = append(out, args)
	}
	return out
}

func buildProbes(probes []*azureapplicationgatewayv1alpha1.AzureApplicationGatewayProbe) network.ApplicationGatewayProbeArray {
	out := make(network.ApplicationGatewayProbeArray, 0, len(probes))
	for _, probe := range probes {
		args := &network.ApplicationGatewayProbeArgs{
			Name:                                pulumi.String(probe.Name),
			Protocol:                            pulumi.String(protocolStrings[probe.Protocol]),
			Interval:                            pulumi.Int(int(probe.Interval)),
			Timeout:                             pulumi.Int(int(probe.Timeout)),
			UnhealthyThreshold:                  pulumi.Int(int(probe.UnhealthyThreshold)),
			PickHostNameFromBackendHttpSettings: pulumi.Bool(probe.PickHostNameFromBackendHttpSettings),
			ProxyProtocolHeaderEnabled:          pulumi.Bool(probe.ProxyProtocolHeaderEnabled),
		}
		if probe.Host != "" {
			args.Host = pulumi.String(probe.Host)
		}
		if probe.Path != "" {
			args.Path = pulumi.String(probe.Path)
		}
		if probe.Port != nil {
			args.Port = pulumi.Int(int(probe.GetPort()))
		}
		if probe.MinimumServers != nil {
			args.MinimumServers = pulumi.Int(int(probe.GetMinimumServers()))
		}
		if probe.Match != nil {
			matchArgs := &network.ApplicationGatewayProbeMatchArgs{
				StatusCodes: pulumi.ToStringArray(probe.Match.StatusCodes),
			}
			if probe.Match.Body != "" {
				matchArgs.Body = pulumi.String(probe.Match.Body)
			}
			args.Match = matchArgs
		}
		out = append(out, args)
	}
	return out
}

func buildSslCertificates(certificates []*azureapplicationgatewayv1alpha1.AzureApplicationGatewaySslCertificate) network.ApplicationGatewaySslCertificateArray {
	out := make(network.ApplicationGatewaySslCertificateArray, 0, len(certificates))
	for _, certificate := range certificates {
		args := &network.ApplicationGatewaySslCertificateArgs{
			Name: pulumi.String(certificate.Name),
		}
		if certificate.KeyVaultSecretId.GetValue() != "" {
			args.KeyVaultSecretId = pulumi.String(certificate.KeyVaultSecretId.GetValue())
		}
		if certificate.Data != "" {
			args.Data = pulumi.String(certificate.Data)
		}
		if certificate.Password != "" {
			args.Password = pulumi.String(certificate.Password)
		}
		out = append(out, args)
	}
	return out
}

func buildTrustedRootCertificates(certificates []*azureapplicationgatewayv1alpha1.AzureApplicationGatewayTrustedRootCertificate) network.ApplicationGatewayTrustedRootCertificateArray {
	out := make(network.ApplicationGatewayTrustedRootCertificateArray, 0, len(certificates))
	for _, certificate := range certificates {
		args := &network.ApplicationGatewayTrustedRootCertificateArgs{
			Name: pulumi.String(certificate.Name),
		}
		if certificate.KeyVaultSecretId.GetValue() != "" {
			args.KeyVaultSecretId = pulumi.String(certificate.KeyVaultSecretId.GetValue())
		}
		if certificate.Data != "" {
			args.Data = pulumi.String(certificate.Data)
		}
		out = append(out, args)
	}
	return out
}

func buildTrustedClientCertificates(certificates []*azureapplicationgatewayv1alpha1.AzureApplicationGatewayTrustedClientCertificate) network.ApplicationGatewayTrustedClientCertificateArray {
	out := make(network.ApplicationGatewayTrustedClientCertificateArray, 0, len(certificates))
	for _, certificate := range certificates {
		out = append(out, &network.ApplicationGatewayTrustedClientCertificateArgs{
			Name: pulumi.String(certificate.Name),
			Data: pulumi.String(certificate.Data),
		})
	}
	return out
}

func buildSslProfiles(profiles []*azureapplicationgatewayv1alpha1.AzureApplicationGatewaySslProfile) network.ApplicationGatewaySslProfileArray {
	out := make(network.ApplicationGatewaySslProfileArray, 0, len(profiles))
	for _, profile := range profiles {
		args := &network.ApplicationGatewaySslProfileArgs{
			Name:                            pulumi.String(profile.Name),
			VerifyClientCertificateIssuerDn: pulumi.Bool(profile.VerifyClientCertificateIssuerDn),
		}
		if len(profile.TrustedClientCertificateNames) > 0 {
			args.TrustedClientCertificateNames = pulumi.ToStringArray(profile.TrustedClientCertificateNames)
		}
		if profile.VerifyClientCertificateRevocation == azureapplicationgatewayv1alpha1.AzureApplicationGatewayClientRevocationCheck_OCSP {
			args.VerifyClientCertificateRevocation = pulumi.String("OCSP")
		}
		if profile.SslPolicy != nil {
			policyArgs := &network.ApplicationGatewaySslProfileSslPolicyArgs{}
			if profile.SslPolicy.PolicyType != azureapplicationgatewayv1alpha1.AzureApplicationGatewaySslPolicyType_azure_application_gateway_ssl_policy_type_unspecified {
				policyArgs.PolicyType = pulumi.String(sslPolicyTypeStrings[profile.SslPolicy.PolicyType])
			}
			if profile.SslPolicy.PolicyName != "" {
				policyArgs.PolicyName = pulumi.String(profile.SslPolicy.PolicyName)
			}
			if profile.SslPolicy.MinProtocolVersion != azureapplicationgatewayv1alpha1.AzureApplicationGatewayTlsProtocol_azure_application_gateway_tls_protocol_unspecified {
				policyArgs.MinProtocolVersion = pulumi.String(tlsProtocolStrings[profile.SslPolicy.MinProtocolVersion])
			}
			if len(profile.SslPolicy.CipherSuites) > 0 {
				policyArgs.CipherSuites = pulumi.ToStringArray(profile.SslPolicy.CipherSuites)
			}
			if len(profile.SslPolicy.DisabledProtocols) > 0 {
				disabled := make(pulumi.StringArray, 0, len(profile.SslPolicy.DisabledProtocols))
				for _, protocol := range profile.SslPolicy.DisabledProtocols {
					disabled = append(disabled, pulumi.String(tlsProtocolStrings[protocol]))
				}
				policyArgs.DisabledProtocols = disabled
			}
			args.SslPolicy = policyArgs
		}
		out = append(out, args)
	}
	return out
}

func buildGlobalSslPolicy(policy *azureapplicationgatewayv1alpha1.AzureApplicationGatewaySslPolicy) *network.ApplicationGatewaySslPolicyArgs {
	args := &network.ApplicationGatewaySslPolicyArgs{}
	if policy.PolicyType != azureapplicationgatewayv1alpha1.AzureApplicationGatewaySslPolicyType_azure_application_gateway_ssl_policy_type_unspecified {
		args.PolicyType = pulumi.String(sslPolicyTypeStrings[policy.PolicyType])
	}
	if policy.PolicyName != "" {
		args.PolicyName = pulumi.String(policy.PolicyName)
	}
	if policy.MinProtocolVersion != azureapplicationgatewayv1alpha1.AzureApplicationGatewayTlsProtocol_azure_application_gateway_tls_protocol_unspecified {
		args.MinProtocolVersion = pulumi.String(tlsProtocolStrings[policy.MinProtocolVersion])
	}
	if len(policy.CipherSuites) > 0 {
		args.CipherSuites = pulumi.ToStringArray(policy.CipherSuites)
	}
	if len(policy.DisabledProtocols) > 0 {
		disabled := make(pulumi.StringArray, 0, len(policy.DisabledProtocols))
		for _, protocol := range policy.DisabledProtocols {
			disabled = append(disabled, pulumi.String(tlsProtocolStrings[protocol]))
		}
		args.DisabledProtocols = disabled
	}
	return args
}

func buildRedirectConfigurations(redirects []*azureapplicationgatewayv1alpha1.AzureApplicationGatewayRedirectConfiguration) network.ApplicationGatewayRedirectConfigurationArray {
	out := make(network.ApplicationGatewayRedirectConfigurationArray, 0, len(redirects))
	for _, redirect := range redirects {
		args := &network.ApplicationGatewayRedirectConfigurationArgs{
			Name:               pulumi.String(redirect.Name),
			RedirectType:       pulumi.String(redirectTypeStrings[redirect.RedirectType]),
			IncludePath:        pulumi.Bool(redirect.IncludePath),
			IncludeQueryString: pulumi.Bool(redirect.IncludeQueryString),
		}
		if redirect.TargetListenerName != "" {
			args.TargetListenerName = pulumi.String(redirect.TargetListenerName)
		}
		if redirect.TargetUrl != "" {
			args.TargetUrl = pulumi.String(redirect.TargetUrl)
		}
		out = append(out, args)
	}
	return out
}

func buildRewriteRuleSets(ruleSets []*azureapplicationgatewayv1alpha1.AzureApplicationGatewayRewriteRuleSet) network.ApplicationGatewayRewriteRuleSetArray {
	out := make(network.ApplicationGatewayRewriteRuleSetArray, 0, len(ruleSets))
	for _, ruleSet := range ruleSets {
		rules := make(network.ApplicationGatewayRewriteRuleSetRewriteRuleArray, 0, len(ruleSet.RewriteRules))
		for _, rule := range ruleSet.RewriteRules {
			ruleArgs := &network.ApplicationGatewayRewriteRuleSetRewriteRuleArgs{
				Name:         pulumi.String(rule.Name),
				RuleSequence: pulumi.Int(int(rule.RuleSequence)),
			}
			if len(rule.Conditions) > 0 {
				conditions := make(network.ApplicationGatewayRewriteRuleSetRewriteRuleConditionArray, 0, len(rule.Conditions))
				for _, condition := range rule.Conditions {
					conditions = append(conditions, &network.ApplicationGatewayRewriteRuleSetRewriteRuleConditionArgs{
						Variable:   pulumi.String(condition.Variable),
						Pattern:    pulumi.String(condition.Pattern),
						IgnoreCase: pulumi.Bool(condition.IgnoreCase),
						Negate:     pulumi.Bool(condition.Negate),
					})
				}
				ruleArgs.Conditions = conditions
			}
			if len(rule.RequestHeaderConfigurations) > 0 {
				headers := make(network.ApplicationGatewayRewriteRuleSetRewriteRuleRequestHeaderConfigurationArray, 0, len(rule.RequestHeaderConfigurations))
				for _, header := range rule.RequestHeaderConfigurations {
					headers = append(headers, &network.ApplicationGatewayRewriteRuleSetRewriteRuleRequestHeaderConfigurationArgs{
						HeaderName:  pulumi.String(header.HeaderName),
						HeaderValue: pulumi.String(header.HeaderValue),
					})
				}
				ruleArgs.RequestHeaderConfigurations = headers
			}
			if len(rule.ResponseHeaderConfigurations) > 0 {
				headers := make(network.ApplicationGatewayRewriteRuleSetRewriteRuleResponseHeaderConfigurationArray, 0, len(rule.ResponseHeaderConfigurations))
				for _, header := range rule.ResponseHeaderConfigurations {
					headers = append(headers, &network.ApplicationGatewayRewriteRuleSetRewriteRuleResponseHeaderConfigurationArgs{
						HeaderName:  pulumi.String(header.HeaderName),
						HeaderValue: pulumi.String(header.HeaderValue),
					})
				}
				ruleArgs.ResponseHeaderConfigurations = headers
			}
			if rule.Url != nil {
				urlArgs := &network.ApplicationGatewayRewriteRuleSetRewriteRuleUrlArgs{
					Reroute: pulumi.Bool(rule.Url.Reroute),
				}
				if rule.Url.Path != "" {
					urlArgs.Path = pulumi.String(rule.Url.Path)
				}
				if rule.Url.QueryString != "" {
					urlArgs.QueryString = pulumi.String(rule.Url.QueryString)
				}
				if rule.Url.Components != azureapplicationgatewayv1alpha1.AzureApplicationGatewayRewriteRuleUrlComponent_azure_application_gateway_rewrite_rule_url_component_unspecified {
					urlArgs.Components = pulumi.String(urlComponentStrings[rule.Url.Components])
				}
				ruleArgs.Url = urlArgs
			}
			rules = append(rules, ruleArgs)
		}
		out = append(out, &network.ApplicationGatewayRewriteRuleSetArgs{
			Name:         pulumi.String(ruleSet.Name),
			RewriteRules: rules,
		})
	}
	return out
}

func buildLayer4Listeners(listeners []*azureapplicationgatewayv1alpha1.AzureApplicationGatewayLayer4Listener) network.ApplicationGatewayListenerArray {
	out := make(network.ApplicationGatewayListenerArray, 0, len(listeners))
	for _, listener := range listeners {
		args := &network.ApplicationGatewayListenerArgs{
			Name:                        pulumi.String(listener.Name),
			FrontendIpConfigurationName: pulumi.String(listener.FrontendIpConfigurationName),
			FrontendPortName:            pulumi.String(listener.FrontendPortName),
			Protocol:                    pulumi.String(protocolStrings[listener.Protocol]),
		}
		if len(listener.HostNames) > 0 {
			args.HostNames = pulumi.ToStringArray(listener.HostNames)
		}
		if listener.SslCertificateName != "" {
			args.SslCertificateName = pulumi.String(listener.SslCertificateName)
		}
		if listener.SslProfileName != "" {
			args.SslProfileName = pulumi.String(listener.SslProfileName)
		}
		out = append(out, args)
	}
	return out
}

func buildLayer4Backends(backends []*azureapplicationgatewayv1alpha1.AzureApplicationGatewayLayer4BackendSettings) network.ApplicationGatewayBackendArray {
	out := make(network.ApplicationGatewayBackendArray, 0, len(backends))
	for _, backend := range backends {
		args := &network.ApplicationGatewayBackendArgs{
			Name:                        pulumi.String(backend.Name),
			Port:                        pulumi.Int(int(backend.Port)),
			Protocol:                    pulumi.String(protocolStrings[backend.Protocol]),
			ClientIpPreservationEnabled: pulumi.Bool(backend.ClientIpPreservationEnabled),
		}
		if backend.HostName != "" {
			args.HostName = pulumi.String(backend.HostName)
		}
		if backend.ProbeName != "" {
			args.ProbeName = pulumi.String(backend.ProbeName)
		}
		// Presence-guarded: unset falls back to Azure's default (30s).
		if backend.TimeoutInSeconds != nil {
			args.TimeoutInSeconds = pulumi.Int(int(backend.GetTimeoutInSeconds()))
		} else {
			args.TimeoutInSeconds = pulumi.Int(30)
		}
		if len(backend.TrustedRootCertificateNames) > 0 {
			args.TrustedRootCertificateNames = pulumi.ToStringArray(backend.TrustedRootCertificateNames)
		}
		out = append(out, args)
	}
	return out
}

func buildLayer4RoutingRules(rules []*azureapplicationgatewayv1alpha1.AzureApplicationGatewayLayer4RoutingRule) network.ApplicationGatewayRoutingRuleArray {
	out := make(network.ApplicationGatewayRoutingRuleArray, 0, len(rules))
	for _, rule := range rules {
		out = append(out, &network.ApplicationGatewayRoutingRuleArgs{
			Name:                   pulumi.String(rule.Name),
			ListenerName:           pulumi.String(rule.ListenerName),
			BackendAddressPoolName: pulumi.String(rule.BackendAddressPoolName),
			BackendName:            pulumi.String(rule.BackendSettingsName),
			Priority:               pulumi.Int(int(rule.Priority)),
		})
	}
	return out
}

func buildCustomErrorConfigurations(configs []*azureapplicationgatewayv1alpha1.AzureApplicationGatewayCustomErrorConfiguration) network.ApplicationGatewayCustomErrorConfigurationArray {
	out := make(network.ApplicationGatewayCustomErrorConfigurationArray, 0, len(configs))
	for _, config := range configs {
		out = append(out, &network.ApplicationGatewayCustomErrorConfigurationArgs{
			StatusCode:         pulumi.String(statusCodeStrings[config.StatusCode]),
			CustomErrorPageUrl: pulumi.String(config.CustomErrorPageUrl),
		})
	}
	return out
}

func buildPrivateLinkConfigurations(configs []*azureapplicationgatewayv1alpha1.AzureApplicationGatewayPrivateLinkConfiguration) network.ApplicationGatewayPrivateLinkConfigurationArray {
	out := make(network.ApplicationGatewayPrivateLinkConfigurationArray, 0, len(configs))
	for _, config := range configs {
		ipConfigs := make(network.ApplicationGatewayPrivateLinkConfigurationIpConfigurationArray, 0, len(config.IpConfigurations))
		for _, ipConfig := range config.IpConfigurations {
			ipArgs := &network.ApplicationGatewayPrivateLinkConfigurationIpConfigurationArgs{
				Name:                       pulumi.String(ipConfig.Name),
				SubnetId:                   pulumi.String(ipConfig.SubnetId.GetValue()),
				PrivateIpAddressAllocation: pulumi.String(ipAllocationStrings[ipConfig.PrivateIpAddressAllocation]),
				Primary:                    pulumi.Bool(ipConfig.Primary),
			}
			if ipConfig.PrivateIpAddress != "" {
				ipArgs.PrivateIpAddress = pulumi.String(ipConfig.PrivateIpAddress)
			}
			ipConfigs = append(ipConfigs, ipArgs)
		}
		out = append(out, &network.ApplicationGatewayPrivateLinkConfigurationArgs{
			Name:             pulumi.String(config.Name),
			IpConfigurations: ipConfigs,
		})
	}
	return out
}
