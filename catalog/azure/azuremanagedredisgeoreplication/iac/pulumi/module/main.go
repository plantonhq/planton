package module

import (
	"github.com/pkg/errors"
	azuremanagedredisgeoreplicationv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremanagedredisgeoreplication/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/managedredis"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremanagedredisgeoreplicationv1alpha1.AzureManagedRedisGeoReplicationIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	// The active geo-replication group: links Managed Redis instances
	// whose default databases declare the same geo_replication_group_name
	// into a multi-primary replica set -- every member accepts writes
	// and Azure merges the datasets with conflict-free semantics.
	//
	// Membership is managed as its own resource because linking mutates
	// the replication state of EVERY member out of band -- it is a
	// group-wide operation performed through one member, not a property
	// of any single instance. ONE resource manages the whole group
	// (linking is reciprocal; never create one per member). Deleting it
	// force-unlinks the members, each keeping its own copy of the data;
	// removing a single ID from the linked list evacuates just that
	// member.
	//
	// Cross-resource contracts Azure enforces at link time: every member
	// carries the SAME group name, is BALANCED_B3 or larger, has no
	// persistence, and uses only the RediSearch/RedisJSON modules. No
	// tags: the group has no ARM object of its own.
	createdGeoReplication, err := managedredis.NewGeoReplication(ctx,
		locals.AzureManagedRedisGeoReplication.Metadata.Name,
		&managedredis.GeoReplicationArgs{
			ManagedRedisId:        pulumi.String(locals.ManagedRedisId),
			LinkedManagedRedisIds: pulumi.ToStringArray(locals.LinkedManagedRedisIds),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create managed redis geo-replication group for %s", locals.ManagedRedisId)
	}

	// Export outputs. The group's resource ID is the managing
	// cluster's ARM ID (the group has no ARM object of its own).
	ctx.Export(OpGeoReplicationId, createdGeoReplication.ID())

	return nil
}
