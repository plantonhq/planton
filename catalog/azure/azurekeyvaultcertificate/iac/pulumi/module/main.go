package module

import (
	"github.com/pkg/errors"
	azurekeyvaultcertificatev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurekeyvaultcertificate/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/keyvault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureKeyVaultCertificate.Spec

	// A Key Vault certificate is a data-plane object: the provider talks
	// to the vault's {name}.vault.azure.net endpoint, not ARM -- which is
	// why creation fails with a 403 when the deploying credential lacks
	// data-plane certificate permissions on the vault, even if it owns the
	// subscription.
	//
	// Enrollment against issuer "Self" completes synchronously; a CA
	// issuer keeps the operation pending until the CA responds, so expect
	// longer creates on integrated-CA policies.
	certificateArgs := &keyvault.CertificateArgs{
		Name:       pulumi.String(spec.Name),
		KeyVaultId: pulumi.String(locals.KeyVaultId),
		Tags:       pulumi.ToStringMap(locals.AzureTags),
	}

	// Import path: bring an existing PFX/PEM bundle (the contents carry
	// the private key). Changing the contents imports a NEW VERSION of the
	// certificate rather than replacing the object.
	if spec.Certificate != nil {
		importArgs := &keyvault.CertificateCertificateArgs{
			Contents: pulumi.String(spec.Certificate.Contents),
		}
		if spec.Certificate.Password != nil {
			importArgs.Password = pulumi.String(spec.Certificate.GetPassword())
		}
		certificateArgs.Certificate = importArgs
	}

	// Generate path (and governance for imports that carry an explicit
	// policy). Everything except lifetime_actions creates a new
	// certificate version when changed; lifetime_actions update in place.
	if spec.CertificatePolicy != nil {
		policy := spec.CertificatePolicy

		// The private key's shape. key_size stays nil for EC keys so
		// Azure derives it from the curve -- identical behavior on both
		// engines.
		keyProperties := &keyvault.CertificateCertificatePolicyKeyPropertiesArgs{
			Exportable: pulumi.Bool(policy.KeyProperties.Exportable),
			KeyType:    pulumi.String(certificateKeyTypeStrings[policy.KeyProperties.KeyType]),
			ReuseKey:   pulumi.Bool(policy.KeyProperties.ReuseKey),
		}
		if policy.KeyProperties.KeySize != nil {
			keyProperties.KeySize = pulumi.Int(int(policy.KeyProperties.GetKeySize()))
		}
		if curve := certificateCurveStrings[policy.KeyProperties.Curve]; curve != "" {
			keyProperties.Curve = pulumi.String(curve)
		}

		policyArgs := &keyvault.CertificateCertificatePolicyArgs{
			IssuerParameters: &keyvault.CertificateCertificatePolicyIssuerParametersArgs{
				Name: pulumi.String(policy.IssuerName),
			},
			KeyProperties: keyProperties,
			SecretProperties: &keyvault.CertificateCertificatePolicySecretPropertiesArgs{
				ContentType: pulumi.String(contentTypeStrings[policy.SecretProperties.ContentType]),
			},
		}

		// Renewal/notification actions as expiry approaches. Exactly one
		// trigger field per action (spec validation enforces Azure's
		// contract).
		if len(policy.LifetimeActions) > 0 {
			lifetimeActions := make(keyvault.CertificateCertificatePolicyLifetimeActionArray, 0, len(policy.LifetimeActions))
			for _, action := range policy.LifetimeActions {
				trigger := &keyvault.CertificateCertificatePolicyLifetimeActionTriggerArgs{}
				if action.Trigger.DaysBeforeExpiry != nil {
					trigger.DaysBeforeExpiry = pulumi.Int(int(action.Trigger.GetDaysBeforeExpiry()))
				}
				if action.Trigger.LifetimePercentage != nil {
					trigger.LifetimePercentage = pulumi.Int(int(action.Trigger.GetLifetimePercentage()))
				}
				lifetimeActions = append(lifetimeActions, &keyvault.CertificateCertificatePolicyLifetimeActionArgs{
					Action: &keyvault.CertificateCertificatePolicyLifetimeActionActionArgs{
						ActionType: pulumi.String(lifetimeActionTypeStrings[action.ActionType]),
					},
					Trigger: trigger,
				})
			}
			policyArgs.LifetimeActions = lifetimeActions
		}

		// X.509 content -- present only when the vault generates the
		// certificate (spec validation requires it then; imports derive
		// it from the bundle).
		if policy.X509CertificateProperties != nil {
			x509 := policy.X509CertificateProperties

			keyUsages := make([]string, 0, len(x509.KeyUsage))
			for _, usage := range x509.KeyUsage {
				keyUsages = append(keyUsages, keyUsageStrings[usage])
			}

			x509Args := &keyvault.CertificateCertificatePolicyX509CertificatePropertiesArgs{
				Subject:          pulumi.String(x509.Subject),
				KeyUsages:        pulumi.ToStringArray(keyUsages),
				ValidityInMonths: pulumi.Int(int(x509.ValidityInMonths)),
			}
			if len(x509.ExtendedKeyUsage) > 0 {
				x509Args.ExtendedKeyUsages = pulumi.ToStringArray(x509.ExtendedKeyUsage)
			}
			if sans := x509.SubjectAlternativeNames; sans != nil {
				sanArgs := &keyvault.CertificateCertificatePolicyX509CertificatePropertiesSubjectAlternativeNamesArgs{}
				if len(sans.DnsNames) > 0 {
					sanArgs.DnsNames = pulumi.ToStringArray(sans.DnsNames)
				}
				if len(sans.Emails) > 0 {
					sanArgs.Emails = pulumi.ToStringArray(sans.Emails)
				}
				if len(sans.Upns) > 0 {
					sanArgs.Upns = pulumi.ToStringArray(sans.Upns)
				}
				x509Args.SubjectAlternativeNames = sanArgs
			}
			policyArgs.X509CertificateProperties = x509Args
		}

		certificateArgs.CertificatePolicy = policyArgs
	}

	createdCertificate, err := keyvault.NewCertificate(ctx,
		spec.Name,
		certificateArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create key vault certificate %s", spec.Name)
	}

	// Export outputs from the created resource.
	ctx.Export(OpCertificateId, createdCertificate.ID())
	ctx.Export(OpVersionlessId, createdCertificate.VersionlessId)
	ctx.Export(OpSecretId, createdCertificate.SecretId)
	ctx.Export(OpVersionlessSecretId, createdCertificate.VersionlessSecretId)
	ctx.Export(OpCertificateName, createdCertificate.Name)
	ctx.Export(OpVersion, createdCertificate.Version)
	ctx.Export(OpThumbprint, createdCertificate.Thumbprint)
	ctx.Export(OpResourceManagerId, createdCertificate.ResourceManagerId)
	ctx.Export(OpResourceManagerVersionlessId, createdCertificate.ResourceManagerVersionlessId)

	return nil
}
