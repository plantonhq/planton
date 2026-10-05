package module

import (
	"strings"

	"github.com/pkg/errors"
	azureservicebustopicv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureservicebustopic/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureServiceBusTopic *azureservicebustopicv1alpha1.AzureServiceBusTopic
	NamespaceId          string
	NamespaceName        string
}

// statusStrings maps the spec's gate-state enum to ARM's wire values. The
// unspecified row deploys Active -- an unmapped enum would send the empty
// string, which the provider rejects. Topics support only Active/Disabled;
// direction gating happens per subscription.
var statusStrings = map[azureservicebustopicv1alpha1.AzureServiceBusTopicStatusValue]string{
	azureservicebustopicv1alpha1.AzureServiceBusTopicStatusValue_azure_service_bus_topic_status_unspecified: "Active",
	azureservicebustopicv1alpha1.AzureServiceBusTopicStatusValue_ACTIVE:                                     "Active",
	azureservicebustopicv1alpha1.AzureServiceBusTopicStatusValue_DISABLED:                                   "Disabled",
}

// parseNamespaceName extracts the namespace name from a Service Bus
// namespace ARM id. The id must END with /namespaces/{name} (matching the
// Terraform module's anchored regex), so a malformed id fails loudly here
// instead of computing a wrong name.
func parseNamespaceName(namespaceId string) (string, error) {
	parts := strings.Split(namespaceId, "/namespaces/")
	if len(parts) != 2 || parts[1] == "" || strings.Contains(parts[1], "/") {
		return "", errors.Errorf("namespace_id %q is not a Service Bus namespace ARM id", namespaceId)
	}
	return parts[1], nil
}

func initializeLocals(ctx *pulumi.Context, iacInput *azureservicebustopicv1alpha1.AzureServiceBusTopicIacInput) *Locals {
	locals := &Locals{}

	locals.AzureServiceBusTopic = iacInput.Target
	locals.NamespaceId = iacInput.Target.Spec.NamespaceId.GetValue()

	// Topics carry no Azure tags: ARM does not support tags on Service Bus
	// entities, so the platform's identity tags live on the parent
	// namespace.

	return locals
}
