package module

import (
	"github.com/pkg/errors"
	azurefrontdoorsecretv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefrontdoorsecret/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cdn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurefrontdoorsecretv1alpha1.AzureFrontDoorSecretIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureFrontDoorSecret.Spec

	// The secret is fully immutable (Azure exposes no update), so every
	// change replaces it. A VERSIONLESS Key Vault certificate id (no
	// trailing version segment) makes Front Door follow the
	// certificate's latest version -- rotation happens in Key Vault, not
	// by touching this resource; a versioned id pins one exact version.
	//
	// Operational prerequisite: Front Door's own service principal (the
	// Microsoft.AzureFrontDoor-Cdn enterprise application) must hold
	// read access on the vault's certificates/secrets before this
	// deploys -- a one-time grant per tenant/vault.
	createdSecret, err := cdn.NewFrontdoorSecret(ctx,
		spec.SecretName,
		&cdn.FrontdoorSecretArgs{
			Name:                  pulumi.String(spec.SecretName),
			CdnFrontdoorProfileId: pulumi.String(locals.ProfileId),
			Secret: &cdn.FrontdoorSecretSecretArgs{
				CustomerCertificates: cdn.FrontdoorSecretSecretCustomerCertificateArray{
					&cdn.FrontdoorSecretSecretCustomerCertificateArgs{
						KeyVaultCertificateId: pulumi.String(locals.KeyVaultCertificateId),
					},
				},
			},
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create front door secret %s", spec.SecretName)
	}

	// Export outputs. secret_id is what
	// AzureFrontDoorCustomDomain's tls.secret_id references; the SANs
	// are read back from the wrapped certificate so operators can
	// confirm coverage of a domain's hostname.
	ctx.Export(OpSecretId, createdSecret.ID())
	ctx.Export(OpSecretName, createdSecret.Name)
	ctx.Export(OpSubjectAlternativeNames, createdSecret.Secret.CustomerCertificates().Index(pulumi.Int(0)).SubjectAlternativeNames())

	return nil
}
