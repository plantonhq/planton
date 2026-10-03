package module

import (
	"strings"

	"github.com/pkg/errors"
	azureservicebussubscriptionv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureservicebussubscription/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureServiceBusSubscription *azureservicebussubscriptionv1alpha1.AzureServiceBusSubscription
	TopicId                     string
	TopicName                   string
	NamespaceName               string
}

// statusStrings maps the spec's gate-state enum to ARM's wire values. The
// unspecified row deploys Active -- an unmapped enum would send the empty
// string, which the provider rejects.
var statusStrings = map[azureservicebussubscriptionv1alpha1.AzureServiceBusSubscriptionStatusValue]string{
	azureservicebussubscriptionv1alpha1.AzureServiceBusSubscriptionStatusValue_azure_service_bus_subscription_status_unspecified: "Active",
	azureservicebussubscriptionv1alpha1.AzureServiceBusSubscriptionStatusValue_ACTIVE:                                            "Active",
	azureservicebussubscriptionv1alpha1.AzureServiceBusSubscriptionStatusValue_DISABLED:                                          "Disabled",
	azureservicebussubscriptionv1alpha1.AzureServiceBusSubscriptionStatusValue_RECEIVE_DISABLED:                                  "ReceiveDisabled",
}

// filterTypeStrings maps the rule's filter-type enum to the provider's
// case-sensitive wire values.
var filterTypeStrings = map[azureservicebussubscriptionv1alpha1.AzureServiceBusFilterType]string{
	azureservicebussubscriptionv1alpha1.AzureServiceBusFilterType_SQL_FILTER:         "SqlFilter",
	azureservicebussubscriptionv1alpha1.AzureServiceBusFilterType_CORRELATION_FILTER: "CorrelationFilter",
}

// parseTopicId extracts the namespace and topic names from a Service Bus
// topic ARM id (…/namespaces/{ns}/topics/{topic}). Malformed ids fail
// loudly here (matching the Terraform module's anchored regexes) instead of
// computing wrong names.
func parseTopicId(topicId string) (namespaceName string, topicName string, err error) {
	topicParts := strings.Split(topicId, "/topics/")
	if len(topicParts) != 2 || topicParts[1] == "" || strings.Contains(topicParts[1], "/") {
		return "", "", errors.Errorf("topic_id %q is not a Service Bus topic ARM id", topicId)
	}
	nsParts := strings.Split(topicParts[0], "/namespaces/")
	if len(nsParts) != 2 || nsParts[1] == "" || strings.Contains(nsParts[1], "/") {
		return "", "", errors.Errorf("topic_id %q does not carry a Service Bus namespace segment", topicId)
	}
	return nsParts[1], topicParts[1], nil
}

func initializeLocals(ctx *pulumi.Context, iacInput *azureservicebussubscriptionv1alpha1.AzureServiceBusSubscriptionIacInput) *Locals {
	locals := &Locals{}

	locals.AzureServiceBusSubscription = iacInput.Target
	locals.TopicId = iacInput.Target.Spec.TopicId.GetValue()

	// Subscriptions carry no Azure tags: ARM does not support tags on
	// Service Bus entities, so the platform's identity tags live on the
	// parent namespace.

	return locals
}
