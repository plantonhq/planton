package module

import (
	"strings"

	azurekeyvaultkeyv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurekeyvaultkey/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureKeyVaultKey *azurekeyvaultkeyv1alpha1.AzureKeyVaultKey

	// KeyVaultId is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal ARM ID.
	KeyVaultId string

	// KeyType and Curve are Azure's exact strings for the spec's enums
	// (Azure is case-sensitive about all of them; the _HSM variants are
	// hyphenated on the wire). Curve stays empty when unspecified so Azure
	// applies its own default (P-256) -- identical behavior on both
	// engines.
	KeyType string
	Curve   string

	// KeyOpts are Azure's camelCase operation strings for the spec's enum
	// list.
	KeyOpts []string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

// The vocabularies below translate the spec's enums to the exact strings
// Azure's key API expects -- the Terraform module carries the same maps in
// locals.tf. A missing entry would silently drop a capability, so each map
// is exhaustive over its enum by construction.

var keyTypeStrings = map[azurekeyvaultkeyv1alpha1.AzureKeyVaultKeyType]string{
	azurekeyvaultkeyv1alpha1.AzureKeyVaultKeyType_RSA:     "RSA",
	azurekeyvaultkeyv1alpha1.AzureKeyVaultKeyType_RSA_HSM: "RSA-HSM",
	azurekeyvaultkeyv1alpha1.AzureKeyVaultKeyType_EC:      "EC",
	azurekeyvaultkeyv1alpha1.AzureKeyVaultKeyType_EC_HSM:  "EC-HSM",
}

var curveStrings = map[azurekeyvaultkeyv1alpha1.AzureKeyVaultKeyCurve]string{
	azurekeyvaultkeyv1alpha1.AzureKeyVaultKeyCurve_P_256:  "P-256",
	azurekeyvaultkeyv1alpha1.AzureKeyVaultKeyCurve_P_256K: "P-256K",
	azurekeyvaultkeyv1alpha1.AzureKeyVaultKeyCurve_P_384:  "P-384",
	azurekeyvaultkeyv1alpha1.AzureKeyVaultKeyCurve_P_521:  "P-521",
}

var keyOperationStrings = map[azurekeyvaultkeyv1alpha1.AzureKeyVaultKeyOperation]string{
	azurekeyvaultkeyv1alpha1.AzureKeyVaultKeyOperation_DECRYPT:    "decrypt",
	azurekeyvaultkeyv1alpha1.AzureKeyVaultKeyOperation_ENCRYPT:    "encrypt",
	azurekeyvaultkeyv1alpha1.AzureKeyVaultKeyOperation_SIGN:       "sign",
	azurekeyvaultkeyv1alpha1.AzureKeyVaultKeyOperation_UNWRAP_KEY: "unwrapKey",
	azurekeyvaultkeyv1alpha1.AzureKeyVaultKeyOperation_VERIFY:     "verify",
	azurekeyvaultkeyv1alpha1.AzureKeyVaultKeyOperation_WRAP_KEY:   "wrapKey",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurekeyvaultkeyv1alpha1.AzureKeyVaultKeyIacInput) *Locals {
	locals := &Locals{}

	locals.AzureKeyVaultKey = iacInput.Target
	target := iacInput.Target

	locals.KeyVaultId = target.Spec.KeyVaultId.GetValue()

	locals.KeyType = keyTypeStrings[target.Spec.KeyType]
	locals.Curve = curveStrings[target.Spec.Curve]

	locals.KeyOpts = make([]string, 0, len(target.Spec.KeyOpts))
	for _, operation := range target.Spec.KeyOpts {
		locals.KeyOpts = append(locals.KeyOpts, keyOperationStrings[operation])
	}

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureKeyVaultKey.String()),
	}

	if target.Metadata.Id != "" {
		locals.AzureTags[azuretagkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.AzureTags[azuretagkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.AzureTags[azuretagkeys.Environment] = target.Metadata.Env
	}

	for k, v := range target.Spec.Tags {
		locals.AzureTags[k] = v
	}

	return locals
}
