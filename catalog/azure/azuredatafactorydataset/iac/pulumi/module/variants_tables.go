package module

import (
	"github.com/pkg/errors"
	azuredatafactorydatasetv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredatafactorydataset/v1alpha1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/datafactory"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The table-form builders (Azure SQL, Cosmos DB, MySQL, PostgreSQL,
// Snowflake, SQL Server) and the raw-JSON custom form.

func createAzureSqlTable(
	ctx *pulumi.Context,
	resourceName string,
	spec *azuredatafactorydatasetv1alpha1.AzureDataFactoryDatasetSpec,
	azureProvider pulumi.ProviderResource,
) (pulumi.StringInput, pulumi.StringInput, error) {
	sqlTable := spec.GetAzureSqlTable()

	// The one variant that references its linked service by ARM ID --
	// Azure requires it to belong to the same factory as the dataset.
	args := &datafactory.DatasetAzureSqlTableArgs{
		Name:                 pulumi.String(spec.Name),
		DataFactoryId:        pulumi.String(spec.DataFactoryId.GetValue()),
		LinkedServiceId:      pulumi.String(sqlTable.LinkedServiceId.GetValue()),
		Description:          descriptionPtr(spec),
		Annotations:          annotationsArray(spec),
		Parameters:           parametersMap(spec),
		AdditionalProperties: additionalPropertiesMap(spec),
		Folder:               folderPtr(spec),
		Schema:               stringPtrWhenSet(sqlTable.Schema),
		Table:                stringPtrWhenSet(sqlTable.Table),
	}

	if len(sqlTable.SchemaColumn) > 0 {
		columns := make(datafactory.DatasetAzureSqlTableSchemaColumnArray, 0, len(sqlTable.SchemaColumn))
		for _, column := range sqlTable.SchemaColumn {
			columns = append(columns, datafactory.DatasetAzureSqlTableSchemaColumnArgs{
				Name:        pulumi.String(column.Name),
				Type:        stringPtrWhenSet(column.Type),
				Description: stringPtrWhenSet(column.Description),
			})
		}
		args.SchemaColumns = columns
	}

	created, err := datafactory.NewDatasetAzureSqlTable(ctx, resourceName, args, pulumi.Provider(azureProvider))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create azure sql table dataset %s", resourceName)
	}
	return created.ID(), created.Name, nil
}

func createCosmosdbSqlapi(
	ctx *pulumi.Context,
	resourceName string,
	spec *azuredatafactorydatasetv1alpha1.AzureDataFactoryDatasetSpec,
	azureProvider pulumi.ProviderResource,
) (pulumi.StringInput, pulumi.StringInput, error) {
	cosmosdb := spec.GetCosmosdbSqlapi()

	args := &datafactory.DatasetCosmosDBApiArgs{
		Name:                 pulumi.String(spec.Name),
		DataFactoryId:        pulumi.String(spec.DataFactoryId.GetValue()),
		LinkedServiceName:    linkedServiceName(spec),
		Description:          descriptionPtr(spec),
		Annotations:          annotationsArray(spec),
		Parameters:           parametersMap(spec),
		AdditionalProperties: additionalPropertiesMap(spec),
		Folder:               folderPtr(spec),
		CollectionName:       stringPtrWhenSet(cosmosdb.CollectionName),
	}

	if len(cosmosdb.SchemaColumn) > 0 {
		columns := make(datafactory.DatasetCosmosDBApiSchemaColumnArray, 0, len(cosmosdb.SchemaColumn))
		for _, column := range cosmosdb.SchemaColumn {
			columns = append(columns, datafactory.DatasetCosmosDBApiSchemaColumnArgs{
				Name:        pulumi.String(column.Name),
				Type:        stringPtrWhenSet(column.Type),
				Description: stringPtrWhenSet(column.Description),
			})
		}
		args.SchemaColumns = columns
	}

	created, err := datafactory.NewDatasetCosmosDBApi(ctx, resourceName, args, pulumi.Provider(azureProvider))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create cosmosdb sqlapi dataset %s", resourceName)
	}
	return created.ID(), created.Name, nil
}

