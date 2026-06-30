package module

import (
	"strconv"

	civoprovider "github.com/plantonhq/planton/apis/dev/planton/provider/civo"
	civovolumev1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civovolume/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/civo/civolabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CivoProviderConfig *civoprovider.CivoProviderConfig
	CivoVolume         *civovolumev1.CivoVolume
	CivoLabels         map[string]string
}

// initializeLocals copies stack‑input fields into the Locals struct and builds
// a reusable label map. Mirrors the style of digital_ocean_volume’s initializeLocals().
func initializeLocals(_ *pulumi.Context, stackInput *civovolumev1.CivoVolumeStackInput) *Locals {
	locals := &Locals{}

	locals.CivoVolume = stackInput.Target

	// Standard Planton labels for Civo resources.
	locals.CivoLabels = map[string]string{
		civolabelkeys.Resource:     strconv.FormatBool(true),
		civolabelkeys.ResourceName: locals.CivoVolume.Metadata.Name,
		civolabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_CivoVolume.String(),
	}

	if locals.CivoVolume.Metadata.Org != "" {
		locals.CivoLabels[civolabelkeys.Organization] = locals.CivoVolume.Metadata.Org
	}

	if locals.CivoVolume.Metadata.Env != "" {
		locals.CivoLabels[civolabelkeys.Environment] = locals.CivoVolume.Metadata.Env
	}

	if locals.CivoVolume.Metadata.Id != "" {
		locals.CivoLabels[civolabelkeys.ResourceId] = locals.CivoVolume.Metadata.Id
	}

	locals.CivoProviderConfig = stackInput.ProviderConfig

	return locals
}
