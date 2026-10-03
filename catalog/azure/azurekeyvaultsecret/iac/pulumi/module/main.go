package module

import (
	"github.com/pkg/errors"
	azurekeyvaultsecretv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurekeyvaultsecret/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/keyvault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurekeyvaultsecretv1alpha1.AzureKeyVaultSecretIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureKeyVaultSecret.Spec

	// A Key Vault secret is a data-plane object: the provider talks to
	// the vault's {name}.vault.azure.net endpoint, not ARM -- which is
	// why creation fails with a 403 when the deploying credential lacks
	// data-plane secret permissions on the vault, even if it owns the
	// subscription.
	//
	// The value arrives already reference-resolved (the spec's
	// sensitive StringValueOrRef); the provider's write-only value_wo
	// variant is deliberately not wired -- it duplicates `value` for a
	// plaintext-in-config problem this module does not have. Changing
	// the value creates a NEW secret version; the versioned outputs
	// move with it, versionless_id keeps resolving to latest.
	secretArgs := &keyvault.SecretArgs{
		Name:       pulumi.String(spec.Name),
		KeyVaultId: pulumi.String(locals.KeyVaultId),
		Value:      pulumi.String(locals.Value),
		Tags:       pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.ContentType != "" {
		secretArgs.ContentType = pulumi.String(spec.ContentType)
	}

	// Advisory attributes -- Key Vault stores and returns them;
	// enforcing them is the consumer's job.
	if spec.NotBeforeDate != nil {
		secretArgs.NotBeforeDate = pulumi.String(spec.GetNotBeforeDate())
	}
	if spec.ExpirationDate != nil {
		secretArgs.ExpirationDate = pulumi.String(spec.GetExpirationDate())
	}

	createdSecret, err := keyvault.NewSecret(ctx,
		spec.Name,
		secretArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create key vault secret %s", spec.Name)
	}

	// Export outputs from the created resource. The secret's
	// VALUE is deliberately never an output -- consumers read it from
	// the vault at runtime via these identifiers.
	ctx.Export(OpSecretId, createdSecret.ID())
	ctx.Export(OpVersionlessId, createdSecret.VersionlessId)
	ctx.Export(OpSecretName, createdSecret.Name)
	ctx.Export(OpVersion, createdSecret.Version)
	ctx.Export(OpResourceId, createdSecret.ResourceId)
	ctx.Export(OpResourceVersionlessId, createdSecret.ResourceVersionlessId)

	return nil
}
