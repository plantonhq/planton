package module

import (
	"encoding/json"
	"sort"

	"github.com/pkg/errors"
	awsbedrockknowledgebasev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbedrockknowledgebase/v1alpha1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/bedrock"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// knowledgeBase creates the Bedrock knowledge base with its derived type
// discriminators, the vector store backend, and the folded data sources,
// and exports outputs.
//
// Nearly the whole surface is create-time only upstream; the provider
// retries the IAM/data-access propagation classes at create -- the module
// adds none.
func knowledgeBase(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec

	args := &bedrock.AgentKnowledgeBaseArgs{
		// Create-time naming basis; doubles as the Name tag. metadata.name
		// on both engines.
		Name:    pulumi.String(locals.KnowledgeBaseName),
		RoleArn: pulumi.String(spec.RoleArn.GetValue()),
		Tags:    pulumi.ToStringMap(locals.AwsTags),
	}

	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}

	// -------------------------------------------------------------------
	// Knowledge-base type (the discriminator is derived from which spec
	// arm is set -- exactly one, per the spec's CEL guards)
	// -------------------------------------------------------------------
	kbConfiguration := &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationArgs{}

	switch {
	case spec.GetVector() != nil:
		kbConfiguration.Type = pulumi.String("VECTOR")
		vector := &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationVectorKnowledgeBaseConfigurationArgs{
			EmbeddingModelArn: pulumi.String(spec.GetVector().EmbeddingModelArn),
		}
		if spec.GetVector().EmbeddingModel != nil {
			m := spec.GetVector().EmbeddingModel
			bedrockModel := &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationVectorKnowledgeBaseConfigurationEmbeddingModelConfigurationBedrockEmbeddingModelConfigurationArgs{}
			if m.Dimensions != 0 {
				bedrockModel.Dimensions = pulumi.Int(int(m.Dimensions))
			}
			if m.EmbeddingDataType != "" {
				bedrockModel.EmbeddingDataType = pulumi.String(m.EmbeddingDataType)
			}
			if m.AudioSegmentationSeconds != 0 {
				bedrockModel.Audio = &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationVectorKnowledgeBaseConfigurationEmbeddingModelConfigurationBedrockEmbeddingModelConfigurationAudioArgs{
					SegmentationConfiguration: &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationVectorKnowledgeBaseConfigurationEmbeddingModelConfigurationBedrockEmbeddingModelConfigurationAudioSegmentationConfigurationArgs{
						FixedLengthDuration: pulumi.Int(int(m.AudioSegmentationSeconds)),
					},
				}
			}
			if m.VideoSegmentationSeconds != 0 {
				bedrockModel.Video = &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationVectorKnowledgeBaseConfigurationEmbeddingModelConfigurationBedrockEmbeddingModelConfigurationVideoArgs{
					SegmentationConfiguration: &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationVectorKnowledgeBaseConfigurationEmbeddingModelConfigurationBedrockEmbeddingModelConfigurationVideoSegmentationConfigurationArgs{
						FixedLengthDuration: pulumi.Int(int(m.VideoSegmentationSeconds)),
					},
				}
			}
			vector.EmbeddingModelConfiguration = &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationVectorKnowledgeBaseConfigurationEmbeddingModelConfigurationArgs{
				BedrockEmbeddingModelConfiguration: bedrockModel,
			}
		}
		// The supplemental S3 location rides a fixed two-level wrapper
		// upstream (storage_location type S3 + s3_location.uri); the spec
		// carries the URI leaf directly.
		if spec.GetVector().SupplementalDataS3Uri != "" {
			vector.SupplementalDataStorageConfiguration = &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationVectorKnowledgeBaseConfigurationSupplementalDataStorageConfigurationArgs{
				StorageLocations: bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationVectorKnowledgeBaseConfigurationSupplementalDataStorageConfigurationStorageLocationArray{
					&bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationVectorKnowledgeBaseConfigurationSupplementalDataStorageConfigurationStorageLocationArgs{
						Type: pulumi.String("S3"),
						S3Location: &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationVectorKnowledgeBaseConfigurationSupplementalDataStorageConfigurationStorageLocationS3LocationArgs{
							Uri: pulumi.String(spec.GetVector().SupplementalDataS3Uri),
						},
					},
				},
			}
		}
		kbConfiguration.VectorKnowledgeBaseConfiguration = vector

	case spec.GetManaged() != nil:
		kbConfiguration.Type = pulumi.String("MANAGED")
		managed := &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationManagedKnowledgeBaseConfigurationArgs{}
		if spec.GetManaged().EmbeddingModelArn != "" {
			managed.EmbeddingModelArn = pulumi.String(spec.GetManaged().EmbeddingModelArn)
		}
		// Derived discriminator, ALWAYS sent: AWS's embeddingModelType is
		// CUSTOM exactly when an embedding-model ARN is brought, MANAGED
		// otherwise. The provider marks the attribute Optional+Computed
		// (UseStateForUnknown); leaving it unset makes the bridge fail the
		// apply with "unexpected unknown property value" AFTER AWS created
		// the knowledge base -- stranding it outside state (live-caught
		// 2026-08-13). Sending the derived value keeps it known at plan.
		if spec.GetManaged().EmbeddingModelArn != "" {
			managed.EmbeddingModelType = pulumi.String("CUSTOM")
		} else {
			managed.EmbeddingModelType = pulumi.String("MANAGED")
		}
		if spec.GetManaged().EmbeddingModel != nil {
			m := spec.GetManaged().EmbeddingModel
			bedrockModel := &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationManagedKnowledgeBaseConfigurationEmbeddingModelConfigurationBedrockEmbeddingModelConfigurationArgs{}
			if m.Dimensions != 0 {
				bedrockModel.Dimensions = pulumi.Int(int(m.Dimensions))
			}
			if m.EmbeddingDataType != "" {
				bedrockModel.EmbeddingDataType = pulumi.String(m.EmbeddingDataType)
			}
			if m.AudioSegmentationSeconds != 0 {
				bedrockModel.Audio = &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationManagedKnowledgeBaseConfigurationEmbeddingModelConfigurationBedrockEmbeddingModelConfigurationAudioArgs{
					SegmentationConfiguration: &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationManagedKnowledgeBaseConfigurationEmbeddingModelConfigurationBedrockEmbeddingModelConfigurationAudioSegmentationConfigurationArgs{
						FixedLengthDuration: pulumi.Int(int(m.AudioSegmentationSeconds)),
					},
				}
			}
			if m.VideoSegmentationSeconds != 0 {
				bedrockModel.Video = &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationManagedKnowledgeBaseConfigurationEmbeddingModelConfigurationBedrockEmbeddingModelConfigurationVideoArgs{
					SegmentationConfiguration: &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationManagedKnowledgeBaseConfigurationEmbeddingModelConfigurationBedrockEmbeddingModelConfigurationVideoSegmentationConfigurationArgs{
						FixedLengthDuration: pulumi.Int(int(m.VideoSegmentationSeconds)),
					},
				}
			}
			managed.EmbeddingModelConfiguration = &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationManagedKnowledgeBaseConfigurationEmbeddingModelConfigurationArgs{
				BedrockEmbeddingModelConfiguration: bedrockModel,
			}
		}
		if spec.GetManaged().KmsKeyArn.GetValue() != "" {
			managed.ServerSideEncryptionConfiguration = &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationManagedKnowledgeBaseConfigurationServerSideEncryptionConfigurationArgs{
				KmsKeyArn: pulumi.String(spec.GetManaged().KmsKeyArn.GetValue()),
			}
		}
		kbConfiguration.ManagedKnowledgeBaseConfiguration = managed

	case spec.GetKendra() != nil:
		kbConfiguration.Type = pulumi.String("KENDRA")
		kbConfiguration.KendraKnowledgeBaseConfiguration = &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationKendraKnowledgeBaseConfigurationArgs{
			KendraIndexArn: pulumi.String(spec.GetKendra().KendraIndexArn),
		}

	case spec.GetSql() != nil:
		kbConfiguration.Type = pulumi.String("SQL")
		// REDSHIFT is the only SQL engine AWS defines -- the module owns
		// the constant.
		redshift := &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationSqlKnowledgeBaseConfigurationRedshiftConfigurationArgs{}

		queryEngine := &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationSqlKnowledgeBaseConfigurationRedshiftConfigurationQueryEngineConfigurationArgs{}
		if spec.GetSql().Provisioned != nil {
			queryEngine.Type = pulumi.String("PROVISIONED")
			provisionedAuth := &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationSqlKnowledgeBaseConfigurationRedshiftConfigurationQueryEngineConfigurationProvisionedConfigurationAuthConfigurationArgs{
				Type: pulumi.String(spec.GetSql().Provisioned.Auth.Type),
			}
			if spec.GetSql().Provisioned.Auth.DatabaseUser != "" {
				provisionedAuth.DatabaseUser = pulumi.String(spec.GetSql().Provisioned.Auth.DatabaseUser)
			}
			if spec.GetSql().Provisioned.Auth.UsernamePasswordSecretArn.GetValue() != "" {
				provisionedAuth.UsernamePasswordSecretArn = pulumi.String(spec.GetSql().Provisioned.Auth.UsernamePasswordSecretArn.GetValue())
			}
			queryEngine.ProvisionedConfiguration = &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationSqlKnowledgeBaseConfigurationRedshiftConfigurationQueryEngineConfigurationProvisionedConfigurationArgs{
				ClusterIdentifier: pulumi.String(spec.GetSql().Provisioned.ClusterIdentifier.GetValue()),
				AuthConfiguration: provisionedAuth,
			}
		}
		if spec.GetSql().Serverless != nil {
			queryEngine.Type = pulumi.String("SERVERLESS")
			serverlessAuth := &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationSqlKnowledgeBaseConfigurationRedshiftConfigurationQueryEngineConfigurationServerlessConfigurationAuthConfigurationArgs{
				Type: pulumi.String(spec.GetSql().Serverless.Auth.Type),
			}
			if spec.GetSql().Serverless.Auth.UsernamePasswordSecretArn.GetValue() != "" {
				serverlessAuth.UsernamePasswordSecretArn = pulumi.String(spec.GetSql().Serverless.Auth.UsernamePasswordSecretArn.GetValue())
			}
			queryEngine.ServerlessConfiguration = &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationSqlKnowledgeBaseConfigurationRedshiftConfigurationQueryEngineConfigurationServerlessConfigurationArgs{
				WorkgroupArn:      pulumi.String(spec.GetSql().Serverless.WorkgroupArn.GetValue()),
				AuthConfiguration: serverlessAuth,
			}
		}
		redshift.QueryEngineConfiguration = queryEngine

		warehouse := &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationSqlKnowledgeBaseConfigurationRedshiftConfigurationStorageConfigurationArgs{}
		if spec.GetSql().Warehouse.DataCatalog != nil {
			warehouse.Type = pulumi.String("AWS_DATA_CATALOG")
			warehouse.AwsDataCatalogConfiguration = &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationSqlKnowledgeBaseConfigurationRedshiftConfigurationStorageConfigurationAwsDataCatalogConfigurationArgs{
				TableNames: pulumi.ToStringArray(spec.GetSql().Warehouse.DataCatalog.TableNames),
			}
		}
		if spec.GetSql().Warehouse.Redshift != nil {
			warehouse.Type = pulumi.String("REDSHIFT")
			warehouse.RedshiftConfiguration = &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationSqlKnowledgeBaseConfigurationRedshiftConfigurationStorageConfigurationRedshiftConfigurationArgs{
				DatabaseName: pulumi.String(spec.GetSql().Warehouse.Redshift.DatabaseName),
			}
		}
		redshift.StorageConfiguration = warehouse

		if spec.GetSql().QueryGeneration != nil {
			qg := &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationSqlKnowledgeBaseConfigurationRedshiftConfigurationQueryGenerationConfigurationArgs{}
			if spec.GetSql().QueryGeneration.ExecutionTimeoutSeconds != 0 {
				qg.ExecutionTimeoutSeconds = pulumi.Int(int(spec.GetSql().QueryGeneration.ExecutionTimeoutSeconds))
			}
			if len(spec.GetSql().QueryGeneration.CuratedQueries) > 0 || len(spec.GetSql().QueryGeneration.Tables) > 0 {
				genContext := &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationSqlKnowledgeBaseConfigurationRedshiftConfigurationQueryGenerationConfigurationGenerationContextArgs{}
				var curated bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationSqlKnowledgeBaseConfigurationRedshiftConfigurationQueryGenerationConfigurationGenerationContextCuratedQueryArray
				for _, q := range spec.GetSql().QueryGeneration.CuratedQueries {
					curated = append(curated, &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationSqlKnowledgeBaseConfigurationRedshiftConfigurationQueryGenerationConfigurationGenerationContextCuratedQueryArgs{
						NaturalLanguage: pulumi.String(q.NaturalLanguage),
						Sql:             pulumi.String(q.Sql),
					})
				}
				genContext.CuratedQueries = curated
				var tables bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationSqlKnowledgeBaseConfigurationRedshiftConfigurationQueryGenerationConfigurationGenerationContextTableArray
				for _, t := range spec.GetSql().QueryGeneration.Tables {
					table := &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationSqlKnowledgeBaseConfigurationRedshiftConfigurationQueryGenerationConfigurationGenerationContextTableArgs{
						Name: pulumi.String(t.Name),
					}
					if t.Description != "" {
						table.Description = pulumi.String(t.Description)
					}
					if t.Inclusion != "" {
						table.Inclusion = pulumi.String(t.Inclusion)
					}
					var columns bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationSqlKnowledgeBaseConfigurationRedshiftConfigurationQueryGenerationConfigurationGenerationContextTableColumnArray
					for _, c := range t.Columns {
						column := &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationSqlKnowledgeBaseConfigurationRedshiftConfigurationQueryGenerationConfigurationGenerationContextTableColumnArgs{
							Name: pulumi.String(c.Name),
						}
						if c.Description != "" {
							column.Description = pulumi.String(c.Description)
						}
						if c.Inclusion != "" {
							column.Inclusion = pulumi.String(c.Inclusion)
						}
						columns = append(columns, column)
					}
					table.Columns = columns
					tables = append(tables, table)
				}
				genContext.Tables = tables
				qg.GenerationContext = genContext
			}
			redshift.QueryGenerationConfiguration = qg
		}

		kbConfiguration.SqlKnowledgeBaseConfiguration = &bedrock.AgentKnowledgeBaseKnowledgeBaseConfigurationSqlKnowledgeBaseConfigurationArgs{
			Type:                  pulumi.String("REDSHIFT"),
			RedshiftConfiguration: redshift,
		}
	}
	args.KnowledgeBaseConfiguration = kbConfiguration

	// -------------------------------------------------------------------
	// Vector store (required with the vector type, absent otherwise --
	// the spec's CEL guards enforce the pairing)
	// -------------------------------------------------------------------
	if spec.Storage != nil {
		storage := &bedrock.AgentKnowledgeBaseStorageConfigurationArgs{}
		switch {
		case spec.Storage.OpensearchServerless != nil:
			s := spec.Storage.OpensearchServerless
			storage.Type = pulumi.String("OPENSEARCH_SERVERLESS")
			storage.OpensearchServerlessConfiguration = &bedrock.AgentKnowledgeBaseStorageConfigurationOpensearchServerlessConfigurationArgs{
				CollectionArn:   pulumi.String(s.CollectionArn.GetValue()),
				VectorIndexName: pulumi.String(s.VectorIndexName),
				FieldMapping: &bedrock.AgentKnowledgeBaseStorageConfigurationOpensearchServerlessConfigurationFieldMappingArgs{
					VectorField:   pulumi.String(s.FieldMapping.VectorField),
					TextField:     pulumi.String(s.FieldMapping.TextField),
					MetadataField: pulumi.String(s.FieldMapping.MetadataField),
				},
			}
		case spec.Storage.OpensearchManaged != nil:
			s := spec.Storage.OpensearchManaged
			storage.Type = pulumi.String("OPENSEARCH_MANAGED_CLUSTER")
			storage.OpensearchManagedClusterConfiguration = &bedrock.AgentKnowledgeBaseStorageConfigurationOpensearchManagedClusterConfigurationArgs{
				DomainArn:       pulumi.String(s.DomainArn.GetValue()),
				DomainEndpoint:  pulumi.String(s.DomainEndpoint),
				VectorIndexName: pulumi.String(s.VectorIndexName),
				FieldMapping: &bedrock.AgentKnowledgeBaseStorageConfigurationOpensearchManagedClusterConfigurationFieldMappingArgs{
					VectorField:   pulumi.String(s.FieldMapping.VectorField),
					TextField:     pulumi.String(s.FieldMapping.TextField),
					MetadataField: pulumi.String(s.FieldMapping.MetadataField),
				},
			}
		case spec.Storage.S3Vectors != nil:
			s := spec.Storage.S3Vectors
			storage.Type = pulumi.String("S3_VECTORS")
			s3Vectors := &bedrock.AgentKnowledgeBaseStorageConfigurationS3VectorsConfigurationArgs{}
			if s.IndexArn != "" {
				s3Vectors.IndexArn = pulumi.String(s.IndexArn)
			}
			if s.IndexName != "" {
				s3Vectors.IndexName = pulumi.String(s.IndexName)
			}
			if arn := s.GetVectorBucketArn().GetValue(); arn != "" {
				s3Vectors.VectorBucketArn = pulumi.String(arn)
			}
			storage.S3VectorsConfiguration = s3Vectors
		case spec.Storage.Rds != nil:
			s := spec.Storage.Rds
			storage.Type = pulumi.String("RDS")
			rdsMapping := &bedrock.AgentKnowledgeBaseStorageConfigurationRdsConfigurationFieldMappingArgs{
				VectorField:     pulumi.String(s.FieldMapping.VectorField),
				TextField:       pulumi.String(s.FieldMapping.TextField),
				MetadataField:   pulumi.String(s.FieldMapping.MetadataField),
				PrimaryKeyField: pulumi.String(s.FieldMapping.PrimaryKeyField),
			}
			if s.FieldMapping.CustomMetadataField != "" {
				rdsMapping.CustomMetadataField = pulumi.String(s.FieldMapping.CustomMetadataField)
			}
			storage.RdsConfiguration = &bedrock.AgentKnowledgeBaseStorageConfigurationRdsConfigurationArgs{
				ResourceArn:          pulumi.String(s.ResourceArn.GetValue()),
				CredentialsSecretArn: pulumi.String(s.CredentialsSecretArn.GetValue()),
				DatabaseName:         pulumi.String(s.DatabaseName),
				TableName:            pulumi.String(s.TableName),
				FieldMapping:         rdsMapping,
			}
		case spec.Storage.Pinecone != nil:
			s := spec.Storage.Pinecone
			storage.Type = pulumi.String("PINECONE")
			pinecone := &bedrock.AgentKnowledgeBaseStorageConfigurationPineconeConfigurationArgs{
				ConnectionString:     pulumi.String(s.ConnectionString),
				CredentialsSecretArn: pulumi.String(s.CredentialsSecretArn.GetValue()),
				FieldMapping: &bedrock.AgentKnowledgeBaseStorageConfigurationPineconeConfigurationFieldMappingArgs{
					TextField:     pulumi.String(s.FieldMapping.TextField),
					MetadataField: pulumi.String(s.FieldMapping.MetadataField),
				},
			}
			if s.Namespace != "" {
				pinecone.Namespace = pulumi.String(s.Namespace)
			}
			storage.PineconeConfiguration = pinecone
		case spec.Storage.MongodbAtlas != nil:
			s := spec.Storage.MongodbAtlas
			storage.Type = pulumi.String("MONGO_DB_ATLAS")
			mongo := &bedrock.AgentKnowledgeBaseStorageConfigurationMongoDbAtlasConfigurationArgs{
				Endpoint:             pulumi.String(s.Endpoint),
				DatabaseName:         pulumi.String(s.DatabaseName),
				CollectionName:       pulumi.String(s.CollectionName),
				VectorIndexName:      pulumi.String(s.VectorIndexName),
				CredentialsSecretArn: pulumi.String(s.CredentialsSecretArn.GetValue()),
				FieldMapping: &bedrock.AgentKnowledgeBaseStorageConfigurationMongoDbAtlasConfigurationFieldMappingArgs{
					VectorField:   pulumi.String(s.FieldMapping.VectorField),
					TextField:     pulumi.String(s.FieldMapping.TextField),
					MetadataField: pulumi.String(s.FieldMapping.MetadataField),
				},
			}
			if s.TextIndexName != "" {
				mongo.TextIndexName = pulumi.String(s.TextIndexName)
			}
			if s.EndpointServiceName != "" {
				mongo.EndpointServiceName = pulumi.String(s.EndpointServiceName)
			}
			storage.MongoDbAtlasConfiguration = mongo
		case spec.Storage.NeptuneAnalytics != nil:
			s := spec.Storage.NeptuneAnalytics
			storage.Type = pulumi.String("NEPTUNE_ANALYTICS")
			storage.NeptuneAnalyticsConfiguration = &bedrock.AgentKnowledgeBaseStorageConfigurationNeptuneAnalyticsConfigurationArgs{
				GraphArn: pulumi.String(s.GraphArn),
				FieldMapping: &bedrock.AgentKnowledgeBaseStorageConfigurationNeptuneAnalyticsConfigurationFieldMappingArgs{
					TextField:     pulumi.String(s.FieldMapping.TextField),
					MetadataField: pulumi.String(s.FieldMapping.MetadataField),
				},
			}
		case spec.Storage.RedisEnterpriseCloud != nil:
			s := spec.Storage.RedisEnterpriseCloud
			storage.Type = pulumi.String("REDIS_ENTERPRISE_CLOUD")
			redis := &bedrock.AgentKnowledgeBaseStorageConfigurationRedisEnterpriseCloudConfigurationArgs{
				Endpoint:             pulumi.String(s.Endpoint),
				VectorIndexName:      pulumi.String(s.VectorIndexName),
				CredentialsSecretArn: pulumi.String(s.CredentialsSecretArn.GetValue()),
			}
			if s.FieldMapping != nil {
				redisMapping := &bedrock.AgentKnowledgeBaseStorageConfigurationRedisEnterpriseCloudConfigurationFieldMappingArgs{}
				if s.FieldMapping.VectorField != "" {
					redisMapping.VectorField = pulumi.String(s.FieldMapping.VectorField)
				}
				if s.FieldMapping.TextField != "" {
					redisMapping.TextField = pulumi.String(s.FieldMapping.TextField)
				}
				if s.FieldMapping.MetadataField != "" {
					redisMapping.MetadataField = pulumi.String(s.FieldMapping.MetadataField)
				}
				redis.FieldMapping = redisMapping
			}
			storage.RedisEnterpriseCloudConfiguration = redis
		}
		args.StorageConfiguration = storage
	}

	createdKnowledgeBase, err := bedrock.NewAgentKnowledgeBase(ctx, locals.KnowledgeBaseName, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "create knowledge base")
	}

	ctx.Export(OpKnowledgeBaseId, createdKnowledgeBase.ID())
	ctx.Export(OpKnowledgeBaseArn, createdKnowledgeBase.Arn)

	// Document connectors keyed by their stable entry names. Iteration is
	// name-sorted for deterministic previews.
	dataSourceIds := pulumi.StringMap{}
	for _, d := range sortedDataSources(spec.DataSources) {
		dataSourceArgs, err := dataSourceArgs(createdKnowledgeBase, d)
		if err != nil {
			return errors.Wrapf(err, "render data source %q", d.Name)
		}
		createdDataSource, err := bedrock.NewAgentDataSource(ctx, "data-source-"+d.Name, dataSourceArgs,
			pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{createdKnowledgeBase}))
		if err != nil {
			return errors.Wrapf(err, "create data source %q", d.Name)
		}
		dataSourceIds[d.Name] = createdDataSource.DataSourceId
	}
	ctx.Export(OpDataSourceIds, dataSourceIds)

	return nil
}

