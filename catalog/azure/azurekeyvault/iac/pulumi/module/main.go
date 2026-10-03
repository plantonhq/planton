package module

import (
	"github.com/pkg/errors"
	azurekeyvaultv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurekeyvault/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/keyvault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurekeyvaultv1alpha1.AzureKeyVaultIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureKeyVault.Spec

	// The vault authenticates against the deploying credential's Azure AD
	// tenant -- a vault cannot be managed cross-tenant, so the tenant is
	// read from the ambient client configuration instead of being modeled
	// as a contradictable spec field.
	clientConfig, err := core.GetClientConfig(ctx, pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrap(err, "failed to read the azure client configuration")
	}

	vaultArgs := &keyvault.KeyVaultArgs{
		Name:              pulumi.String(spec.VaultName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		TenantId:          pulumi.String(clientConfig.TenantId),
		SkuName:           pulumi.String(locals.SkuName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// RBAC is the spec's recommended default; access_policy entries below
	// are only honored by Azure when this is false (ARM stores but ignores
	// policies on an RBAC-mode vault). Presence-guarded: IaC inputs built
	// from a manifest do NOT materialize proto defaults, so an unset field
	// falls back to the spec default (true) explicitly.
	if spec.RbacAuthorizationEnabled != nil {
		vaultArgs.RbacAuthorizationEnabled = pulumi.Bool(spec.GetRbacAuthorizationEnabled())
	} else {
		vaultArgs.RbacAuthorizationEnabled = pulumi.Bool(true)
	}

	// Legacy access-policy grants, declared inline so the vault owns its
	// complete grant list (mixing inline policies with the standalone
	// access-policy resource causes perpetual drift).
	if len(spec.AccessPolicies) > 0 {
		accessPolicies := make(keyvault.KeyVaultAccessPolicyArray, 0, len(spec.AccessPolicies))
		for _, policy := range spec.AccessPolicies {
			policyArgs := &keyvault.KeyVaultAccessPolicyArgs{
				ObjectId:               pulumi.String(policy.ObjectId.GetValue()),
				KeyPermissions:         permissionStrings(policy.KeyPermissions, keyPermissionStrings),
				SecretPermissions:      permissionStrings(policy.SecretPermissions, secretPermissionStrings),
				CertificatePermissions: permissionStrings(policy.CertificatePermissions, certificatePermissionStrings),
				StoragePermissions:     permissionStrings(policy.StoragePermissions, storagePermissionStrings),
			}
			// An unset tenant falls back to the vault's own tenant --
			// access policies cannot span tenants in practice.
			if policy.TenantId != nil {
				policyArgs.TenantId = pulumi.String(policy.GetTenantId())
			} else {
				policyArgs.TenantId = pulumi.String(clientConfig.TenantId)
			}
			if policy.ApplicationId != nil {
				policyArgs.ApplicationId = pulumi.String(policy.GetApplicationId())
			}
			accessPolicies = append(accessPolicies, policyArgs)
		}
		vaultArgs.AccessPolicies = accessPolicies
	}

	// Resource-manager integration switches (Azure defaults: all false).
	vaultArgs.EnabledForDeployment = pulumi.Bool(spec.EnabledForDeployment)
	vaultArgs.EnabledForDiskEncryption = pulumi.Bool(spec.EnabledForDiskEncryption)
	vaultArgs.EnabledForTemplateDeployment = pulumi.Bool(spec.EnabledForTemplateDeployment)

	// Presence-guarded true-default optional: unset falls back to Azure's
	// default (public access on).
	if spec.PublicNetworkAccessEnabled != nil {
		vaultArgs.PublicNetworkAccessEnabled = pulumi.Bool(spec.GetPublicNetworkAccessEnabled())
	} else {
		vaultArgs.PublicNetworkAccessEnabled = pulumi.Bool(true)
	}

	// Purge protection is irreversible once enabled; with it on, destroying
	// the vault schedules deletion for the end of the soft-delete retention
	// window instead of purging (the provider's default behavior purges
	// soft-deleted vaults on destroy when purge protection is off).
	vaultArgs.PurgeProtectionEnabled = pulumi.Bool(spec.PurgeProtectionEnabled)

	// Presence-guarded: unset falls back to Azure's default (90 days) --
	// the same value the Terraform module's optional(number, 90) encodes.
	if spec.SoftDeleteRetentionDays != nil {
		vaultArgs.SoftDeleteRetentionDays = pulumi.Int(int(spec.GetSoftDeleteRetentionDays()))
	} else {
		vaultArgs.SoftDeleteRetentionDays = pulumi.Int(90)
	}

	// Public-endpoint firewall. azurerm requires default_action and bypass
	// whenever the block is present; an unspecified bypass materializes
	// Azure's own default (AzureServices) -- identical on both engines.
	if spec.NetworkAcls != nil {
		networkAcls := &keyvault.KeyVaultNetworkAclsArgs{}

		// default_action is required whenever the block is present (spec
		// validation enforces it), so there is no null fallback to invent.
		if spec.NetworkAcls.DefaultAction == azurekeyvaultv1alpha1.AzureKeyVaultNetworkAclsDefaultAction_ALLOW {
			networkAcls.DefaultAction = pulumi.String("Allow")
		} else {
			networkAcls.DefaultAction = pulumi.String("Deny")
		}

		if spec.NetworkAcls.Bypass == azurekeyvaultv1alpha1.AzureKeyVaultNetworkAclsBypass_NONE {
			networkAcls.Bypass = pulumi.String("None")
		} else {
			networkAcls.Bypass = pulumi.String("AzureServices")
		}

		if len(spec.NetworkAcls.IpRules) > 0 {
			networkAcls.IpRules = pulumi.ToStringArray(spec.NetworkAcls.IpRules)
		}

		if len(spec.NetworkAcls.VirtualNetworkSubnetIds) > 0 {
			subnetIds := make([]string, 0, len(spec.NetworkAcls.VirtualNetworkSubnetIds))
			for _, subnetId := range spec.NetworkAcls.VirtualNetworkSubnetIds {
				subnetIds = append(subnetIds, subnetId.GetValue())
			}
			networkAcls.VirtualNetworkSubnetIds = pulumi.ToStringArray(subnetIds)
		}

		vaultArgs.NetworkAcls = networkAcls
	}

	createdVault, err := keyvault.NewKeyVault(ctx,
		spec.VaultName,
		vaultArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create key vault %s", spec.VaultName)
	}

	// Export outputs from the created resource.
	ctx.Export(OpKeyVaultId, createdVault.ID())
	ctx.Export(OpKeyVaultName, createdVault.Name)
	ctx.Export(OpVaultUri, createdVault.VaultUri)
	ctx.Export(OpTenantId, createdVault.TenantId)
	ctx.Export(OpResourceGroupName, createdVault.ResourceGroupName)

	return nil
}

// permissionStrings translates a list of spec permission enums to the
// data-plane strings Azure expects, through the exhaustive vocabulary maps
// in locals.go.
func permissionStrings[E comparable](permissions []E, vocabulary map[E]string) pulumi.StringArray {
	if len(permissions) == 0 {
		return nil
	}
	out := make(pulumi.StringArray, 0, len(permissions))
	for _, permission := range permissions {
		out = append(out, pulumi.String(vocabulary[permission]))
	}
	return out
}
