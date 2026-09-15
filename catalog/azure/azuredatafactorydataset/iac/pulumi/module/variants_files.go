package module

import (
	"github.com/pkg/errors"
	azuredatafactorydatasetv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredatafactorydataset/v1alpha1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/datafactory"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The file-format builders: azure blob (flat path form), binary,
// delimited text, HTTP, JSON, and Parquet.

func createAzureBlob(
	ctx *pulumi.Context,
	resourceName string,
	spec *azuredatafactorydatasetv1alpha1.AzureDataFactoryDatasetSpec,
	azureProvider pulumi.ProviderResource,
) (pulumi.StringInput, pulumi.StringInput, error) {
	blob := spec.GetAzureBlob()

	args := &datafactory.DatasetAzureBlobArgs{
		Name:                   pulumi.String(spec.Name),
		DataFactoryId:          pulumi.String(spec.DataFactoryId.GetValue()),
		LinkedServiceName:      linkedServiceName(spec),
		Description:            descriptionPtr(spec),
		Annotations:            annotationsArray(spec),
		Parameters:             parametersMap(spec),
		AdditionalProperties:   additionalPropertiesMap(spec),
		Folder:                 folderPtr(spec),
		Path:                   stringPtrWhenSet(blob.Path),
		Filename:               stringPtrWhenSet(blob.Filename),
		DynamicPathEnabled:     boolPtrWhenTrue(blob.DynamicPathEnabled),
		DynamicFilenameEnabled: boolPtrWhenTrue(blob.DynamicFilenameEnabled),
	}

	if len(blob.SchemaColumn) > 0 {
		columns := make(datafactory.DatasetAzureBlobSchemaColumnArray, 0, len(blob.SchemaColumn))
		for _, column := range blob.SchemaColumn {
			columns = append(columns, datafactory.DatasetAzureBlobSchemaColumnArgs{
				Name:        pulumi.String(column.Name),
				Type:        stringPtrWhenSet(column.Type),
				Description: stringPtrWhenSet(column.Description),
			})
		}
		args.SchemaColumns = columns
	}

	created, err := datafactory.NewDatasetAzureBlob(ctx, resourceName, args, pulumi.Provider(azureProvider))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create azure blob dataset %s", resourceName)
	}
	return created.ID(), created.Name, nil
}

func createBinary(
	ctx *pulumi.Context,
	resourceName string,
	spec *azuredatafactorydatasetv1alpha1.AzureDataFactoryDatasetSpec,
	azureProvider pulumi.ProviderResource,
) (pulumi.StringInput, pulumi.StringInput, error) {
	binary := spec.GetBinary()

	args := &datafactory.DatasetBinaryArgs{
		Name:                 pulumi.String(spec.Name),
		DataFactoryId:        pulumi.String(spec.DataFactoryId.GetValue()),
		LinkedServiceName:    linkedServiceName(spec),
		Description:          descriptionPtr(spec),
		Annotations:          annotationsArray(spec),
		Parameters:           parametersMap(spec),
		AdditionalProperties: additionalPropertiesMap(spec),
		Folder:               folderPtr(spec),
	}

	// The binary format's HTTP location requires path and filename
	// (the spec enforces both non-empty).
	if binary.HttpServerLocation != nil {
		args.HttpServerLocation = datafactory.DatasetBinaryHttpServerLocationArgs{
			RelativeUrl:            pulumi.String(binary.HttpServerLocation.RelativeUrl),
			Path:                   pulumi.String(binary.HttpServerLocation.Path),
			DynamicPathEnabled:     boolPtrWhenTrue(binary.HttpServerLocation.DynamicPathEnabled),
			Filename:               pulumi.String(binary.HttpServerLocation.Filename),
			DynamicFilenameEnabled: boolPtrWhenTrue(binary.HttpServerLocation.DynamicFilenameEnabled),
		}
	}
	if binary.AzureBlobStorageLocation != nil {
		args.AzureBlobStorageLocation = datafactory.DatasetBinaryAzureBlobStorageLocationArgs{
			Container:               pulumi.String(binary.AzureBlobStorageLocation.Container),
			DynamicContainerEnabled: boolPtrWhenTrue(binary.AzureBlobStorageLocation.DynamicContainerEnabled),
			Path:                    stringPtrWhenSet(binary.AzureBlobStorageLocation.Path),
			DynamicPathEnabled:      boolPtrWhenTrue(binary.AzureBlobStorageLocation.DynamicPathEnabled),
			Filename:                stringPtrWhenSet(binary.AzureBlobStorageLocation.Filename),
			DynamicFilenameEnabled:  boolPtrWhenTrue(binary.AzureBlobStorageLocation.DynamicFilenameEnabled),
		}
	}
	if binary.SftpServerLocation != nil {
		args.SftpServerLocation = datafactory.DatasetBinarySftpServerLocationArgs{
			Path:                   pulumi.String(binary.SftpServerLocation.Path),
			DynamicPathEnabled:     boolPtrWhenTrue(binary.SftpServerLocation.DynamicPathEnabled),
			Filename:               pulumi.String(binary.SftpServerLocation.Filename),
			DynamicFilenameEnabled: boolPtrWhenTrue(binary.SftpServerLocation.DynamicFilenameEnabled),
		}
	}
	if binary.Compression != nil {
		args.Compression = datafactory.DatasetBinaryCompressionArgs{
			Type:  pulumi.String(binary.Compression.Type),
			Level: stringPtrWhenSet(binary.Compression.Level),
		}
	}

	created, err := datafactory.NewDatasetBinary(ctx, resourceName, args, pulumi.Provider(azureProvider))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create binary dataset %s", resourceName)
	}
	return created.ID(), created.Name, nil
}

