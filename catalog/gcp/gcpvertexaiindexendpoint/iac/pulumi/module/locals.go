package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpvertexaiindexendpointv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaiindexendpoint/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// computeSelfLinkPrefix is the URL prefix a compute self-link carries in
// front of the relative resource path the Vertex AI API expects.
const computeSelfLinkPrefix = "https://www.googleapis.com/compute/v1/"

type Locals struct {
	GcpProviderConfig        *gcpprovider.GcpProviderConfig
	GcpVertexAiIndexEndpoint *gcpvertexaiindexendpointv1alpha1.GcpVertexAiIndexEndpoint
	GcpLabels                map[string]string
	Network                  string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpvertexaiindexendpointv1alpha1.GcpVertexAiIndexEndpointIacInput) *Locals {
	locals := &Locals{}
	locals.GcpVertexAiIndexEndpoint = iacInput.Target
	locals.GcpProviderConfig = iacInput.ProviderConfig

	// The Vertex AI API expects the RELATIVE network form
	// projects/{project}/global/networks/{name} and rejects full compute
	// self-link URLs. GcpVpcNetwork references resolve to the self-link
	// (the kind's canonical output), so both literal URLs and references
	// are normalized. Stripping is a no-op for values already in relative
	// form — identical normalization to the Terraform module.
	if locals.GcpVertexAiIndexEndpoint.Spec.Network != nil {
		locals.Network = strings.TrimPrefix(
			locals.GcpVertexAiIndexEndpoint.Spec.Network.GetValue(), computeSelfLinkPrefix)
	}

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpVertexAiIndexEndpoint.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = strings.ToLower(locals.GcpVertexAiIndexEndpoint.Metadata.Name)
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpVertexAiIndexEndpoint.String())

	if locals.GcpVertexAiIndexEndpoint.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpVertexAiIndexEndpoint.Metadata.Org
	}
	if locals.GcpVertexAiIndexEndpoint.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpVertexAiIndexEndpoint.Metadata.Env
	}
	if locals.GcpVertexAiIndexEndpoint.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpVertexAiIndexEndpoint.Metadata.Id
	}

	return locals
}
