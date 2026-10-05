package module

import (
	"github.com/pkg/errors"
	azurecontainerappenvironmentcertificatev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecontainerappenvironmentcertificate/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/containerapp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources stores a bring-your-own TLS certificate on a Container App
// Environment -- either an inline PFX upload or a Key Vault reference the
// environment keeps current across renewals. The certificate is shared by
// every app in the environment; AzureContainerAppCustomDomain bindings
// reference it by the certificate_id output.
//
// Lifecycle notes worth knowing before operating this resource:
//   - Only tags update in place; every other change replaces the
//     certificate (and briefly re-binds any custom domain using it).
//   - Azure never returns the PFX blob on reads, so blob drift is
//     invisible -- rotating an inline certificate means updating the spec.
//   - The Key Vault path requires the environment's managed identity (the
//     one named in certificate_key_vault.identity) to already hold read
//     access to the vault's secrets; Azure checks it at deploy time.
func Resources(ctx *pulumi.Context, iacInput *azurecontainerappenvironmentcertificatev1alpha1.AzureContainerAppEnvironmentCertificateIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureContainerAppEnvironmentCertificate.Spec

	certificateArgs := &containerapp.EnvironmentCertificateArgs{
		Name:                      pulumi.String(spec.CertificateName),
		ContainerAppEnvironmentId: pulumi.String(locals.EnvironmentId),
		Tags:                      pulumi.ToStringMap(locals.AzureTags),
	}

	// Spec validation guarantees exactly one source: the inline PFX (with
	// its possibly-empty password -- passwordless PFX bundles are legal,
	// and Azure expects the password argument alongside the blob either
	// way) or the Key Vault reference. The unused arguments stay nil so
	// the provider's own exactly-one-of contract is satisfied.
	if spec.CertificateBlobBase64 != "" {
		certificateArgs.CertificateBlobBase64 = pulumi.String(spec.CertificateBlobBase64)
		certificateArgs.CertificatePassword = pulumi.String(spec.CertificatePassword)
	}

	if spec.CertificateKeyVault != nil {
		keyVaultArgs := &containerapp.EnvironmentCertificateCertificateKeyVaultArgs{
			KeyVaultSecretId: pulumi.String(spec.CertificateKeyVault.KeyVaultSecretId.GetValue()),
		}
		// Unset deploys "System" -- Azure's own default identity for the
		// vault read; the explicit fallback keeps both engines identical
		// on iac-input paths.
		if spec.CertificateKeyVault.Identity != nil {
			keyVaultArgs.Identity = pulumi.String(spec.CertificateKeyVault.Identity.GetValue())
		} else {
			keyVaultArgs.Identity = pulumi.String("System")
		}
		certificateArgs.CertificateKeyVault = keyVaultArgs
	}

	createdCertificate, err := containerapp.NewEnvironmentCertificate(ctx,
		spec.CertificateName,
		certificateArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create container app environment certificate %s", spec.CertificateName)
	}

	// certificate_id is the binding seam AzureContainerAppCustomDomain
	// consumes; the certificate facts support expiry monitoring for
	// manually-rotated inline certificates.
	ctx.Export(OpCertificateId, createdCertificate.ID())
	ctx.Export(OpSubjectName, createdCertificate.SubjectName)
	ctx.Export(OpIssuer, createdCertificate.Issuer)
	ctx.Export(OpIssueDate, createdCertificate.IssueDate)
	ctx.Export(OpExpirationDate, createdCertificate.ExpirationDate)
	ctx.Export(OpThumbprint, createdCertificate.Thumbprint)

	return nil
}
