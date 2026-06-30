package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	gcppubsubtopicv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcppubsubtopic/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpPubSubTopic    *gcppubsubtopicv1.GcpPubSubTopic
	GcpLabels         map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcppubsubtopicv1.GcpPubSubTopicStackInput) *Locals {
	locals := &Locals{}
	locals.GcpPubSubTopic = stackInput.Target
	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     "true",
		gcplabelkeys.ResourceName: locals.GcpPubSubTopic.Spec.TopicName,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpPubSubTopic.String()),
	}

	if locals.GcpPubSubTopic.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpPubSubTopic.Metadata.Org
	}
	if locals.GcpPubSubTopic.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpPubSubTopic.Metadata.Env
	}
	if locals.GcpPubSubTopic.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpPubSubTopic.Metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