func createDelimitedText(
	ctx *pulumi.Context,
	resourceName string,
	spec *azuredatafactorydatasetv1alpha1.AzureDataFactoryDatasetSpec,
	azureProvider pulumi.ProviderResource,
) (pulumi.StringInput, pulumi.StringInput, error) {
	delimitedText := spec.GetDelimitedText()

	args := &datafactory.DatasetDelimitedTextArgs{
		Name:                 pulumi.String(spec.Name),
		DataFactoryId:        pulumi.String(spec.DataFactoryId.GetValue()),
		LinkedServiceName:    linkedServiceName(spec),
		Description:          descriptionPtr(spec),
		Annotations:          annotationsArray(spec),
		Parameters:           parametersMap(spec),
		AdditionalProperties: additionalPropertiesMap(spec),
		Folder:               folderPtr(spec),
		// Omitted parse settings fall back to the provider's own
		// defaults ("," column delimiter, '"' quote, "\" escape).
		ColumnDelimiter:  stringPtrWhenSet(delimitedText.ColumnDelimiter),
		RowDelimiter:     stringPtrWhenSet(delimitedText.RowDelimiter),
		QuoteCharacter:   stringPtrWhenSet(delimitedText.QuoteCharacter),
		EscapeCharacter:  stringPtrWhenSet(delimitedText.EscapeCharacter),
		Encoding:         stringPtrWhenSet(delimitedText.Encoding),
		NullValue:        stringPtrWhenSet(delimitedText.NullValue),
		CompressionCodec: stringPtrWhenSet(delimitedText.CompressionCodec),
		CompressionLevel: stringPtrWhenSet(delimitedText.CompressionLevel),
	}

	// Platform default applied in the module: first row is not a
	// header unless the spec turns it on (the provider's own default,
	// sent explicitly for a readable plan).
	if delimitedText.FirstRowAsHeader != nil {
		args.FirstRowAsHeader = pulumi.BoolPtr(*delimitedText.FirstRowAsHeader)
	}

	// The delimited text format's HTTP location requires path and
	// filename (the spec enforces both non-empty).
	if delimitedText.HttpServerLocation != nil {
		args.HttpServerLocation = datafactory.DatasetDelimitedTextHttpServerLocationArgs{
			RelativeUrl:            pulumi.String(delimitedText.HttpServerLocation.RelativeUrl),
			Path:                   pulumi.String(delimitedText.HttpServerLocation.Path),
			DynamicPathEnabled:     boolPtrWhenTrue(delimitedText.HttpServerLocation.DynamicPathEnabled),
			Filename:               pulumi.String(delimitedText.HttpServerLocation.Filename),
			DynamicFilenameEnabled: boolPtrWhenTrue(delimitedText.HttpServerLocation.DynamicFilenameEnabled),
		}
	}
	if delimitedText.AzureBlobStorageLocation != nil {
		args.AzureBlobStorageLocation = datafactory.DatasetDelimitedTextAzureBlobStorageLocationArgs{
			Container:               pulumi.String(delimitedText.AzureBlobStorageLocation.Container),
			DynamicContainerEnabled: boolPtrWhenTrue(delimitedText.AzureBlobStorageLocation.DynamicContainerEnabled),
			Path:                    stringPtrWhenSet(delimitedText.AzureBlobStorageLocation.Path),
			DynamicPathEnabled:      boolPtrWhenTrue(delimitedText.AzureBlobStorageLocation.DynamicPathEnabled),
			Filename:                stringPtrWhenSet(delimitedText.AzureBlobStorageLocation.Filename),
			DynamicFilenameEnabled:  boolPtrWhenTrue(delimitedText.AzureBlobStorageLocation.DynamicFilenameEnabled),
		}
	}
	if delimitedText.AzureBlobFsLocation != nil {
		args.AzureBlobFsLocation = datafactory.DatasetDelimitedTextAzureBlobFsLocationArgs{
			FileSystem:               stringPtrWhenSet(delimitedText.AzureBlobFsLocation.FileSystem),
			DynamicFileSystemEnabled: boolPtrWhenTrue(delimitedText.AzureBlobFsLocation.DynamicFileSystemEnabled),
			Path:                     stringPtrWhenSet(delimitedText.AzureBlobFsLocation.Path),
			DynamicPathEnabled:       boolPtrWhenTrue(delimitedText.AzureBlobFsLocation.DynamicPathEnabled),
			Filename:                 stringPtrWhenSet(delimitedText.AzureBlobFsLocation.Filename),
			DynamicFilenameEnabled:   boolPtrWhenTrue(delimitedText.AzureBlobFsLocation.DynamicFilenameEnabled),
		}
	}

	if len(delimitedText.SchemaColumn) > 0 {
		columns := make(datafactory.DatasetDelimitedTextSchemaColumnArray, 0, len(delimitedText.SchemaColumn))
		for _, column := range delimitedText.SchemaColumn {
			columns = append(columns, datafactory.DatasetDelimitedTextSchemaColumnArgs{
				Name:        pulumi.String(column.Name),
				Type:        stringPtrWhenSet(column.Type),
				Description: stringPtrWhenSet(column.Description),
			})
		}
		args.SchemaColumns = columns
	}

	created, err := datafactory.NewDatasetDelimitedText(ctx, resourceName, args, pulumi.Provider(azureProvider))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create delimited text dataset %s", resourceName)
	}
	return created.ID(), created.Name, nil
}

