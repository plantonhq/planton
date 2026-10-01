package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpdeliverypipelinev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdeliverypipeline/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig   *gcpprovider.GcpProviderConfig
	GcpDeliveryPipeline *gcpdeliverypipelinev1alpha1.GcpDeliveryPipeline

	// AttributionLabels is the planton-ai_* set; GcpLabels is the spec's
	// labels with the attribution set merged on top.
	AttributionLabels map[string]string
	GcpLabels         map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpdeliverypipelinev1alpha1.GcpDeliveryPipelineStackInput) *Locals {
	locals := &Locals{}
	locals.GcpDeliveryPipeline = stackInput.Target
	metadata := locals.GcpDeliveryPipeline.Metadata

	locals.AttributionLabels = map[string]string{
		gcplabelkeys.Resource:     "true",
		gcplabelkeys.ResourceName: metadata.Name,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpDeliveryPipeline.String()),
	}
	if metadata.Org != "" {
		locals.AttributionLabels[gcplabelkeys.Organization] = metadata.Org
	}
	if metadata.Env != "" {
		locals.AttributionLabels[gcplabelkeys.Environment] = metadata.Env
	}
	if metadata.Id != "" {
		locals.AttributionLabels[gcplabelkeys.ResourceId] = metadata.Id
	}

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module.
	locals.GcpLabels = mergeLabels(locals.GcpDeliveryPipeline.Spec.Labels, locals.AttributionLabels)

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}

// mergeLabels returns user labels with the attribution labels laid over
// them, the Terraform module's merge(user, attribution).
func mergeLabels(user, attribution map[string]string) map[string]string {
	merged := map[string]string{}
	for key, value := range user {
		merged[key] = value
	}
	for key, value := range attribution {
		merged[key] = value
	}
	return merged
}
