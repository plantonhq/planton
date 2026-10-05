package module

import (
	"strconv"
	"strings"

	digitaloceanprovider "github.com/plantonhq/planton/catalog/digitalocean"
	digitaloceanappv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanapp/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/digitaloceanlabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals gathers convenient references for the rest of the module.
// App Platform has no tag surface, so the label map is kept for
// consistency with sibling DigitalOcean modules and is not attached
// to the created app.
type Locals struct {
	DigitalOceanProviderConfig *digitaloceanprovider.DigitalOceanProviderConfig
	DigitalOceanApp            *digitaloceanappv1alpha1.DigitalOceanApp
	DigitalOceanLabels         map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *digitaloceanappv1alpha1.DigitalOceanAppIacInput) *Locals {
	locals := &Locals{}
	locals.DigitalOceanApp = iacInput.Target

	locals.DigitalOceanLabels = map[string]string{
		digitaloceanlabelkeys.Resource:     strconv.FormatBool(true),
		digitaloceanlabelkeys.ResourceName: locals.DigitalOceanApp.Metadata.Name,
		digitaloceanlabelkeys.ResourceKind: catalogkind.CatalogKind_DigitalOceanApp.String(),
	}

	if locals.DigitalOceanApp.Metadata.Org != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Organization] = locals.DigitalOceanApp.Metadata.Org
	}
	if locals.DigitalOceanApp.Metadata.Env != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Environment] = locals.DigitalOceanApp.Metadata.Env
	}
	if locals.DigitalOceanApp.Metadata.Id != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.ResourceId] = locals.DigitalOceanApp.Metadata.Id
	}

	locals.DigitalOceanProviderConfig = iacInput.ProviderConfig
	return locals
}

func stripScheme(url string) string {
	url = strings.TrimPrefix(url, "https://")
	return strings.TrimPrefix(url, "http://")
}
