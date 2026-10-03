package module

import (
	azuremanagedredisgeoreplicationv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremanagedredisgeoreplication/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureManagedRedisGeoReplication *azuremanagedredisgeoreplicationv1alpha1.AzureManagedRedisGeoReplication
	ManagedRedisId                  string
	LinkedManagedRedisIds           []string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azuremanagedredisgeoreplicationv1alpha1.AzureManagedRedisGeoReplicationIacInput) *Locals {
	locals := &Locals{}

	locals.AzureManagedRedisGeoReplication = iacInput.Target
	locals.ManagedRedisId = iacInput.Target.Spec.ManagedRedisId.GetValue()

	for _, linkedId := range iacInput.Target.Spec.LinkedManagedRedisIds {
		locals.LinkedManagedRedisIds = append(locals.LinkedManagedRedisIds, linkedId.GetValue())
	}

	// No Azure tags: the geo-replication group has no ARM object of its
	// own (its state lives on every member's default database), so the
	// platform's identity tags live on the member clusters.

	return locals
}
