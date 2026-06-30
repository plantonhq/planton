package module

import (
	"strconv"

	civoprovider "github.com/plantonhq/planton/apis/dev/planton/provider/civo"
	civobucketv1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civobucket/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/civo/civolabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CivoProviderConfig *civoprovider.CivoProviderConfig
	CivoBucket         *civobucketv1.CivoBucket
	CivoLabels         map[string]string
}

// initializeLocals copies stack‑input fields into Locals and builds a label map.
func initializeLocals(_ *pulumi.Context, stackInput *civobucketv1.CivoBucketStackInput) *Locals {
	var locals Locals

	locals.CivoBucket = stackInput.Target

	// Standard Planton labels for Civo resources.
	locals.CivoLabels = map[string]string{
		civolabelkeys.Resource:     strconv.FormatBool(true),
		civolabelkeys.ResourceName: locals.CivoBucket.Metadata.Name,
		civolabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_CivoBucket.String(),
	}

	if locals.CivoBucket.Metadata.Org != "" {
		locals.CivoLabels[civolabelkeys.Organization] = locals.CivoBucket.Metadata.Org
	}
	if locals.CivoBucket.Metadata.Env != "" {
		locals.CivoLabels[civolabelkeys.Environment] = locals.CivoBucket.Metadata.Env
	}
	if locals.CivoBucket.Metadata.Id != "" {
		locals.CivoLabels[civolabelkeys.ResourceId] = locals.CivoBucket.Metadata.Id
	}

	locals.CivoProviderConfig = stackInput.ProviderConfig
	return &locals
}
