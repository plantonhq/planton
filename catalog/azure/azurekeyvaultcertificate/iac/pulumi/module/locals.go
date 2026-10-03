package module

import (
	"strings"

	azurekeyvaultcertificatev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurekeyvaultcertificate/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureKeyVaultCertificate *azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificate

	// KeyVaultId is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal ARM ID.
	KeyVaultId string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

// The vocabularies below translate the spec's enums to the exact strings
// Azure's certificate API expects (case-sensitive about all of them -- note
// the lowerCamel key-usage extensions vs the UpperCamel action types). The
// Terraform module carries the same maps in locals.tf. A missing entry would
// silently drop a policy setting, so each map is exhaustive over its enum by
// construction.

var certificateKeyTypeStrings = map[azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyType]string{
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyType_RSA:     "RSA",
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyType_RSA_HSM: "RSA-HSM",
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyType_EC:      "EC",
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyType_EC_HSM:  "EC-HSM",
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyType_OCT:     "oct",
}

var certificateCurveStrings = map[azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyCurve]string{
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyCurve_P_256:  "P-256",
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyCurve_P_256K: "P-256K",
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyCurve_P_384:  "P-384",
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyCurve_P_521:  "P-521",
}

var lifetimeActionTypeStrings = map[azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateLifetimeActionType]string{
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateLifetimeActionType_AUTO_RENEW:     "AutoRenew",
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateLifetimeActionType_EMAIL_CONTACTS: "EmailContacts",
}

// The secret face's media type: what consumers reading the certificate's
// secret get back.
var contentTypeStrings = map[azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateContentType]string{
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateContentType_PKCS12: "application/x-pkcs12",
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateContentType_PEM:    "application/x-pem-file",
}

var keyUsageStrings = map[azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyUsage]string{
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyUsage_CRL_SIGN:          "cRLSign",
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyUsage_DATA_ENCIPHERMENT: "dataEncipherment",
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyUsage_DECIPHER_ONLY:     "decipherOnly",
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyUsage_DIGITAL_SIGNATURE: "digitalSignature",
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyUsage_ENCIPHER_ONLY:     "encipherOnly",
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyUsage_KEY_AGREEMENT:     "keyAgreement",
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyUsage_KEY_CERT_SIGN:     "keyCertSign",
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyUsage_KEY_ENCIPHERMENT:  "keyEncipherment",
	azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateKeyUsage_NON_REPUDIATION:   "nonRepudiation",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurekeyvaultcertificatev1alpha1.AzureKeyVaultCertificateIacInput) *Locals {
	locals := &Locals{}

	locals.AzureKeyVaultCertificate = iacInput.Target
	target := iacInput.Target

	locals.KeyVaultId = target.Spec.KeyVaultId.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureKeyVaultCertificate.String()),
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
