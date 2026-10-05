package module

import (
	azuremanagedredisaccesspolicyassignmentv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremanagedredisaccesspolicyassignment/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureManagedRedisAccessPolicyAssignment *azuremanagedredisaccesspolicyassignmentv1alpha1.AzureManagedRedisAccessPolicyAssignment
	ManagedRedisId                          string
	ObjectId                                string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azuremanagedredisaccesspolicyassignmentv1alpha1.AzureManagedRedisAccessPolicyAssignmentIacInput) *Locals {
	locals := &Locals{}

	locals.AzureManagedRedisAccessPolicyAssignment = iacInput.Target
	locals.ManagedRedisId = iacInput.Target.Spec.ManagedRedisId.GetValue()
	locals.ObjectId = iacInput.Target.Spec.ObjectId.GetValue()

	// No Azure tags: ARM does not support tags on access policy
	// assignments (database children), so the platform's identity tags
	// live on the cluster.

	return locals
}
