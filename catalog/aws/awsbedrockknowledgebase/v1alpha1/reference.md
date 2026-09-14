# AwsBedrockKnowledgeBase

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

AwsBedrockKnowledgeBaseSpec defines the desired configuration for an
Amazon Bedrock knowledge base - the RAG (retrieval-augmented generation)
store that ingests documents from data sources, embeds them, and answers
retrieval queries for agents and applications.

The knowledge base's name is taken from `metadata.name` (letters,
digits, hyphen, underscore; AWS rejects spaces and dots).

Exactly ONE knowledge-base type is configured:
  - `vector`: embeddings-based retrieval over a vector store you own
    (set `storage` to say which one) - the classic RAG shape;
  - `managed`: embeddings-based retrieval where AWS manages the vector
    store for you - no `storage` block at all;
  - `kendra`: retrieval delegated to an existing Amazon Kendra index;
  - `sql`: natural-language-to-SQL over a Redshift warehouse (no
    embeddings, no vector store).

Almost every field below is create-time only - changing anything except
the description, the role, and data sources replaces the knowledge base
(AWS re-ingests from the data sources afterwards). Knowledge bases are
free to create; embedding and retrieval bill per use, and a managed
vector store bills for what AWS provisions behind it.

## Example

```yaml
# Canonical AwsBedrockKnowledgeBase example (hack/dev manifest and refgen
# Example source): a VECTOR knowledge base on S3 Vectors with a tuned
# embedding model, an S3 data source with fixed-size chunking, and a web
# crawler source. Literal ARNs stand in for composed references so the
# offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockKnowledgeBase
metadata:
  name: product-docs
  id: product-docs
  org: test-org
  env: dev
spec:
  region: us-west-2
  description: Product documentation retrieval store
  roleArn:
    value: arn:aws:iam::123456789012:role/bedrock-kb-role
  vector:
    embeddingModelArn: arn:aws:bedrock:us-west-2::foundation-model/amazon.titan-embed-text-v2:0
    embeddingModel:
      dimensions: 256
      embeddingDataType: FLOAT32
  storage:
    s3Vectors:
      indexArn: arn:aws:s3vectors:us-west-2:123456789012:bucket/product-docs-vectors/index/docs-index
  dataSources:
    - name: manuals
      description: Product manuals bucket
      dataDeletionPolicy: DELETE
      s3:
        bucketArn:
          value: arn:aws:s3:::product-docs-manuals
        inclusionPrefix: manuals/
      vectorIngestion:
        chunking:
          strategy: FIXED_SIZE
          fixedSize:
            maxTokens: 300
            overlapPercentage: 20
    - name: docs-site
      description: Public documentation site crawl
      web:
        seedUrls:
          - https://docs.example.com
        scope: HOST_ONLY
        maxPages: 500
        rateLimit: 60
      vectorIngestion:
        chunking:
          strategy: HIERARCHICAL
          hierarchical:
            overlapTokens: 60
            levels:
              - maxTokens: 1500
              - maxTokens: 300
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.vector` | `AwsBedrockKnowledgeBaseVectorConfig` |  |  |  |
| `spec.vector.embeddingModelArn` | `string` | yes |  |  |
| `spec.vector.embeddingModel` | `AwsBedrockKnowledgeBaseEmbeddingModelConfig` |  |  |  |
| `spec.vector.embeddingModel.dimensions` | `int32` |  |  |  |
| `spec.vector.embeddingModel.embeddingDataType` | `string` |  |  |  |
| `spec.vector.embeddingModel.audioSegmentationSeconds` | `int32` |  |  |  |
| `spec.vector.embeddingModel.videoSegmentationSeconds` | `int32` |  |  |  |
| `spec.vector.supplementalDataS3Uri` | `string` |  |  |  |
| `spec.managed` | `AwsBedrockKnowledgeBaseManagedConfig` |  |  |  |
| `spec.managed.embeddingModelArn` | `string` |  |  |  |
| `spec.managed.embeddingModel` | `AwsBedrockKnowledgeBaseEmbeddingModelConfig` |  |  |  |
| `spec.managed.embeddingModel.dimensions` | `int32` |  |  |  |
| `spec.managed.embeddingModel.embeddingDataType` | `string` |  |  |  |
| `spec.managed.embeddingModel.audioSegmentationSeconds` | `int32` |  |  |  |
| `spec.managed.embeddingModel.videoSegmentationSeconds` | `int32` |  |  |  |
| `spec.managed.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.kendra` | `AwsBedrockKnowledgeBaseKendraConfig` |  |  |  |
| `spec.kendra.kendraIndexArn` | `string` |  |  |  |
| `spec.sql` | `AwsBedrockKnowledgeBaseSqlConfig` |  |  |  |
| `spec.sql.provisioned` | `AwsBedrockKnowledgeBaseRedshiftProvisioned` |  |  |  |
| `spec.sql.provisioned.clusterIdentifier` | `string \| valueFrom` | yes |  | AwsRedshiftCluster (`status.outputs.cluster_identifier`) |
| `spec.sql.provisioned.auth` | `AwsBedrockKnowledgeBaseRedshiftProvisionedAuth` | yes |  |  |
| `spec.sql.provisioned.auth.type` | `string` |  |  |  |
| `spec.sql.provisioned.auth.databaseUser` | `string` |  |  |  |
| `spec.sql.provisioned.auth.usernamePasswordSecretArn` | `string \| valueFrom` |  |  | AwsSecretsManagerSecret (`status.outputs.secret_arn`) |
| `spec.sql.serverless` | `AwsBedrockKnowledgeBaseRedshiftServerless` |  |  |  |
| `spec.sql.serverless.workgroupArn` | `string \| valueFrom` | yes |  | AwsRedshiftServerlessWorkgroup (`status.outputs.arn`) |
| `spec.sql.serverless.auth` | `AwsBedrockKnowledgeBaseRedshiftServerlessAuth` | yes |  |  |
| `spec.sql.serverless.auth.type` | `string` |  |  |  |
| `spec.sql.serverless.auth.usernamePasswordSecretArn` | `string \| valueFrom` |  |  | AwsSecretsManagerSecret (`status.outputs.secret_arn`) |
| `spec.sql.warehouse` | `AwsBedrockKnowledgeBaseRedshiftStorage` | yes |  |  |
| `spec.sql.warehouse.dataCatalog` | `AwsBedrockKnowledgeBaseDataCatalogStorage` |  |  |  |
| `spec.sql.warehouse.dataCatalog.tableNames` | `[]string` | yes |  |  |
| `spec.sql.warehouse.redshift` | `AwsBedrockKnowledgeBaseRedshiftDatabaseStorage` |  |  |  |
| `spec.sql.warehouse.redshift.databaseName` | `string` | yes |  |  |
| `spec.sql.queryGeneration` | `AwsBedrockKnowledgeBaseQueryGeneration` |  |  |  |
| `spec.sql.queryGeneration.executionTimeoutSeconds` | `int32` |  |  |  |
| `spec.sql.queryGeneration.curatedQueries` | `[]AwsBedrockKnowledgeBaseCuratedQuery` |  |  |  |
| `spec.sql.queryGeneration.curatedQueries[].naturalLanguage` | `string` | yes |  |  |
| `spec.sql.queryGeneration.curatedQueries[].sql` | `string` | yes |  |  |
| `spec.sql.queryGeneration.tables` | `[]AwsBedrockKnowledgeBaseQueryTable` |  |  |  |
| `spec.sql.queryGeneration.tables[].name` | `string` | yes |  |  |
| `spec.sql.queryGeneration.tables[].description` | `string` |  |  |  |
| `spec.sql.queryGeneration.tables[].inclusion` | `string` |  |  |  |
| `spec.sql.queryGeneration.tables[].columns` | `[]AwsBedrockKnowledgeBaseQueryColumn` |  |  |  |
| `spec.sql.queryGeneration.tables[].columns[].name` | `string` | yes |  |  |
| `spec.sql.queryGeneration.tables[].columns[].description` | `string` |  |  |  |
| `spec.sql.queryGeneration.tables[].columns[].inclusion` | `string` |  |  |  |
| `spec.storage` | `AwsBedrockKnowledgeBaseStorage` |  |  |  |
| `spec.storage.opensearchServerless` | `AwsBedrockKnowledgeBaseOpenSearchServerlessStorage` |  |  |  |
| `spec.storage.opensearchServerless.collectionArn` | `string \| valueFrom` | yes |  | AwsOpenSearchServerlessCollection (`status.outputs.collection_arn`) |
| `spec.storage.opensearchServerless.vectorIndexName` | `string` | yes |  |  |
| `spec.storage.opensearchServerless.fieldMapping` | `AwsBedrockKnowledgeBaseFieldMapping` | yes |  |  |
| `spec.storage.opensearchServerless.fieldMapping.vectorField` | `string` | yes |  |  |
| `spec.storage.opensearchServerless.fieldMapping.textField` | `string` | yes |  |  |
| `spec.storage.opensearchServerless.fieldMapping.metadataField` | `string` | yes |  |  |
| `spec.storage.opensearchManaged` | `AwsBedrockKnowledgeBaseOpenSearchManagedStorage` |  |  |  |
| `spec.storage.opensearchManaged.domainArn` | `string \| valueFrom` | yes |  | AwsOpenSearchDomain (`status.outputs.domain_arn`) |
| `spec.storage.opensearchManaged.domainEndpoint` | `string` |  |  |  |
| `spec.storage.opensearchManaged.vectorIndexName` | `string` | yes |  |  |
| `spec.storage.opensearchManaged.fieldMapping` | `AwsBedrockKnowledgeBaseFieldMapping` | yes |  |  |
| `spec.storage.opensearchManaged.fieldMapping.vectorField` | `string` | yes |  |  |
| `spec.storage.opensearchManaged.fieldMapping.textField` | `string` | yes |  |  |
| `spec.storage.opensearchManaged.fieldMapping.metadataField` | `string` | yes |  |  |
| `spec.storage.s3Vectors` | `AwsBedrockKnowledgeBaseS3VectorsStorage` |  |  |  |
| `spec.storage.s3Vectors.indexArn` | `string` |  |  |  |
| `spec.storage.s3Vectors.indexName` | `string` |  |  |  |
| `spec.storage.s3Vectors.vectorBucketArn` | `string` |  |  |  |
| `spec.storage.rds` | `AwsBedrockKnowledgeBaseRdsStorage` |  |  |  |
| `spec.storage.rds.resourceArn` | `string \| valueFrom` | yes |  | AwsRdsCluster (`status.outputs.arn`) |
| `spec.storage.rds.credentialsSecretArn` | `string \| valueFrom` | yes |  | AwsSecretsManagerSecret (`status.outputs.secret_arn`) |
| `spec.storage.rds.databaseName` | `string` | yes |  |  |
| `spec.storage.rds.tableName` | `string` | yes |  |  |
| `spec.storage.rds.fieldMapping` | `AwsBedrockKnowledgeBaseRdsFieldMapping` | yes |  |  |
| `spec.storage.rds.fieldMapping.vectorField` | `string` | yes |  |  |
| `spec.storage.rds.fieldMapping.textField` | `string` | yes |  |  |
| `spec.storage.rds.fieldMapping.metadataField` | `string` | yes |  |  |
| `spec.storage.rds.fieldMapping.primaryKeyField` | `string` | yes |  |  |
| `spec.storage.rds.fieldMapping.customMetadataField` | `string` |  |  |  |
| `spec.storage.pinecone` | `AwsBedrockKnowledgeBasePineconeStorage` |  |  |  |
| `spec.storage.pinecone.connectionString` | `string` | yes |  |  |
| `spec.storage.pinecone.credentialsSecretArn` | `string \| valueFrom` | yes |  | AwsSecretsManagerSecret (`status.outputs.secret_arn`) |
| `spec.storage.pinecone.namespace` | `string` |  |  |  |
| `spec.storage.pinecone.fieldMapping` | `AwsBedrockKnowledgeBaseTextMetadataFieldMapping` | yes |  |  |
| `spec.storage.pinecone.fieldMapping.textField` | `string` | yes |  |  |
| `spec.storage.pinecone.fieldMapping.metadataField` | `string` | yes |  |  |
| `spec.storage.mongodbAtlas` | `AwsBedrockKnowledgeBaseMongoDbAtlasStorage` |  |  |  |
| `spec.storage.mongodbAtlas.endpoint` | `string` | yes |  |  |
| `spec.storage.mongodbAtlas.databaseName` | `string` | yes |  |  |
| `spec.storage.mongodbAtlas.collectionName` | `string` | yes |  |  |
| `spec.storage.mongodbAtlas.vectorIndexName` | `string` | yes |  |  |
| `spec.storage.mongodbAtlas.textIndexName` | `string` |  |  |  |
| `spec.storage.mongodbAtlas.endpointServiceName` | `string` |  |  |  |
| `spec.storage.mongodbAtlas.credentialsSecretArn` | `string \| valueFrom` | yes |  | AwsSecretsManagerSecret (`status.outputs.secret_arn`) |
| `spec.storage.mongodbAtlas.fieldMapping` | `AwsBedrockKnowledgeBaseFieldMapping` | yes |  |  |
| `spec.storage.mongodbAtlas.fieldMapping.vectorField` | `string` | yes |  |  |
| `spec.storage.mongodbAtlas.fieldMapping.textField` | `string` | yes |  |  |
| `spec.storage.mongodbAtlas.fieldMapping.metadataField` | `string` | yes |  |  |
| `spec.storage.neptuneAnalytics` | `AwsBedrockKnowledgeBaseNeptuneAnalyticsStorage` |  |  |  |
| `spec.storage.neptuneAnalytics.graphArn` | `string` |  |  |  |
| `spec.storage.neptuneAnalytics.fieldMapping` | `AwsBedrockKnowledgeBaseTextMetadataFieldMapping` | yes |  |  |
| `spec.storage.neptuneAnalytics.fieldMapping.textField` | `string` | yes |  |  |
| `spec.storage.neptuneAnalytics.fieldMapping.metadataField` | `string` | yes |  |  |
| `spec.storage.redisEnterpriseCloud` | `AwsBedrockKnowledgeBaseRedisEnterpriseCloudStorage` |  |  |  |
| `spec.storage.redisEnterpriseCloud.endpoint` | `string` | yes |  |  |
| `spec.storage.redisEnterpriseCloud.vectorIndexName` | `string` | yes |  |  |
| `spec.storage.redisEnterpriseCloud.credentialsSecretArn` | `string \| valueFrom` | yes |  | AwsSecretsManagerSecret (`status.outputs.secret_arn`) |
| `spec.storage.redisEnterpriseCloud.fieldMapping` | `AwsBedrockKnowledgeBaseRedisFieldMapping` |  |  |  |
| `spec.storage.redisEnterpriseCloud.fieldMapping.vectorField` | `string` |  |  |  |
| `spec.storage.redisEnterpriseCloud.fieldMapping.textField` | `string` |  |  |  |
| `spec.storage.redisEnterpriseCloud.fieldMapping.metadataField` | `string` |  |  |  |
| `spec.dataSources` | `[]AwsBedrockKnowledgeBaseDataSource` |  |  |  |
| `spec.dataSources[].name` | `string` | yes |  |  |
| `spec.dataSources[].description` | `string` |  |  |  |
| `spec.dataSources[].dataDeletionPolicy` | `string` |  |  |  |
| `spec.dataSources[].kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.dataSources[].s3` | `AwsBedrockKnowledgeBaseS3DataSource` |  |  |  |
| `spec.dataSources[].s3.bucketArn` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_arn`) |
| `spec.dataSources[].s3.inclusionPrefix` | `string` | yes |  |  |
| `spec.dataSources[].s3.bucketOwnerAccountId` | `string` |  |  |  |
| `spec.dataSources[].web` | `AwsBedrockKnowledgeBaseWebDataSource` |  |  |  |
| `spec.dataSources[].web.seedUrls` | `[]string` | yes |  |  |
| `spec.dataSources[].web.scope` | `string` |  |  |  |
| `spec.dataSources[].web.inclusionFilters` | `[]string` |  |  |  |
| `spec.dataSources[].web.exclusionFilters` | `[]string` |  |  |  |
| `spec.dataSources[].web.userAgent` | `string` | yes |  |  |
| `spec.dataSources[].web.maxPages` | `int32` |  |  |  |
| `spec.dataSources[].web.rateLimit` | `int32` |  |  |  |
| `spec.dataSources[].confluence` | `AwsBedrockKnowledgeBaseConfluenceDataSource` |  |  |  |
| `spec.dataSources[].confluence.hostUrl` | `string` |  |  |  |
| `spec.dataSources[].confluence.authType` | `string` |  |  |  |
| `spec.dataSources[].confluence.credentialsSecretArn` | `string \| valueFrom` | yes |  | AwsSecretsManagerSecret (`status.outputs.secret_arn`) |
| `spec.dataSources[].confluence.filters` | `[]AwsBedrockKnowledgeBaseCrawlFilter` |  |  |  |
| `spec.dataSources[].confluence.filters[].objectType` | `string` | yes |  |  |
| `spec.dataSources[].confluence.filters[].inclusionFilters` | `[]string` |  |  |  |
| `spec.dataSources[].confluence.filters[].exclusionFilters` | `[]string` |  |  |  |
| `spec.dataSources[].salesforce` | `AwsBedrockKnowledgeBaseSalesforceDataSource` |  |  |  |
| `spec.dataSources[].salesforce.hostUrl` | `string` | yes |  |  |
| `spec.dataSources[].salesforce.credentialsSecretArn` | `string \| valueFrom` | yes |  | AwsSecretsManagerSecret (`status.outputs.secret_arn`) |
| `spec.dataSources[].salesforce.filters` | `[]AwsBedrockKnowledgeBaseCrawlFilter` |  |  |  |
| `spec.dataSources[].salesforce.filters[].objectType` | `string` | yes |  |  |
| `spec.dataSources[].salesforce.filters[].inclusionFilters` | `[]string` |  |  |  |
| `spec.dataSources[].salesforce.filters[].exclusionFilters` | `[]string` |  |  |  |
| `spec.dataSources[].sharepoint` | `AwsBedrockKnowledgeBaseSharePointDataSource` |  |  |  |
| `spec.dataSources[].sharepoint.siteUrls` | `[]string` | yes |  |  |
| `spec.dataSources[].sharepoint.domain` | `string` | yes |  |  |
| `spec.dataSources[].sharepoint.tenantId` | `string` |  |  |  |
| `spec.dataSources[].sharepoint.authType` | `string` |  |  |  |
| `spec.dataSources[].sharepoint.credentialsSecretArn` | `string \| valueFrom` | yes |  | AwsSecretsManagerSecret (`status.outputs.secret_arn`) |
| `spec.dataSources[].sharepoint.filters` | `[]AwsBedrockKnowledgeBaseCrawlFilter` |  |  |  |
| `spec.dataSources[].sharepoint.filters[].objectType` | `string` | yes |  |  |
| `spec.dataSources[].sharepoint.filters[].inclusionFilters` | `[]string` |  |  |  |
| `spec.dataSources[].sharepoint.filters[].exclusionFilters` | `[]string` |  |  |  |
| `spec.dataSources[].managedConnector` | `AwsBedrockKnowledgeBaseManagedConnectorDataSource` |  |  |  |
| `spec.dataSources[].managedConnector.connectorParameters` | `object` |  |  |  |
| `spec.dataSources[].managedConnector.deletionProtection` | `AwsBedrockKnowledgeBaseDeletionProtection` |  |  |  |
| `spec.dataSources[].managedConnector.deletionProtection.enabled` | `bool` |  |  |  |
| `spec.dataSources[].managedConnector.deletionProtection.thresholdPercent` | `int32` |  |  |  |
| `spec.dataSources[].managedConnector.mediaExtraction` | `AwsBedrockKnowledgeBaseMediaExtraction` |  |  |  |
| `spec.dataSources[].managedConnector.mediaExtraction.audio` | `bool` |  |  |  |
| `spec.dataSources[].managedConnector.mediaExtraction.image` | `bool` |  |  |  |
| `spec.dataSources[].managedConnector.mediaExtraction.video` | `bool` |  |  |  |
| `spec.dataSources[].vectorIngestion` | `AwsBedrockKnowledgeBaseVectorIngestion` |  |  |  |
| `spec.dataSources[].vectorIngestion.chunking` | `AwsBedrockKnowledgeBaseChunking` |  |  |  |
| `spec.dataSources[].vectorIngestion.chunking.strategy` | `string` |  |  |  |
| `spec.dataSources[].vectorIngestion.chunking.fixedSize` | `AwsBedrockKnowledgeBaseFixedSizeChunking` |  |  |  |
| `spec.dataSources[].vectorIngestion.chunking.fixedSize.maxTokens` | `int32` |  |  |  |
| `spec.dataSources[].vectorIngestion.chunking.fixedSize.overlapPercentage` | `int32` |  |  |  |
| `spec.dataSources[].vectorIngestion.chunking.hierarchical` | `AwsBedrockKnowledgeBaseHierarchicalChunking` |  |  |  |
| `spec.dataSources[].vectorIngestion.chunking.hierarchical.overlapTokens` | `int32` |  |  |  |
| `spec.dataSources[].vectorIngestion.chunking.hierarchical.levels` | `[]AwsBedrockKnowledgeBaseChunkingLevel` | yes |  |  |
| `spec.dataSources[].vectorIngestion.chunking.hierarchical.levels[].maxTokens` | `int32` |  |  |  |
| `spec.dataSources[].vectorIngestion.chunking.semantic` | `AwsBedrockKnowledgeBaseSemanticChunking` |  |  |  |
| `spec.dataSources[].vectorIngestion.chunking.semantic.breakpointPercentileThreshold` | `int32` |  |  |  |
| `spec.dataSources[].vectorIngestion.chunking.semantic.bufferSize` | `int32` |  |  |  |
| `spec.dataSources[].vectorIngestion.chunking.semantic.maxTokens` | `int32` |  |  |  |
| `spec.dataSources[].vectorIngestion.parsing` | `AwsBedrockKnowledgeBaseParsing` |  |  |  |
| `spec.dataSources[].vectorIngestion.parsing.strategy` | `string` |  |  |  |
| `spec.dataSources[].vectorIngestion.parsing.multimodal` | `bool` |  |  |  |
| `spec.dataSources[].vectorIngestion.parsing.foundationModel` | `AwsBedrockKnowledgeBaseFoundationModelParsing` |  |  |  |
| `spec.dataSources[].vectorIngestion.parsing.foundationModel.modelArn` | `string` | yes |  |  |
| `spec.dataSources[].vectorIngestion.parsing.foundationModel.multimodal` | `bool` |  |  |  |
| `spec.dataSources[].vectorIngestion.parsing.foundationModel.parsingPrompt` | `string` |  |  |  |
| `spec.dataSources[].vectorIngestion.customTransformation` | `AwsBedrockKnowledgeBaseCustomTransformation` |  |  |  |
| `spec.dataSources[].vectorIngestion.customTransformation.intermediateS3Uri` | `string` |  |  |  |
| `spec.dataSources[].vectorIngestion.customTransformation.lambdaArn` | `string \| valueFrom` | yes |  | AwsLambda (`status.outputs.function_arn`) |

## Field Details

### spec.region

`string` · required

The AWS region where the knowledge base will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

Human-readable description shown in the Bedrock console. Updates in
place.

- rule: {"string":{"maxLen":"200"}}

### spec.roleArn

`string | valueFrom` · required

IAM role the Bedrock service assumes to operate the knowledge base
(read data sources, call the embedding model, read/write the vector
store). The role must trust bedrock.amazonaws.com.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.vector

`AwsBedrockKnowledgeBaseVectorConfig`

Embeddings-based retrieval over a vector store you own - pair with
`storage`.

### spec.vector.embeddingModelArn

`string` · required

ARN of the embedding foundation model (e.g. Titan Text Embeddings V2:
arn:aws:bedrock:<region>::foundation-model/amazon.titan-embed-text-v2:0).

- rule: {"string":{"minLen":"1"}}

### spec.vector.embeddingModel

`AwsBedrockKnowledgeBaseEmbeddingModelConfig`

Tune the embedding model's output.

### spec.vector.embeddingModel.dimensions

`int32`

Vector dimensionality (the model must support it - e.g. Titan V2
supports 256, 512, 1024). Omitted = the model's default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.vector.embeddingModel.embeddingDataType

`string`

Vector element type: FLOAT32 (default) or BINARY (compact binary
embeddings; the model and the vector store must both support it).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["FLOAT32","BINARY"]}}

### spec.vector.embeddingModel.audioSegmentationSeconds

`int32`

Segment ingested AUDIO into fixed-length chunks of this many seconds
before embedding (multimodal models only).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.vector.embeddingModel.videoSegmentationSeconds

`int32`

Segment ingested VIDEO into fixed-length chunks of this many seconds
before embedding (multimodal models only).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.vector.supplementalDataS3Uri

`string`

S3 location (s3://bucket/prefix) where AWS stores supplemental data
for multimodal retrieval (images extracted from documents). Required
when parsing extracts multimodal content.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^s3://[a-z0-9.-]+(/.*)?$"}}

### spec.managed

`AwsBedrockKnowledgeBaseManagedConfig`

Embeddings-based retrieval with the vector store fully managed by
AWS - no `storage` block.

### spec.managed.embeddingModelArn

`string`

ARN of the embedding foundation model. Omitted = AWS uses its
service-managed embedding model. Setting an ARN switches the
embedding-model type to CUSTOM; the modules derive and always send
that discriminator (AWS's embeddingModelType is CUSTOM exactly when
an ARN is brought, MANAGED otherwise - a leaf that must agree with
structure is drift surface, not configuration).

### spec.managed.embeddingModel

`AwsBedrockKnowledgeBaseEmbeddingModelConfig`

Tune the embedding model's output.

### spec.managed.embeddingModel.dimensions

`int32`

Vector dimensionality (the model must support it - e.g. Titan V2
supports 256, 512, 1024). Omitted = the model's default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.managed.embeddingModel.embeddingDataType

`string`

Vector element type: FLOAT32 (default) or BINARY (compact binary
embeddings; the model and the vector store must both support it).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["FLOAT32","BINARY"]}}

### spec.managed.embeddingModel.audioSegmentationSeconds

`int32`

Segment ingested AUDIO into fixed-length chunks of this many seconds
before embedding (multimodal models only).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.managed.embeddingModel.videoSegmentationSeconds

`int32`

Segment ingested VIDEO into fixed-length chunks of this many seconds
before embedding (multimodal models only).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.managed.kmsKeyArn

`string | valueFrom`

Customer-managed KMS key for encrypting the managed store. Without
it, AWS uses a service-managed key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.kendra

`AwsBedrockKnowledgeBaseKendraConfig`

Retrieval delegated to an existing Amazon Kendra index.

### spec.kendra.kendraIndexArn

`string`

ARN of the Kendra index that answers retrieval queries.

- rule: {"string":{"pattern":"^arn:aws[a-z-]*:kendra:[a-z0-9-]+:[0-9]{12}:index/.+$"}}

### spec.sql

`AwsBedrockKnowledgeBaseSqlConfig`

Natural-language-to-SQL over Amazon Redshift.

- rule: exactly one of provisioned or serverless must be configured

### spec.sql.provisioned

`AwsBedrockKnowledgeBaseRedshiftProvisioned`

The Redshift compute that runs generated queries - exactly one arm.

### spec.sql.provisioned.clusterIdentifier

`string | valueFrom` · required

The Redshift cluster that runs generated queries.

- references: AwsRedshiftCluster (`status.outputs.cluster_identifier`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsRedshiftCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_identifier}} -- a bare string does not parse

### spec.sql.provisioned.auth

`AwsBedrockKnowledgeBaseRedshiftProvisionedAuth` · required

How Bedrock authenticates to the cluster.

- rule: {"required":true}
- rule: USERNAME requires database_user; USERNAME_PASSWORD requires username_password_secret_arn; IAM takes neither

### spec.sql.provisioned.auth.type

`string`

IAM (role-based), USERNAME (database user via IAM temporary
credentials), or USERNAME_PASSWORD (Secrets Manager secret).

- rule: {"string":{"in":["IAM","USERNAME","USERNAME_PASSWORD"]}}

### spec.sql.provisioned.auth.databaseUser

`string`

Database user for the USERNAME type.

### spec.sql.provisioned.auth.usernamePasswordSecretArn

`string | valueFrom`

Secrets Manager secret holding username/password for the
USERNAME_PASSWORD type.

- references: AwsSecretsManagerSecret (`status.outputs.secret_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecretsManagerSecret, name: <that resource's name>, fieldPath: status.outputs.secret_arn}} -- a bare string does not parse

