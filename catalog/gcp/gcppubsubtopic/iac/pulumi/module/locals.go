package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcppubsubtopicv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcppubsubtopic/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpPubSubTopic    *gcppubsubtopicv1alpha1.GcpPubSubTopic
	GcpLabels         map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcppubsubtopicv1alpha1.GcpPubSubTopicIacInput) *Locals {
	locals := &Locals{}
	locals.GcpPubSubTopic = iacInput.Target

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpPubSubTopic.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.GcpPubSubTopic.Spec.TopicName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpPubSubTopic.String())

	if locals.GcpPubSubTopic.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpPubSubTopic.Metadata.Org
	}
	if locals.GcpPubSubTopic.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpPubSubTopic.Metadata.Env
	}
	if locals.GcpPubSubTopic.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpPubSubTopic.Metadata.Id
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
