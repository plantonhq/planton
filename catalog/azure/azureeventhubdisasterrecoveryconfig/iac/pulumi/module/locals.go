package module

import (
	"regexp"

	"github.com/pkg/errors"
	azureeventhubdisasterrecoveryconfigv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventhubdisasterrecoveryconfig/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureEventHubDisasterRecoveryConfig *azureeventhubdisasterrecoveryconfigv1alpha1.AzureEventHubDisasterRecoveryConfig
	PrimaryNamespaceId                  string
	PrimaryResourceGroupName            string
	PrimaryNamespaceName                string
	PartnerNamespaceId                  string
}

// The provider addresses the PRIMARY side by discrete names (namespace
// name + resource group) rather than by ARM ID, so the resolved primary
// namespace id is parsed into those parts. The anchored pattern rejects
// a malformed id loudly instead of sending garbage to the API.
var primaryNamespaceIdPattern = regexp.MustCompile(
	`/resourceGroups/(?P<rg>[^/]+)/providers/Microsoft.EventHub/namespaces/(?P<ns>[^/]+)$`)

func initializeLocals(ctx *pulumi.Context, iacInput *azureeventhubdisasterrecoveryconfigv1alpha1.AzureEventHubDisasterRecoveryConfigIacInput) (*Locals, error) {
	locals := &Locals{}

	locals.AzureEventHubDisasterRecoveryConfig = iacInput.Target
	spec := iacInput.Target.Spec

	locals.PrimaryNamespaceId = spec.PrimaryNamespaceId.GetValue()
	locals.PartnerNamespaceId = spec.PartnerNamespaceId.GetValue()

	matches := primaryNamespaceIdPattern.FindStringSubmatch(locals.PrimaryNamespaceId)
	if matches == nil {
		return nil, errors.Errorf(
			"primary_namespace_id %q is not a valid Event Hubs namespace ARM id"+
				" (expected .../resourceGroups/{rg}/providers/Microsoft.EventHub/namespaces/{ns})",
			locals.PrimaryNamespaceId)
	}
	locals.PrimaryResourceGroupName = matches[primaryNamespaceIdPattern.SubexpIndex("rg")]
	locals.PrimaryNamespaceName = matches[primaryNamespaceIdPattern.SubexpIndex("ns")]

	// Geo-DR configs carry no Azure tags: ARM does not support tags on
	// disasterRecoveryConfigs, so the platform's identity tags live on the
	// paired namespaces.

	return locals, nil
}
