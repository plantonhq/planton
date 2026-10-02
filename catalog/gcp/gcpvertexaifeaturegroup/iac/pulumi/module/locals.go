package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpvertexaifeaturegroupv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaifeaturegroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig       *gcpprovider.GcpProviderConfig
	GcpVertexAiFeatureGroup *gcpvertexaifeaturegroupv1alpha1.GcpVertexAiFeatureGroup
	GcpLabels               map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpvertexaifeaturegroupv1alpha1.GcpVertexAiFeatureGroupStackInput) *Locals {
	locals := &Locals{}
	locals.GcpVertexAiFeatureGroup = stackInput.Target
	metadata := locals.GcpVertexAiFeatureGroup.Metadata

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module. The same
	// set lands on the group and on every registered feature.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpVertexAiFeatureGroup.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.GcpVertexAiFeatureGroup.Spec.FeatureGroupId
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpVertexAiFeatureGroup.String())

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
