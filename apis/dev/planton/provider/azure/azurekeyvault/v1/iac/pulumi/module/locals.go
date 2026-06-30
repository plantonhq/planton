package module

import (
	"strings"

	azurekeyvaultv1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azurekeyvault/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureKeyVault     *azurekeyvaultv1.AzureKeyVault
	ResourceGroupName string
	VaultName         string
	AzureTags         map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurekeyvaultv1.AzureKeyVaultStackInput) *Locals {
	locals := &Locals{}

	locals.AzureKeyVault = stackInput.Target
	target := stackInput.Target

	// The resource_group field is a StringValueOrRef. The platform middleware resolves
	// valueFrom references before IaC modules run, so .GetValue() always returns the
	// resolved literal string.
	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Create vault name from metadata.name
	// Azure Key Vault names must be 3-24 characters, alphanumeric and hyphens only, globally unique
	// Format: {prefix}-{sanitized-name} where prefix could be "kv" or metadata id
	vaultName := target.Metadata.Name
	// Replace dots and underscores with hyphens for Azure compatibility
	vaultName = strings.ReplaceAll(vaultName, ".", "-")
	vaultName = strings.ReplaceAll(vaultName, "_", "-")
	// Ensure it's not too long (Azure limit is 24 characters)
	if len(vaultName) > 24 {
		vaultName = vaultName[:24]
	}
	locals.VaultName = vaultName

	// Create Azure tags for resource tagging
	locals.AzureTags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AzureKeyVault.String()),
	}

	if target.Metadata.Id != "" {
		locals.AzureTags["resource_id"] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.AzureTags["organization"] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.AzureTags["environment"] = target.Metadata.Env
	}

	return locals
}

// getSku converts proto enum to Azure SDK SKU string
func getSku(sku azurekeyvaultv1.AzureKeyVaultSku) string {
	switch sku {
	case azurekeyvaultv1.AzureKeyVaultSku_PREMIUM:
		return "premium"
	case azurekeyvaultv1.AzureKeyVaultSku_STANDARD, azurekeyvaultv1.AzureKeyVaultSku_SKU_UNSPECIFIED:
		return "standard"
	default:
		return "standard"
	}
}

// getNetworkDefaultAction converts proto enum to Azure SDK string
func getNetworkDefaultAction(action azurekeyvaultv1.AzureKeyVaultNetworkAction) string {
	switch action {
	case azurekeyvaultv1.AzureKeyVaultNetworkAction_ALLOW:
		return "Allow"
	case azurekeyvaultv1.AzureKeyVaultNetworkAction_DENY, azurekeyvaultv1.AzureKeyVaultNetworkAction_ACTION_UNSPECIFIED:
		return "Deny"
	default:
		return "Deny"
	}
}
