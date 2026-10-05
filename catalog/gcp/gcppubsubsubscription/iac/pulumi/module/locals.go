package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcppubsubsubscriptionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcppubsubsubscription/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig     *gcpprovider.GcpProviderConfig
	GcpPubSubSubscription *gcppubsubsubscriptionv1alpha1.GcpPubSubSubscription
	GcpLabels             map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcppubsubsubscriptionv1alpha1.GcpPubSubSubscriptionIacInput) *Locals {
	locals := &Locals{}
	locals.GcpPubSubSubscription = iacInput.Target

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpPubSubSubscription.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.GcpPubSubSubscription.Spec.SubscriptionName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpPubSubSubscription.String())

	if locals.GcpPubSubSubscription.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpPubSubSubscription.Metadata.Org
	}
	if locals.GcpPubSubSubscription.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpPubSubSubscription.Metadata.Env
	}
	if locals.GcpPubSubSubscription.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpPubSubSubscription.Metadata.Id
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
