package module

import (
	azureservicebusauthorizationrulev1alpha1 "github.com/plantonhq/planton/catalog/azure/azureservicebusauthorizationrule/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureServiceBusAuthorizationRule *azureservicebusauthorizationrulev1alpha1.AzureServiceBusAuthorizationRule

	// Exactly one of the three is non-empty (spec-enforced XOR) -- the
	// scope discriminator that picks which provider resource materializes.
	NamespaceId string
	QueueId     string
	TopicId     string

	// The rights trio -- Azure defaults all three to false; the spec's
	// CELs guarantee at least one is true and manage implies listen+send.
	Listen bool
	Send   bool
	Manage bool
}

func initializeLocals(ctx *pulumi.Context, iacInput *azureservicebusauthorizationrulev1alpha1.AzureServiceBusAuthorizationRuleIacInput) *Locals {
	locals := &Locals{}

	locals.AzureServiceBusAuthorizationRule = iacInput.Target
	spec := iacInput.Target.Spec

	locals.NamespaceId = spec.NamespaceId.GetValue()
	locals.QueueId = spec.QueueId.GetValue()
	locals.TopicId = spec.TopicId.GetValue()

	locals.Listen = spec.GetListen()
	locals.Send = spec.GetSend()
	locals.Manage = spec.GetManage()

	// Authorization rules carry no Azure tags: ARM does not support tags
	// on Service Bus entities, so the platform's identity tags live on the
	// parent namespace.

	return locals
}
