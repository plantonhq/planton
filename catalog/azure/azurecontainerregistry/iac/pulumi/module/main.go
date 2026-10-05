package module

import (
	"sort"

	"github.com/pkg/errors"
	azurecontainerregistryv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecontainerregistry/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/containerservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurecontainerregistryv1alpha1.AzureContainerRegistryIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureContainerRegistry.Spec

	// Lifecycle notes worth knowing before operating this resource:
	// - Name and region are the registry's identity -- changing either
	//   replaces it and its CONTENTS DO NOT MIGRATE; every image would
	//   need re-pushing. Zone redundancy and CMK encryption are likewise
	//   fixed at creation.
	// - The SKU changes in place, but downgrading requires every
	//   Premium-only feature (geo-replication, network rules, policies,
	//   CMK) to be unset first -- the same ordering ARM enforces.
	registryArgs := &containerservice.RegistryArgs{
		Name:              pulumi.String(spec.RegistryName),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Location:          pulumi.String(spec.Region),
		Sku:               pulumi.String(locals.Sku),
		AdminEnabled:      pulumi.Bool(spec.AdminUserEnabled),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Presence-guarded true-default optional bools: an absent spec value
	// explicitly falls back to the proto/Azure default so iac-input
	// paths that bypass the manifest loader deploy identically to the
	// Terraform module's optional(bool, true) defaults.
	if spec.PublicNetworkAccessEnabled != nil {
		registryArgs.PublicNetworkAccessEnabled = pulumi.Bool(spec.GetPublicNetworkAccessEnabled())
	} else {
		registryArgs.PublicNetworkAccessEnabled = pulumi.Bool(true)
	}
	if spec.ExportPolicyEnabled != nil {
		registryArgs.ExportPolicyEnabled = pulumi.Bool(spec.GetExportPolicyEnabled())
	} else {
		registryArgs.ExportPolicyEnabled = pulumi.Bool(true)
	}

	registryArgs.ZoneRedundancyEnabled = pulumi.Bool(spec.ZoneRedundancyEnabled)
	registryArgs.AnonymousPullEnabled = pulumi.Bool(spec.AnonymousPullEnabled)
	registryArgs.DataEndpointEnabled = pulumi.Bool(spec.DataEndpointEnabled)
	registryArgs.QuarantinePolicyEnabled = pulumi.Bool(spec.QuarantinePolicyEnabled)

	// Unset keeps untagged manifests forever (Azure's default); a value
	// (including 0 = purge immediately) turns the retention policy on.
	if spec.RetentionPolicyInDays != nil {
		registryArgs.RetentionPolicyInDays = pulumi.Int(int(spec.GetRetentionPolicyInDays()))
	}

	// Only an explicit bypass choice is sent, so an unspecified spec and
	// Azure's default (AzureServices) deploy identically on both engines.
	if locals.NetworkRuleBypassOption != "" {
		registryArgs.NetworkRuleBypassOption = pulumi.String(locals.NetworkRuleBypassOption)
	}

	// The public-registry allowlist (Premium). ARM only supports Allow
	// rules, so the action is constant and only the CIDR ranges vary.
	if spec.NetworkRuleSet != nil {
		defaultAction := "Allow"
		if spec.NetworkRuleSet.DefaultAction == azurecontainerregistryv1alpha1.AzureContainerRegistryNetworkRuleDefaultAction_DENY {
			defaultAction = "Deny"
		}
		ipRules := containerservice.RegistryNetworkRuleSetIpRuleArray{}
		for _, ipRule := range spec.NetworkRuleSet.IpRules {
			ipRules = append(ipRules, containerservice.RegistryNetworkRuleSetIpRuleArgs{
				Action:  pulumi.String("Allow"),
				IpRange: pulumi.String(ipRule.IpRange),
			})
		}
		registryArgs.NetworkRuleSet = containerservice.RegistryNetworkRuleSetArgs{
			DefaultAction: pulumi.String(defaultAction),
			IpRules:       ipRules,
		}
	}

	// Additional regions the registry replicates to (Premium). azurerm
	// expects alphabetical location order; sorting here keeps manifests
	// order-insensitive instead of surfacing ARM's quirk to users.
	if len(spec.Georeplications) > 0 {
		sorted := make([]*azurecontainerregistryv1alpha1.AzureContainerRegistryGeoreplication, len(spec.Georeplications))
		copy(sorted, spec.Georeplications)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].Location < sorted[j].Location })

		georeplications := containerservice.RegistryGeoreplicationArray{}
		for _, replication := range sorted {
			georeplications = append(georeplications, containerservice.RegistryGeoreplicationArgs{
				Location:              pulumi.String(replication.Location),
				ZoneRedundancyEnabled: pulumi.Bool(replication.ZoneRedundancyEnabled),
				// The classic pulumi-azure v6 SDK is bridged from azurerm 4.x,
				// whose argument name is regional_endpoint_enabled; it carries
				// the same ARM property (RegionEndpointEnabled) that azurerm 5
				// renamed to global_endpoint_routing_enabled.
				RegionalEndpointEnabled: pulumi.Bool(replication.GlobalEndpointRoutingEnabled),
				Tags:                    pulumi.ToStringMap(replication.Tags),
			})
		}
		registryArgs.Georeplications = georeplications
	}

	// The registry's managed identity; a USER_ASSIGNED identity is what
	// unwraps the CMK encryption key at boot.
	if locals.IdentityType != "" {
		registryArgs.Identity = containerservice.RegistryIdentityArgs{
			Type:        pulumi.String(locals.IdentityType),
			IdentityIds: pulumi.ToStringArray(locals.IdentityIds),
		}
	}

	// Customer-managed-key encryption (Premium; fixed at creation). The
	// identity_client_id must belong to an identity listed in the identity
	// block that holds get/wrapKey/unwrapKey on the key's vault.
	if spec.Encryption != nil {
		registryArgs.Encryption = containerservice.RegistryEncryptionArgs{
			IdentityClientId: pulumi.String(spec.Encryption.IdentityClientId.GetValue()),
			KeyVaultKeyId:    pulumi.String(spec.Encryption.KeyVaultKeyId.GetValue()),
		}
	}

	createdRegistry, err := containerservice.NewRegistry(ctx,
		spec.RegistryName,
		registryArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create container registry %s", spec.RegistryName)
	}

	// Export outputs from the created resource. The admin
	// credentials and the system-assigned principal id are empty strings
	// when their features are off, matching the Terraform module's try()
	// fallbacks.
	ctx.Export(OpContainerRegistryId, createdRegistry.ID())
	ctx.Export(OpContainerRegistryName, createdRegistry.Name)
	ctx.Export(OpLoginServer, createdRegistry.LoginServer)
	ctx.Export(OpAdminUsername, createdRegistry.AdminUsername)
	ctx.Export(OpAdminPassword, pulumi.ToSecret(createdRegistry.AdminPassword))
	ctx.Export(OpSystemAssignedIdentityPrincipalId, createdRegistry.Identity.ApplyT(func(identity *containerservice.RegistryIdentity) string {
		if identity == nil || identity.PrincipalId == nil {
			return ""
		}
		return *identity.PrincipalId
	}).(pulumi.StringOutput))
	ctx.Export(OpDataEndpointHostNames, createdRegistry.DataEndpointHostNames)

	return nil
}
