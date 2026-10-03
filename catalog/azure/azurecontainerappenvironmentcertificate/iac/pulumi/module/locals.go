package module

import (
	"strings"

	azurecontainerappenvironmentcertificatev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecontainerappenvironmentcertificate/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureContainerAppEnvironmentCertificate *azurecontainerappenvironmentcertificatev1alpha1.AzureContainerAppEnvironmentCertificate
	EnvironmentId                           string
	AzureTags                               map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurecontainerappenvironmentcertificatev1alpha1.AzureContainerAppEnvironmentCertificateIacInput) *Locals {
	locals := &Locals{}

	locals.AzureContainerAppEnvironmentCertificate = iacInput.Target

	target := iacInput.Target

	// container_app_environment_id is a StringValueOrRef. The platform
	// middleware resolves valueFrom references before IaC modules run, so
	// .GetValue() always returns the resolved literal ARM ID.
	locals.EnvironmentId = target.Spec.ContainerAppEnvironmentId.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureContainerAppEnvironmentCertificate.String()),
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

	for key, value := range target.Spec.Tags {
		locals.AzureTags[key] = value
	}

	return locals
}
