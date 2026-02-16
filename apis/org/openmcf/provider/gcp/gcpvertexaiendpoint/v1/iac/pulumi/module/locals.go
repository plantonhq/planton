package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp"
	gcpvertexaiendpointv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpvertexaiendpoint/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig   *gcpprovider.GcpProviderConfig
	GcpVertexAiEndpoint *gcpvertexaiendpointv1.GcpVertexAiEndpoint
	GcpLabels           map[string]string
	DisplayName         string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpvertexaiendpointv1.GcpVertexAiEndpointStackInput) *Locals {
	locals := &Locals{}
	locals.GcpVertexAiEndpoint = stackInput.Target
	locals.GcpProviderConfig = stackInput.ProviderConfig

	locals.DisplayName = locals.GcpVertexAiEndpoint.Spec.DisplayName

	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     "true",
		gcplabelkeys.ResourceName: strings.ToLower(locals.GcpVertexAiEndpoint.Metadata.Name),
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpVertexAiEndpoint.String()),
	}

	if locals.GcpVertexAiEndpoint.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpVertexAiEndpoint.Metadata.Org
	}
	if locals.GcpVertexAiEndpoint.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpVertexAiEndpoint.Metadata.Env
	}
	if locals.GcpVertexAiEndpoint.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpVertexAiEndpoint.Metadata.Id
	}

	return locals
}