### spec.sql.serverless

`AwsBedrockKnowledgeBaseRedshiftServerless`

Redshift Serverless workgroup arm.

### spec.sql.serverless.workgroupArn

`string | valueFrom` · required

The Redshift Serverless workgroup that runs generated queries.

- references: AwsRedshiftServerlessWorkgroup (`status.outputs.arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsRedshiftServerlessWorkgroup, name: <that resource's name>, fieldPath: status.outputs.arn}} -- a bare string does not parse

### spec.sql.serverless.auth

`AwsBedrockKnowledgeBaseRedshiftServerlessAuth` · required

How Bedrock authenticates to the workgroup.

- rule: {"required":true}
- rule: USERNAME_PASSWORD requires username_password_secret_arn; IAM takes none

### spec.sql.serverless.auth.type

`string`

IAM (role-based) or USERNAME_PASSWORD (Secrets Manager secret).

- rule: {"string":{"in":["IAM","USERNAME_PASSWORD"]}}

### spec.sql.serverless.auth.usernamePasswordSecretArn

`string | valueFrom`

Secrets Manager secret holding username/password for the
USERNAME_PASSWORD type.

- references: AwsSecretsManagerSecret (`status.outputs.secret_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecretsManagerSecret, name: <that resource's name>, fieldPath: status.outputs.secret_arn}} -- a bare string does not parse

### spec.sql.warehouse

`AwsBedrockKnowledgeBaseRedshiftStorage` · required

Where the warehouse's table metadata lives - exactly one arm.

- rule: {"required":true}
- rule: exactly one of data_catalog or redshift must be configured

### spec.sql.warehouse.dataCatalog

`AwsBedrockKnowledgeBaseDataCatalogStorage`

Read table metadata from AWS Glue Data Catalog tables.

### spec.sql.warehouse.dataCatalog.tableNames

`[]string` · required

Glue table names the query generator may use (at least one).

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.sql.warehouse.redshift

`AwsBedrockKnowledgeBaseRedshiftDatabaseStorage`

Read table metadata directly from a Redshift database.

### spec.sql.warehouse.redshift.databaseName

`string` · required

Database name (1-200 characters).

- rule: {"string":{"minLen":"1","maxLen":"200"}}

### spec.sql.queryGeneration

`AwsBedrockKnowledgeBaseQueryGeneration`

Teach the query generator about your schema (curated examples, table
and column descriptions, inclusion/exclusion).

### spec.sql.queryGeneration.executionTimeoutSeconds

`int32`

Abort query generation after this many seconds (1-200).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":200,"gte":1}}

### spec.sql.queryGeneration.curatedQueries

`[]AwsBedrockKnowledgeBaseCuratedQuery`

Curated question-to-SQL examples that anchor generation (at most 10).

- rule: {"repeated":{"maxItems":"10"}}

### spec.sql.queryGeneration.curatedQueries[].naturalLanguage

`string` · required

The natural-language question.

- rule: {"string":{"minLen":"1"}}

### spec.sql.queryGeneration.curatedQueries[].sql

`string` · required

The SQL that correctly answers it.

- rule: {"string":{"minLen":"1"}}

### spec.sql.queryGeneration.tables

`[]AwsBedrockKnowledgeBaseQueryTable`

Table-level guidance: descriptions and inclusion/exclusion (at most
50).

- rule: {"repeated":{"maxItems":"50"}}

### spec.sql.queryGeneration.tables[].name

`string` · required

Table name (as the query engine sees it, e.g.
"database.schema.table").

- rule: {"string":{"minLen":"1"}}

### spec.sql.queryGeneration.tables[].description

`string`

What the table contains (1-200 characters when set) - context for the
query generator.

- rule: {"string":{"maxLen":"200"}}

### spec.sql.queryGeneration.tables[].inclusion

`string`

INCLUDE or EXCLUDE this table from generation. Omitted = included.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["INCLUDE","EXCLUDE"]}}

