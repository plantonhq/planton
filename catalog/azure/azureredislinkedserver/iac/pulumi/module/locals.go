package module

import (
	azureredislinkedserverv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureredislinkedserver/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureRedisLinkedServer *azureredislinkedserverv1alpha1.AzureRedisLinkedServer
	TargetRedisCacheId     string
	LinkedRedisCacheId     string
}

// serverRoleStrings maps the spec's role enum to ARM's capitalized values.
var serverRoleStrings = map[azureredislinkedserverv1alpha1.AzureRedisLinkedServerRole]string{
	azureredislinkedserverv1alpha1.AzureRedisLinkedServerRole_PRIMARY:   "Primary",
	azureredislinkedserverv1alpha1.AzureRedisLinkedServerRole_SECONDARY: "Secondary",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azureredislinkedserverv1alpha1.AzureRedisLinkedServerIacInput) *Locals {
	locals := &Locals{}

	locals.AzureRedisLinkedServer = iacInput.Target
	locals.TargetRedisCacheId = iacInput.Target.Spec.TargetRedisCacheId.GetValue()
	locals.LinkedRedisCacheId = iacInput.Target.Spec.LinkedRedisCacheId.GetValue()

	// No Azure tags: ARM does not support tags on linked servers, so the
	// platform's identity tags live on the caches themselves.

	return locals
}
