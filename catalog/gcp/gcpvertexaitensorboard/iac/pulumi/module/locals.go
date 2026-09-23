package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpvertexaitensorboardv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaitensorboard/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig      *gcpprovider.GcpProviderConfig
	GcpVertexAiTensorboard *gcpvertexaitensorboardv1alpha1.GcpVertexAiTensorboard
	GcpLabels              map[string]string

	// DisplayName is spec.display_name when set, otherwise metadata.name --
	// Google requires one, and the Terraform module applies the same
	// fallback in locals.tf.
	DisplayName string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpvertexaitensorboardv1alpha1.GcpVertexAiTensorboardStackInput) *Locals {
	locals := &Locals{}
	locals.GcpVertexAiTensorboard = stackInput.Target
	metadata := locals.GcpVertexAiTensorboard.Metadata

	locals.DisplayName = locals.GcpVertexAiTensorboard.Spec.DisplayName
	if locals.DisplayName == "" {
		locals.DisplayName = metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module. The same
	// set lands on the TensorBoard and on every experiment and run.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpVertexAiTensorboard.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = metadata.Name
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpVertexAiTensorboard.String())

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
