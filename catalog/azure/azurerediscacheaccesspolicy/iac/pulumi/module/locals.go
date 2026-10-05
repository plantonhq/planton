package module

import (
	azurerediscacheaccesspolicyv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurerediscacheaccesspolicy/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureRedisCacheAccessPolicy *azurerediscacheaccesspolicyv1alpha1.AzureRedisCacheAccessPolicy
	RedisCacheId                string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurerediscacheaccesspolicyv1alpha1.AzureRedisCacheAccessPolicyIacInput) *Locals {
	locals := &Locals{}

	locals.AzureRedisCacheAccessPolicy = iacInput.Target
	locals.RedisCacheId = iacInput.Target.Spec.RedisCacheId.GetValue()

	// No Azure tags: ARM does not support tags on access policies (cache
	// children), so the platform's identity tags live on the cache.

	return locals
}
