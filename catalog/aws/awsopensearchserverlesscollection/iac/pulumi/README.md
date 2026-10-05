# Pulumi Module: AWS OpenSearch Serverless Collection

Provisions an Amazon OpenSearch Serverless collection using Pulumi (Go),
together with the collection-scoped policies that make it usable.

## Resources Created

- `opensearch.ServerlessSecurityPolicy` (type `encryption`) — Always
  created: AWS rejects CreateCollection without a matching encryption
  policy. AWS-owned key by default; the referenced customer-managed key
  otherwise. The collection depends on it (create after, destroy before).
- `opensearch.ServerlessCollection` — The collection: type, standby
  replicas, optional collection-group membership, optional vector
  acceleration.
- `opensearch.ServerlessSecurityPolicy` (type `network`) — Always
  created: public reachability by default (SigV4-authenticated; not
  public data), or restricted to OpenSearch Serverless VPC endpoints.
- `opensearch.ServerlessAccessPolicy` (type `data`) — Only when
  `data_access` rules are declared.
- `opensearch.ServerlessLifecyclePolicy` (type `retention`) — Only when
  `retention_rules` are declared.

## How It Works

The module receives an `AwsOpenSearchServerlessCollectionIacInput` (the
manifest plus provider credentials), builds the AWS provider through the
shared builder, and renders the collection plus its policies from the
spec. The policy JSON documents are built from the same typed spec fields
the Terraform module renders through `jsonencode` — arm-for-arm, scoped
to exactly this collection. The materialized defaults (`type`,
`standby_replicas`) are sent explicitly so the manifest's intent is
visible in state on both engines.

## Outputs

| Name | Description |
|------|-------------|
| `collection_id` | Unique ID of the collection |
| `collection_arn` | ARN of the collection (the vector-store ARN for Bedrock knowledge bases) |
| `collection_name` | Name of the collection (matches metadata.name) |
| `collection_endpoint` | OpenSearch API endpoint (HTTPS, SigV4) |
| `dashboard_endpoint` | OpenSearch Dashboards endpoint |
| `kms_key_arn` | The key encrypting the collection (AWS-owned or customer-managed) |
