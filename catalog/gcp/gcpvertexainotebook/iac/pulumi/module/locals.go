package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpvertexainotebookv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexainotebook/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig   *gcpprovider.GcpProviderConfig
	GcpVertexAiNotebook *gcpvertexainotebookv1alpha1.GcpVertexAiNotebook
	GcpLabels           map[string]string
	InstanceName        string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpvertexainotebookv1alpha1.GcpVertexAiNotebookIacInput) *Locals {
	locals := &Locals{}
	locals.GcpVertexAiNotebook = iacInput.Target
	locals.GcpProviderConfig = iacInput.ProviderConfig

	// Determine the instance name: explicit instance_name or fall back to metadata.name.
	locals.InstanceName = locals.GcpVertexAiNotebook.Spec.InstanceName
	if locals.InstanceName == "" {
		locals.InstanceName = locals.GcpVertexAiNotebook.Metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpVertexAiNotebook.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.InstanceName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpVertexAiNotebook.String())

	if locals.GcpVertexAiNotebook.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpVertexAiNotebook.Metadata.Org
	}
	if locals.GcpVertexAiNotebook.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpVertexAiNotebook.Metadata.Env
	}
	if locals.GcpVertexAiNotebook.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpVertexAiNotebook.Metadata.Id
	}

	return locals
}
