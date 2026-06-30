package module

import (
	"strconv"

	civoprovider "github.com/plantonhq/planton/apis/dev/planton/provider/civo"
	civokubernetesnodepoolv1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civokubernetesnodepool/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals aggregates handy references for the rest of the module.
//
// Keeping this struct tiny and flat mirrors the way Terraform modules rely on
// a handful of “locals” rather than deep, complex helpers.
type Locals struct {
	CivoProviderConfig     *civoprovider.CivoProviderConfig
	CivoKubernetesNodePool *civokubernetesnodepoolv1.CivoKubernetesNodePool
	CivoLabels             map[string]string
}

// initializeLocals prepares convenience values exactly once.
//
// It mirrors the digital_ocean_kubernetes_node_pool pattern but stays “plain Go”
// (no generics or advanced reflection), so Terraform‑first engineers can follow
// along easily.
func initializeLocals(_ *pulumi.Context, stackInput *civokubernetesnodepoolv1.CivoKubernetesNodePoolStackInput) *Locals {
	locals := &Locals{}

	locals.CivoKubernetesNodePool = stackInput.Target
	locals.CivoProviderConfig = stackInput.ProviderConfig

	// Standard Planton labels.  No external helper package required: keep it obvious.
	locals.CivoLabels = map[string]string{
		"resource":      strconv.FormatBool(true),
		"resource_name": locals.CivoKubernetesNodePool.Metadata.Name,
		"resource_kind": cloudresourcekind.CloudResourceKind_CivoKubernetesNodePool.String(),
	}

	if locals.CivoKubernetesNodePool.Metadata.Org != "" {
		locals.CivoLabels["org"] = locals.CivoKubernetesNodePool.Metadata.Org
	}
	if locals.CivoKubernetesNodePool.Metadata.Env != "" {
		locals.CivoLabels["env"] = locals.CivoKubernetesNodePool.Metadata.Env
	}
	if locals.CivoKubernetesNodePool.Metadata.Id != "" {
		locals.CivoLabels["resource_id"] = locals.CivoKubernetesNodePool.Metadata.Id
	}

	return locals
}
