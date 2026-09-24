# A Bedrock knowledge base names the S3 vector bucket that holds its vectors

## What changed

- **`AwsBedrockKnowledgeBaseS3VectorsStorage.vector_bucket_arn` is a typed reference to `AwsS3VectorBucket`** (default wiring: the bucket's `vector_bucket_arn` output), returning at a new field number with the old one reserved. Every other vector-store arm of the knowledge base -- the OpenSearch Serverless collection, the OpenSearch domain, the Aurora cluster -- already named its store by reference; S3 Vectors, the store the bucket kind's own comment calls "the natural backend for Bedrock knowledge bases", was the one arm that took a plain string.
- **`index_arn` stays a literal**, and its comment says why: a bucket's index ARNs are a map output keyed by index name, so no single field path addresses one. Address a Planton-managed bucket through `vector_bucket_arn` + `index_name`.
- The `s3_vectors_addressing` rule reads presence of the reference (`has()`) instead of comparing a string; the Pulumi module reads the literal through the reference; the preset's explainer and the catalog page's Consumes table name the bucket.

## Why

A reference is how the catalog says one resource depends on another: the console offers a picker, the set lane deploys the bucket first, and the diagram draws the line. A knowledge base over S3 Vectors had none of the three while every sibling store had all of them.

## How to check

```bash
go test ./catalog/aws/awsbedrockknowledgebase/v1alpha1/       # the addressing rule on both shapes, with the reference
go build ./catalog/aws/awsbedrockknowledgebase/iac/pulumi/module/
grep -n 'reserved 3;' catalog/aws/awsbedrockknowledgebase/v1alpha1/spec.proto
```
