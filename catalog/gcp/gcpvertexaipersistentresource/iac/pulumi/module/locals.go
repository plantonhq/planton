package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpvertexaipersistentresourcev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaipersistentresource/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig             *gcpprovider.GcpProviderConfig
	GcpVertexAiPersistentResource *gcpvertexaipersistentresourcev1alpha1.GcpVertexAiPersistentResource
	GcpLabels                     map[string]string

	// PersistentResourceId is spec.persistent_resource_id when set,
	// otherwise metadata.name -- the same fallback the Terraform module
	// applies in locals.tf.
	PersistentResourceId string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpvertexaipersistentresourcev1alpha1.GcpVertexAiPersistentResourceStackInput) *Locals {
	locals := &Locals{}
	locals.GcpVertexAiPersistentResource = stackInput.Target
	metadata := locals.GcpVertexAiPersistentResource.Metadata

	locals.PersistentResourceId = locals.GcpVertexAiPersistentResource.Spec.PersistentResourceId
	if locals.PersistentResourceId == "" {
		locals.PersistentResourceId = metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpVertexAiPersistentResource.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.PersistentResourceId
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpVertexAiPersistentResource.String())

	if metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = metadata.Org
	}
	if metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = metadata.Env
	}
	if metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
