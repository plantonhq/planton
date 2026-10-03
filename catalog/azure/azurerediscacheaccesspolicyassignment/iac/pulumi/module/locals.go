package module

import (
	azurerediscacheaccesspolicyassignmentv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurerediscacheaccesspolicyassignment/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureRedisCacheAccessPolicyAssignment *azurerediscacheaccesspolicyassignmentv1alpha1.AzureRedisCacheAccessPolicyAssignment
	RedisCacheId                          string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurerediscacheaccesspolicyassignmentv1alpha1.AzureRedisCacheAccessPolicyAssignmentIacInput) *Locals {
	locals := &Locals{}

	locals.AzureRedisCacheAccessPolicyAssignment = iacInput.Target
	locals.RedisCacheId = iacInput.Target.Spec.RedisCacheId.GetValue()

	// No Azure tags: ARM does not support tags on access policy
	// assignments (cache children), so the platform's identity tags live
	// on the cache.

	return locals
}