func createCustom(
	ctx *pulumi.Context,
	resourceName string,
	spec *azuredatafactorydatasetv1alpha1.AzureDataFactoryDatasetSpec,
	azureProvider pulumi.ProviderResource,
) (pulumi.StringInput, pulumi.StringInput, error) {
	custom := spec.GetCustom()

	linkedService := datafactory.CustomDatasetLinkedServiceArgs{
		Name: pulumi.String(custom.LinkedService.Name.GetValue()),
	}
	if len(custom.LinkedService.Parameters) > 0 {
		linkedService.Parameters = pulumi.ToStringMap(custom.LinkedService.Parameters)
	}

	args := &datafactory.CustomDatasetArgs{
		Name:                 pulumi.String(spec.Name),
		DataFactoryId:        pulumi.String(spec.DataFactoryId.GetValue()),
		LinkedService:        linkedService,
		Type:                 pulumi.String(custom.Type),
		TypePropertiesJson:   pulumi.String(custom.TypePropertiesJson),
		SchemaJson:           stringPtrWhenSet(custom.SchemaJson),
		Description:          descriptionPtr(spec),
		Annotations:          annotationsArray(spec),
		Parameters:           parametersMap(spec),
		AdditionalProperties: additionalPropertiesMap(spec),
		Folder:               folderPtr(spec),
	}

	created, err := datafactory.NewCustomDataset(ctx, resourceName, args, pulumi.Provider(azureProvider))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create custom dataset %s", resourceName)
	}
	return created.ID(), created.Name, nil
}

func createMysql(
	ctx *pulumi.Context,
	resourceName string,
	spec *azuredatafactorydatasetv1alpha1.AzureDataFactoryDatasetSpec,
	azureProvider pulumi.ProviderResource,
) (pulumi.StringInput, pulumi.StringInput, error) {
	mysql := spec.GetMysql()

	args := &datafactory.DatasetMysqlArgs{
		Name:                 pulumi.String(spec.Name),
		DataFactoryId:        pulumi.String(spec.DataFactoryId.GetValue()),
		LinkedServiceName:    linkedServiceName(spec),
		Description:          descriptionPtr(spec),
		Annotations:          annotationsArray(spec),
		Parameters:           parametersMap(spec),
		AdditionalProperties: additionalPropertiesMap(spec),
		Folder:               folderPtr(spec),
		TableName:            stringPtrWhenSet(mysql.TableName),
	}

	if len(mysql.SchemaColumn) > 0 {
		columns := make(datafactory.DatasetMysqlSchemaColumnArray, 0, len(mysql.SchemaColumn))
		for _, column := range mysql.SchemaColumn {
			columns = append(columns, datafactory.DatasetMysqlSchemaColumnArgs{
				Name:        pulumi.String(column.Name),
				Type:        stringPtrWhenSet(column.Type),
				Description: stringPtrWhenSet(column.Description),
			})
		}
		args.SchemaColumns = columns
	}

	created, err := datafactory.NewDatasetMysql(ctx, resourceName, args, pulumi.Provider(azureProvider))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create mysql dataset %s", resourceName)
	}
	return created.ID(), created.Name, nil
}

func createPostgresql(
	ctx *pulumi.Context,
	resourceName string,
	spec *azuredatafactorydatasetv1alpha1.AzureDataFactoryDatasetSpec,
	azureProvider pulumi.ProviderResource,
) (pulumi.StringInput, pulumi.StringInput, error) {
	postgresql := spec.GetPostgresql()

	args := &datafactory.DatasetPostgresqlArgs{
		Name:                 pulumi.String(spec.Name),
		DataFactoryId:        pulumi.String(spec.DataFactoryId.GetValue()),
		LinkedServiceName:    linkedServiceName(spec),
		Description:          descriptionPtr(spec),
		Annotations:          annotationsArray(spec),
		Parameters:           parametersMap(spec),
		AdditionalProperties: additionalPropertiesMap(spec),
		Folder:               folderPtr(spec),
		TableName:            stringPtrWhenSet(postgresql.TableName),
	}

	if len(postgresql.SchemaColumn) > 0 {
		columns := make(datafactory.DatasetPostgresqlSchemaColumnArray, 0, len(postgresql.SchemaColumn))
		for _, column := range postgresql.SchemaColumn {
			columns = append(columns, datafactory.DatasetPostgresqlSchemaColumnArgs{
				Name:        pulumi.String(column.Name),
				Type:        stringPtrWhenSet(column.Type),
				Description: stringPtrWhenSet(column.Description),
			})
		}
		args.SchemaColumns = columns
	}

	created, err := datafactory.NewDatasetPostgresql(ctx, resourceName, args, pulumi.Provider(azureProvider))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create postgresql dataset %s", resourceName)
	}
	return created.ID(), created.Name, nil
}

