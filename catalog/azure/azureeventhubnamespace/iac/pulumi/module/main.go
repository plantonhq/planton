package module

import (
	"github.com/pkg/errors"
	azureeventhubnamespacev1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventhubnamespace/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/eventhub"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureeventhubnamespacev1alpha1.AzureEventHubNamespaceStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureEventHubNamespace.Spec

	// The namespace is the container and billing boundary for
	// high-throughput event streaming. Hubs, consumer groups, SAS rules,
	// schema groups, the geo-DR pairing, and CMK encryption are all
	// first-class kinds that reference this namespace's ARM id -- nothing
	// is bundled here.
	namespaceArgs := &eventhub.EventHubNamespaceArgs{
		// ForceNew: the name is the public DNS identity
		// ({name}.servicebus.windows.net -- Event Hubs shares the Service
		// Bus DNS zone) and the Kafka bootstrap host; changing it replaces
		// the namespace and every entity in it.
		Name:              pulumi.String(spec.NamespaceName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		// BASIC <-> STANDARD updates in place; moving into or out of
		// PREMIUM replaces the namespace (Azure cannot convert across the
		// reserved/multi-tenant boundary).
		Sku:  pulumi.String(skuStrings[spec.Sku]),
		Tags: pulumi.ToStringMap(locals.AzureTags),
	}

	// Throughput units (BASIC/STANDARD) or processing units (PREMIUM).
	// Sent only when present so Azure's default (1) applies otherwise.
	if spec.Capacity != nil {
		namespaceArgs.Capacity = pulumi.IntPtr(int(spec.GetCapacity()))
	}

	// STANDARD's elastic scaling: Azure grows TUs up to the ceiling under
	// load but never shrinks them back. Azure validates the ceiling/enable
	// pairing at apply time. Presence-guarded: stack inputs built from a
	// manifest materialize proto defaults, but direct paths do not.
	if spec.AutoInflateEnabled != nil {
		namespaceArgs.AutoInflateEnabled = pulumi.BoolPtr(spec.GetAutoInflateEnabled())
	}
	if spec.MaximumThroughputUnits != nil {
		namespaceArgs.MaximumThroughputUnits = pulumi.IntPtr(int(spec.GetMaximumThroughputUnits()))
	}

	// ForceNew: a namespace cannot move on or off a dedicated cluster in
	// place. Placement buys single-tenant capacity, 1024-partition hubs,
	// 90-day retention, and CMK eligibility.
	if spec.DedicatedClusterId.GetValue() != "" {
		namespaceArgs.DedicatedClusterId = pulumi.String(spec.DedicatedClusterId.GetValue())
	}

	// Managed identity -- required for identity-based capture auth and for
	// CMK (the unwrapping identity must be attached here), usable anywhere
	// the namespace itself authenticates to other services.
	if spec.Identity != nil {
		identityArgs := &eventhub.EventHubNamespaceIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(spec.Identity.UserAssignedIdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.UserAssignedIdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		namespaceArgs.Identity = identityArgs
	}

	// False = keyless posture: every SAS rule's keys (including the root
	// rule surfaced in this module's outputs) stop being usable credentials.
	// Presence-guarded to Azure's default (true).
	namespaceArgs.LocalAuthenticationEnabled = pulumi.Bool(presenceGuardedBool(spec.LocalAuthenticationEnabled, true))
	namespaceArgs.PublicNetworkAccessEnabled = pulumi.Bool(presenceGuardedBool(spec.PublicNetworkAccessEnabled, true))

	// The namespace firewall (not available on BASIC -- front-loaded as a
	// spec CEL). The provider models it as an inline block riding a
	// SEPARATE ARM operation after the namespace create; the block-level
	// public-access dial must agree with the namespace-level one (Azure
	// validates the pair server-side; the spec CEL front-loads it).
	if spec.NetworkRuleSets != nil {
		ruleSetArgs := &eventhub.EventHubNamespaceNetworkRulesetsArgs{
			DefaultAction:              pulumi.String(networkDefaultActionStrings[spec.NetworkRuleSets.DefaultAction]),
			PublicNetworkAccessEnabled: pulumi.Bool(presenceGuardedBool(spec.NetworkRuleSets.PublicNetworkAccessEnabled, true)),
		}
		if spec.NetworkRuleSets.TrustedServiceAccessEnabled != nil {
			ruleSetArgs.TrustedServiceAccessEnabled = pulumi.BoolPtr(spec.NetworkRuleSets.GetTrustedServiceAccessEnabled())
		}
		if len(spec.NetworkRuleSets.IpRules) > 0 {
			ipRules := eventhub.EventHubNamespaceNetworkRulesetsIpRuleArray{}
			for _, ipMask := range spec.NetworkRuleSets.IpRules {
				// Each entry is an allow rule: Azure's per-rule action
				// accepts exactly one value (Allow), so the spec models
				// just the mask.
				ipRules = append(ipRules, &eventhub.EventHubNamespaceNetworkRulesetsIpRuleArgs{
					IpMask: pulumi.String(ipMask),
					Action: pulumi.String("Allow"),
				})
			}
			ruleSetArgs.IpRules = ipRules
		}
		if len(spec.NetworkRuleSets.VirtualNetworkRules) > 0 {
			vnetRules := eventhub.EventHubNamespaceNetworkRulesetsVirtualNetworkRuleArray{}
			for _, vnetRule := range spec.NetworkRuleSets.VirtualNetworkRules {
				ruleArgs := &eventhub.EventHubNamespaceNetworkRulesetsVirtualNetworkRuleArgs{
					SubnetId: pulumi.String(vnetRule.SubnetId.GetValue()),
				}
				if vnetRule.IgnoreMissingVirtualNetworkServiceEndpoint != nil {
					ruleArgs.IgnoreMissingVirtualNetworkServiceEndpoint = pulumi.BoolPtr(vnetRule.GetIgnoreMissingVirtualNetworkServiceEndpoint())
				}
				vnetRules = append(vnetRules, ruleArgs)
			}
			ruleSetArgs.VirtualNetworkRules = vnetRules
		}
		namespaceArgs.NetworkRulesets = ruleSetArgs
	}

	createdNamespace, err := eventhub.NewEventHubNamespace(ctx,
		spec.NamespaceName,
		namespaceArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Event Hubs namespace %s", spec.NamespaceName)
	}

	// Export stack outputs. The root SAS rule's credential faces are
	// quick-start/break-glass credentials; production workloads mint
	// least-privilege rules with AzureEventHubAuthorizationRule or go
	// keyless (local_authentication_enabled false). The alias faces are
	// only populated when a geo-DR pairing exists.
	ctx.Export(OpNamespaceId, createdNamespace.ID())
	ctx.Export(OpNamespaceName, createdNamespace.Name)
	ctx.Export(OpDefaultPrimaryConnectionString, pulumi.ToSecret(createdNamespace.DefaultPrimaryConnectionString))
	ctx.Export(OpDefaultSecondaryConnectionString, pulumi.ToSecret(createdNamespace.DefaultSecondaryConnectionString))
	ctx.Export(OpDefaultPrimaryKey, pulumi.ToSecret(createdNamespace.DefaultPrimaryKey))
	ctx.Export(OpDefaultSecondaryKey, pulumi.ToSecret(createdNamespace.DefaultSecondaryKey))
	ctx.Export(OpDefaultPrimaryConnectionStringAlias, pulumi.ToSecret(createdNamespace.DefaultPrimaryConnectionStringAlias))
	ctx.Export(OpDefaultSecondaryConnectionStringAlias, pulumi.ToSecret(createdNamespace.DefaultSecondaryConnectionStringAlias))
	// Empty unless SYSTEM_ASSIGNED is enabled -- mirrors the TF module's
	// try(identity[0].principal_id, "").
	ctx.Export(OpIdentityPrincipalId, createdNamespace.Identity.PrincipalId().ApplyT(func(principalId *string) string {
		if principalId == nil {
			return ""
		}
		return *principalId
	}).(pulumi.StringOutput))

	return nil
}

// presenceGuardedBool returns the field's value when set and the proto
// default otherwise -- default materialization is middleware behavior, not a
// wire guarantee.
func presenceGuardedBool(field *bool, defaultValue bool) bool {
	if field == nil {
		return defaultValue
	}
	return *field
}
