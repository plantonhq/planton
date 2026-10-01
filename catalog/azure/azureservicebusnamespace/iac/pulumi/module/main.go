package module

import (
	"github.com/pkg/errors"
	azureservicebusnamespacev1alpha1 "github.com/plantonhq/planton/catalog/azure/azureservicebusnamespace/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/servicebus"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureservicebusnamespacev1alpha1.AzureServiceBusNamespaceStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureServiceBusNamespace.Spec

	// The namespace is the container and billing boundary for enterprise
	// messaging. Queues, topics, subscriptions, SAS rules, and the geo-DR
	// pairing are all first-class kinds that reference this namespace's ARM
	// id -- nothing is bundled here.
	namespaceArgs := &servicebus.NamespaceArgs{
		// ForceNew: the name is the public DNS identity
		// ({name}.servicebus.windows.net); changing it replaces the namespace
		// and every entity in it.
		Name:              pulumi.String(spec.NamespaceName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Sku:               pulumi.String(skuStrings[spec.Sku]),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// PREMIUM pairings: the spec's CELs mirror Azure's create-time contract
	// (capacity {1,2,4,8,16} and partitions {1,2,4} required on PREMIUM,
	// forbidden otherwise), so the module sends them only when present.
	// Partitions are ForceNew -- the layout is fixed at creation.
	if spec.Capacity != nil {
		namespaceArgs.Capacity = pulumi.IntPtr(int(spec.GetCapacity()))
	}
	if spec.PremiumMessagingPartitions != nil {
		namespaceArgs.PremiumMessagingPartitions = pulumi.IntPtr(int(spec.GetPremiumMessagingPartitions()))
	}

	// False = keyless posture: every SAS rule's keys (including the root
	// rule surfaced in this module's outputs) stop being usable credentials.
	// Presence-guarded to Azure's default (true): stack inputs built from a
	// manifest materialize proto defaults, but direct paths do not.
	namespaceArgs.LocalAuthEnabled = pulumi.Bool(presenceGuardedBool(spec.LocalAuthEnabled, true))
	namespaceArgs.PublicNetworkAccessEnabled = pulumi.Bool(presenceGuardedBool(spec.PublicNetworkAccessEnabled, true))

	// Managed identity -- required for CMK (the unwrapping identity must be
	// attached here), usable anywhere the namespace itself authenticates to
	// other services.
	if spec.Identity != nil {
		identityArgs := &servicebus.NamespaceIdentityArgs{
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

	// Customer-managed-key encryption (PREMIUM only). Azure cannot remove
	// CMK once set -- dropping this block forces namespace replacement (the
	// provider encodes that as a ForceNew diff).
	if spec.CustomerManagedKey != nil {
		cmkArgs := &servicebus.NamespaceCustomerManagedKeyTypeArgs{
			KeyVaultKeyId: pulumi.String(spec.CustomerManagedKey.KeyVaultKeyId.GetValue()),
			IdentityId:    pulumi.String(spec.CustomerManagedKey.UserAssignedIdentityId.GetValue()),
		}
		if spec.CustomerManagedKey.InfrastructureEncryptionEnabled != nil {
			cmkArgs.InfrastructureEncryptionEnabled = pulumi.BoolPtr(spec.CustomerManagedKey.GetInfrastructureEncryptionEnabled())
		}
		namespaceArgs.CustomerManagedKey = cmkArgs
	}

	// The namespace firewall (PREMIUM only). The provider models it as an
	// inline block riding a SEPARATE ARM operation after the namespace
	// create; Azure rejects DENY with no admitted sources (front-loaded as a
	// spec CEL), and the block-level public-access dial must agree with the
	// namespace-level one (validated server-side).
	if spec.NetworkRuleSet != nil {
		ruleSetArgs := &servicebus.NamespaceNetworkRuleSetArgs{
			DefaultAction:              pulumi.String(networkDefaultActionStrings[spec.NetworkRuleSet.DefaultAction]),
			PublicNetworkAccessEnabled: pulumi.Bool(presenceGuardedBool(spec.NetworkRuleSet.PublicNetworkAccessEnabled, true)),
		}
		if spec.NetworkRuleSet.TrustedServicesAllowed != nil {
			ruleSetArgs.TrustedServicesAllowed = pulumi.BoolPtr(spec.NetworkRuleSet.GetTrustedServicesAllowed())
		}
		if len(spec.NetworkRuleSet.IpRules) > 0 {
			ruleSetArgs.IpRules = pulumi.ToStringArray(spec.NetworkRuleSet.IpRules)
		}
		if len(spec.NetworkRuleSet.NetworkRules) > 0 {
			networkRules := servicebus.NamespaceNetworkRuleSetNetworkRuleArray{}
			for _, networkRule := range spec.NetworkRuleSet.NetworkRules {
				ruleArgs := &servicebus.NamespaceNetworkRuleSetNetworkRuleArgs{
					SubnetId: pulumi.String(networkRule.SubnetId.GetValue()),
				}
				if networkRule.IgnoreMissingVnetServiceEndpoint != nil {
					ruleArgs.IgnoreMissingVnetServiceEndpoint = pulumi.BoolPtr(networkRule.GetIgnoreMissingVnetServiceEndpoint())
				}
				networkRules = append(networkRules, ruleArgs)
			}
			ruleSetArgs.NetworkRules = networkRules
		}
		namespaceArgs.NetworkRuleSet = ruleSetArgs
	}

	createdNamespace, err := servicebus.NewNamespace(ctx,
		spec.NamespaceName,
		namespaceArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Service Bus namespace %s", spec.NamespaceName)
	}

	// Export stack outputs. The root SAS rule's four credential faces are
	// quick-start/break-glass credentials; production workloads mint
	// least-privilege rules with AzureServiceBusAuthorizationRule or go
	// keyless (local_auth_enabled false).
	ctx.Export(OpNamespaceId, createdNamespace.ID())
	ctx.Export(OpNamespaceName, createdNamespace.Name)
	ctx.Export(OpEndpoint, createdNamespace.Endpoint)
	ctx.Export(OpDefaultPrimaryConnectionString, pulumi.ToSecret(createdNamespace.DefaultPrimaryConnectionString))
	ctx.Export(OpDefaultSecondaryConnectionString, pulumi.ToSecret(createdNamespace.DefaultSecondaryConnectionString))
	ctx.Export(OpDefaultPrimaryKey, pulumi.ToSecret(createdNamespace.DefaultPrimaryKey))
	ctx.Export(OpDefaultSecondaryKey, pulumi.ToSecret(createdNamespace.DefaultSecondaryKey))
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
