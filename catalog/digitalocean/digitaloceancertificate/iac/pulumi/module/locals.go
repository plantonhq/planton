package module

import (
	"strconv"

	digitaloceanprovider "github.com/plantonhq/planton/catalog/digitalocean"
	digitaloceancertificatev1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceancertificate/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/digitaloceanlabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals aggregates handy references for the rest of the module.
type Locals struct {
	DigitalOceanProviderConfig *digitaloceanprovider.DigitalOceanProviderConfig
	DigitalOceanCertificate    *digitaloceancertificatev1alpha1.DigitalOceanCertificate
	DigitalOceanLabels         map[string]string
}

// initializeLocals copies stack‑input fields into Locals and builds a reusable label map.
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceancertificatev1alpha1.DigitalOceanCertificateIacInput) *Locals {
	var locals Locals

	locals.DigitalOceanCertificate = iacInput.Target

	// Standard Planton labels for DigitalOcean resources.
	locals.DigitalOceanLabels = map[string]string{
		digitaloceanlabelkeys.Resource:     strconv.FormatBool(true),
		digitaloceanlabelkeys.ResourceName: locals.DigitalOceanCertificate.Metadata.Name,
		digitaloceanlabelkeys.ResourceKind: catalogkind.CatalogKind_DigitalOceanCertificate.String(),
	}

	if locals.DigitalOceanCertificate.Metadata.Org != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Organization] = locals.DigitalOceanCertificate.Metadata.Org
	}

	if locals.DigitalOceanCertificate.Metadata.Env != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Environment] = locals.DigitalOceanCertificate.Metadata.Env
	}

	if locals.DigitalOceanCertificate.Metadata.Id != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.ResourceId] = locals.DigitalOceanCertificate.Metadata.Id
	}

	locals.DigitalOceanProviderConfig = iacInput.ProviderConfig

	return &locals
}
