package module

import (
	"github.com/pkg/errors"
	azuremanagedredisv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremanagedredis/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/managedredis"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremanagedredisv1alpha1.AzureManagedRedisIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureManagedRedis.Spec

	// The Managed Redis instance: the cluster (compute, load balancer,
	// TLS endpoint) plus its default database (the Redis process), which
	// Azure maps 1-to-1 -- one resource here, exactly as ARM provisions
	// them. Provisioning polls the cluster to its Running state and then
	// creates the database; expect tens of minutes end to end.
	instanceArgs := &managedredis.ManagedRedisArgs{
		Name:              pulumi.String(spec.ClusterName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		// One value picks the tier family and memory size. Azure
		// validates in-place SKU changes against the live instance at
		// apply time and replaces the instance when the target is not
		// scalable.
		SkuName: pulumi.String(locals.SkuName),
		Tags:    pulumi.ToStringMap(locals.AzureTags),
	}

	// Presence-guarded proto defaults: IaC inputs never materialize
	// them, so an unset field must deploy the spec's documented default,
	// not the Go zero value.

	// HA runs a replica and carries the zone-redundant SLA; disabling it
	// halves cost for dev/test. Fixed at creation.
	highAvailability := true
	if spec.HighAvailabilityEnabled != nil {
		highAvailability = spec.GetHighAvailabilityEnabled()
	}
	instanceArgs.HighAvailabilityEnabled = pulumi.BoolPtr(highAvailability)

	// false forces all traffic through Private Link
	// (AzurePrivateEndpoint) -- Managed Redis has no VNet injection or
	// IP firewall. The provider models this as an Enabled/Disabled
	// string; the spec's bool maps onto it.
	publicNetworkAccess := "Enabled"
	if spec.PublicNetworkAccessEnabled != nil && !spec.GetPublicNetworkAccessEnabled() {
		publicNetworkAccess = "Disabled"
	}
	instanceArgs.PublicNetworkAccess = pulumi.StringPtr(publicNetworkAccess)

	// Customer-managed-key encryption. The key id is the VERSIONED Key
	// Vault id (rotation = updating the reference); the same identity
	// must also be attached below -- an ARM pairing enforced at apply
	// time.
	if spec.CustomerManagedKey != nil {
		instanceArgs.CustomerManagedKey = &managedredis.ManagedRedisCustomerManagedKeyArgs{
			KeyVaultKeyId:          pulumi.String(spec.CustomerManagedKey.KeyVaultKeyId.GetValue()),
			UserAssignedIdentityId: pulumi.String(spec.CustomerManagedKey.UserAssignedIdentityId.GetValue()),
		}
	}

	// The managed identity -- what customer-managed-key encryption
	// authenticates to Key Vault with.
	if spec.Identity != nil {
		identityArgs := &managedredis.ManagedRedisIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(spec.Identity.UserAssignedIdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, id := range spec.Identity.UserAssignedIdentityIds {
				identityIds = append(identityIds, pulumi.String(id.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		instanceArgs.Identity = identityArgs
	}

	// The Redis process itself. Required at create (Azure rejects a
	// database-less cluster). The database enums deploy Azure's own
	// defaults explicitly so both engines send identical bodies;
	// changing clustering_policy, geo_replication_group_name, or the
	// module set recreates the DATABASE in place (data loss, brief
	// unavailability) without replacing the cluster.
	database := spec.DefaultDatabase
	databaseArgs := &managedredis.ManagedRedisDefaultDatabaseArgs{
		// Keyless-first: keys are OFF by default; Entra grants
		// (AzureManagedRedisAccessPolicyAssignment) are how clients
		// connect.
		AccessKeysAuthenticationEnabled: pulumi.BoolPtr(database.AccessKeysAuthenticationEnabled),
	}

	// Enum defaults materialized explicitly (Encrypted / OSSCluster /
	// VolatileLRU are Azure's own defaults).
	clientProtocol := "Encrypted"
	if database.ClientProtocol != azuremanagedredisv1alpha1.AzureManagedRedisClientProtocol_azure_managed_redis_client_protocol_unspecified {
		clientProtocol = clientProtocolStrings[database.ClientProtocol]
	}
	databaseArgs.ClientProtocol = pulumi.StringPtr(clientProtocol)

	clusteringPolicy := "OSSCluster"
	if database.ClusteringPolicy != azuremanagedredisv1alpha1.AzureManagedRedisClusteringPolicy_azure_managed_redis_clustering_policy_unspecified {
		clusteringPolicy = clusteringPolicyStrings[database.ClusteringPolicy]
	}
	databaseArgs.ClusteringPolicy = pulumi.StringPtr(clusteringPolicy)

	evictionPolicy := "VolatileLRU"
	if database.EvictionPolicy != azuremanagedredisv1alpha1.AzureManagedRedisEvictionPolicy_azure_managed_redis_eviction_policy_unspecified {
		evictionPolicy = evictionPolicyStrings[database.EvictionPolicy]
	}
	databaseArgs.EvictionPolicy = pulumi.StringPtr(evictionPolicy)

	// Joining a named ACTIVE geo-replication group; membership is
	// linked by AzureManagedRedisGeoReplication.
	if database.GeoReplicationGroupName != "" {
		databaseArgs.GeoReplicationGroupName = pulumi.StringPtr(database.GeoReplicationGroupName)
	}

	// Redis modules (search, JSON, bloom, time series) -- capabilities
	// classic Redis never had.
	if len(database.Modules) > 0 {
		moduleArray := managedredis.ManagedRedisDefaultDatabaseModuleArray{}
		for _, specModule := range database.Modules {
			moduleArgs := &managedredis.ManagedRedisDefaultDatabaseModuleArgs{
				Name: pulumi.String(specModule.Name),
			}
			if specModule.Args != "" {
				moduleArgs.Args = pulumi.StringPtr(specModule.Args)
			}
			moduleArray = append(moduleArray, moduleArgs)
		}
		databaseArgs.Modules = moduleArray
	}

	// Setting a frequency ENABLES the matching persistence method. AOF
	// and RDB are mutually exclusive, and both conflict with
	// geo-replication (spec-enforced, mirroring Azure's own contract).
	if database.PersistenceAppendOnlyFileBackupFrequency != nil {
		databaseArgs.PersistenceAppendOnlyFileBackupFrequency = pulumi.StringPtr(database.GetPersistenceAppendOnlyFileBackupFrequency())
	}
	if database.PersistenceRedisDatabaseBackupFrequency != nil {
		databaseArgs.PersistenceRedisDatabaseBackupFrequency = pulumi.StringPtr(database.GetPersistenceRedisDatabaseBackupFrequency())
	}

	instanceArgs.DefaultDatabase = databaseArgs

	createdInstance, err := managedredis.NewManagedRedis(ctx,
		spec.ClusterName,
		instanceArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create managed redis instance %s", spec.ClusterName)
	}

	// The database-derived outputs ride the default_database block; each
	// applier handles the block's absence defensively (it is always
	// present on create, but the typed pointer output requires it).
	databaseId := createdInstance.DefaultDatabase.ApplyT(func(database *managedredis.ManagedRedisDefaultDatabase) string {
		if database == nil || database.Id == nil {
			return ""
		}
		return *database.Id
	}).(pulumi.StringOutput)

	databasePort := createdInstance.DefaultDatabase.ApplyT(func(database *managedredis.ManagedRedisDefaultDatabase) int {
		if database == nil || database.Port == nil {
			return 0
		}
		return *database.Port
	}).(pulumi.IntOutput)

	// The keys are secret-bearing: they are the database password. Both
	// stay empty under the keyless default (access keys disabled).
	primaryAccessKey := createdInstance.DefaultDatabase.ApplyT(func(database *managedredis.ManagedRedisDefaultDatabase) string {
		if database == nil || database.PrimaryAccessKey == nil {
			return ""
		}
		return *database.PrimaryAccessKey
	}).(pulumi.StringOutput)

	secondaryAccessKey := createdInstance.DefaultDatabase.ApplyT(func(database *managedredis.ManagedRedisDefaultDatabase) string {
		if database == nil || database.SecondaryAccessKey == nil {
			return ""
		}
		return *database.SecondaryAccessKey
	}).(pulumi.StringOutput)

	// The system-assigned identity's principal, when one exists -- what
	// RBAC grants target.
	identityPrincipalId := createdInstance.Identity.ApplyT(func(identity *managedredis.ManagedRedisIdentity) string {
		if identity == nil || identity.PrincipalId == nil {
			return ""
		}
		return *identity.PrincipalId
	}).(pulumi.StringOutput)

	// Export outputs. The access keys are secret-bearing.
	ctx.Export(OpManagedRedisId, createdInstance.ID())
	ctx.Export(OpManagedRedisName, createdInstance.Name)
	ctx.Export(OpRegion, createdInstance.Location)
	ctx.Export(OpResourceGroupName, createdInstance.ResourceGroupName)
	ctx.Export(OpHostname, createdInstance.Hostname)
	ctx.Export(OpDatabaseId, databaseId)
	ctx.Export(OpPort, databasePort)
	ctx.Export(OpPrimaryAccessKey, pulumi.ToSecret(primaryAccessKey))
	ctx.Export(OpSecondaryAccessKey, pulumi.ToSecret(secondaryAccessKey))
	ctx.Export(OpIdentityPrincipalId, identityPrincipalId)

	return nil
}
