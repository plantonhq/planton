package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp"
	gcppubsubsubscriptionv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcppubsubsubscription/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig      *gcpprovider.GcpProviderConfig
	GcpPubSubSubscription  *gcppubsubsubscriptionv1.GcpPubSubSubscription
	GcpLabels              map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcppubsubsubscriptionv1.GcpPubSubSubscriptionStackInput) *Locals {
	locals := &Locals{}
	locals.GcpPubSubSubscription = stackInput.Target
	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     "true",
		gcplabelkeys.ResourceName: locals.GcpPubSubSubscription.Spec.SubscriptionName,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpPubSubSubscription.String()),
	}

	if locals.GcpPubSubSubscription.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpPubSubSubscription.Metadata.Org
	}
	if locals.GcpPubSubSubscription.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpPubSubSubscription.Metadata.Env
	}
	if locals.GcpPubSubSubscription.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpPubSubSubscription.Metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
