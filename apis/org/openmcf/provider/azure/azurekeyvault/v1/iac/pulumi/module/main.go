package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurekeyvaultv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurekeyvault/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/keyvault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurekeyvaultv1.AzureKeyVaultStackInput) error {
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

	// Get the spec from locals
	spec := locals.AzureKeyVault.Spec

	// Build network ACLs configuration
	networkAcls := &keyvault.KeyVaultNetworkAclsArgs{
		DefaultAction: pulumi.String("Deny"), // Default to secure
		Bypass:        pulumi.String("AzureServices"),
	}

	if spec.NetworkAcls != nil {
		networkAcls.DefaultAction = pulumi.String(getNetworkDefaultAction(spec.NetworkAcls.GetDefaultAction()))

		if spec.NetworkAcls.GetBypassAzureServices() {
			networkAcls.Bypass = pulumi.String("AzureServices")
		} else {
			networkAcls.Bypass = pulumi.String("None")
		}

		// Add IP rules
		if len(spec.NetworkAcls.IpRules) > 0 {
			ipRules := make(pulumi.StringArray, 0)
			for _, ipRule := range spec.NetworkAcls.IpRules {
				ipRules = append(ipRules, pulumi.String(ipRule))
			}
			networkAcls.IpRules = ipRules
		}

		// Add VNet rules
		if len(spec.NetworkAcls.VirtualNetworkSubnetIds) > 0 {
			vnetRules := make(pulumi.StringArray, 0)
			for _, subnetId := range spec.NetworkAcls.VirtualNetworkSubnetIds {
				vnetRules = append(vnetRules, pulumi.String(subnetId))
			}
			networkAcls.VirtualNetworkSubnetIds = vnetRules
		}
	}

	// Create the Key Vault
	vault, err := keyvault.NewKeyVault(ctx,
		locals.VaultName,
		&keyvault.KeyVaultArgs{
			Name:              pulumi.String(locals.VaultName),
			Location:          pulumi.String(spec.Region),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			TenantId:          pulumi.String(azureProviderConfig.TenantId),
			SkuName:           pulumi.String(getSku(spec.GetSku())),

			// Security settings
			EnableRbacAuthorization:      pulumi.Bool(spec.GetEnableRbacAuthorization()),
			PurgeProtectionEnabled:       pulumi.Bool(spec.GetEnablePurgeProtection()),
			SoftDeleteRetentionDays:      pulumi.Int(int(spec.GetSoftDeleteRetentionDays())),
			EnabledForDeployment:         pulumi.Bool(false),
			EnabledForDiskEncryption:     pulumi.Bool(false),
			EnabledForTemplateDeployment: pulumi.Bool(false),

			// Network ACLs
			NetworkAcls: networkAcls,

			// Tags
			Tags: pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Key Vault %s", locals.VaultName)
	}

	// Create secrets (placeholder entries - actual values must be set separately)
	secretIdMap := make(map[string]pulumi.StringOutput)

	for _, secretName := range spec.SecretNames {
		// Create empty secret (value must be set separately via Azure SDK/CLI)
		secret, err := keyvault.NewSecret(ctx,
			fmt.Sprintf("secret-%s", secretName),
			&keyvault.SecretArgs{
				Name:       pulumi.String(secretName),
				KeyVaultId: vault.ID(),
				Value:      pulumi.String(""), // Empty placeholder - must be set separately
				Tags:       pulumi.ToStringMap(locals.AzureTags),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(vault))
		if err != nil {
			return errors.Wrapf(err, "failed to create secret %s", secretName)
		}

		secretIdMap[secretName] = secret.ID().ToStringOutput()
	}

	// Export stack outputs
	ctx.Export(OpVaultId, vault.ID())
	ctx.Export(OpVaultName, vault.Name)
	ctx.Export(OpVaultUri, vault.VaultUri)
	ctx.Export(OpRegion, pulumi.String(spec.Region))
	ctx.Export(OpResourceGroup, pulumi.String(locals.ResourceGroupName))

	// Export secret ID map
	if len(secretIdMap) > 0 {
		secretMap := pulumi.StringMap{}
		for name, id := range secretIdMap {
			secretMap[name] = id.ToStringOutput()
		}
		ctx.Export(OpSecretIdMap, secretMap)
	}

	return nil
}