### spec.sql.queryGeneration.tables[].columns

`[]AwsBedrockKnowledgeBaseQueryColumn`

Column-level guidance within this table.

### spec.sql.queryGeneration.tables[].columns[].name

`string` · required

Column name.

- rule: {"string":{"minLen":"1"}}

### spec.sql.queryGeneration.tables[].columns[].description

`string`

What the column contains (1-200 characters when set).

- rule: {"string":{"maxLen":"200"}}

### spec.sql.queryGeneration.tables[].columns[].inclusion

`string`

INCLUDE or EXCLUDE this column. Omitted = included.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["INCLUDE","EXCLUDE"]}}

### spec.storage

`AwsBedrockKnowledgeBaseStorage`

Where the `vector` type stores and queries embeddings - exactly one
backend.

- rule: exactly one vector store backend must be configured

### spec.storage.opensearchServerless

`AwsBedrockKnowledgeBaseOpenSearchServerlessStorage`

Amazon OpenSearch Serverless collection (VECTORSEARCH type). The
vector index must already exist in the collection - AWS does not
create it.

### spec.storage.opensearchServerless.collectionArn

`string | valueFrom` · required

The VECTORSEARCH-type collection.

- references: AwsOpenSearchServerlessCollection (`status.outputs.collection_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsOpenSearchServerlessCollection, name: <that resource's name>, fieldPath: status.outputs.collection_arn}} -- a bare string does not parse

### spec.storage.opensearchServerless.vectorIndexName

`string` · required

Name of the pre-created vector index inside the collection.

- rule: {"string":{"minLen":"1"}}

### spec.storage.opensearchServerless.fieldMapping

`AwsBedrockKnowledgeBaseFieldMapping` · required

Index field names Bedrock uses.

- rule: {"required":true}

### spec.storage.opensearchServerless.fieldMapping.vectorField

`string` · required

Field holding the embedding vectors.

- rule: {"string":{"minLen":"1"}}

### spec.storage.opensearchServerless.fieldMapping.textField

`string` · required

Field holding the raw chunk text.

- rule: {"string":{"minLen":"1"}}

### spec.storage.opensearchServerless.fieldMapping.metadataField

`string` · required

Field holding chunk metadata (source attribution).

- rule: {"string":{"minLen":"1"}}

### spec.storage.opensearchManaged

`AwsBedrockKnowledgeBaseOpenSearchManagedStorage`

Self-managed Amazon OpenSearch domain.

### spec.storage.opensearchManaged.domainArn

`string | valueFrom` · required

The OpenSearch domain.

- references: AwsOpenSearchDomain (`status.outputs.domain_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsOpenSearchDomain, name: <that resource's name>, fieldPath: status.outputs.domain_arn}} -- a bare string does not parse

### spec.storage.opensearchManaged.domainEndpoint

`string`

The domain's HTTPS endpoint.

- rule: {"string":{"pattern":"^https://.*$"}}

### spec.storage.opensearchManaged.vectorIndexName

`string` · required

Name of the pre-created vector index in the domain.

- rule: {"string":{"minLen":"1"}}

### spec.storage.opensearchManaged.fieldMapping

`AwsBedrockKnowledgeBaseFieldMapping` · required

Index field names Bedrock uses.

- rule: {"required":true}

### spec.storage.opensearchManaged.fieldMapping.vectorField

`string` · required

Field holding the embedding vectors.

- rule: {"string":{"minLen":"1"}}

### spec.storage.opensearchManaged.fieldMapping.textField

`string` · required

Field holding the raw chunk text.

- rule: {"string":{"minLen":"1"}}

### spec.storage.opensearchManaged.fieldMapping.metadataField

`string` · required

Field holding chunk metadata (source attribution).

- rule: {"string":{"minLen":"1"}}

### spec.storage.s3Vectors

`AwsBedrockKnowledgeBaseS3VectorsStorage`

Amazon S3 Vectors - purpose-built, pay-per-use vector storage (the
lowest-cost self-contained option).

- rule: set index_arn alone, or vector_bucket_arn with index_name

### spec.storage.s3Vectors.indexArn

`string`

ARN of an existing S3 vector index.

### spec.storage.s3Vectors.indexName

`string`

Name of the vector index inside `vector_bucket_arn`.

### spec.storage.s3Vectors.vectorBucketArn

`string`

ARN of the S3 vector bucket holding `index_name`.

### spec.storage.rds

`AwsBedrockKnowledgeBaseRdsStorage`

Aurora PostgreSQL with pgvector.

### spec.storage.rds.resourceArn

`string | valueFrom` · required

The Aurora cluster.

- references: AwsRdsCluster (`status.outputs.arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsRdsCluster, name: <that resource's name>, fieldPath: status.outputs.arn}} -- a bare string does not parse

### spec.storage.rds.credentialsSecretArn

`string | valueFrom` · required

Secrets Manager secret with the database credentials.

- references: AwsSecretsManagerSecret (`status.outputs.secret_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecretsManagerSecret, name: <that resource's name>, fieldPath: status.outputs.secret_arn}} -- a bare string does not parse

### spec.storage.rds.databaseName

`string` · required

Database name.

- rule: {"string":{"minLen":"1"}}

### spec.storage.rds.tableName

`string` · required

Table name holding the vector data.

- rule: {"string":{"minLen":"1"}}

### spec.storage.rds.fieldMapping

`AwsBedrockKnowledgeBaseRdsFieldMapping` · required

Column names Bedrock uses.

- rule: {"required":true}

### spec.storage.rds.fieldMapping.vectorField

`string` · required

Column holding the embedding vectors.

- rule: {"string":{"minLen":"1"}}

### spec.storage.rds.fieldMapping.textField

`string` · required

Column holding the raw chunk text.

- rule: {"string":{"minLen":"1"}}

### spec.storage.rds.fieldMapping.metadataField

`string` · required

Column holding chunk metadata.

- rule: {"string":{"minLen":"1"}}

### spec.storage.rds.fieldMapping.primaryKeyField

`string` · required

Primary-key column.

- rule: {"string":{"minLen":"1"}}

### spec.storage.rds.fieldMapping.customMetadataField

`string`

Column for custom document metadata (filterable attributes).

### spec.storage.pinecone

`AwsBedrockKnowledgeBasePineconeStorage`

Pinecone (SaaS vector database).

### spec.storage.pinecone.connectionString

`string` · required

Pinecone index connection string (host URL).

- rule: {"string":{"minLen":"1"}}

### spec.storage.pinecone.credentialsSecretArn

`string | valueFrom` · required

Secrets Manager secret with the Pinecone API key.

- references: AwsSecretsManagerSecret (`status.outputs.secret_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecretsManagerSecret, name: <that resource's name>, fieldPath: status.outputs.secret_arn}} -- a bare string does not parse

### spec.storage.pinecone.namespace

`string`

Pinecone namespace to isolate this knowledge base's vectors.

### spec.storage.pinecone.fieldMapping

`AwsBedrockKnowledgeBaseTextMetadataFieldMapping` · required

Field names Bedrock uses (Pinecone stores vectors natively - only
text and metadata fields are mapped).

- rule: {"required":true}

### spec.storage.pinecone.fieldMapping.textField

`string` · required

Field holding the raw chunk text.

- rule: {"string":{"minLen":"1"}}

### spec.storage.pinecone.fieldMapping.metadataField

`string` · required

Field holding chunk metadata.

- rule: {"string":{"minLen":"1"}}

### spec.storage.mongodbAtlas

`AwsBedrockKnowledgeBaseMongoDbAtlasStorage`

MongoDB Atlas Vector Search.

### spec.storage.mongodbAtlas.endpoint

`string` · required

Atlas cluster endpoint (mongodb+srv host).

- rule: {"string":{"minLen":"1"}}

### spec.storage.mongodbAtlas.databaseName

`string` · required

Database name.

- rule: {"string":{"minLen":"1"}}

### spec.storage.mongodbAtlas.collectionName

`string` · required

Collection name.

- rule: {"string":{"minLen":"1"}}

### spec.storage.mongodbAtlas.vectorIndexName

`string` · required

Atlas Vector Search index name.

- rule: {"string":{"minLen":"1"}}

### spec.storage.mongodbAtlas.textIndexName

`string`

Atlas Search index for hybrid text search.

### spec.storage.mongodbAtlas.endpointServiceName

`string`

PrivateLink endpoint service name when Atlas is reached privately.

### spec.storage.mongodbAtlas.credentialsSecretArn

`string | valueFrom` · required

Secrets Manager secret with the Atlas credentials.

- references: AwsSecretsManagerSecret (`status.outputs.secret_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecretsManagerSecret, name: <that resource's name>, fieldPath: status.outputs.secret_arn}} -- a bare string does not parse

### spec.storage.mongodbAtlas.fieldMapping

`AwsBedrockKnowledgeBaseFieldMapping` · required

Field names Bedrock uses.

- rule: {"required":true}

### spec.storage.mongodbAtlas.fieldMapping.vectorField

`string` · required

Field holding the embedding vectors.

- rule: {"string":{"minLen":"1"}}

### spec.storage.mongodbAtlas.fieldMapping.textField

`string` · required

Field holding the raw chunk text.

- rule: {"string":{"minLen":"1"}}

### spec.storage.mongodbAtlas.fieldMapping.metadataField

`string` · required

Field holding chunk metadata (source attribution).

- rule: {"string":{"minLen":"1"}}

### spec.storage.neptuneAnalytics

`AwsBedrockKnowledgeBaseNeptuneAnalyticsStorage`

Amazon Neptune Analytics graph (GraphRAG).

### spec.storage.neptuneAnalytics.graphArn

`string`

ARN of the Neptune Analytics graph.

- rule: {"string":{"pattern":"^arn:aws[a-z-]*:neptune-graph:[a-z0-9-]+:[0-9]{12}:graph/.+$"}}

### spec.storage.neptuneAnalytics.fieldMapping

`AwsBedrockKnowledgeBaseTextMetadataFieldMapping` · required

Field names Bedrock uses (the graph stores vectors natively).

- rule: {"required":true}

### spec.storage.neptuneAnalytics.fieldMapping.textField

`string` · required

Field holding the raw chunk text.

- rule: {"string":{"minLen":"1"}}

### spec.storage.neptuneAnalytics.fieldMapping.metadataField

`string` · required

Field holding chunk metadata.

- rule: {"string":{"minLen":"1"}}

### spec.storage.redisEnterpriseCloud

`AwsBedrockKnowledgeBaseRedisEnterpriseCloudStorage`

Redis Enterprise Cloud.

### spec.storage.redisEnterpriseCloud.endpoint

`string` · required

Redis Enterprise Cloud database endpoint.

- rule: {"string":{"minLen":"1"}}

### spec.storage.redisEnterpriseCloud.vectorIndexName

`string` · required

Vector index name in the database.

- rule: {"string":{"minLen":"1"}}

### spec.storage.redisEnterpriseCloud.credentialsSecretArn

`string | valueFrom` · required

Secrets Manager secret with the Redis credentials.

- references: AwsSecretsManagerSecret (`status.outputs.secret_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecretsManagerSecret, name: <that resource's name>, fieldPath: status.outputs.secret_arn}} -- a bare string does not parse

### spec.storage.redisEnterpriseCloud.fieldMapping

`AwsBedrockKnowledgeBaseRedisFieldMapping`

Field names Bedrock uses (Redis accepts partial mappings - AWS
derives what is omitted).

### spec.storage.redisEnterpriseCloud.fieldMapping.vectorField

`string`

Field holding the embedding vectors.

### spec.storage.redisEnterpriseCloud.fieldMapping.textField

`string`

Field holding the raw chunk text.

### spec.storage.redisEnterpriseCloud.fieldMapping.metadataField

`string`

Field holding chunk metadata.

### spec.dataSources

`[]AwsBedrockKnowledgeBaseDataSource`

Where the knowledge base ingests documents from. Each entry is one
connector (S3 bucket, website crawl, Confluence, Salesforce,
SharePoint, or an AWS-managed connector) with its own chunking and
parsing configuration.

### spec.dataSources[].name

`string` · required

Data source name (1-100 characters; letters, digits, hyphen,
underscore). The for_each key on both engines and the key in the
`data_source_ids` output map; also the name in AWS. Renaming replaces
the data source.

- rule: {"string":{"minLen":"1","maxLen":"100","pattern":"^([0-9a-zA-Z][_-]?){1,100}$"}}

### spec.dataSources[].description

`string`

What this source contains (1-200 characters when set).

- rule: {"string":{"maxLen":"200"}}

### spec.dataSources[].dataDeletionPolicy

`string`

What happens to already-ingested vectors when this data source is
deleted: RETAIN keeps them queryable; DELETE purges them. Omitted =
AWS default (DELETE).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["RETAIN","DELETE"]}}

### spec.dataSources[].kmsKeyArn

`string | valueFrom`

Customer-managed KMS key for encrypting transient ingestion data.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.dataSources[].s3

`AwsBedrockKnowledgeBaseS3DataSource`

Ingest objects from an S3 bucket.

### spec.dataSources[].s3.bucketArn

`string | valueFrom` · required

The bucket to ingest from.

- references: AwsS3Bucket (`status.outputs.bucket_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_arn}} -- a bare string does not parse

### spec.dataSources[].s3.inclusionPrefix

`string` · required

Only ingest objects under this key prefix (1-300 characters; AWS
accepts at most one prefix).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"1","maxLen":"300"}}

