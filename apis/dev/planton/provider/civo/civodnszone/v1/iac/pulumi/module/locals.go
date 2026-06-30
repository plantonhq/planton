package module

import (
	"strconv"

	civoprovider "github.com/plantonhq/planton/apis/dev/planton/provider/civo"
	civodnszonev1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civodnszone/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/civo/civolabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles quick references that multiple files need.
type Locals struct {
	CivoProviderConfig *civoprovider.CivoProviderConfig
	CivoDnsZone        *civodnszonev1.CivoDnsZone
	CivoLabels         map[string]string
}

// initializeLocals mirrors the pattern used by other Planton modules.
func initializeLocals(_ *pulumi.Context, stackInput *civodnszonev1.CivoDnsZoneStackInput) *Locals {
	locals := &Locals{}
	locals.CivoDnsZone = stackInput.Target

	// Standard Planton labels for Civo resources.
	locals.CivoLabels = map[string]string{
		civolabelkeys.Resource:     strconv.FormatBool(true),
		civolabelkeys.ResourceName: locals.CivoDnsZone.Metadata.Name,
		civolabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_CivoDnsZone.String(),
	}

	if locals.CivoDnsZone.Metadata.Org != "" {
		locals.CivoLabels[civolabelkeys.Organization] = locals.CivoDnsZone.Metadata.Org
	}
	if locals.CivoDnsZone.Metadata.Env != "" {
		locals.CivoLabels[civolabelkeys.Environment] = locals.CivoDnsZone.Metadata.Env
	}
	if locals.CivoDnsZone.Metadata.Id != "" {
		locals.CivoLabels[civolabelkeys.ResourceId] = locals.CivoDnsZone.Metadata.Id
	}

	locals.CivoProviderConfig = stackInput.ProviderConfig

	return locals
}
