package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcolabruntimetemplatev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcolabruntimetemplate/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig       *gcpprovider.GcpProviderConfig
	GcpColabRuntimeTemplate *gcpcolabruntimetemplatev1alpha1.GcpColabRuntimeTemplate
	GcpLabels               map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpcolabruntimetemplatev1alpha1.GcpColabRuntimeTemplateStackInput) *Locals {
	locals := &Locals{}
	locals.GcpColabRuntimeTemplate = stackInput.Target
	metadata := locals.GcpColabRuntimeTemplate.Metadata

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpColabRuntimeTemplate.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = metadata.Name
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpColabRuntimeTemplate.String())

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
