package module

import (
	"strconv"

	digitaloceanprovider "github.com/plantonhq/planton/catalog/digitalocean"
	digitaloceandnsrecordv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceandnsrecord/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/digitaloceanlabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds quick references used by other files.
type Locals struct {
	DigitalOceanProviderConfig *digitaloceanprovider.DigitalOceanProviderConfig
	DigitalOceanDnsRecord      *digitaloceandnsrecordv1alpha1.DigitalOceanDnsRecord
	DigitalOceanLabels         map[string]string
}

// initializeLocals sets up local values from IaC input.
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceandnsrecordv1alpha1.DigitalOceanDnsRecordIacInput) *Locals {
	locals := &Locals{}

	locals.DigitalOceanDnsRecord = iacInput.Target

	// Standard Planton labels for DigitalOcean resources.
	locals.DigitalOceanLabels = map[string]string{
		digitaloceanlabelkeys.Resource:     strconv.FormatBool(true),
		digitaloceanlabelkeys.ResourceName: locals.DigitalOceanDnsRecord.Metadata.Name,
		digitaloceanlabelkeys.ResourceKind: catalogkind.CatalogKind_DigitalOceanDnsRecord.String(),
	}

	if locals.DigitalOceanDnsRecord.Metadata.Org != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Organization] = locals.DigitalOceanDnsRecord.Metadata.Org
	}
	if locals.DigitalOceanDnsRecord.Metadata.Env != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Environment] = locals.DigitalOceanDnsRecord.Metadata.Env
	}
	if locals.DigitalOceanDnsRecord.Metadata.Id != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.ResourceId] = locals.DigitalOceanDnsRecord.Metadata.Id
	}

	locals.DigitalOceanProviderConfig = iacInput.ProviderConfig

	return locals
}
