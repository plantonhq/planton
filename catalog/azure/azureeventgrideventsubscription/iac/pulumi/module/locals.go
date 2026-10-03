package module

import (
	"fmt"
	"strings"

	azureeventgrideventsubscriptionv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventgrideventsubscription/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// No tags in these locals: the provider carries no tags argument on
// event subscriptions (Event Grid stores free-form labels instead --
// the spec's labels field).
type Locals struct {
	AzureEventgridEventSubscription *azureeventgrideventsubscriptionv1alpha1.AzureEventgridEventSubscription
}

func initializeLocals(ctx *pulumi.Context, iacInput *azureeventgrideventsubscriptionv1alpha1.AzureEventgridEventSubscriptionIacInput) *Locals {
	locals := &Locals{}

	locals.AzureEventgridEventSubscription = iacInput.Target

	return locals
}

// parseSystemTopicId splits a system topic's ARM id into its resource
// group and topic name -- the provider addresses system-topic
// subscriptions by (resource group, system topic name) while the spec
// takes the topic's ARM id (the composable reference shape). Segment
// names are matched case-insensitively (ARM treats them that way), so
// ids composed by hand survive the split.
//
// The id's shape is /subscriptions/{sub}/resourceGroups/{rg}
// /providers/Microsoft.EventGrid/systemTopics/{name}.
func parseSystemTopicId(systemTopicId string) (resourceGroup string, systemTopicName string, err error) {
	segments := strings.Split(systemTopicId, "/")
	for i, segment := range segments {
		if i+1 >= len(segments) {
			break
		}
		switch strings.ToLower(segment) {
		case "resourcegroups":
			resourceGroup = segments[i+1]
		case "systemtopics":
			systemTopicName = segments[i+1]
		}
	}
	if resourceGroup == "" || systemTopicName == "" {
		return "", "", fmt.Errorf(
			"system_topic_id %q does not look like a system topic ARM id (expected .../resourceGroups/{rg}/providers/Microsoft.EventGrid/systemTopics/{name})",
			systemTopicId)
	}
	return resourceGroup, systemTopicName, nil
}
