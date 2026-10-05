package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurecosmosdbmongocollectionv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecosmosdbmongocollection/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cosmosdb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurecosmosdbmongocollectionv1alpha1.AzureCosmosdbMongoCollectionIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureCosmosdbMongoCollection.Spec

	// The provider addresses Cosmos children by the (resource group,
	// account, database, name) tuple rather than an ARM ID, so all
	// three parents are parsed from the resolved database ID -- the
	// spec models a single parent reference and the module derives the
	// rest (no redundant, contradictable state).
	databaseName, accountName, resourceGroupName, err := parseMongoDatabaseId(locals.MongoDatabaseId)
	if err != nil {
		return err
	}

	// No Azure tags: ARM does not support tags on Cosmos child
	// resources, so the platform's identity tags live on the account.
	collectionArgs := &cosmosdb.MongoCollectionArgs{
		Name:              pulumi.String(spec.CollectionName),
		ResourceGroupName: pulumi.String(resourceGroupName),
		AccountName:       pulumi.String(accountName),
		DatabaseName:      pulumi.String(databaseName),
	}

	// The shard key -- the MongoDB face of the partition key, fixed at
	// creation. Sent only when set: empty creates an unsharded
	// collection confined to one physical partition.
	if spec.ShardKey != "" {
		collectionArgs.ShardKey = pulumi.String(spec.ShardKey)
	}

	// Dedicated throughput. Sent only when set: serverless accounts
	// reject provisioned throughput, and unset means the collection
	// shares the database's throughput. The spec enforces the
	// fixed-XOR-autoscale contract.
	if spec.Throughput != nil {
		collectionArgs.Throughput = pulumi.Int(int(spec.GetThroughput()))
	}
	if spec.AutoscaleMaxThroughput != nil {
		collectionArgs.AutoscaleSettings = cosmosdb.MongoCollectionAutoscaleSettingsArgs{
			MaxThroughput: pulumi.Int(int(spec.GetAutoscaleMaxThroughput())),
		}
	}

	// Document TTL (implemented by Cosmos DB as an expireAfter index on
	// _ts): -1 turns TTL on with per-document expiry only; never 0 (the
	// spec rejects it -- ARM's contract).
	if spec.DefaultTtlSeconds != nil {
		collectionArgs.DefaultTtlSeconds = pulumi.Int(int(spec.GetDefaultTtlSeconds()))
	}

	// Analytical-store TTL (requires analytical storage on the
	// account): -1 keeps analytical data forever.
	if spec.AnalyticalStorageTtl != nil {
		collectionArgs.AnalyticalStorageTtl = pulumi.Int(int(spec.GetAnalyticalStorageTtl()))
	}

	// Indexes, including the ["_id"] unique index Azure requires on
	// every Mongo collection (spec-enforced) -- declared explicitly,
	// never injected.
	if len(spec.Indexes) > 0 {
		indexArray := cosmosdb.MongoCollectionIndexArray{}
		for _, index := range spec.Indexes {
			indexArray = append(indexArray, cosmosdb.MongoCollectionIndexArgs{
				Keys:   pulumi.ToStringArray(index.Keys),
				Unique: pulumi.Bool(index.GetUnique()),
			})
		}
		collectionArgs.Indices = indexArray
	}

	createdCollection, err := cosmosdb.NewMongoCollection(ctx,
		fmt.Sprintf("%s-%s", databaseName, spec.CollectionName),
		collectionArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create cosmosdb mongo collection %s", spec.CollectionName)
	}

	// Export outputs. No endpoint or credential outputs on
	// purpose: connectivity and the MongoDB connection strings live on
	// the ACCOUNT; the collection is addressed inside that connection
	// by database and collection name.
	ctx.Export(OpMongoCollectionId, createdCollection.ID())
	ctx.Export(OpMongoCollectionName, createdCollection.Name)
	ctx.Export(OpMongoDatabaseName, pulumi.String(databaseName))
	ctx.Export(OpCosmosdbAccountName, pulumi.String(accountName))

	return nil
}