func createHttp(
	ctx *pulumi.Context,
	resourceName string,
	spec *azuredatafactorydatasetv1alpha1.AzureDataFactoryDatasetSpec,
	azureProvider pulumi.ProviderResource,
) (pulumi.StringInput, pulumi.StringInput, error) {
	http := spec.GetHttp()

	args := &datafactory.DatasetHttpArgs{
		Name:                 pulumi.String(spec.Name),
		DataFactoryId:        pulumi.String(spec.DataFactoryId.GetValue()),
		LinkedServiceName:    linkedServiceName(spec),
		Description:          descriptionPtr(spec),
		Annotations:          annotationsArray(spec),
		Parameters:           parametersMap(spec),
		AdditionalProperties: additionalPropertiesMap(spec),
		Folder:               folderPtr(spec),
		RelativeUrl:          stringPtrWhenSet(http.RelativeUrl),
		RequestBody:          stringPtrWhenSet(http.RequestBody),
		RequestMethod:        stringPtrWhenSet(http.RequestMethod),
	}

	if len(http.SchemaColumn) > 0 {
		columns := make(datafactory.DatasetHttpSchemaColumnArray, 0, len(http.SchemaColumn))
		for _, column := range http.SchemaColumn {
			columns = append(columns, datafactory.DatasetHttpSchemaColumnArgs{
				Name:        pulumi.String(column.Name),
				Type:        stringPtrWhenSet(column.Type),
				Description: stringPtrWhenSet(column.Description),
			})
		}
		args.SchemaColumns = columns
	}

	created, err := datafactory.NewDatasetHttp(ctx, resourceName, args, pulumi.Provider(azureProvider))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create http dataset %s", resourceName)
	}
	return created.ID(), created.Name, nil
}

