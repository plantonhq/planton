package module

import (
	"github.com/pkg/errors"
	azurekeyvaultkeyv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurekeyvaultkey/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/keyvault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurekeyvaultkeyv1alpha1.AzureKeyVaultKeyIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureKeyVaultKey.Spec

	// A Key Vault key is a data-plane object: the provider talks to the
	// vault's {name}.vault.azure.net endpoint, not ARM -- which is why
	// creation fails with a 403 when the deploying credential lacks
	// data-plane key permissions on the vault, even if it owns the
	// subscription.
	//
	// Type, size, and curve are fixed at creation -- Azure key material is
	// immutable by design; changing any of them replaces the key (and
	// every consumer re-encrypts through the new key on its next unwrap).
	keyArgs := &keyvault.KeyArgs{
		Name:       pulumi.String(spec.Name),
		KeyVaultId: pulumi.String(locals.KeyVaultId),
		KeyType:    pulumi.String(locals.KeyType),
		// The capability boundary: Azure rejects any operation not listed.
		KeyOpts: pulumi.ToStringArray(locals.KeyOpts),
		Tags:    pulumi.ToStringMap(locals.AzureTags),
	}

	// key_size for RSA, curve for EC (spec validation enforces the
	// pairing); the unset side stays nil so Azure applies its own defaults
	// -- identical behavior on both engines.
	if spec.KeySize != nil {
		keyArgs.KeySize = pulumi.Int(int(spec.GetKeySize()))
	}
	if locals.Curve != "" {
		keyArgs.Curve = pulumi.String(locals.Curve)
	}

	if spec.NotBeforeDate != nil {
		keyArgs.NotBeforeDate = pulumi.String(spec.GetNotBeforeDate())
	}
	if spec.ExpirationDate != nil {
		keyArgs.ExpirationDate = pulumi.String(spec.GetExpirationDate())
	}

	// Rotation policy rides the key resource itself (Azure models it as a
	// sub-resource updated in place). expire_after stamps an expiry on
	// every NEW version; the automatic block is what actually rotates.
	if spec.RotationPolicy != nil {
		rotationPolicy := &keyvault.KeyRotationPolicyArgs{}
		if spec.RotationPolicy.ExpireAfter != nil {
			rotationPolicy.ExpireAfter = pulumi.String(spec.RotationPolicy.GetExpireAfter())
		}
		if spec.RotationPolicy.NotifyBeforeExpiry != nil {
			rotationPolicy.NotifyBeforeExpiry = pulumi.String(spec.RotationPolicy.GetNotifyBeforeExpiry())
		}
		if spec.RotationPolicy.Automatic != nil {
			automatic := &keyvault.KeyRotationPolicyAutomaticArgs{}
			if spec.RotationPolicy.Automatic.TimeAfterCreation != nil {
				automatic.TimeAfterCreation = pulumi.String(spec.RotationPolicy.Automatic.GetTimeAfterCreation())
			}
			if spec.RotationPolicy.Automatic.TimeBeforeExpiry != nil {
				automatic.TimeBeforeExpiry = pulumi.String(spec.RotationPolicy.Automatic.GetTimeBeforeExpiry())
			}
			rotationPolicy.Automatic = automatic
		}
		keyArgs.RotationPolicy = rotationPolicy
	}

	createdKey, err := keyvault.NewKey(ctx,
		spec.Name,
		keyArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create key vault key %s", spec.Name)
	}

	// Export outputs from the created resource.
	ctx.Export(OpKeyId, createdKey.ID())
	ctx.Export(OpVersionlessId, createdKey.VersionlessId)
	ctx.Export(OpKeyName, createdKey.Name)
	ctx.Export(OpVersion, createdKey.Version)
	ctx.Export(OpResourceId, createdKey.ResourceId)
	ctx.Export(OpResourceVersionlessId, createdKey.ResourceVersionlessId)
	ctx.Export(OpPublicKeyPem, createdKey.PublicKeyPem)
	ctx.Export(OpPublicKeyOpenssh, createdKey.PublicKeyOpenssh)

	return nil
}
