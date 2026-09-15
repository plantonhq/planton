# AWS Bedrock Knowledge Base

Deploys an Amazon Bedrock knowledge base — the RAG store that ingests documents from data sources, embeds them, and answers retrieval queries for agents, flows, and applications. Exactly one knowledge-base type is configured: vector (embeddings over a store you own, across eight backends), managed (AWS runs the vector store), kendra (retrieval delegated to an existing Kendra index), or sql (natural-language-to-SQL over Redshift, with no embeddings at all). Data sources fold into the same component — S3, web crawler, Confluence, Salesforce, SharePoint, or AWS-managed connectors, each with its own chunking, parsing, and transformation pipeline. The cost drivers are embedding tokens at ingestion, retrieval at query time, and the store behind the managed type.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Bedrock Knowledge Base** — of the type you declare, with the modules deriving AWS's type discriminator from which arm (`vector`, `managed`, `kendra`, `sql`) is set; the `vector` type additionally binds the `storage` backend you point it at
- **Data Sources** — created only when `dataSources` entries exist: one connector per entry, keyed by your stable entry names, each carrying its chunking strategy, parsing configuration, and optional Lambda transformation

Deploying creates the knowledge base and its data sources; document ingestion (`StartIngestionJob` syncs) is a separate step run from the console, CLI, or your pipeline.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Bedrock knowledge-base permissions (`bedrock:CreateKnowledgeBase`, `bedrock:CreateDataSource`, and their read/update/delete siblings, plus `iam:PassRole` on the role). Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### AWS Account

- **Bedrock knowledge bases available in the target region.**
- **An IAM role trusting `bedrock.amazonaws.com`** — referenced by `roleArn`, with data-source read access, embedding-model invoke, and data-plane access to the vector store. AWS validates all of it at CreateKnowledgeBase (with a propagation retry window) — a missing permission is a deploy failure, not a query failure.
- **The vector store pre-created** — only for the `vector` type: an S3 Vectors index, an OpenSearch Serverless collection WITH a vector index already inside it (AWS does not create the index), an Aurora pgvector table, or the SaaS backend of your choice.
- **Credentials in Secrets Manager** — only for SaaS stores (Pinecone, MongoDB Atlas, Redis) and SaaS connectors (Confluence, Salesforce, SharePoint); the spec references secrets, never raw keys.

## Deploy

### Console