func createJson(
	ctx *pulumi.Context,
	resourceName string,
	spec *azuredatafactorydatasetv1alpha1.AzureDataFactoryDatasetSpec,
	azureProvider pulumi.ProviderResource,
) (pulumi.StringInput, pulumi.StringInput, error) {
	jsonDataset := spec.GetJson()

	args := &datafactory.DatasetJsonArgs{
		Name:                 pulumi.String(spec.Name),
		DataFactoryId:        pulumi.String(spec.DataFactoryId.GetValue()),
		LinkedServiceName:    linkedServiceName(spec),
		Description:          descriptionPtr(spec),
		Annotations:          annotationsArray(spec),
		Parameters:           parametersMap(spec),
		AdditionalProperties: additionalPropertiesMap(spec),
		Folder:               folderPtr(spec),
		Encoding:             stringPtrWhenSet(jsonDataset.Encoding),
	}

	// The JSON format requires path and filename in BOTH location
	// shapes (the spec enforces all four non-empty).
	if jsonDataset.HttpServerLocation != nil {
		args.HttpServerLocation = datafactory.DatasetJsonHttpServerLocationArgs{
			RelativeUrl:            pulumi.String(jsonDataset.HttpServerLocation.RelativeUrl),
			Path:                   pulumi.String(jsonDataset.HttpServerLocation.Path),
			DynamicPathEnabled:     boolPtrWhenTrue(jsonDataset.HttpServerLocation.DynamicPathEnabled),
			Filename:               pulumi.String(jsonDataset.HttpServerLocation.Filename),
			DynamicFilenameEnabled: boolPtrWhenTrue(jsonDataset.HttpServerLocation.DynamicFilenameEnabled),
		}
	}
	if jsonDataset.AzureBlobStorageLocation != nil {
		args.AzureBlobStorageLocation = datafactory.DatasetJsonAzureBlobStorageLocationArgs{
			Container:               pulumi.String(jsonDataset.AzureBlobStorageLocation.Container),
			DynamicContainerEnabled: boolPtrWhenTrue(jsonDataset.AzureBlobStorageLocation.DynamicContainerEnabled),
			Path:                    pulumi.String(jsonDataset.AzureBlobStorageLocation.Path),
			DynamicPathEnabled:      boolPtrWhenTrue(jsonDataset.AzureBlobStorageLocation.DynamicPathEnabled),
			Filename:                pulumi.String(jsonDataset.AzureBlobStorageLocation.Filename),
			DynamicFilenameEnabled:  boolPtrWhenTrue(jsonDataset.AzureBlobStorageLocation.DynamicFilenameEnabled),
		}
	}

	if len(jsonDataset.SchemaColumn) > 0 {
		columns := make(datafactory.DatasetJsonSchemaColumnArray, 0, len(jsonDataset.SchemaColumn))
		for _, column := range jsonDataset.SchemaColumn {
			columns = append(columns, datafactory.DatasetJsonSchemaColumnArgs{
				Name:        pulumi.String(column.Name),
				Type:        stringPtrWhenSet(column.Type),
				Description: stringPtrWhenSet(column.Description),
			})
		}
		args.SchemaColumns = columns
	}

	created, err := datafactory.NewDatasetJson(ctx, resourceName, args, pulumi.Provider(azureProvider))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create json dataset %s", resourceName)
	}
	return created.ID(), created.Name, nil
}

