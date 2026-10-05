package module

import (
	"strings"

	"github.com/pkg/errors"
	azureservicebusqueuev1alpha1 "github.com/plantonhq/planton/catalog/azure/azureservicebusqueue/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureServiceBusQueue *azureservicebusqueuev1alpha1.AzureServiceBusQueue
	NamespaceId          string
	NamespaceName        string
}

// statusStrings maps the spec's gate-state enum to ARM's wire values. The
// unspecified row deploys Active -- an unmapped enum would send the empty
// string, which the provider rejects.
var statusStrings = map[azureservicebusqueuev1alpha1.AzureServiceBusEntityStatus]string{
	azureservicebusqueuev1alpha1.AzureServiceBusEntityStatus_azure_service_bus_entity_status_unspecified: "Active",
	azureservicebusqueuev1alpha1.AzureServiceBusEntityStatus_ACTIVE:                                      "Active",
	azureservicebusqueuev1alpha1.AzureServiceBusEntityStatus_DISABLED:                                    "Disabled",
	azureservicebusqueuev1alpha1.AzureServiceBusEntityStatus_SEND_DISABLED:                               "SendDisabled",
	azureservicebusqueuev1alpha1.AzureServiceBusEntityStatus_RECEIVE_DISABLED:                            "ReceiveDisabled",
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

func initializeLocals(ctx *pulumi.Context, iacInput *azureservicebusqueuev1alpha1.AzureServiceBusQueueIacInput) *Locals {
	locals := &Locals{}

	locals.AzureServiceBusQueue = iacInput.Target
	locals.NamespaceId = iacInput.Target.Spec.NamespaceId.GetValue()

	// Queues carry no Azure tags: ARM does not support tags on Service Bus
	// entities, so the platform's identity tags live on the parent
	// namespace.

	return locals
}
