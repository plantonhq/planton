package module

import (
	azurefrontdoorcustomdomainv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefrontdoorcustomdomain/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureFrontDoorCustomDomain *azurefrontdoorcustomdomainv1alpha1.AzureFrontDoorCustomDomain
	ProfileId                  string
}

// certificateTypeStrings maps the certificate-provenance enum to ARM's
// values (unspecified deploys ManagedCertificate, Azure's default).
var certificateTypeStrings = map[azurefrontdoorcustomdomainv1alpha1.AzureFrontDoorCustomDomainCertificateType]string{
	azurefrontdoorcustomdomainv1alpha1.AzureFrontDoorCustomDomainCertificateType_MANAGED_CERTIFICATE:  "ManagedCertificate",
	azurefrontdoorcustomdomainv1alpha1.AzureFrontDoorCustomDomainCertificateType_CUSTOMER_CERTIFICATE: "CustomerCertificate",
}

// cipherSuiteSetTypeStrings maps the cipher-suite policy enum to ARM's
// values. The year-versioned names are ARM's own vocabulary.
var cipherSuiteSetTypeStrings = map[azurefrontdoorcustomdomainv1alpha1.AzureFrontDoorCustomDomainCipherSuiteSetType]string{
	azurefrontdoorcustomdomainv1alpha1.AzureFrontDoorCustomDomainCipherSuiteSetType_TLS12_2022: "TLS12_2022",
	azurefrontdoorcustomdomainv1alpha1.AzureFrontDoorCustomDomainCipherSuiteSetType_TLS12_2023: "TLS12_2023",
	azurefrontdoorcustomdomainv1alpha1.AzureFrontDoorCustomDomainCipherSuiteSetType_CUSTOMIZED: "Customized",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurefrontdoorcustomdomainv1alpha1.AzureFrontDoorCustomDomainIacInput) *Locals {
	locals := &Locals{}

	locals.AzureFrontDoorCustomDomain = iacInput.Target
	locals.ProfileId = iacInput.Target.Spec.ProfileId.GetValue()

	// No Azure tags: ARM does not support tags on Front Door custom
	// domains, so the platform's identity tags live on the profile.

	return locals
}