func createParquet(
	ctx *pulumi.Context,
	resourceName string,
	spec *azuredatafactorydatasetv1alpha1.AzureDataFactoryDatasetSpec,
	azureProvider pulumi.ProviderResource,
) (pulumi.StringInput, pulumi.StringInput, error) {
	parquet := spec.GetParquet()

	args := &datafactory.DatasetParquetArgs{
		Name:                 pulumi.String(spec.Name),
		DataFactoryId:        pulumi.String(spec.DataFactoryId.GetValue()),
		LinkedServiceName:    linkedServiceName(spec),
		Description:          descriptionPtr(spec),
		Annotations:          annotationsArray(spec),
		Parameters:           parametersMap(spec),
		AdditionalProperties: additionalPropertiesMap(spec),
		Folder:               folderPtr(spec),
		// compression_level exists in the provider's schema for this
		// resource but its create/update code never reads it (dead
		// argument) -- deliberately not modeled; recorded in
		// iac/provider-parity.yaml.
		CompressionCodec: stringPtrWhenSet(parquet.CompressionCodec),
	}

	// Parquet's HTTP location requires the filename; the folder path
	// is optional (the one format where it may be omitted).
	if parquet.HttpServerLocation != nil {
		args.HttpServerLocation = datafactory.DatasetParquetHttpServerLocationArgs{
			RelativeUrl:            pulumi.String(parquet.HttpServerLocation.RelativeUrl),
			Path:                   stringPtrWhenSet(parquet.HttpServerLocation.Path),
			DynamicPathEnabled:     boolPtrWhenTrue(parquet.HttpServerLocation.DynamicPathEnabled),
			Filename:               pulumi.String(parquet.HttpServerLocation.Filename),
			DynamicFilenameEnabled: boolPtrWhenTrue(parquet.HttpServerLocation.DynamicFilenameEnabled),
		}
	}
	if parquet.AzureBlobStorageLocation != nil {
		args.AzureBlobStorageLocation = datafactory.DatasetParquetAzureBlobStorageLocationArgs{
			Container:               pulumi.String(parquet.AzureBlobStorageLocation.Container),
			DynamicContainerEnabled: boolPtrWhenTrue(parquet.AzureBlobStorageLocation.DynamicContainerEnabled),
			Path:                    stringPtrWhenSet(parquet.AzureBlobStorageLocation.Path),
			DynamicPathEnabled:      boolPtrWhenTrue(parquet.AzureBlobStorageLocation.DynamicPathEnabled),
			Filename:                stringPtrWhenSet(parquet.AzureBlobStorageLocation.Filename),
			DynamicFilenameEnabled:  boolPtrWhenTrue(parquet.AzureBlobStorageLocation.DynamicFilenameEnabled),
		}
	}
	if parquet.AzureBlobFsLocation != nil {
		args.AzureBlobFsLocation = datafactory.DatasetParquetAzureBlobFsLocationArgs{
			FileSystem:               stringPtrWhenSet(parquet.AzureBlobFsLocation.FileSystem),
			DynamicFileSystemEnabled: boolPtrWhenTrue(parquet.AzureBlobFsLocation.DynamicFileSystemEnabled),
			Path:                     stringPtrWhenSet(parquet.AzureBlobFsLocation.Path),
			DynamicPathEnabled:       boolPtrWhenTrue(parquet.AzureBlobFsLocation.DynamicPathEnabled),
			Filename:                 stringPtrWhenSet(parquet.AzureBlobFsLocation.Filename),
			DynamicFilenameEnabled:   boolPtrWhenTrue(parquet.AzureBlobFsLocation.DynamicFilenameEnabled),
		}
	}

	if len(parquet.SchemaColumn) > 0 {
		columns := make(datafactory.DatasetParquetSchemaColumnArray, 0, len(parquet.SchemaColumn))
		for _, column := range parquet.SchemaColumn {
			columns = append(columns, datafactory.DatasetParquetSchemaColumnArgs{
				Name:        pulumi.String(column.Name),
				Type:        stringPtrWhenSet(column.Type),
				Description: stringPtrWhenSet(column.Description),
			})
		}
		args.SchemaColumns = columns
	}

	created, err := datafactory.NewDatasetParquet(ctx, resourceName, args, pulumi.Provider(azureProvider))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create parquet dataset %s", resourceName)
	}
	return created.ID(), created.Name, nil
}