Open the deployment store, find **AWS Bedrock Knowledge Base**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: region and role, the knowledge-base type and store, and the data sources. Start from the **S3 Docs on S3 Vectors** preset in the [Presets](#presets) tab for the lowest-operations RAG shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockKnowledgeBase
metadata:
  name: product-docs
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: bedrock-kb-role
      fieldPath: status.outputs.role_arn
  vector:
    embeddingModelArn: arn:aws:bedrock:us-west-2::foundation-model/amazon.titan-embed-text-v2:0
    embeddingModel:
      dimensions: 256
  storage:
    s3Vectors:
      indexArn: arn:aws:s3vectors:us-west-2:123456789012:bucket/docs-vectors/index/docs
  dataSources:
    - name: manuals
      dataDeletionPolicy: DELETE
      s3:
        bucketArn:
          value: arn:aws:s3:::product-manuals
      vectorIngestion:
        chunking:
          strategy: FIXED_SIZE
          fixedSize:
            maxTokens: 300
            overlapPercentage: 20
```

```shell
planton apply -f knowledge-base.yaml
```

This creates a vector knowledge base on S3 Vectors with Titan V2 embeddings at 256 dimensions and one S3 data source chunked into 300-token windows — ready for its first ingestion sync. A Stack Job tracks the provisioning in real time.

### InfraChart

When the knowledge base deploys alongside its role and source bucket in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: bedrock-kb-role
      fieldPath: status.outputs.role_arn
  vector:
    embeddingModelArn: arn:aws:bedrock:us-west-2::foundation-model/amazon.titan-embed-text-v2:0
    embeddingModel:
      dimensions: 256
  storage:
    s3Vectors:
      indexArn: arn:aws:s3vectors:us-west-2:123456789012:bucket/docs-vectors/index/docs
  dataSources:
    - name: manuals
      s3:
        bucketArn:
          valueFrom:
            kind: AwsS3Bucket
            name: product-manuals
            fieldPath: status.outputs.bucket_arn
```

The InfraPipeline resolves the dependency graph, deploys the role and bucket first, then creates the knowledge base over them.

## Key Configuration

These are the most important decisions when configuring a knowledge base. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Pick the type first** — `vector` is the classic RAG shape and the only one that takes a `storage` block; `managed` hands the store to AWS (and bills for what AWS provisions behind it); `kendra` delegates retrieval to an index you already operate; `sql` answers questions by generating Redshift queries — no embeddings, no ingestion. Exactly one arm, and the choice is create-time: changing it replaces the knowledge base.

**Pick the store by operational appetite** — S3 Vectors is pay-per-use with no infrastructure to run (the cheapest start); OpenSearch Serverless gives sub-second recall at OCU cost but you create and manage the vector index yourself — AWS will not create it; Aurora pgvector suits teams already operating Postgres; the SaaS backends (Pinecone, MongoDB Atlas, Redis) carry their credentials as Secrets Manager references.

**Dimensions are a contract** — the embedding model's `dimensions` (Titan V2: 256, 512, or 1024) must equal the vector index's dimension. A mismatch fails at the first ingestion sync, not at create — the most delayed failure in this component.

**Almost everything replaces** — only the description, the role, and data-source membership update in place; type, storage, embedding, and ingestion-pipeline changes replace the knowledge base, and AWS re-ingests every source afterwards. Budget sync time for any such change.

**Ingestion is a separate step** — the deploy creates the structure; documents arrive via `StartIngestionJob` syncs per data source. Wire the sync into your content pipeline, and watch the job counts (scanned/indexed/failed) rather than assuming apply success means searchable documents.

**Chunking is the retrieval-quality lever** — FIXED_SIZE (the AWS default at 300 tokens) is the safe start; HIERARCHICAL retrieves small child chunks but hands the model their larger parent for context, which helps long structured documents; SEMANTIC splits at meaning boundaries at extra embedding cost; NONE treats each document as one chunk — right only for already-small documents. Chunking is create-time per data source.

**`dataDeletionPolicy` decides what teardown leaves behind** — DELETE (the AWS default) purges ingested vectors when the data source is removed, keeping destroys clean; RETAIN leaves them queryable, which orphans store contents past the manifest's knowledge.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `roleArn` | `status.outputs.role_arn` |
| **AwsS3Bucket** | `dataSources[].s3.bucketArn` | `status.outputs.bucket_arn` |
| **AwsOpenSearchServerlessCollection** | `storage.opensearchServerless.collectionArn` | `status.outputs.collection_arn` |
| **AwsOpenSearchDomain** | `storage.opensearchManaged.domainArn` | `status.outputs.domain_arn` |
| **AwsRdsCluster** | `storage.rds.resourceArn` | `status.outputs.arn` |
| **AwsS3VectorBucket** | `storage.s3Vectors.vectorBucketArn` | `status.outputs.vector_bucket_arn` |
| **AwsSecretsManagerSecret** | SaaS store and connector `credentialsSecretArn` fields | `status.outputs.secret_arn` |
| **AwsRedshiftCluster** | `sql.provisioned.clusterIdentifier` | `status.outputs.cluster_identifier` |
| **AwsRedshiftServerlessWorkgroup** | `sql.serverless.workgroupArn` | `status.outputs.arn` |
| **AwsKmsKey** | `managed.kmsKeyArn`, `dataSources[].kmsKeyArn` | `status.outputs.key_arn` |
| **AwsLambda** | `dataSources[].vectorIngestion.customTransformation.lambdaArn` | `status.outputs.function_arn` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `knowledge_base_id` | The unique knowledge base identifier | An AwsBedrockAgent's `knowledgeBaseAssociations[].knowledgeBaseId`; a flow's knowledge-base node |
| `knowledge_base_arn` | The knowledge base's ARN | IAM policies scoping retrieval and management access |
| `data_source_ids` | Data source IDs keyed by each `dataSources` entry's name | Ingestion pipelines targeting `StartIngestionJob` at a specific source |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**S3 documents on S3 Vectors** — the lowest-operations RAG store: pay-per-use vectors, no index infrastructure, an S3 bucket as the document source with DELETE deletion policy for clean teardowns. The right first knowledge base for most teams. Start from the **S3 Docs on S3 Vectors** preset.

**Managed store with a web crawl** — AWS runs the vector store while a web-crawler data source ingests your public site, scoped by host and rate-limited. Trades store-level control for zero store operations — you never see dimensions or indexes. Start from the **Managed Store with Web Crawl** preset.

**Retrieval for an agent** — the knowledge base as an agent's grounding source: associate it through the agent's `knowledgeBaseAssociations`, invest in the association description (the model reads it to decide when to retrieve), and let the chart's reference wiring order knowledge base before agent.

## Works With

- [**AWS Bedrock Agent**](/cloud-catalog/aws-bedrock-agent) — queries this knowledge base through its `knowledgeBaseAssociations`
- [**AWS Bedrock Flow**](/cloud-catalog/aws-bedrock-flow) — knowledge-base nodes retrieve (and optionally generate) from it
- [**AWS IAM Role**](/cloud-catalog/aws-iam-role) — the operating role, wired via `roleArn`
- [**AWS S3 Bucket**](/cloud-catalog/aws-s3-bucket) — document sources for S3 data sources
- [**AWS OpenSearch Serverless Collection**](/cloud-catalog/aws-open-search-serverless-collection) — the sub-second-recall vector store option
- [**AWS RDS Cluster**](/cloud-catalog/aws-rds-cluster) — the Aurora pgvector store option
- [**AWS Secrets Manager Secret**](/cloud-catalog/aws-secrets-manager-secret) — credentials for SaaS stores and connectors
- [**AWS Lambda**](/cloud-catalog/aws-lambda) — custom chunk transformation between ingestion steps
