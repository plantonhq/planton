package module

import (
	"github.com/pkg/errors"
	azuredatafactorylinkedservicev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredatafactorylinkedservice/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// One kind, 23 provider resources: Azure stores every connection type
// in the SAME factory-scoped linked-service namespace, so the spec's
// variant block selects which resource is created. Shared fields
// (name, factory, description, annotations, parameters,
// additional_properties, the integration runtime) travel identically
// on every type; each builder adds only its variant's own arguments.
func Resources(ctx *pulumi.Context, iacInput *azuredatafactorylinkedservicev1alpha1.AzureDataFactoryLinkedServiceIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureDataFactoryLinkedService.Spec
	resourceName := locals.AzureDataFactoryLinkedService.Metadata.Name

	var linkedServiceId pulumi.StringInput
	var linkedServiceName pulumi.StringInput

	switch {
	case spec.AzureBlobStorage != nil:
		linkedServiceId, linkedServiceName, err = createAzureBlobStorage(ctx, resourceName, spec, azureProvider)
	case spec.AzureDatabricks != nil:
		linkedServiceId, linkedServiceName, err = createAzureDatabricks(ctx, resourceName, spec, azureProvider)
	case spec.AzureFileStorage != nil:
		linkedServiceId, linkedServiceName, err = createAzureFileStorage(ctx, resourceName, spec, azureProvider)
	case spec.AzureFunction != nil:
		linkedServiceId, linkedServiceName, err = createAzureFunction(ctx, resourceName, spec, azureProvider)
	case spec.AzureSearch != nil:
		linkedServiceId, linkedServiceName, err = createAzureSearch(ctx, resourceName, spec, azureProvider)
	case spec.AzureSqlDatabase != nil:
		linkedServiceId, linkedServiceName, err = createAzureSqlDatabase(ctx, resourceName, spec, azureProvider)
	case spec.AzureTableStorage != nil:
		linkedServiceId, linkedServiceName, err = createAzureTableStorage(ctx, resourceName, spec, azureProvider)
	case spec.Cosmosdb != nil:
		linkedServiceId, linkedServiceName, err = createCosmosdb(ctx, resourceName, spec, azureProvider)
	case spec.CosmosdbMongoapi != nil:
		linkedServiceId, linkedServiceName, err = createCosmosdbMongoapi(ctx, resourceName, spec, azureProvider)
	case spec.Custom != nil:
		linkedServiceId, linkedServiceName, err = createCustom(ctx, resourceName, spec, azureProvider)
	case spec.DataLakeStorageGen2 != nil:
		linkedServiceId, linkedServiceName, err = createDataLakeStorageGen2(ctx, resourceName, spec, azureProvider)
	case spec.KeyVault != nil:
		linkedServiceId, linkedServiceName, err = createKeyVault(ctx, resourceName, spec, azureProvider)
	case spec.Kusto != nil:
		linkedServiceId, linkedServiceName, err = createKusto(ctx, resourceName, spec, azureProvider)
	case spec.Mysql != nil:
		linkedServiceId, linkedServiceName, err = createMysql(ctx, resourceName, spec, azureProvider)
	case spec.Odata != nil:
		linkedServiceId, linkedServiceName, err = createOdata(ctx, resourceName, spec, azureProvider)
	case spec.Odbc != nil:
		linkedServiceId, linkedServiceName, err = createOdbc(ctx, resourceName, spec, azureProvider)
	case spec.Postgresql != nil:
		linkedServiceId, linkedServiceName, err = createPostgresql(ctx, resourceName, spec, azureProvider)
	case spec.Sftp != nil:
		linkedServiceId, linkedServiceName, err = createSftp(ctx, resourceName, spec, azureProvider)
	case spec.Snowflake != nil:
		linkedServiceId, linkedServiceName, err = createSnowflake(ctx, resourceName, spec, azureProvider)
	case spec.SqlManagedInstance != nil:
		linkedServiceId, linkedServiceName, err = createSqlManagedInstance(ctx, resourceName, spec, azureProvider)
	case spec.SqlServer != nil:
		linkedServiceId, linkedServiceName, err = createSqlServer(ctx, resourceName, spec, azureProvider)
	case spec.Synapse != nil:
		linkedServiceId, linkedServiceName, err = createSynapse(ctx, resourceName, spec, azureProvider)
	case spec.Web != nil:
		linkedServiceId, linkedServiceName, err = createWeb(ctx, resourceName, spec, azureProvider)
	default:
		// The spec's exactly-one CEL makes this unreachable; the guard
		// keeps a broken input loud instead of silently exporting nothing.
		return errors.New("exactly one linked service variant block must be set")
	}
	if err != nil {
		return err
	}

	ctx.Export(OpLinkedServiceId, linkedServiceId)
	ctx.Export(OpLinkedServiceName, linkedServiceName)

	return nil
}
