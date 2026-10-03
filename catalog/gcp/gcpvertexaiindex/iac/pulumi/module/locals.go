package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpvertexaiindexv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaiindex/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpVertexAiIndex  *gcpvertexaiindexv1alpha1.GcpVertexAiIndex
	GcpLabels         map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpvertexaiindexv1alpha1.GcpVertexAiIndexIacInput) *Locals {
	locals := &Locals{}
	locals.GcpVertexAiIndex = iacInput.Target
	locals.GcpProviderConfig = iacInput.ProviderConfig

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpVertexAiIndex.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = strings.ToLower(locals.GcpVertexAiIndex.Metadata.Name)
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpVertexAiIndex.String())

	if locals.GcpVertexAiIndex.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpVertexAiIndex.Metadata.Org
	}
	if locals.GcpVertexAiIndex.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpVertexAiIndex.Metadata.Env
	}
	if locals.GcpVertexAiIndex.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpVertexAiIndex.Metadata.Id
	}

	return locals
}
