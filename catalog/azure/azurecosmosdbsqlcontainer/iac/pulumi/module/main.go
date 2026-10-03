package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurecosmosdbsqlcontainerv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecosmosdbsqlcontainer/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cosmosdb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurecosmosdbsqlcontainerv1alpha1.AzureCosmosdbSqlContainerIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureCosmosdbSqlContainer.Spec

	// The provider addresses Cosmos children by the (resource group,
	// account, database, name) tuple rather than an ARM ID, so all
	// three parents are parsed from the resolved database ID -- the
	// spec models a single parent reference and the module derives the
	// rest (no redundant, contradictable state).
	databaseName, accountName, resourceGroupName, err := parseSqlDatabaseId(locals.SqlDatabaseId)
	if err != nil {
		return err
	}

	// Unspecified kind materializes Hash -- azurerm's own default.
	// MULTI_HASH hierarchical keys require version 2; the spec enforces
	// the pairing before the deploy ever runs.
	partitionKeyKind := "Hash"
	if spec.PartitionKeyKind != azurecosmosdbsqlcontainerv1alpha1.AzureCosmosdbSqlContainerPartitionKeyKind_azure_cosmosdb_sql_container_partition_key_kind_unspecified {
		partitionKeyKind = partitionKeyKindStrings[spec.PartitionKeyKind]
	}

	// No Azure tags: ARM does not support tags on Cosmos child
	// resources, so the platform's identity tags live on the account.
	containerArgs := &cosmosdb.SqlContainerArgs{
		Name:              pulumi.String(spec.ContainerName),
		ResourceGroupName: pulumi.String(resourceGroupName),
		AccountName:       pulumi.String(accountName),
		DatabaseName:      pulumi.String(databaseName),
		// The partition key -- fixed at creation, the container's most
		// consequential design decision.
		PartitionKeyPaths: pulumi.ToStringArray(spec.PartitionKeyPaths),
		PartitionKeyKind:  pulumi.String(partitionKeyKind),
	}

	if spec.PartitionKeyVersion != nil {
		containerArgs.PartitionKeyVersion = pulumi.Int(int(spec.GetPartitionKeyVersion()))
	}

	// Dedicated throughput. Sent only when set: serverless accounts
	// reject provisioned throughput, and unset means the container
	// shares the database's throughput. The spec enforces the
	// fixed-XOR-autoscale contract.
	if spec.Throughput != nil {
		containerArgs.Throughput = pulumi.Int(int(spec.GetThroughput()))
	}
	if spec.AutoscaleMaxThroughput != nil {
		containerArgs.AutoscaleSettings = cosmosdb.SqlContainerAutoscaleSettingsArgs{
			MaxThroughput: pulumi.Int(int(spec.GetAutoscaleMaxThroughput())),
		}
	}

	// Document TTL: -1 turns TTL on with per-document expiry only; a
	// positive value expires documents after their last write.
	if spec.DefaultTtl != nil {
		containerArgs.DefaultTtl = pulumi.Int(int(spec.GetDefaultTtl()))
	}

	// Analytical-store TTL (requires analytical storage on the
	// account): -1 keeps analytical data forever. Disabling on an
	// existing container forces a replacement -- ARM's contract,
	// documented on the spec.
	if spec.AnalyticalStorageTtl != nil {
		containerArgs.AnalyticalStorageTtl = pulumi.Int(int(spec.GetAnalyticalStorageTtl()))
	}

	// Unique key constraints, scoped to the logical partition. Fixed at
	// creation.
	if len(spec.UniqueKeys) > 0 {
		uniqueKeyArray := cosmosdb.SqlContainerUniqueKeyArray{}
		for _, uniqueKey := range spec.UniqueKeys {
			uniqueKeyArray = append(uniqueKeyArray, cosmosdb.SqlContainerUniqueKeyArgs{
				Paths: pulumi.ToStringArray(uniqueKey.Paths),
			})
		}
		containerArgs.UniqueKeys = uniqueKeyArray
	}

	// The indexing policy -- the main lever for write RU cost and query
	// performance, updatable in place. When any included/excluded path
	// is declared the policy replaces Azure's index-everything default
	// wholesale.
	if spec.IndexingPolicy != nil {
		indexingMode := "consistent"
		if spec.IndexingPolicy.IndexingMode != azurecosmosdbsqlcontainerv1alpha1.AzureCosmosdbSqlContainerIndexingMode_azure_cosmosdb_sql_container_indexing_mode_unspecified {
			indexingMode = indexingModeStrings[spec.IndexingPolicy.IndexingMode]
		}
		policyArgs := cosmosdb.SqlContainerIndexingPolicyArgs{
			IndexingMode: pulumi.String(indexingMode),
		}
		if len(spec.IndexingPolicy.IncludedPaths) > 0 {
			includedArray := cosmosdb.SqlContainerIndexingPolicyIncludedPathArray{}
			for _, includedPath := range spec.IndexingPolicy.IncludedPaths {
				includedArray = append(includedArray, cosmosdb.SqlContainerIndexingPolicyIncludedPathArgs{
					Path: pulumi.String(includedPath.Path),
				})
			}
			policyArgs.IncludedPaths = includedArray
		}
		if len(spec.IndexingPolicy.ExcludedPaths) > 0 {
			excludedArray := cosmosdb.SqlContainerIndexingPolicyExcludedPathArray{}
			for _, excludedPath := range spec.IndexingPolicy.ExcludedPaths {
				excludedArray = append(excludedArray, cosmosdb.SqlContainerIndexingPolicyExcludedPathArgs{
					Path: pulumi.String(excludedPath.Path),
				})
			}
			policyArgs.ExcludedPaths = excludedArray
		}
		if len(spec.IndexingPolicy.CompositeIndexes) > 0 {
			compositeArray := cosmosdb.SqlContainerIndexingPolicyCompositeIndexArray{}
			for _, compositeIndex := range spec.IndexingPolicy.CompositeIndexes {
				entryArray := cosmosdb.SqlContainerIndexingPolicyCompositeIndexIndexArray{}
				for _, entry := range compositeIndex.Entries {
					order := "Ascending"
					if entry.Order != azurecosmosdbsqlcontainerv1alpha1.AzureCosmosdbSqlContainerCompositeIndexOrder_azure_cosmosdb_sql_container_composite_index_order_unspecified {
						order = compositeIndexOrderStrings[entry.Order]
					}
					entryArray = append(entryArray, cosmosdb.SqlContainerIndexingPolicyCompositeIndexIndexArgs{
						Path:  pulumi.String(entry.Path),
						Order: pulumi.String(order),
					})
				}
				compositeArray = append(compositeArray, cosmosdb.SqlContainerIndexingPolicyCompositeIndexArgs{
					Indices: entryArray,
				})
			}
			policyArgs.CompositeIndices = compositeArray
		}
		if len(spec.IndexingPolicy.SpatialIndexes) > 0 {
			spatialArray := cosmosdb.SqlContainerIndexingPolicySpatialIndexArray{}
			for _, spatialIndex := range spec.IndexingPolicy.SpatialIndexes {
				spatialArray = append(spatialArray, cosmosdb.SqlContainerIndexingPolicySpatialIndexArgs{
					Path: pulumi.String(spatialIndex.Path),
				})
			}
			policyArgs.SpatialIndices = spatialArray
		}
		containerArgs.IndexingPolicy = policyArgs
	}

	// Conflict resolution for multi-region-write accounts. Fixed at
	// creation; the per-mode field pairing is enforced by the spec.
	if spec.ConflictResolutionPolicy != nil {
		conflictArgs := cosmosdb.SqlContainerConflictResolutionPolicyArgs{
			Mode: pulumi.String(conflictResolutionModeStrings[spec.ConflictResolutionPolicy.Mode]),
		}
		if spec.ConflictResolutionPolicy.ConflictResolutionPath != "" {
			conflictArgs.ConflictResolutionPath = pulumi.String(spec.ConflictResolutionPolicy.ConflictResolutionPath)
		}
		if spec.ConflictResolutionPolicy.ConflictResolutionProcedure != "" {
			conflictArgs.ConflictResolutionProcedure = pulumi.String(spec.ConflictResolutionPolicy.ConflictResolutionProcedure)
		}
		containerArgs.ConflictResolutionPolicy = conflictArgs
	}

	createdContainer, err := cosmosdb.NewSqlContainer(ctx,
		fmt.Sprintf("%s-%s", databaseName, spec.ContainerName),
		containerArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create cosmosdb sql container %s", spec.ContainerName)
	}

	// Export outputs. No endpoint or credential outputs on
	// purpose: connectivity and keys live on the ACCOUNT; the container
	// is addressed inside that connection by database and container
	// name.
	ctx.Export(OpSqlContainerId, createdContainer.ID())
	ctx.Export(OpSqlContainerName, createdContainer.Name)
	ctx.Export(OpSqlDatabaseName, pulumi.String(databaseName))
	ctx.Export(OpCosmosdbAccountName, pulumi.String(accountName))

	return nil
}
