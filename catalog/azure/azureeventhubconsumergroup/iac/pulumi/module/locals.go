package module

import (
	"regexp"

	"github.com/pkg/errors"
	azureeventhubconsumergroupv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventhubconsumergroup/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureEventHubConsumerGroup *azureeventhubconsumergroupv1alpha1.AzureEventHubConsumerGroup
	EventHubId                 string
	ResourceGroupName          string
	NamespaceName              string
	EventHubName               string
}

// eventHubIdPattern matches an event hub ARM id of the shape
// /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.EventHub/namespaces/{ns}/eventhubs/{hub}
// with segment literals in ARM's canonical camelCase, anchored at the end
// (matching the Terraform module's regex) so a malformed id fails loudly
// instead of computing wrong names.
var eventHubIdPattern = regexp.MustCompile(
	`/resourceGroups/([^/]+)/providers/Microsoft\.EventHub/namespaces/([^/]+)/eventhubs/([^/]+)$`)

// parseEventHubId extracts the resource group, namespace, and event hub
// names from the hub's ARM id. azurerm still addresses consumer groups by
// those discrete names rather than the parent's ARM id; deriving them from
// the spec's single parent reference keeps the spec on the ARM-id grain
// with no redundant fields that could contradict each other.
func parseEventHubId(eventHubId string) (resourceGroupName, namespaceName, eventHubName string, err error) {
	matches := eventHubIdPattern.FindStringSubmatch(eventHubId)
	if matches == nil {
		return "", "", "",
			errors.Errorf("event_hub_id %q is not an Event Hub ARM id "+
				"(expected .../resourceGroups/{rg}/providers/Microsoft.EventHub/namespaces/{ns}/eventhubs/{hub})",
				eventHubId)
	}
	return matches[1], matches[2], matches[3], nil
}

func initializeLocals(ctx *pulumi.Context, iacInput *azureeventhubconsumergroupv1alpha1.AzureEventHubConsumerGroupIacInput) *Locals {
	locals := &Locals{}

	locals.AzureEventHubConsumerGroup = iacInput.Target
	locals.EventHubId = iacInput.Target.Spec.EventHubId.GetValue()

	// Consumer groups carry no Azure tags: ARM does not support tags on
	// Event Hubs entities, so the platform's identity tags live on the
	// parent namespace.

	return locals
}
