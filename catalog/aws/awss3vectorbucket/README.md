# AwsS3VectorBucket

One S3 vector bucket with its vector indexes — purpose-built storage for AI embeddings, queried by similarity instead of key. The natural backend for Bedrock knowledge bases and any RAG stack that wants pay-per-query vector search without running a vector database.

## Highlights

- **The Bedrock KB catalog kind**: AwsBedrockKnowledgeBase's `s3_vectors` arm points at an index defined here — bucket + index + knowledge base is a three-resource chart.
- **Everything on an index is fixed for life**: dimension, distance metric, non-filterable keys — an index is replaced, not edited, so the spec teaches sizing dimension to the embedding model (Titan v2: 1024) BEFORE the first vector lands.
- **The metadata budget made visible**: bulky payload fields (source text, URIs) belong in `non_filterable_metadata_keys` — filterable metadata has a hard per-vector size budget, taught on the field.
- **Secure and honest by default**: `data_type` is module-pinned to float32 (the provider's single-value enum, recorded exclusion); `force_destroy` is config-only and declared so in the import catalog.

## Both Engines

Both modules key indexes by name and export the same outputs: `vector_bucket_arn` (import ID), `index_arns` (keyed by index name).

## Chart Wiring

`kms_key_arn` references an AwsKmsKey. The `index_arns` map is what a Bedrock knowledge base's s3_vectors arm consumes; the `vector_bucket_arn` is what policies and cross-account grants reference.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