// dataSourceArgs renders one data source entry -- the connector arm plus
// the ingestion pipeline.
func dataSourceArgs(kb *bedrock.AgentKnowledgeBase, d *awsbedrockknowledgebasev1alpha1.AwsBedrockKnowledgeBaseDataSource) (*bedrock.AgentDataSourceArgs, error) {
	args := &bedrock.AgentDataSourceArgs{
		Name:            pulumi.String(d.Name),
		KnowledgeBaseId: kb.ID(),
	}
	if d.Description != "" {
		args.Description = pulumi.String(d.Description)
	}
	if d.DataDeletionPolicy != "" {
		args.DataDeletionPolicy = pulumi.String(d.DataDeletionPolicy)
	}
	if d.KmsKeyArn.GetValue() != "" {
		args.ServerSideEncryptionConfiguration = &bedrock.AgentDataSourceServerSideEncryptionConfigurationArgs{
			KmsKeyArn: pulumi.String(d.KmsKeyArn.GetValue()),
		}
	}

	configuration := &bedrock.AgentDataSourceDataSourceConfigurationArgs{}
	switch {
	case d.GetS3() != nil:
		configuration.Type = pulumi.String("S3")
		s3 := &bedrock.AgentDataSourceDataSourceConfigurationS3ConfigurationArgs{
			BucketArn: pulumi.String(d.GetS3().BucketArn.GetValue()),
		}
		if d.GetS3().InclusionPrefix != "" {
			s3.InclusionPrefixes = pulumi.StringArray{pulumi.String(d.GetS3().InclusionPrefix)}
		}
		if d.GetS3().BucketOwnerAccountId != "" {
			s3.BucketOwnerAccountId = pulumi.String(d.GetS3().BucketOwnerAccountId)
		}
		configuration.S3Configuration = s3

	case d.GetWeb() != nil:
		configuration.Type = pulumi.String("WEB")
		var seedUrls bedrock.AgentDataSourceDataSourceConfigurationWebConfigurationSourceConfigurationUrlConfigurationSeedUrlArray
		for _, u := range d.GetWeb().SeedUrls {
			seedUrls = append(seedUrls, &bedrock.AgentDataSourceDataSourceConfigurationWebConfigurationSourceConfigurationUrlConfigurationSeedUrlArgs{
				Url: pulumi.String(u),
			})
		}
		web := &bedrock.AgentDataSourceDataSourceConfigurationWebConfigurationArgs{
			SourceConfiguration: &bedrock.AgentDataSourceDataSourceConfigurationWebConfigurationSourceConfigurationArgs{
				UrlConfiguration: &bedrock.AgentDataSourceDataSourceConfigurationWebConfigurationSourceConfigurationUrlConfigurationArgs{
					SeedUrls: seedUrls,
				},
			},
		}
		crawler := &bedrock.AgentDataSourceDataSourceConfigurationWebConfigurationCrawlerConfigurationArgs{}
		if d.GetWeb().Scope != "" {
			crawler.Scope = pulumi.String(d.GetWeb().Scope)
		}
		if len(d.GetWeb().InclusionFilters) > 0 {
			crawler.InclusionFilters = pulumi.ToStringArray(d.GetWeb().InclusionFilters)
		}
		if len(d.GetWeb().ExclusionFilters) > 0 {
			crawler.ExclusionFilters = pulumi.ToStringArray(d.GetWeb().ExclusionFilters)
		}
		if d.GetWeb().UserAgent != "" {
			crawler.UserAgent = pulumi.String(d.GetWeb().UserAgent)
		}
		if d.GetWeb().MaxPages != 0 || d.GetWeb().RateLimit != 0 {
			limits := &bedrock.AgentDataSourceDataSourceConfigurationWebConfigurationCrawlerConfigurationCrawlerLimitsArgs{}
			if d.GetWeb().MaxPages != 0 {
				limits.MaxPages = pulumi.Int(int(d.GetWeb().MaxPages))
			}
			if d.GetWeb().RateLimit != 0 {
				limits.RateLimit = pulumi.Int(int(d.GetWeb().RateLimit))
			}
			crawler.CrawlerLimits = limits
		}
		web.CrawlerConfiguration = crawler
		configuration.WebConfiguration = web

	case d.GetConfluence() != nil:
		configuration.Type = pulumi.String("CONFLUENCE")
		confluence := &bedrock.AgentDataSourceDataSourceConfigurationConfluenceConfigurationArgs{
			SourceConfiguration: &bedrock.AgentDataSourceDataSourceConfigurationConfluenceConfigurationSourceConfigurationArgs{
				// SAAS is the only Confluence host type AWS defines -- the
				// module owns the constant.
				HostType:             pulumi.String("SAAS"),
				HostUrl:              pulumi.String(d.GetConfluence().HostUrl),
				AuthType:             pulumi.String(d.GetConfluence().AuthType),
				CredentialsSecretArn: pulumi.String(d.GetConfluence().CredentialsSecretArn.GetValue()),
			},
		}
		if len(d.GetConfluence().Filters) > 0 {
			var filters bedrock.AgentDataSourceDataSourceConfigurationConfluenceConfigurationCrawlerConfigurationFilterConfigurationPatternObjectFilterFilterArray
			for _, f := range d.GetConfluence().Filters {
				filter := &bedrock.AgentDataSourceDataSourceConfigurationConfluenceConfigurationCrawlerConfigurationFilterConfigurationPatternObjectFilterFilterArgs{
					ObjectType: pulumi.String(f.ObjectType),
				}
				if len(f.InclusionFilters) > 0 {
					filter.InclusionFilters = pulumi.ToStringArray(f.InclusionFilters)
				}
				if len(f.ExclusionFilters) > 0 {
					filter.ExclusionFilters = pulumi.ToStringArray(f.ExclusionFilters)
				}
				filters = append(filters, filter)
			}
			confluence.CrawlerConfiguration = &bedrock.AgentDataSourceDataSourceConfigurationConfluenceConfigurationCrawlerConfigurationArgs{
				FilterConfiguration: &bedrock.AgentDataSourceDataSourceConfigurationConfluenceConfigurationCrawlerConfigurationFilterConfigurationArgs{
					// PATTERN is the only filter type AWS defines -- the
					// module owns the constant.
					Type: pulumi.String("PATTERN"),
					PatternObjectFilters: bedrock.AgentDataSourceDataSourceConfigurationConfluenceConfigurationCrawlerConfigurationFilterConfigurationPatternObjectFilterArray{
						&bedrock.AgentDataSourceDataSourceConfigurationConfluenceConfigurationCrawlerConfigurationFilterConfigurationPatternObjectFilterArgs{
							Filters: filters,
						},
					},
				},
			}
		}
		configuration.ConfluenceConfiguration = confluence

	case d.GetSalesforce() != nil:
		configuration.Type = pulumi.String("SALESFORCE")
		salesforce := &bedrock.AgentDataSourceDataSourceConfigurationSalesforceConfigurationArgs{
			SourceConfiguration: &bedrock.AgentDataSourceDataSourceConfigurationSalesforceConfigurationSourceConfigurationArgs{
				// OAUTH2_CLIENT_CREDENTIALS is the only Salesforce auth
				// type AWS defines -- the module owns the constant.
				AuthType:             pulumi.String("OAUTH2_CLIENT_CREDENTIALS"),
				HostUrl:              pulumi.String(d.GetSalesforce().HostUrl),
				CredentialsSecretArn: pulumi.String(d.GetSalesforce().CredentialsSecretArn.GetValue()),
			},
		}
		if len(d.GetSalesforce().Filters) > 0 {
			var filters bedrock.AgentDataSourceDataSourceConfigurationSalesforceConfigurationCrawlerConfigurationFilterConfigurationPatternObjectFilterFilterArray
			for _, f := range d.GetSalesforce().Filters {
				filter := &bedrock.AgentDataSourceDataSourceConfigurationSalesforceConfigurationCrawlerConfigurationFilterConfigurationPatternObjectFilterFilterArgs{
					ObjectType: pulumi.String(f.ObjectType),
				}
				if len(f.InclusionFilters) > 0 {
					filter.InclusionFilters = pulumi.ToStringArray(f.InclusionFilters)
				}
				if len(f.ExclusionFilters) > 0 {
					filter.ExclusionFilters = pulumi.ToStringArray(f.ExclusionFilters)
				}
				filters = append(filters, filter)
			}
			salesforce.CrawlerConfiguration = &bedrock.AgentDataSourceDataSourceConfigurationSalesforceConfigurationCrawlerConfigurationArgs{
				FilterConfiguration: &bedrock.AgentDataSourceDataSourceConfigurationSalesforceConfigurationCrawlerConfigurationFilterConfigurationArgs{
					Type: pulumi.String("PATTERN"),
					PatternObjectFilters: bedrock.AgentDataSourceDataSourceConfigurationSalesforceConfigurationCrawlerConfigurationFilterConfigurationPatternObjectFilterArray{
						&bedrock.AgentDataSourceDataSourceConfigurationSalesforceConfigurationCrawlerConfigurationFilterConfigurationPatternObjectFilterArgs{
							Filters: filters,
						},
					},
				},
			}
		}
		configuration.SalesforceConfiguration = salesforce

	case d.GetSharepoint() != nil:
		configuration.Type = pulumi.String("SHAREPOINT")
		sourceConfiguration := &bedrock.AgentDataSourceDataSourceConfigurationSharePointConfigurationSourceConfigurationArgs{
			// ONLINE is the only SharePoint host type AWS defines -- the
			// module owns the constant.
			HostType:             pulumi.String("ONLINE"),
			SiteUrls:             pulumi.ToStringArray(d.GetSharepoint().SiteUrls),
			Domain:               pulumi.String(d.GetSharepoint().Domain),
			AuthType:             pulumi.String(d.GetSharepoint().AuthType),
			CredentialsSecretArn: pulumi.String(d.GetSharepoint().CredentialsSecretArn.GetValue()),
		}
		if d.GetSharepoint().TenantId != "" {
			sourceConfiguration.TenantId = pulumi.String(d.GetSharepoint().TenantId)
		}
		sharepoint := &bedrock.AgentDataSourceDataSourceConfigurationSharePointConfigurationArgs{
			SourceConfiguration: sourceConfiguration,
		}
		if len(d.GetSharepoint().Filters) > 0 {
			var filters bedrock.AgentDataSourceDataSourceConfigurationSharePointConfigurationCrawlerConfigurationFilterConfigurationPatternObjectFilterFilterArray
			for _, f := range d.GetSharepoint().Filters {
				filter := &bedrock.AgentDataSourceDataSourceConfigurationSharePointConfigurationCrawlerConfigurationFilterConfigurationPatternObjectFilterFilterArgs{
					ObjectType: pulumi.String(f.ObjectType),
				}
				if len(f.InclusionFilters) > 0 {
					filter.InclusionFilters = pulumi.ToStringArray(f.InclusionFilters)
				}
				if len(f.ExclusionFilters) > 0 {
					filter.ExclusionFilters = pulumi.ToStringArray(f.ExclusionFilters)
				}
				filters = append(filters, filter)
			}
			sharepoint.CrawlerConfiguration = &bedrock.AgentDataSourceDataSourceConfigurationSharePointConfigurationCrawlerConfigurationArgs{
				FilterConfiguration: &bedrock.AgentDataSourceDataSourceConfigurationSharePointConfigurationCrawlerConfigurationFilterConfigurationArgs{
					Type: pulumi.String("PATTERN"),
					PatternObjectFilters: bedrock.AgentDataSourceDataSourceConfigurationSharePointConfigurationCrawlerConfigurationFilterConfigurationPatternObjectFilterArray{
						&bedrock.AgentDataSourceDataSourceConfigurationSharePointConfigurationCrawlerConfigurationFilterConfigurationPatternObjectFilterArgs{
							Filters: filters,
						},
					},
				},
			}
		}
		configuration.SharePointConfiguration = sharepoint

	case d.GetManagedConnector() != nil:
		configuration.Type = pulumi.String("MANAGED_KNOWLEDGE_BASE_CONNECTOR")
		managed := &bedrock.AgentDataSourceDataSourceConfigurationManagedKnowledgeBaseConnectorConfigurationArgs{}
		if d.GetManagedConnector().ConnectorParameters != nil {
			parametersJson, err := json.Marshal(d.GetManagedConnector().ConnectorParameters.AsMap())
			if err != nil {
				return nil, errors.Wrap(err, "marshal connector parameters")
			}
			managed.ConnectorParameters = pulumi.String(string(parametersJson))
		}
		if d.GetManagedConnector().DeletionProtection != nil {
			protection := &bedrock.AgentDataSourceDataSourceConfigurationManagedKnowledgeBaseConnectorConfigurationDeletionProtectionConfigurationArgs{
				DeletionProtectionStatus: pulumi.String(enabledOrDisabled(d.GetManagedConnector().DeletionProtection.Enabled)),
			}
			if d.GetManagedConnector().DeletionProtection.ThresholdPercent != 0 {
				protection.DeletionProtectionThreshold = pulumi.Int(int(d.GetManagedConnector().DeletionProtection.ThresholdPercent))
			}
			managed.DeletionProtectionConfiguration = protection
		}
		if d.GetManagedConnector().MediaExtraction != nil {
			managed.MediaExtractionConfiguration = &bedrock.AgentDataSourceDataSourceConfigurationManagedKnowledgeBaseConnectorConfigurationMediaExtractionConfigurationArgs{
				AudioExtractionConfiguration: &bedrock.AgentDataSourceDataSourceConfigurationManagedKnowledgeBaseConnectorConfigurationMediaExtractionConfigurationAudioExtractionConfigurationArgs{
					AudioExtractionStatus: pulumi.String(enabledOrDisabled(d.GetManagedConnector().MediaExtraction.Audio)),
				},
				ImageExtractionConfiguration: &bedrock.AgentDataSourceDataSourceConfigurationManagedKnowledgeBaseConnectorConfigurationMediaExtractionConfigurationImageExtractionConfigurationArgs{
					ImageExtractionStatus: pulumi.String(enabledOrDisabled(d.GetManagedConnector().MediaExtraction.Image)),
				},
				VideoExtractionConfiguration: &bedrock.AgentDataSourceDataSourceConfigurationManagedKnowledgeBaseConnectorConfigurationMediaExtractionConfigurationVideoExtractionConfigurationArgs{
					VideoExtractionStatus: pulumi.String(enabledOrDisabled(d.GetManagedConnector().MediaExtraction.Video)),
				},
			}
		}
		configuration.ManagedKnowledgeBaseConnectorConfiguration = managed
	}
	args.DataSourceConfiguration = configuration

	if d.VectorIngestion != nil {
		ingestion := &bedrock.AgentDataSourceVectorIngestionConfigurationArgs{}
		if d.VectorIngestion.Chunking != nil {
			chunking := &bedrock.AgentDataSourceVectorIngestionConfigurationChunkingConfigurationArgs{
				ChunkingStrategy: pulumi.String(d.VectorIngestion.Chunking.Strategy),
			}
			if d.VectorIngestion.Chunking.FixedSize != nil {
				chunking.FixedSizeChunkingConfiguration = &bedrock.AgentDataSourceVectorIngestionConfigurationChunkingConfigurationFixedSizeChunkingConfigurationArgs{
					MaxTokens:         pulumi.Int(int(d.VectorIngestion.Chunking.FixedSize.MaxTokens)),
					OverlapPercentage: pulumi.Int(int(d.VectorIngestion.Chunking.FixedSize.OverlapPercentage)),
				}
			}
			if d.VectorIngestion.Chunking.Hierarchical != nil {
				var levels bedrock.AgentDataSourceVectorIngestionConfigurationChunkingConfigurationHierarchicalChunkingConfigurationLevelConfigurationArray
				for _, l := range d.VectorIngestion.Chunking.Hierarchical.Levels {
					levels = append(levels, &bedrock.AgentDataSourceVectorIngestionConfigurationChunkingConfigurationHierarchicalChunkingConfigurationLevelConfigurationArgs{
						MaxTokens: pulumi.Int(int(l.MaxTokens)),
					})
				}
				chunking.HierarchicalChunkingConfiguration = &bedrock.AgentDataSourceVectorIngestionConfigurationChunkingConfigurationHierarchicalChunkingConfigurationArgs{
					OverlapTokens:       pulumi.Int(int(d.VectorIngestion.Chunking.Hierarchical.OverlapTokens)),
					LevelConfigurations: levels,
				}
			}
			if d.VectorIngestion.Chunking.Semantic != nil {
				chunking.SemanticChunkingConfiguration = &bedrock.AgentDataSourceVectorIngestionConfigurationChunkingConfigurationSemanticChunkingConfigurationArgs{
					BreakpointPercentileThreshold: pulumi.Int(int(d.VectorIngestion.Chunking.Semantic.BreakpointPercentileThreshold)),
					BufferSize:                    pulumi.Int(int(d.VectorIngestion.Chunking.Semantic.BufferSize)),
					// The provider spells this one singular.
					MaxToken: pulumi.Int(int(d.VectorIngestion.Chunking.Semantic.MaxTokens)),
				}
			}
			ingestion.ChunkingConfiguration = chunking
		}
		if d.VectorIngestion.Parsing != nil {
			parsing := &bedrock.AgentDataSourceVectorIngestionConfigurationParsingConfigurationArgs{
				ParsingStrategy: pulumi.String(d.VectorIngestion.Parsing.Strategy),
			}
			// MULTIMODAL is the only parsing modality AWS defines -- the
			// spec models it as a bool and the module owns the constant.
			if d.VectorIngestion.Parsing.Strategy == "BEDROCK_DATA_AUTOMATION" && d.VectorIngestion.Parsing.Multimodal {
				parsing.BedrockDataAutomationConfiguration = &bedrock.AgentDataSourceVectorIngestionConfigurationParsingConfigurationBedrockDataAutomationConfigurationArgs{
					ParsingModality: pulumi.String("MULTIMODAL"),
				}
			}
			if d.VectorIngestion.Parsing.FoundationModel != nil {
				foundationModel := &bedrock.AgentDataSourceVectorIngestionConfigurationParsingConfigurationBedrockFoundationModelConfigurationArgs{
					ModelArn: pulumi.String(d.VectorIngestion.Parsing.FoundationModel.ModelArn),
				}
				if d.VectorIngestion.Parsing.FoundationModel.Multimodal {
					foundationModel.ParsingModality = pulumi.String("MULTIMODAL")
				}
				if d.VectorIngestion.Parsing.FoundationModel.ParsingPrompt != "" {
					foundationModel.ParsingPrompt = &bedrock.AgentDataSourceVectorIngestionConfigurationParsingConfigurationBedrockFoundationModelConfigurationParsingPromptArgs{
						ParsingPromptString: pulumi.String(d.VectorIngestion.Parsing.FoundationModel.ParsingPrompt),
					}
				}
				parsing.BedrockFoundationModelConfiguration = foundationModel
			}
			ingestion.ParsingConfiguration = parsing
		}
		if d.VectorIngestion.CustomTransformation != nil {
			ingestion.CustomTransformationConfiguration = &bedrock.AgentDataSourceVectorIngestionConfigurationCustomTransformationConfigurationArgs{
				IntermediateStorage: &bedrock.AgentDataSourceVectorIngestionConfigurationCustomTransformationConfigurationIntermediateStorageArgs{
					S3Location: &bedrock.AgentDataSourceVectorIngestionConfigurationCustomTransformationConfigurationIntermediateStorageS3LocationArgs{
						Uri: pulumi.String(d.VectorIngestion.CustomTransformation.IntermediateS3Uri),
					},
				},
				Transformation: &bedrock.AgentDataSourceVectorIngestionConfigurationCustomTransformationConfigurationTransformationArgs{
					// POST_CHUNKING is the only transformation step AWS
					// defines -- the module owns the constant.
					StepToApply: pulumi.String("POST_CHUNKING"),
					TransformationFunction: &bedrock.AgentDataSourceVectorIngestionConfigurationCustomTransformationConfigurationTransformationTransformationFunctionArgs{
						TransformationLambdaConfiguration: &bedrock.AgentDataSourceVectorIngestionConfigurationCustomTransformationConfigurationTransformationTransformationFunctionTransformationLambdaConfigurationArgs{
							LambdaArn: pulumi.String(d.VectorIngestion.CustomTransformation.LambdaArn.GetValue()),
						},
					},
				},
			}
		}
		args.VectorIngestionConfiguration = ingestion
	}

	return args, nil
}

func enabledOrDisabled(enabled bool) string {
	if enabled {
		return "ENABLED"
	}
	return "DISABLED"
}

func sortedDataSources(in []*awsbedrockknowledgebasev1alpha1.AwsBedrockKnowledgeBaseDataSource) []*awsbedrockknowledgebasev1alpha1.AwsBedrockKnowledgeBaseDataSource {
	out := append([]*awsbedrockknowledgebasev1alpha1.AwsBedrockKnowledgeBaseDataSource{}, in...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
