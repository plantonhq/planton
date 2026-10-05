package module

import (
	"strconv"

	digitaloceandropletautoscalepoolv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceandropletautoscalepool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/digitaloceanlabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles common pointers and label maps used across the module.
type Locals struct {
	DigitalOceanDropletAutoscalePool *digitaloceandropletautoscalepoolv1alpha1.DigitalOceanDropletAutoscalePool
	DigitalOceanLabels               map[string]string
}

// initializeLocals copies iac-input fields into the Locals struct and
// builds the standard Planton label set, which lands on every member
// droplet as tags.
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceandropletautoscalepoolv1alpha1.DigitalOceanDropletAutoscalePoolIacInput) *Locals {
	locals := &Locals{}

	locals.DigitalOceanDropletAutoscalePool = iacInput.Target

	locals.DigitalOceanLabels = map[string]string{
		digitaloceanlabelkeys.Resource:     strconv.FormatBool(true),
		digitaloceanlabelkeys.ResourceName: locals.DigitalOceanDropletAutoscalePool.Metadata.Name,
		digitaloceanlabelkeys.ResourceKind: catalogkind.CatalogKind_DigitalOceanDropletAutoscalePool.String(),
	}

	if locals.DigitalOceanDropletAutoscalePool.Metadata.Org != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Organization] = locals.DigitalOceanDropletAutoscalePool.Metadata.Org
	}

	if locals.DigitalOceanDropletAutoscalePool.Metadata.Env != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Environment] = locals.DigitalOceanDropletAutoscalePool.Metadata.Env
	}

	if locals.DigitalOceanDropletAutoscalePool.Metadata.Id != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.ResourceId] = locals.DigitalOceanDropletAutoscalePool.Metadata.Id
	}

	return locals
}
