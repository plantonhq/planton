package module

import (
	"strconv"

	digitaloceanprovider "github.com/plantonhq/planton/catalog/digitalocean"
	digitaloceanfunctionv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanfunction/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/digitaloceanlabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	DigitalOceanProviderConfig *digitaloceanprovider.DigitalOceanProviderConfig
	DigitalOceanFunction       *digitaloceanfunctionv1alpha1.DigitalOceanFunction
	DigitalOceanLabels         map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *digitaloceanfunctionv1alpha1.DigitalOceanFunctionIacInput) *Locals {
	locals := &Locals{}
	locals.DigitalOceanFunction = iacInput.Target

	locals.DigitalOceanLabels = map[string]string{
		digitaloceanlabelkeys.Resource:     strconv.FormatBool(true),
		digitaloceanlabelkeys.ResourceName: locals.DigitalOceanFunction.Metadata.Name,
		digitaloceanlabelkeys.ResourceKind: catalogkind.CatalogKind_DigitalOceanFunction.String(),
	}

	if locals.DigitalOceanFunction.Metadata.Org != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Organization] = locals.DigitalOceanFunction.Metadata.Org
	}
	if locals.DigitalOceanFunction.Metadata.Env != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Environment] = locals.DigitalOceanFunction.Metadata.Env
	}
	if locals.DigitalOceanFunction.Metadata.Id != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.ResourceId] = locals.DigitalOceanFunction.Metadata.Id
	}

	locals.DigitalOceanProviderConfig = iacInput.ProviderConfig
	return locals
}