### spec.dataSources[].s3.bucketOwnerAccountId

`string`

Bucket owner's account ID when the bucket lives in another account.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]{12}$"}}

### spec.dataSources[].web

`AwsBedrockKnowledgeBaseWebDataSource`

Crawl websites from seed URLs.

### spec.dataSources[].web.seedUrls

`[]string` · required

Where the crawl starts (1-100 URLs, http/https).

- rule: {"repeated":{"minItems":"1","maxItems":"100","items":{"string":{"pattern":"^https?://[A-Za-z0-9][^\\s]*$"}}}}

### spec.dataSources[].web.scope

`string`

How far the crawl may wander from the seeds: HOST_ONLY stays on each
seed's host; SUBDOMAINS also crawls its subdomains. Omitted = AWS
default (host plus deeper paths of the seed).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["HOST_ONLY","SUBDOMAINS"]}}

### spec.dataSources[].web.inclusionFilters

`[]string`

Only crawl URLs matching these regex patterns (at most 25, each up to
1000 characters).

- rule: {"repeated":{"maxItems":"25","items":{"string":{"minLen":"1","maxLen":"1000"}}}}

### spec.dataSources[].web.exclusionFilters

`[]string`

Never crawl URLs matching these regex patterns (at most 25).

- rule: {"repeated":{"maxItems":"25","items":{"string":{"minLen":"1","maxLen":"1000"}}}}

