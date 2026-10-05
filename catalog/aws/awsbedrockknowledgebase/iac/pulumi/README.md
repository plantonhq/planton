# Pulumi Module: AWS Bedrock Knowledge Base

Provisions an Amazon Bedrock knowledge base and its folded data sources
using Pulumi (Go).

## Resources Created

- `bedrock.AgentKnowledgeBase` — The knowledge base: exactly one type arm
  (vector / managed / kendra / sql) with AWS's `type` discriminator
  derived from the configured arm, plus the vector store backend for the
  vector type (eight backends, same derivation).
- `bedrock.AgentDataSource` — One per `spec.data_sources` entry (resource
  name `data-source-<entry name>`): the connector arm (S3, web crawl,
  Confluence, Salesforce, SharePoint, managed connector) plus the
  chunk-parse-transform ingestion pipeline.

## Notable Behavior

- Behavioral parity with the Terraform module is the contract: identical
  send conditions, identical derived discriminators, identical constants
  (REDSHIFT SQL engine, PATTERN crawl filters, SAAS/ONLINE host types,
  OAUTH2_CLIENT_CREDENTIALS Salesforce auth, MULTIMODAL parsing modality,
  POST_CHUNKING transformation step, S3 supplemental storage), identical
  outputs.
- Nearly the whole surface is create-time only upstream — plan/preview
  shows replacements for type/storage/ingestion changes; AWS re-ingests
  afterwards.
- Data source entries iterate name-sorted for deterministic previews.

## Usage

The module is executed by the Planton platform through the entrypoint in
`main.go`, which loads the `AwsBedrockKnowledgeBaseIacInput`.
