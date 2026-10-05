# AWS S3 Vector Bucket

Deploys an S3 vector bucket with its similarity-search indexes — purpose-built storage for AI embeddings, queried by similarity instead of key. It is the natural backend for Bedrock knowledge bases (a knowledge base's s3_vectors arm points at an index defined here) and for any RAG stack that wants vector search without operating a vector database. Every index property is fixed for life — an index is replaced, not edited — so the dimension must match the embedding model before the first vector lands.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **S3 Vector Bucket** — named after the resource, with the declared encryption at rest (S3-managed keys when `encryption` is unset) and force-destroy posture
- **Vector Bucket Policy** — created only when `policy` is set; who can put and query vectors, including cross-account Bedrock knowledge bases
- **Vector Indexes** — one per `indexes[]` entry: dimension, distance metric, non-filterable metadata keys, and optional per-index encryption (the vector data type is module-pinned to float32 — the provider's enum holds exactly that one value)
- **AWS Tags** — resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with credentials for the target AWS account, including S3 Vectors permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **Your embedding model's output dimension** — not an AWS resource but the one fact you must have first: the index `dimension` must equal it exactly and can never change.
- **A KMS key** (only for `aws:kms` encryption) — reference an AwsKmsKey Infra Component or pass a literal key ARN. Encryption is fixed for life of the bucket.

## Deploy

### Console

Open the deployment store, find **AWS S3 Vector Bucket**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Bedrock Knowledge Base Backend** preset in the [Presets](#presets) tab for the RAG starter: one index sized for Titan Text v2 with the bulky payload fields kept non-filterable.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsS3VectorBucket
metadata:
  name: kb-vectors
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  indexes:
    - name: kb-embeddings
      dimension: 1024
      distanceMetric: cosine
      nonFilterableMetadataKeys:
        - source_text
        - source_uri
```

```shell
planton apply -f vector-bucket.yaml
```

This creates a vector bucket with one 1024-dimension cosine index whose chunk text and source URI ride along as non-filterable metadata. An Infra Job tracks the provisioning in real time.

### InfraChart

When the bucket deploys alongside its encryption key in one chart, wire the key reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  encryption:
    sseType: aws:kms
    kmsKeyArn:
      valueFrom:
        kind: AwsKmsKey
        name: embeddings-key
        fieldPath: status.outputs.key_arn
  indexes:
    - name: kb-embeddings
      dimension: 1024
      distanceMetric: cosine
```

The InfraPipeline resolves the dependency graph, provisions the key first, then creates the encrypted vector bucket.

## Key Configuration

These are the most important decisions when configuring a vector bucket. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Dimension is decided by the model, forever** — an index's `dimension` must equal the embedding model's output exactly (Titan Text v2: 1024/512/256; Cohere: 1024). Mismatched puts are rejected vector-by-vector, and the dimension never changes: model migration means a new index plus a re-embed. Budget for that when choosing models — and different dimensions are ALWAYS different indexes.

**The distance metric is part of the model choice** — `cosine` for normalized text embeddings (the common case), `euclidean` when the model was trained for it. A wrong metric returns plausible-looking but degraded results — the worst failure mode, because nothing errors. Match the model card, not intuition.

**Non-filterable keys are the cost and latency lever** — every filterable metadata byte rides the index's query path and counts against a hard per-vector budget. Bulky payloads (source text chunks, URIs) belong in `nonFilterableMetadataKeys` (up to 10). Getting this wrong doesn't error either — it makes queries slower and oversized puts fail.

**Encryption is fixed for life** — the bucket's encryption (and any per-index override) is chosen once. Rotating to a different key means a new bucket and a re-embed or copy, so pick the key posture before the first vector, not after the millionth.

**Deleting embeddings deletes real money** — embeddings cost compute to produce. `forceDestroy: false` (the default) makes a non-empty bucket refuse teardown; keep it that way outside scratch environments, and treat re-embedding cost as part of any bucket-replacing change.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** (optional) | `encryption.kmsKeyArn` | `status.outputs.key_arn` |
| **AwsKmsKey** (optional, per index) | `indexes[].encryption.kmsKeyArn` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `vector_bucket_arn` | The vector bucket's ARN | Resource policies and cross-account access statements |
| `index_arns` | Each index's ARN, keyed by index name | The Bedrock knowledge base's s3_vectors arm — the stable reference charts should pass instead of a hardcoded ARN |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Bedrock knowledge base backend** — one index sized for the chosen embedding model, with chunk text and source URI kept non-filterable, consumed by an AwsBedrockKnowledgeBase through the `index_arns` output. This is the pay-per-query RAG shape: no vector database cluster to run, at the cost of AWS-defined query semantics. Start from the **Bedrock Knowledge Base Backend** preset.

**One encrypted bucket, several models** — a KMS-encrypted bucket holding one index per embedding model (full-size document embeddings and compact code embeddings, say), because different dimensions are always different indexes. The key is chosen once, for life of the bucket. Start from the **Encrypted Multi-Index Store** preset.

## Works With

- [**AWS Bedrock Knowledge Base**](/infra-catalog/aws-bedrock-knowledge-base) — the main consumer: its s3_vectors arm points at an index here via the `index_arns` output
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — the customer-managed key for at-rest encryption, wired via `encryption.kmsKeyArn` at the bucket or per index