func createSnowflake(
	ctx *pulumi.Context,
	resourceName string,
	spec *azuredatafactorydatasetv1alpha1.AzureDataFactoryDatasetSpec,
	azureProvider pulumi.ProviderResource,
) (pulumi.StringInput, pulumi.StringInput, error) {
	snowflake := spec.GetSnowflake()

	args := &datafactory.DatasetSnowflakeArgs{
		Name:                 pulumi.String(spec.Name),
		DataFactoryId:        pulumi.String(spec.DataFactoryId.GetValue()),
		LinkedServiceName:    linkedServiceName(spec),
		Description:          descriptionPtr(spec),
		Annotations:          annotationsArray(spec),
		Parameters:           parametersMap(spec),
		AdditionalProperties: additionalPropertiesMap(spec),
		Folder:               folderPtr(spec),
		TableName:            stringPtrWhenSet(snowflake.TableName),
		SchemaName:           stringPtrWhenSet(snowflake.SchemaName),
	}

	if len(snowflake.SchemaColumn) > 0 {
		columns := make(datafactory.DatasetSnowflakeSchemaColumnArray, 0, len(snowflake.SchemaColumn))
		for _, column := range snowflake.SchemaColumn {
			// Precision and scale are always sent (0 when undeclared)
			// -- the provider expands them unconditionally on both
			// engines.
			columns = append(columns, datafactory.DatasetSnowflakeSchemaColumnArgs{
				Name:      pulumi.String(column.Name),
				Type:      stringPtrWhenSet(column.Type),
				Precision: pulumi.IntPtr(int(column.Precision)),
				Scale:     pulumi.IntPtr(int(column.Scale)),
			})
		}
		args.SchemaColumns = columns
	}

	created, err := datafactory.NewDatasetSnowflake(ctx, resourceName, args, pulumi.Provider(azureProvider))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create snowflake dataset %s", resourceName)
	}
	return created.ID(), created.Name, nil
}

func createSqlServerTable(
	ctx *pulumi.Context,
	resourceName string,
	spec *azuredatafactorydatasetv1alpha1.AzureDataFactoryDatasetSpec,
	azureProvider pulumi.ProviderResource,
) (pulumi.StringInput, pulumi.StringInput, error) {
	sqlServer := spec.GetSqlServerTable()

	args := &datafactory.DatasetSqlServerTableArgs{
		Name:                 pulumi.String(spec.Name),
		DataFactoryId:        pulumi.String(spec.DataFactoryId.GetValue()),
		LinkedServiceName:    linkedServiceName(spec),
		Description:          descriptionPtr(spec),
		Annotations:          annotationsArray(spec),
		Parameters:           parametersMap(spec),
		AdditionalProperties: additionalPropertiesMap(spec),
		Folder:               folderPtr(spec),
		TableName:            stringPtrWhenSet(sqlServer.TableName),
	}

	if len(sqlServer.SchemaColumn) > 0 {
		columns := make(datafactory.DatasetSqlServerTableSchemaColumnArray, 0, len(sqlServer.SchemaColumn))
		for _, column := range sqlServer.SchemaColumn {
			columns = append(columns, datafactory.DatasetSqlServerTableSchemaColumnArgs{
				Name:        pulumi.String(column.Name),
				Type:        stringPtrWhenSet(column.Type),
				Description: stringPtrWhenSet(column.Description),
			})
		}
		args.SchemaColumns = columns
	}

	created, err := datafactory.NewDatasetSqlServerTable(ctx, resourceName, args, pulumi.Provider(azureProvider))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create sql server table dataset %s", resourceName)
	}
	return created.ID(), created.Name, nil
}
