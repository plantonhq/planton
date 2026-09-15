# S3 Docs on S3 Vectors

This preset creates a vector knowledge base over Amazon S3 Vectors — the
pay-per-use vector store with no infrastructure to run — ingesting a
documentation bucket's `manuals/` prefix with fixed-size chunking and
Titan Text Embeddings V2 at 256 dimensions.

## When to Use

- The lowest-cost, lowest-friction RAG starting point: no OpenSearch
  collection, no OCU floor, no index servers
- Documentation, runbooks, or policy corpora that live in S3

## What You Get

- A VECTOR-type knowledge base whose embedding dimensions match the S3
  Vectors index shape (create the index with dimension 256 to pair)
- One S3 data source with DELETE deletion policy — removing the source
  purges its vectors, keeping teardown clean

## Customize

- Point `indexArn` at your S3 Vectors index, or address a
  Planton-managed bucket with `vectorBucketArn` (a reference to an
  `AwsS3VectorBucket`, or a literal ARN under `value:`) + `indexName`
- Raise `dimensions` to 512/1024 for higher recall (recreate the index
  to match)
- Switch chunking to `HIERARCHICAL` (parent/child levels) for long,
  structured documents
