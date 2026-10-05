package module

import (
	azurecontainerappcustomdomainv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecontainerappcustomdomain/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureContainerAppCustomDomain *azurecontainerappcustomdomainv1alpha1.AzureContainerAppCustomDomain
	ContainerAppId                string
	CertificateId                 string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurecontainerappcustomdomainv1alpha1.AzureContainerAppCustomDomainIacInput) *Locals {
	locals := &Locals{}

	locals.AzureContainerAppCustomDomain = iacInput.Target

	target := iacInput.Target

	// container_app_id and the optional certificate id are
	// StringValueOrRef fields. The platform middleware resolves valueFrom
	// references before IaC modules run, so .GetValue() always returns
	// the resolved literal ARM ID.
	locals.ContainerAppId = target.Spec.ContainerAppId.GetValue()

	if target.Spec.ContainerAppEnvironmentCertificateId != nil {
		locals.CertificateId = target.Spec.ContainerAppEnvironmentCertificateId.GetValue()
	}

	// No tag map: Azure models the binding as an entry in the app's
	// ingress configuration, not a taggable ARM resource.

	return locals
}
