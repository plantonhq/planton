package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpvertexaiagentenginev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaiagentengine/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig      *gcpprovider.GcpProviderConfig
	GcpVertexAiAgentEngine *gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngine
	GcpLabels              map[string]string

	// DisplayName is the agent's display name: spec.display_name when set,
	// otherwise metadata.name -- the same fallback the Terraform module
	// applies in locals.tf. Google requires a display name; Planton's
	// object name is the honest default.
	DisplayName string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngineStackInput) *Locals {
	locals := &Locals{}
	locals.GcpVertexAiAgentEngine = stackInput.Target

	locals.DisplayName = locals.GcpVertexAiAgentEngine.Spec.DisplayName
	if locals.DisplayName == "" {
		locals.DisplayName = locals.GcpVertexAiAgentEngine.Metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpVertexAiAgentEngine.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.GcpVertexAiAgentEngine.Metadata.Name
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpVertexAiAgentEngine.String())

	if locals.GcpVertexAiAgentEngine.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpVertexAiAgentEngine.Metadata.Org
	}
	if locals.GcpVertexAiAgentEngine.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpVertexAiAgentEngine.Metadata.Env
	}
	if locals.GcpVertexAiAgentEngine.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpVertexAiAgentEngine.Metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
