package module

import (
	azurefrontdoorsecretv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefrontdoorsecret/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureFrontDoorSecret  *azurefrontdoorsecretv1alpha1.AzureFrontDoorSecret
	ProfileId             string
	KeyVaultCertificateId string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurefrontdoorsecretv1alpha1.AzureFrontDoorSecretIacInput) *Locals {
	locals := &Locals{}

	locals.AzureFrontDoorSecret = iacInput.Target
	locals.ProfileId = iacInput.Target.Spec.ProfileId.GetValue()
	locals.KeyVaultCertificateId = iacInput.Target.Spec.KeyVaultCertificateId.GetValue()

	// No Azure tags: ARM does not support tags on Front Door secrets,
	// so the platform's identity tags live on the profile.

	return locals
}
