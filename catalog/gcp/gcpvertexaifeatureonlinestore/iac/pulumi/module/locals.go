package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpvertexaifeatureonlinestorev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaifeatureonlinestore/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig             *gcpprovider.GcpProviderConfig
	GcpVertexAiFeatureOnlineStore *gcpvertexaifeatureonlinestorev1alpha1.GcpVertexAiFeatureOnlineStore
	GcpLabels                     map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpvertexaifeatureonlinestorev1alpha1.GcpVertexAiFeatureOnlineStoreStackInput) *Locals {
	locals := &Locals{}
	locals.GcpVertexAiFeatureOnlineStore = stackInput.Target
	metadata := locals.GcpVertexAiFeatureOnlineStore.Metadata

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module. The same
	// set lands on the store and on every feature view.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpVertexAiFeatureOnlineStore.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.GcpVertexAiFeatureOnlineStore.Spec.FeatureOnlineStoreId
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpVertexAiFeatureOnlineStore.String())

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
