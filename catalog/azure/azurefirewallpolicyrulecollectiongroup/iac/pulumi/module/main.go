package module

import (
	"github.com/pkg/errors"
	azurefirewallpolicyrulecollectiongroupv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefirewallpolicyrulecollectiongroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurefirewallpolicyrulecollectiongroupv1alpha1.AzureFirewallPolicyRuleCollectionGroupIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureFirewallPolicyRuleCollectionGroup.Spec

	groupArgs := &network.FirewallPolicyRuleCollectionGroupArgs{
		Name:             pulumi.String(spec.Name),
		FirewallPolicyId: pulumi.String(locals.FirewallPolicyId),
		Priority:         pulumi.Int(int(spec.Priority)),
	}

	// Collections are sent in spec order; Azure evaluates by PRIORITY
	// (not list position), and across types always DNAT -> network ->
	// application regardless of priorities.
	if len(spec.ApplicationRuleCollections) > 0 {
		collections := network.FirewallPolicyRuleCollectionGroupApplicationRuleCollectionArray{}
		for _, collection := range spec.ApplicationRuleCollections {
			rules := network.FirewallPolicyRuleCollectionGroupApplicationRuleCollectionRuleArray{}
			for _, rule := range collection.Rules {
				ruleArgs := &network.FirewallPolicyRuleCollectionGroupApplicationRuleCollectionRuleArgs{
					Name:                 pulumi.String(rule.Name),
					SourceAddresses:      pulumi.ToStringArray(rule.SourceAddresses),
					SourceIpGroups:       pulumi.ToStringArray(refValues(rule.SourceIpGroups)),
					DestinationAddresses: pulumi.ToStringArray(rule.DestinationAddresses),
					DestinationFqdns:     pulumi.ToStringArray(rule.DestinationFqdns),
					DestinationUrls:      pulumi.ToStringArray(rule.DestinationUrls),
					DestinationFqdnTags:  pulumi.ToStringArray(rule.DestinationFqdnTags),
					WebCategories:        pulumi.ToStringArray(rule.WebCategories),
					// terminate_tls decrypts for inspection (Premium +
					// policy TLS certificate); sent explicitly on both
					// engines so the payloads stay identical.
					TerminateTls: pulumi.Bool(rule.TerminateTls),
				}
				if rule.Description != "" {
					ruleArgs.Description = pulumi.String(rule.Description)
				}
				if len(rule.Protocols) > 0 {
					protocols := network.FirewallPolicyRuleCollectionGroupApplicationRuleCollectionRuleProtocolArray{}
					for _, protocol := range rule.Protocols {
						protocols = append(protocols, &network.FirewallPolicyRuleCollectionGroupApplicationRuleCollectionRuleProtocolArgs{
							Type: pulumi.String(applicationProtocolTypeWireValue(protocol.Type)),
							Port: pulumi.Int(int(protocol.Port)),
						})
					}
					ruleArgs.Protocols = protocols
				}
				if len(rule.HttpHeaders) > 0 {
					headers := network.FirewallPolicyRuleCollectionGroupApplicationRuleCollectionRuleHttpHeaderArray{}
					for _, header := range rule.HttpHeaders {
						headers = append(headers, &network.FirewallPolicyRuleCollectionGroupApplicationRuleCollectionRuleHttpHeaderArgs{
							Name:  pulumi.String(header.Name),
							Value: pulumi.String(header.Value),
						})
					}
					ruleArgs.HttpHeaders = headers
				}
				rules = append(rules, ruleArgs)
			}
			collections = append(collections, &network.FirewallPolicyRuleCollectionGroupApplicationRuleCollectionArgs{
				Name:     pulumi.String(collection.Name),
				Priority: pulumi.Int(int(collection.Priority)),
				Action:   pulumi.String(filterActionWireValue(collection.Action)),
				Rules:    rules,
			})
		}
		groupArgs.ApplicationRuleCollections = collections
	}

	if len(spec.NetworkRuleCollections) > 0 {
		collections := network.FirewallPolicyRuleCollectionGroupNetworkRuleCollectionArray{}
		for _, collection := range spec.NetworkRuleCollections {
			rules := network.FirewallPolicyRuleCollectionGroupNetworkRuleCollectionRuleArray{}
			for _, rule := range collection.Rules {
				ruleArgs := &network.FirewallPolicyRuleCollectionGroupNetworkRuleCollectionRuleArgs{
					Name:                 pulumi.String(rule.Name),
					Protocols:            pulumi.ToStringArray(protocolWireValues(rule.Protocols)),
					SourceAddresses:      pulumi.ToStringArray(rule.SourceAddresses),
					SourceIpGroups:       pulumi.ToStringArray(refValues(rule.SourceIpGroups)),
					DestinationAddresses: pulumi.ToStringArray(rule.DestinationAddresses),
					DestinationIpGroups:  pulumi.ToStringArray(refValues(rule.DestinationIpGroups)),
					DestinationFqdns:     pulumi.ToStringArray(rule.DestinationFqdns),
					DestinationPorts:     pulumi.ToStringArray(rule.DestinationPorts),
				}
				if rule.Description != "" {
					ruleArgs.Description = pulumi.String(rule.Description)
				}
				rules = append(rules, ruleArgs)
			}
			collections = append(collections, &network.FirewallPolicyRuleCollectionGroupNetworkRuleCollectionArgs{
				Name:     pulumi.String(collection.Name),
				Priority: pulumi.Int(int(collection.Priority)),
				Action:   pulumi.String(filterActionWireValue(collection.Action)),
				Rules:    rules,
			})
		}
		groupArgs.NetworkRuleCollections = collections
	}

	if len(spec.NatRuleCollections) > 0 {
		collections := network.FirewallPolicyRuleCollectionGroupNatRuleCollectionArray{}
		for _, collection := range spec.NatRuleCollections {
			rules := network.FirewallPolicyRuleCollectionGroupNatRuleCollectionRuleArray{}
			for _, rule := range collection.Rules {
				ruleArgs := &network.FirewallPolicyRuleCollectionGroupNatRuleCollectionRuleArgs{
					Name:            pulumi.String(rule.Name),
					Protocols:       pulumi.ToStringArray(protocolWireValues(rule.Protocols)),
					SourceAddresses: pulumi.ToStringArray(rule.SourceAddresses),
					SourceIpGroups:  pulumi.ToStringArray(refValues(rule.SourceIpGroups)),
					TranslatedPort:  pulumi.Int(int(rule.TranslatedPort)),
				}
				if rule.Description != "" {
					ruleArgs.Description = pulumi.String(rule.Description)
				}
				if rule.DestinationAddress != "" {
					ruleArgs.DestinationAddress = pulumi.String(rule.DestinationAddress)
				}
				// ARM caps DNAT destination ports at ONE entry (a port or
				// range); the provider models it as a singular string.
				if len(rule.DestinationPorts) > 0 {
					ruleArgs.DestinationPorts = pulumi.String(rule.DestinationPorts[0])
				}
				// Exactly one translation target (spec CEL); the unset one
				// is omitted from the payload.
				if rule.TranslatedAddress != "" {
					ruleArgs.TranslatedAddress = pulumi.String(rule.TranslatedAddress)
				}
				if rule.TranslatedFqdn != "" {
					ruleArgs.TranslatedFqdn = pulumi.String(rule.TranslatedFqdn)
				}
				rules = append(rules, ruleArgs)
			}
			collections = append(collections, &network.FirewallPolicyRuleCollectionGroupNatRuleCollectionArgs{
				Name:     pulumi.String(collection.Name),
				Priority: pulumi.Int(int(collection.Priority)),
				// The DNAT action vocabulary has exactly one value -- a
				// constant, not a knob -- so both engines send "Dnat"
				// unconditionally (the provider's own schema literal; ARM
				// normalizes case).
				Action: pulumi.String("Dnat"),
				Rules:  rules,
			})
		}
		groupArgs.NatRuleCollections = collections
	}

	createdGroup, err := network.NewFirewallPolicyRuleCollectionGroup(ctx,
		spec.Name,
		groupArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create firewall policy rule collection group %s", spec.Name)
	}

	ctx.Export(OpRuleCollectionGroupId, createdGroup.ID())
	ctx.Export(OpRuleCollectionGroupName, createdGroup.Name)

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

// protocolWireValues maps a repeated protocol enum to the wire vocabulary.
func protocolWireValues(protocols []azurefirewallpolicyrulecollectiongroupv1alpha1.AzureFirewallPolicyRuleProtocol) []string {
	values := make([]string, 0, len(protocols))
	for _, protocol := range protocols {
		values = append(values, ruleProtocolWireValue(protocol))
	}
	return values
}
