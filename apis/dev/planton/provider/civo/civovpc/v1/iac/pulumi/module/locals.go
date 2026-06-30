package module

import (
	"strconv"

	civoprovider "github.com/plantonhq/planton/apis/dev/planton/provider/civo"
	civovpcv1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civovpc/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/civo/civolabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CivoProviderConfig *civoprovider.CivoProviderConfig
	CivoVpc            *civovpcv1.CivoVpc
	CivoLabels         map[string]string
}

// initializeLocals copies stack‑input fields into the Locals struct and builds
// a reusable label map. Mirrors DigitalOcean module style.
func initializeLocals(_ *pulumi.Context, stackInput *civovpcv1.CivoVpcStackInput) *Locals {
	locals := &Locals{}

	locals.CivoVpc = stackInput.Target

	// Standard Planton labels for Civo resources.
	locals.CivoLabels = map[string]string{
		civolabelkeys.Resource:     strconv.FormatBool(true),
		civolabelkeys.ResourceName: locals.CivoVpc.Metadata.Name,
		civolabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_CivoVpc.String(),
	}

	if locals.CivoVpc.Metadata.Org != "" {
		locals.CivoLabels[civolabelkeys.Organization] = locals.CivoVpc.Metadata.Org
	}

	if locals.CivoVpc.Metadata.Env != "" {
		locals.CivoLabels[civolabelkeys.Environment] = locals.CivoVpc.Metadata.Env
	}

	if locals.CivoVpc.Metadata.Id != "" {
		locals.CivoLabels[civolabelkeys.ResourceId] = locals.CivoVpc.Metadata.Id
	}

	locals.CivoProviderConfig = stackInput.ProviderConfig

	return locals
}
