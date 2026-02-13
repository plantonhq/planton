package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurecosmosdbaccountv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurecosmosdbaccount/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cosmosdb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurecosmosdbaccountv1.AzureCosmosdbAccountStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	azureProviderConfig := stackInput.ProviderConfig

	// Create azure provider using the credentials from the input
	azureProvider, err := azure.NewProvider(ctx,
		"azure",
		&azure.ProviderArgs{
			ClientId:       pulumi.String(azureProviderConfig.ClientId),
			ClientSecret:   pulumi.String(azureProviderConfig.ClientSecret),
			SubscriptionId: pulumi.String(azureProviderConfig.SubscriptionId),
			TenantId:       pulumi.String(azureProviderConfig.TenantId),
		})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureCosmosdbAccount.Spec

	// Resolve kind with default
	kind := spec.GetKind()
	if kind == "" {
		kind = "GlobalDocumentDB"
	}

	// Build capabilities array; if kind is MongoDB, auto-add EnableMongo if not present
	capabilities := spec.GetCapabilities()
	hasEnableMongo := false
	for _, c := range capabilities {
		if c == "EnableMongo" {
			hasEnableMongo = true
			break
		}
	}
	if kind == "MongoDB" && !hasEnableMongo {
		capabilities = append([]string{"EnableMongo"}, capabilities...)
	}

	// Build consistency policy block
	consistencyPolicy := &cosmosdb.AccountConsistencyPolicyArgs{
		ConsistencyLevel: pulumi.String("Session"),
	}
	if spec.ConsistencyPolicy != nil {
		cl := spec.ConsistencyPolicy.GetConsistencyLevel()
		if cl != "" {
			consistencyPolicy.ConsistencyLevel = pulumi.String(cl)
		}
		if cl == "BoundedStaleness" {
			mi := spec.ConsistencyPolicy.GetMaxIntervalInSeconds()
			if mi > 0 {
				consistencyPolicy.MaxIntervalInSeconds = pulumi.Int(int(mi))
			} else {
				consistencyPolicy.MaxIntervalInSeconds = pulumi.Int(5)
			}
			mp := spec.ConsistencyPolicy.GetMaxStalenessPrefix()
			if mp > 0 {
				consistencyPolicy.MaxStalenessPrefix = pulumi.Int(int(mp))
			} else {
				consistencyPolicy.MaxStalenessPrefix = pulumi.Int(100)
			}
		}
	}

	// Build geo_location array
	geoLocations := make(cosmosdb.AccountGeoLocationArray, 0, len(spec.GeoLocations))
	for _, gl := range spec.GeoLocations {
		geoLocations = append(geoLocations, &cosmosdb.AccountGeoLocationArgs{
			Location:         pulumi.String(gl.Location),
			FailoverPriority: pulumi.Int(int(gl.FailoverPriority)),
			ZoneRedundant:    pulumi.BoolPtr(gl.GetZoneRedundant()),
		})
	}

	// Build capabilities array for Pulumi
	capabilityArgs := make(cosmosdb.AccountCapabilityArray, 0, len(capabilities))
	for _, cap := range capabilities {
		capabilityArgs = append(capabilityArgs, &cosmosdb.AccountCapabilityArgs{
			Name: pulumi.String(cap),
		})
	}

	// Build virtual network rules
	var vnetRules cosmosdb.AccountVirtualNetworkRuleArray
	for _, rule := range spec.GetVirtualNetworkRules() {
		if rule != nil && rule.SubnetId != nil {
			subnetId := rule.SubnetId.GetValue()
			if subnetId != "" {
				vnetRules = append(vnetRules, &cosmosdb.AccountVirtualNetworkRuleArgs{
					Id: pulumi.String(subnetId),
				})
			}
		}
	}

	// Build IP range filters (Pulumi uses IpRangeFilters)
	var ipRangeFilters pulumi.StringArray
	for _, ip := range spec.GetIpRangeFilter() {
		ipRangeFilters = append(ipRangeFilters, pulumi.String(ip))
	}

	// Build backup policy
	var backupArgs *cosmosdb.AccountBackupArgs
	if spec.Backup != nil {
		backupArgs = &cosmosdb.AccountBackupArgs{
			Type: pulumi.String(spec.Backup.Type),
		}
		if spec.Backup.Type == "Periodic" {
			if spec.Backup.IntervalInMinutes != nil {
				backupArgs.IntervalInMinutes = pulumi.Int(int(*spec.Backup.IntervalInMinutes))
			}
			if spec.Backup.RetentionInHours != nil {
				backupArgs.RetentionInHours = pulumi.Int(int(*spec.Backup.RetentionInHours))
			}
			if spec.Backup.StorageRedundancy != nil && *spec.Backup.StorageRedundancy != "" {
				backupArgs.StorageRedundancy = pulumi.String(*spec.Backup.StorageRedundancy)
			}
		}
		if spec.Backup.Type == "Continuous" && spec.Backup.Tier != nil && *spec.Backup.Tier != "" {
			backupArgs.Tier = pulumi.String(*spec.Backup.Tier)
		}
	}

	// Build account args
	accountArgs := &cosmosdb.AccountArgs{
		Name:                              pulumi.String(spec.Name),
		Location:                         pulumi.String(spec.Region),
		ResourceGroupName:                pulumi.String(locals.ResourceGroupName),
		OfferType:                        pulumi.String("Standard"),
		Kind:                              pulumi.String(kind),
		ConsistencyPolicy:                 consistencyPolicy,
		GeoLocations:                     geoLocations,
		Capabilities:                     capabilityArgs,
		FreeTierEnabled:                  pulumi.BoolPtr(spec.GetFreeTierEnabled()),
		AutomaticFailoverEnabled:         pulumi.BoolPtr(spec.GetAutomaticFailoverEnabled()),
		MultipleWriteLocationsEnabled:    pulumi.BoolPtr(spec.GetMultipleWriteLocationsEnabled()),
		PublicNetworkAccessEnabled:       pulumi.BoolPtr(spec.GetPublicNetworkAccessEnabled()),
		IsVirtualNetworkFilterEnabled:    pulumi.BoolPtr(spec.GetIsVirtualNetworkFilterEnabled()),
		Tags:                             pulumi.ToStringMap(locals.AzureTags),
	}

	if len(vnetRules) > 0 {
		accountArgs.VirtualNetworkRules = vnetRules
	}
	if len(ipRangeFilters) > 0 {
		accountArgs.IpRangeFilters = ipRangeFilters
	}
	if backupArgs != nil {
		accountArgs.Backup = backupArgs
	}
	if kind == "MongoDB" && spec.GetMongoServerVersion() != "" {
		accountArgs.MongoServerVersion = pulumi.StringPtr(spec.GetMongoServerVersion())
	}

	// Create Cosmos DB Account
	account, err := cosmosdb.NewAccount(ctx,
		spec.Name,
		accountArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Cosmos DB Account %s", spec.Name)
	}

	databaseIdMap := make(map[string]pulumi.StringOutput)

	if kind == "GlobalDocumentDB" || kind == "" {
		// SQL API: create sql_databases and containers
		for _, db := range spec.GetSqlDatabases() {
			dbArgs := &cosmosdb.SqlDatabaseArgs{
				Name:               pulumi.String(db.Name),
				AccountName:        account.Name,
				ResourceGroupName:  pulumi.String(locals.ResourceGroupName),
			}
			if db.Throughput != nil {
				dbArgs.Throughput = pulumi.IntPtr(int(*db.Throughput))
			}
			if db.AutoscaleMaxThroughput != nil {
				dbArgs.AutoscaleSettings = &cosmosdb.SqlDatabaseAutoscaleSettingsArgs{
					MaxThroughput: pulumi.Int(int(*db.AutoscaleMaxThroughput)),
				}
			}

			database, err := cosmosdb.NewSqlDatabase(ctx,
				fmt.Sprintf("%s-%s", spec.Name, db.Name),
				dbArgs,
				pulumi.Provider(azureProvider),
				pulumi.DependsOn([]pulumi.Resource{account}))
			if err != nil {
				return errors.Wrapf(err, "failed to create SQL database %s", db.Name)
			}
			databaseIdMap[db.Name] = database.ID().ToStringOutput()

			for _, c := range db.GetContainers() {
				pkKind := c.GetPartitionKeyKind()
				if pkKind == "" {
					pkKind = "Hash"
				}
				pkVersion := 1
				if pkKind == "MultiHash" {
					pkVersion = 2
				}
				containerArgs := &cosmosdb.SqlContainerArgs{
					Name:                pulumi.String(c.Name),
					AccountName:         account.Name,
					ResourceGroupName:   pulumi.String(locals.ResourceGroupName),
					DatabaseName:        database.Name,
					PartitionKeyPaths:   pulumi.ToStringArray(c.PartitionKeyPaths),
					PartitionKeyKind:    pulumi.String(pkKind),
					PartitionKeyVersion: pulumi.Int(pkVersion),
				}
				if c.Throughput != nil {
					containerArgs.Throughput = pulumi.IntPtr(int(*c.Throughput))
				}
				if c.AutoscaleMaxThroughput != nil {
					containerArgs.AutoscaleSettings = &cosmosdb.SqlContainerAutoscaleSettingsArgs{
						MaxThroughput: pulumi.Int(int(*c.AutoscaleMaxThroughput)),
					}
				}
				if c.DefaultTtl != nil {
					containerArgs.DefaultTtl = pulumi.IntPtr(int(*c.DefaultTtl))
				}

				_, err = cosmosdb.NewSqlContainer(ctx,
					fmt.Sprintf("%s-%s-%s", spec.Name, db.Name, c.Name),
					containerArgs,
					pulumi.Provider(azureProvider),
					pulumi.DependsOn([]pulumi.Resource{database}))
				if err != nil {
					return errors.Wrapf(err, "failed to create SQL container %s", c.Name)
				}
			}
		}
	} else if kind == "MongoDB" {
		// MongoDB API: create mongo_databases and collections
		for _, db := range spec.GetMongoDatabases() {
			dbArgs := &cosmosdb.MongoDatabaseArgs{
				Name:              pulumi.String(db.Name),
				AccountName:       account.Name,
				ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			}
			if db.Throughput != nil {
				dbArgs.Throughput = pulumi.IntPtr(int(*db.Throughput))
			}
			if db.AutoscaleMaxThroughput != nil {
				dbArgs.AutoscaleSettings = &cosmosdb.MongoDatabaseAutoscaleSettingsArgs{
					MaxThroughput: pulumi.Int(int(*db.AutoscaleMaxThroughput)),
				}
			}

			database, err := cosmosdb.NewMongoDatabase(ctx,
				fmt.Sprintf("%s-%s", spec.Name, db.Name),
				dbArgs,
				pulumi.Provider(azureProvider),
				pulumi.DependsOn([]pulumi.Resource{account}))
			if err != nil {
				return errors.Wrapf(err, "failed to create MongoDB database %s", db.Name)
			}
			databaseIdMap[db.Name] = database.ID().ToStringOutput()

			for _, col := range db.GetCollections() {
				collectionArgs := &cosmosdb.MongoCollectionArgs{
					Name:              pulumi.String(col.Name),
					AccountName:       account.Name,
					ResourceGroupName: pulumi.String(locals.ResourceGroupName),
					DatabaseName:      database.Name,
					ShardKey:          pulumi.String(col.ShardKey),
				}
				if col.Throughput != nil {
					collectionArgs.Throughput = pulumi.IntPtr(int(*col.Throughput))
				}
				if col.AutoscaleMaxThroughput != nil {
					collectionArgs.AutoscaleSettings = &cosmosdb.MongoCollectionAutoscaleSettingsArgs{
						MaxThroughput: pulumi.Int(int(*col.AutoscaleMaxThroughput)),
					}
				}
				if col.DefaultTtlSeconds != nil {
					collectionArgs.DefaultTtlSeconds = pulumi.IntPtr(int(*col.DefaultTtlSeconds))
				}
				if len(col.GetIndexes()) > 0 {
					indexArgs := make(cosmosdb.MongoCollectionIndexArray, 0, len(col.Indexes))
					for _, idx := range col.Indexes {
						indexArgs = append(indexArgs, &cosmosdb.MongoCollectionIndexArgs{
							Keys:   pulumi.ToStringArray(idx.Keys),
							Unique: pulumi.BoolPtr(idx.GetUnique()),
						})
					}
					collectionArgs.Indices = indexArgs
				}

				_, err = cosmosdb.NewMongoCollection(ctx,
					fmt.Sprintf("%s-%s-%s", spec.Name, db.Name, col.Name),
					collectionArgs,
					pulumi.Provider(azureProvider),
					pulumi.DependsOn([]pulumi.Resource{database}))
				if err != nil {
					return errors.Wrapf(err, "failed to create MongoDB collection %s", col.Name)
				}
			}
		}
	}

	// Export outputs
	ctx.Export(OpAccountId, account.ID())
	ctx.Export(OpAccountName, account.Name)
	ctx.Export(OpEndpoint, account.Endpoint)
	ctx.Export(OpPrimaryKey, account.PrimaryKey)
	ctx.Export(OpPrimaryConnectionString, account.PrimarySqlConnectionString)
	ctx.Export(OpPrimaryMongodbConnectionString, account.PrimaryMongodbConnectionString)

	if len(databaseIdMap) > 0 {
		dbIdMapOutput := pulumi.StringMap{}
		for name, id := range databaseIdMap {
			dbIdMapOutput[name] = id
		}
		ctx.Export(OpDatabaseIds, dbIdMapOutput)
	}

	return nil
}