### spec.dataSources[].web.userAgent

`string` · required

Custom User-Agent suffix for the crawler (15-40 characters).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"15","maxLen":"40"}}

### spec.dataSources[].web.maxPages

`int32`

Stop after crawling this many pages.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.dataSources[].web.rateLimit

`int32`

Crawl at most this many URLs per minute per host (1-300).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":300,"gte":1}}

### spec.dataSources[].confluence

`AwsBedrockKnowledgeBaseConfluenceDataSource`

Ingest Confluence spaces and pages.

### spec.dataSources[].confluence.hostUrl

`string`

The Confluence Cloud instance URL (https://<company>.atlassian.net).

- rule: {"string":{"pattern":"^https://[A-Za-z0-9][^\\s]*$"}}

### spec.dataSources[].confluence.authType

`string`

How Bedrock authenticates: BASIC (username + API token) or
OAUTH2_CLIENT_CREDENTIALS.

- rule: {"string":{"in":["BASIC","OAUTH2_CLIENT_CREDENTIALS"]}}

### spec.dataSources[].confluence.credentialsSecretArn

`string | valueFrom` · required

Secrets Manager secret with the Confluence credentials (shape depends
on auth_type - see the AWS connector documentation).

- references: AwsSecretsManagerSecret (`status.outputs.secret_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecretsManagerSecret, name: <that resource's name>, fieldPath: status.outputs.secret_arn}} -- a bare string does not parse

### spec.dataSources[].confluence.filters

`[]AwsBedrockKnowledgeBaseCrawlFilter`

Narrow what gets ingested, per content type.

### spec.dataSources[].confluence.filters[].objectType

`string` · required

The connector-specific content type this filter applies to (e.g.
Confluence "Space"/"Page", Salesforce "Account", SharePoint "Page").

- rule: {"string":{"minLen":"1"}}

### spec.dataSources[].confluence.filters[].inclusionFilters

`[]string`

Only ingest objects matching these regex patterns (at most 25).

- rule: {"repeated":{"maxItems":"25","items":{"string":{"minLen":"1","maxLen":"1000"}}}}

### spec.dataSources[].confluence.filters[].exclusionFilters

`[]string`

Never ingest objects matching these regex patterns (at most 25).

- rule: {"repeated":{"maxItems":"25","items":{"string":{"minLen":"1","maxLen":"1000"}}}}

### spec.dataSources[].salesforce

`AwsBedrockKnowledgeBaseSalesforceDataSource`

Ingest Salesforce objects.

### spec.dataSources[].salesforce.hostUrl

`string` · required

The Salesforce instance URL (1-256 characters).

- rule: {"string":{"minLen":"1","maxLen":"256","pattern":"^https://[A-Za-z0-9][^\\s]*$"}}

### spec.dataSources[].salesforce.credentialsSecretArn

`string | valueFrom` · required

Secrets Manager secret with the Salesforce connected-app credentials.

- references: AwsSecretsManagerSecret (`status.outputs.secret_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecretsManagerSecret, name: <that resource's name>, fieldPath: status.outputs.secret_arn}} -- a bare string does not parse

### spec.dataSources[].salesforce.filters

`[]AwsBedrockKnowledgeBaseCrawlFilter`

Narrow what gets ingested, per object type.

### spec.dataSources[].salesforce.filters[].objectType

`string` · required

The connector-specific content type this filter applies to (e.g.
Confluence "Space"/"Page", Salesforce "Account", SharePoint "Page").

- rule: {"string":{"minLen":"1"}}

### spec.dataSources[].salesforce.filters[].inclusionFilters

`[]string`

Only ingest objects matching these regex patterns (at most 25).

- rule: {"repeated":{"maxItems":"25","items":{"string":{"minLen":"1","maxLen":"1000"}}}}

### spec.dataSources[].salesforce.filters[].exclusionFilters

`[]string`

Never ingest objects matching these regex patterns (at most 25).

- rule: {"repeated":{"maxItems":"25","items":{"string":{"minLen":"1","maxLen":"1000"}}}}

### spec.dataSources[].sharepoint

`AwsBedrockKnowledgeBaseSharePointDataSource`

Ingest SharePoint sites.

### spec.dataSources[].sharepoint.siteUrls

`[]string` · required

Site URLs to ingest (1-100, https).

- rule: {"repeated":{"minItems":"1","maxItems":"100","items":{"string":{"pattern":"^https://[A-Za-z0-9][^\\s]*$"}}}}

### spec.dataSources[].sharepoint.domain

`string` · required

The SharePoint domain (1-50 characters).

- rule: {"string":{"minLen":"1","maxLen":"50"}}

### spec.dataSources[].sharepoint.tenantId

`string`

Azure AD tenant ID (UUID).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"}}

### spec.dataSources[].sharepoint.authType

`string`

How Bedrock authenticates: OAUTH2_CLIENT_CREDENTIALS or
OAUTH2_SHAREPOINT_APP_ONLY_CLIENT_CREDENTIALS.

- rule: {"string":{"in":["OAUTH2_CLIENT_CREDENTIALS","OAUTH2_SHAREPOINT_APP_ONLY_CLIENT_CREDENTIALS"]}}

### spec.dataSources[].sharepoint.credentialsSecretArn

`string | valueFrom` · required

Secrets Manager secret with the Azure app credentials.

- references: AwsSecretsManagerSecret (`status.outputs.secret_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecretsManagerSecret, name: <that resource's name>, fieldPath: status.outputs.secret_arn}} -- a bare string does not parse

### spec.dataSources[].sharepoint.filters

`[]AwsBedrockKnowledgeBaseCrawlFilter`

Narrow what gets ingested, per content type.

### spec.dataSources[].sharepoint.filters[].objectType

`string` · required

The connector-specific content type this filter applies to (e.g.
Confluence "Space"/"Page", Salesforce "Account", SharePoint "Page").

- rule: {"string":{"minLen":"1"}}

### spec.dataSources[].sharepoint.filters[].inclusionFilters

`[]string`

Only ingest objects matching these regex patterns (at most 25).

- rule: {"repeated":{"maxItems":"25","items":{"string":{"minLen":"1","maxLen":"1000"}}}}

### spec.dataSources[].sharepoint.filters[].exclusionFilters

`[]string`

Never ingest objects matching these regex patterns (at most 25).

- rule: {"repeated":{"maxItems":"25","items":{"string":{"minLen":"1","maxLen":"1000"}}}}

### spec.dataSources[].managedConnector

`AwsBedrockKnowledgeBaseManagedConnectorDataSource`

AWS-managed knowledge-base connector (connector-specific parameters
as a JSON document).

### spec.dataSources[].managedConnector.connectorParameters

`object`

Connector-specific parameters as a JSON document (shape defined by
the connector - see the AWS connector catalog).

### spec.dataSources[].managedConnector.deletionProtection

`AwsBedrockKnowledgeBaseDeletionProtection`

Refuse a sync that would delete more than `threshold` percent of the
already-ingested documents (guards against accidental mass deletion
upstream).

### spec.dataSources[].managedConnector.deletionProtection.enabled

`bool`

Turn the guard on or off.

### spec.dataSources[].managedConnector.deletionProtection.thresholdPercent

`int32`

Abort a sync that would delete more than this percent of documents
(0-100).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":100,"gte":0}}

### spec.dataSources[].managedConnector.mediaExtraction

`AwsBedrockKnowledgeBaseMediaExtraction`

Extract text from non-text media during ingestion.

### spec.dataSources[].managedConnector.mediaExtraction.audio

`bool`

Extract text from audio files.

### spec.dataSources[].managedConnector.mediaExtraction.image

`bool`

Extract text from images.

### spec.dataSources[].managedConnector.mediaExtraction.video

`bool`

Extract text from video files.

### spec.dataSources[].vectorIngestion

`AwsBedrockKnowledgeBaseVectorIngestion`

How documents are split into chunks, parsed, and transformed before
embedding. Create-time only - changing it replaces the data source.

### spec.dataSources[].vectorIngestion.chunking

`AwsBedrockKnowledgeBaseChunking`

How documents split into chunks before embedding.

- rule: configure exactly the block matching strategy (FIXED_SIZE -> fixed_size, HIERARCHICAL -> hierarchical, SEMANTIC -> semantic, NONE -> none)

### spec.dataSources[].vectorIngestion.chunking.strategy

`string`

FIXED_SIZE (token windows with overlap), HIERARCHICAL (parent/child
levels), SEMANTIC (split at meaning boundaries), or NONE (one chunk
per document). Omitted at the data-source level = AWS default
(FIXED_SIZE, 300 tokens).

- rule: {"string":{"in":["FIXED_SIZE","HIERARCHICAL","SEMANTIC","NONE"]}}

### spec.dataSources[].vectorIngestion.chunking.fixedSize

`AwsBedrockKnowledgeBaseFixedSizeChunking`

FIXED_SIZE settings.

### spec.dataSources[].vectorIngestion.chunking.fixedSize.maxTokens

`int32`

Maximum tokens per chunk.

- rule: {"int32":{"gte":1}}

### spec.dataSources[].vectorIngestion.chunking.fixedSize.overlapPercentage

`int32`

Percent of each chunk shared with the next (1-99).

- rule: {"int32":{"lte":99,"gte":1}}

### spec.dataSources[].vectorIngestion.chunking.hierarchical

`AwsBedrockKnowledgeBaseHierarchicalChunking`

HIERARCHICAL settings.

### spec.dataSources[].vectorIngestion.chunking.hierarchical.overlapTokens

`int32`

Tokens shared between adjacent child chunks.

- rule: {"int32":{"gte":1}}

### spec.dataSources[].vectorIngestion.chunking.hierarchical.levels

`[]AwsBedrockKnowledgeBaseChunkingLevel` · required

Exactly two levels: the parent level first, the child level second.

- rule: {"repeated":{"minItems":"2","maxItems":"2"}}

### spec.dataSources[].vectorIngestion.chunking.hierarchical.levels[].maxTokens

`int32`

Maximum tokens per chunk at this level (1-8192).

- rule: {"int32":{"lte":8192,"gte":1}}

### spec.dataSources[].vectorIngestion.chunking.semantic

`AwsBedrockKnowledgeBaseSemanticChunking`

SEMANTIC settings.

### spec.dataSources[].vectorIngestion.chunking.semantic.breakpointPercentileThreshold

`int32`

Split where sentence similarity drops below this percentile (50-99;
higher = fewer, larger chunks).

- rule: {"int32":{"lte":99,"gte":50}}

### spec.dataSources[].vectorIngestion.chunking.semantic.bufferSize

`int32`

Sentences to group when computing similarity (0 or 1).

- rule: {"int32":{"lte":1,"gte":0}}

### spec.dataSources[].vectorIngestion.chunking.semantic.maxTokens

`int32`

Maximum tokens per chunk.

- rule: {"int32":{"gte":1}}

### spec.dataSources[].vectorIngestion.parsing

`AwsBedrockKnowledgeBaseParsing`

How documents are parsed into text.

- rule: foundation_model is required with BEDROCK_FOUNDATION_MODEL and forbidden otherwise
- rule: multimodal requires strategy BEDROCK_DATA_AUTOMATION

### spec.dataSources[].vectorIngestion.parsing.strategy

`string`

BEDROCK_DATA_AUTOMATION, BEDROCK_FOUNDATION_MODEL, or SMART_PARSING.

- rule: {"string":{"in":["BEDROCK_DATA_AUTOMATION","BEDROCK_FOUNDATION_MODEL","SMART_PARSING"]}}

### spec.dataSources[].vectorIngestion.parsing.multimodal

`bool`

Parse multimodal content (images embedded in documents) with Bedrock
Data Automation. Only meaningful with BEDROCK_DATA_AUTOMATION.

### spec.dataSources[].vectorIngestion.parsing.foundationModel

`AwsBedrockKnowledgeBaseFoundationModelParsing`

BEDROCK_FOUNDATION_MODEL settings (required with that strategy).

### spec.dataSources[].vectorIngestion.parsing.foundationModel.modelArn

`string` · required

ARN of the parsing foundation model.

- rule: {"string":{"minLen":"1"}}

### spec.dataSources[].vectorIngestion.parsing.foundationModel.multimodal

`bool`

Parse multimodal content (images embedded in documents).

### spec.dataSources[].vectorIngestion.parsing.foundationModel.parsingPrompt

`string`

Custom parsing prompt. Omitted = AWS's built-in parsing prompt.

### spec.dataSources[].vectorIngestion.customTransformation

`AwsBedrockKnowledgeBaseCustomTransformation`

Run a Lambda between pipeline steps to reshape chunks.

### spec.dataSources[].vectorIngestion.customTransformation.intermediateS3Uri

`string`

S3 location (s3://bucket/prefix) AWS uses to hand chunks to the
Lambda and read its output.

- rule: {"string":{"pattern":"^s3://[a-z0-9.-]+(/.*)?$"}}

### spec.dataSources[].vectorIngestion.customTransformation.lambdaArn

`string | valueFrom` · required

The transformation Lambda (runs after chunking - POST_CHUNKING, the
only step AWS defines; the modules send the constant).

- references: AwsLambda (`status.outputs.function_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

## Validation Rules

- `storage_iff_vector`: storage is required with the vector type and forbidden with managed/kendra/sql (AWS manages or delegates their storage)
- `data_source_names_unique`: data_sources entries must have unique names

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBedrockKnowledgeBase, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.knowledge_base_id` | `string` | The unique knowledge base identifier (e.g. "EMDPPAYPZI") - the join key agents and flows use. |
| `status.outputs.knowledge_base_arn` | `string` | The Amazon Resource Name of the knowledge base - the canonical key for IAM policies. |
| `status.outputs.data_source_ids` | `map<string, string>` | Data source IDs keyed by each `data_sources` entry's name. Example: {"docs": "GWCMFMQF6T"}. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.managed.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.sql.provisioned.clusterIdentifier` | AwsRedshiftCluster | `status.outputs.cluster_identifier` |
| `spec.sql.provisioned.auth.usernamePasswordSecretArn` | AwsSecretsManagerSecret | `status.outputs.secret_arn` |
| `spec.sql.serverless.workgroupArn` | AwsRedshiftServerlessWorkgroup | `status.outputs.arn` |
| `spec.sql.serverless.auth.usernamePasswordSecretArn` | AwsSecretsManagerSecret | `status.outputs.secret_arn` |
| `spec.storage.opensearchServerless.collectionArn` | AwsOpenSearchServerlessCollection | `status.outputs.collection_arn` |
| `spec.storage.opensearchManaged.domainArn` | AwsOpenSearchDomain | `status.outputs.domain_arn` |
| `spec.storage.rds.resourceArn` | AwsRdsCluster | `status.outputs.arn` |
| `spec.storage.rds.credentialsSecretArn` | AwsSecretsManagerSecret | `status.outputs.secret_arn` |
| `spec.storage.pinecone.credentialsSecretArn` | AwsSecretsManagerSecret | `status.outputs.secret_arn` |
| `spec.storage.mongodbAtlas.credentialsSecretArn` | AwsSecretsManagerSecret | `status.outputs.secret_arn` |
| `spec.storage.redisEnterpriseCloud.credentialsSecretArn` | AwsSecretsManagerSecret | `status.outputs.secret_arn` |
| `spec.dataSources[].kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.dataSources[].s3.bucketArn` | AwsS3Bucket | `status.outputs.bucket_arn` |
| `spec.dataSources[].confluence.credentialsSecretArn` | AwsSecretsManagerSecret | `status.outputs.secret_arn` |
| `spec.dataSources[].salesforce.credentialsSecretArn` | AwsSecretsManagerSecret | `status.outputs.secret_arn` |
| `spec.dataSources[].sharepoint.credentialsSecretArn` | AwsSecretsManagerSecret | `status.outputs.secret_arn` |
| `spec.dataSources[].vectorIngestion.customTransformation.lambdaArn` | AwsLambda | `status.outputs.function_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBedrockAgent | `spec.knowledgeBaseAssociations[].knowledgeBaseId` | `status.outputs.knowledge_base_id` |
| AwsBedrockFlow | `spec.definition.nodes[].knowledgeBase.knowledgeBaseId` | `status.outputs.knowledge_base_id` |

## See Also

- [Overview](../README.md)
