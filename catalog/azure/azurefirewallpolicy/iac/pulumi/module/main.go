package module

import (
	"github.com/pkg/errors"
	azurefirewallpolicyv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefirewallpolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurefirewallpolicyv1alpha1.AzureFirewallPolicyIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureFirewallPolicy.Spec

	policyArgs := &network.FirewallPolicyArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		// The sku and threat-intelligence mode are always sent explicitly
		// (Standard/Alert when unspecified) -- Azure's own defaults, made
		// deterministic so both engines produce identical payloads.
		Sku:                    pulumi.String(skuWireValue(spec.Sku)),
		ThreatIntelligenceMode: pulumi.String(threatIntelModeWireValue(spec.ThreatIntelligenceMode)),
		SqlRedirectAllowed:     pulumi.Bool(spec.SqlRedirectAllowed),
		Tags:                   pulumi.ToStringMap(locals.AzureTags),
	}

	// The base policy makes this one a CHILD: the base's rules and
	// settings apply beneath this policy's own. The tier of both policies
	// must match -- ARM validates the pairing at apply time.
	if spec.BasePolicyId.GetValue() != "" {
		policyArgs.BasePolicyId = pulumi.String(spec.BasePolicyId.GetValue())
	}

	if spec.ThreatIntelligenceAllowlist != nil {
		policyArgs.ThreatIntelligenceAllowlist = &network.FirewallPolicyThreatIntelligenceAllowlistArgs{
			IpAddresses: pulumi.ToStringArray(spec.ThreatIntelligenceAllowlist.IpAddresses),
			Fqdns:       pulumi.ToStringArray(spec.ThreatIntelligenceAllowlist.Fqdns),
		}
	}

	if spec.Dns != nil {
		policyArgs.Dns = &network.FirewallPolicyDnsArgs{
			Servers:      pulumi.ToStringArray(spec.Dns.Servers),
			ProxyEnabled: pulumi.Bool(spec.Dns.ProxyEnabled),
		}
	}

	// IDPS is Premium-only (spec validation front-loads the gate). The
	// engine mode is sent only when specified -- Azure defaults an
	// unspecified mode to Off.
	if spec.IntrusionDetection != nil {
		idpsArgs := &network.FirewallPolicyIntrusionDetectionArgs{
			PrivateRanges: pulumi.ToStringArray(spec.IntrusionDetection.PrivateRanges),
		}

		if mode := idpsStateWireValue(spec.IntrusionDetection.Mode); mode != "" {
			idpsArgs.Mode = pulumi.String(mode)
		}

		if len(spec.IntrusionDetection.SignatureOverrides) > 0 {
			overrides := network.FirewallPolicyIntrusionDetectionSignatureOverrideArray{}
			for _, override := range spec.IntrusionDetection.SignatureOverrides {
				overrideArgs := &network.FirewallPolicyIntrusionDetectionSignatureOverrideArgs{}
				if override.Id != "" {
					overrideArgs.Id = pulumi.String(override.Id)
				}
				if state := idpsStateWireValue(override.State); state != "" {
					overrideArgs.State = pulumi.String(state)
				}
				overrides = append(overrides, overrideArgs)
			}
			idpsArgs.SignatureOverrides = overrides
		}

		if len(spec.IntrusionDetection.TrafficBypass) > 0 {
			bypasses := network.FirewallPolicyIntrusionDetectionTrafficBypassArray{}
			for _, bypass := range spec.IntrusionDetection.TrafficBypass {
				bypasses = append(bypasses, &network.FirewallPolicyIntrusionDetectionTrafficBypassArgs{
					Name:                 pulumi.String(bypass.Name),
					Description:          pulumi.String(bypass.Description),
					Protocol:             pulumi.String(bypassProtocolWireValue(bypass.Protocol)),
					SourceAddresses:      pulumi.ToStringArray(bypass.SourceAddresses),
					SourceIpGroups:       pulumi.ToStringArray(refValues(bypass.SourceIpGroups)),
					DestinationAddresses: pulumi.ToStringArray(bypass.DestinationAddresses),
					DestinationIpGroups:  pulumi.ToStringArray(refValues(bypass.DestinationIpGroups)),
					DestinationPorts:     pulumi.ToStringArray(bypass.DestinationPorts),
				})
			}
			idpsArgs.TrafficBypasses = bypasses
		}

		policyArgs.IntrusionDetection = idpsArgs
	}

	if spec.Identity != nil {
		policyArgs.Identity = &network.FirewallPolicyIdentityArgs{
			Type:        pulumi.String(identityTypeWireValue(spec.Identity.Type)),
			IdentityIds: pulumi.ToStringArray(refValues(spec.Identity.UserAssignedIdentityIds)),
		}
	}

	// TLS inspection is Premium-only (gate front-loaded in the spec). The
	// referenced Key Vault SECRET id defaults to the certificate's
	// versionless secret face so renewals are picked up automatically;
	// the policy's user-assigned identity must be able to read it.
	if spec.TlsCertificate != nil {
		policyArgs.TlsCertificate = &network.FirewallPolicyTlsCertificateArgs{
			KeyVaultSecretId: pulumi.String(spec.TlsCertificate.KeyVaultSecretId.GetValue()),
			Name:             pulumi.String(spec.TlsCertificate.Name),
		}
	}

	if spec.Insights != nil {
		insightsArgs := &network.FirewallPolicyInsightsArgs{
			// enabled is presence-tracked in the spec (an explicit false
			// keeps the workspace wiring with analysis paused); validation
			// guarantees it is set when the block is present.
			Enabled:                        pulumi.Bool(spec.Insights.GetEnabled()),
			DefaultLogAnalyticsWorkspaceId: pulumi.String(spec.Insights.DefaultLogAnalyticsWorkspaceId.GetValue()),
		}

		if spec.Insights.RetentionInDays != nil {
			insightsArgs.RetentionInDays = pulumi.Int(int(spec.Insights.GetRetentionInDays()))
		}

		if len(spec.Insights.LogAnalyticsWorkspaces) > 0 {
			workspaces := network.FirewallPolicyInsightsLogAnalyticsWorkspaceArray{}
			for _, workspace := range spec.Insights.LogAnalyticsWorkspaces {
				workspaces = append(workspaces, &network.FirewallPolicyInsightsLogAnalyticsWorkspaceArgs{
					Id:               pulumi.String(workspace.WorkspaceId.GetValue()),
					FirewallLocation: pulumi.String(workspace.FirewallLocation),
				})
			}
			insightsArgs.LogAnalyticsWorkspaces = workspaces
		}

		policyArgs.Insights = insightsArgs
	}

	if spec.ExplicitProxy != nil {
		proxyArgs := &network.FirewallPolicyExplicitProxyArgs{
			Enabled:       pulumi.Bool(spec.ExplicitProxy.Enabled),
			EnablePacFile: pulumi.Bool(spec.ExplicitProxy.EnablePacFile),
		}
		// Ports are presence-tracked: an unset port is omitted rather than
		// sent as 0 (the provider treats 0 as a value, not absence).
		if spec.ExplicitProxy.HttpPort != nil {
			proxyArgs.HttpPort = pulumi.Int(int(spec.ExplicitProxy.GetHttpPort()))
		}
		if spec.ExplicitProxy.HttpsPort != nil {
			proxyArgs.HttpsPort = pulumi.Int(int(spec.ExplicitProxy.GetHttpsPort()))
		}
		if spec.ExplicitProxy.PacFilePort != nil {
			proxyArgs.PacFilePort = pulumi.Int(int(spec.ExplicitProxy.GetPacFilePort()))
		}
		if spec.ExplicitProxy.PacFile != "" {
			proxyArgs.PacFile = pulumi.String(spec.ExplicitProxy.PacFile)
		}
		policyArgs.ExplicitProxy = proxyArgs
	}

	if len(spec.PrivateIpRanges) > 0 {
		policyArgs.PrivateIpRanges = pulumi.ToStringArray(spec.PrivateIpRanges)
	}

	// Azure only records "Enabled" for auto-learn -- disabling is done by
	// omission on the wire, so the flag is sent only when true (sending
	// an explicit false would still read back as absent and churn state).
	if spec.AutoLearnPrivateRangesEnabled {
		policyArgs.AutoLearnPrivateRangesEnabled = pulumi.Bool(true)
	}

	createdPolicy, err := network.NewFirewallPolicy(ctx,
		spec.Name,
		policyArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create firewall policy %s", spec.Name)
	}

	ctx.Export(OpFirewallPolicyId, createdPolicy.ID())
	ctx.Export(OpFirewallPolicyName, createdPolicy.Name)
	// The system-assigned principal id -- empty when the policy has no
	// system identity. Grant it Key Vault secret read access when TLS
	// inspection rides the system identity.
	ctx.Export(OpIdentityPrincipalId, createdPolicy.Identity.ApplyT(func(identity *network.FirewallPolicyIdentity) string {
		if identity == nil || identity.PrincipalId == nil {
			return ""
		}
		return *identity.PrincipalId
	}).(pulumi.StringOutput))

	return nil
}

// refValues resolves a repeated StringValueOrRef into its literal values
// (the platform middleware resolves valueFrom references before modules
// run, so GetValue() always returns the resolved literal).
func refValues(refs []*foreignkeyv1.StringValueOrRef) []string {
	values := make([]string, 0, len(refs))
	for _, ref := range refs {
		values = append(values, ref.GetValue())
	}
	return values
}
