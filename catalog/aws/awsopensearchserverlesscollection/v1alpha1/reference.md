# AwsOpenSearchServerlessCollection

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsOpenSearchServerlessCollectionSpec defines the desired configuration
for an Amazon OpenSearch Serverless collection - a fully managed,
auto-scaling OpenSearch workspace (no domains, no nodes, capacity billed
in OpenSearch Compute Units) for search, time-series, or vector
workloads.

The collection's name is taken from `metadata.name` (3-32 characters,
lowercase letters, numbers, and hyphens, starting with a letter -
AWS-enforced at create). The name is ForceNew.

OpenSearch Serverless separates the collection from three account-level
POLICY objects that attach to collections by name-pattern matching:
encryption security policies, network security policies, and data access
policies. This component scopes all three to exactly this collection -
the modules render each policy with rules matching only this
collection's name, so one manifest owns one collection and everything
that makes it usable. (Account-wide pattern policies shared by many
collections are a different tool and deliberately out of this spec.)

A collection CANNOT be created without a matching encryption policy -
the modules always render one (AWS-owned key by default; set
`encryption.kms_key_arn` for a customer-managed key). Network access
defaults to public reachability of the collection AND dashboard
endpoints - note "public" here means network reachability only: every
request must still be SigV4-signed and authorized by a data access rule,
so an omitted `network` block is the AWS console's own easy-create
posture, not an open database. Without at least one `data_access` rule
nothing can read or write data (IAM permissions alone are not
sufficient in OpenSearch Serverless).

Create-time-immutable (ForceNew) fields: the name (metadata.name),
`type`, `standby_replicas`, `collection_group_name`, and the encryption
key choice. Network, data-access, and retention rules update in place.

Credentials, region, and deployment workflow live outside this spec in
IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsOpenSearchServerlessCollection
metadata:
  name: test-app-search
  id: test-app-search
  org: test-org
  env: dev
spec:
  region: us-west-2
  description: Full-surface OpenSearch Serverless hack manifest
  type: SEARCH
  standbyReplicas: DISABLED
  encryption:
    kmsKeyArn:
      value: arn:aws:kms:us-west-2:123456789012:key/abc-123
  network:
    allowFromPublic: true
    includeDashboards: true
  dataAccess:
    - principals:
        - value: arn:aws:iam::123456789012:role/app-role
      indexPermissions:
        - aoss:ReadDocument
        - aoss:WriteDocument
        - aoss:CreateIndex
        - aoss:DescribeIndex
      indexPatterns:
        - "*"
    - principals:
        - value: arn:aws:iam::123456789012:role/admin-role
      collectionPermissions:
        - "aoss:*"
      indexPermissions:
        - "aoss:*"
  retentionRules:
    - indexPatterns:
        - "logs-*"
      minIndexRetention: 30d
    - indexPatterns:
        - "audit-*"
      unlimited: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.type` | `string` |  | `TIMESERIES` |  |
| `spec.standbyReplicas` | `string` |  | `ENABLED` |  |
| `spec.collectionGroupName` | `string` |  |  |  |
| `spec.serverlessVectorAcceleration` | `string` |  |  |  |
| `spec.encryption` | `AwsOpenSearchServerlessCollectionEncryption` |  |  |  |
| `spec.encryption.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.network` | `AwsOpenSearchServerlessCollectionNetwork` |  |  |  |
| `spec.network.allowFromPublic` | `bool` |  | `true` |  |
| `spec.network.vpcEndpointIds` | `[]string` |  |  |  |
| `spec.network.includeDashboards` | `bool` |  | `true` |  |
| `spec.dataAccess` | `[]AwsOpenSearchServerlessCollectionAccessRule` |  |  |  |
| `spec.dataAccess[].principals` | `[]string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.dataAccess[].collectionPermissions` | `[]string` |  |  |  |
| `spec.dataAccess[].indexPermissions` | `[]string` |  |  |  |
| `spec.dataAccess[].indexPatterns` | `[]string` |  |  |  |
| `spec.retentionRules` | `[]AwsOpenSearchServerlessCollectionRetentionRule` |  |  |  |
| `spec.retentionRules[].indexPatterns` | `[]string` | yes |  |  |
| `spec.retentionRules[].minIndexRetention` | `string` |  |  |  |
| `spec.retentionRules[].unlimited` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the collection will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

Human-readable description of the collection (up to 1000 characters).
Updates in place.

- rule: {"string":{"maxLen":"1000"}}

### spec.type

`string` · optional (explicit presence)

Workload type of the collection. "SEARCH" for full-text/application
search, "TIMESERIES" for log and observability analytics (the AWS
default), "VECTORSEARCH" for vector embeddings (the type Bedrock
knowledge bases require). ForceNew - changing it destroys and
recreates the collection.

- default: `TIMESERIES`
- rule: {"string":{"in":["SEARCH","TIMESERIES","VECTORSEARCH"]}}

### spec.standbyReplicas

`string` · optional (explicit presence)

Whether the collection keeps warm standby replicas in a second AZ for
higher availability. "ENABLED" (the AWS default; production posture,
doubles the minimum OCU floor) or "DISABLED" (half the floor cost -
the right choice for dev/test). ForceNew - fixed at create time.

- default: `ENABLED`
- rule: {"string":{"in":["ENABLED","DISABLED"]}}

### spec.collectionGroupName

`string`

Name of an existing collection group to place this collection in
(collection groups pool OCU capacity limits across member collections).
Leave empty for an ungrouped collection. ForceNew - the group choice
is fixed at create time. The group's own standby-replicas setting must
match the collection's.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-z][0-9a-z-]{2,31}$"}}

### spec.serverlessVectorAcceleration

`string`

Serverless vector acceleration for VECTORSEARCH collections - AWS
manages GPU-accelerated capacity for vector workloads. Values:
"ENABLED", "DISABLED". Leave empty for the AWS default. Only
meaningful on VECTORSEARCH collections.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ENABLED","DISABLED"]}}

### spec.encryption

`AwsOpenSearchServerlessCollectionEncryption`

Encryption key choice for the collection. OpenSearch Serverless always
encrypts at rest and requires a matching encryption security policy
BEFORE the collection can be created - the modules always render one
scoped to exactly this collection. Omit the block for the AWS-owned
key; set `kms_key_arn` for a customer-managed key. The key choice is
ForceNew at AWS (an existing collection cannot change keys).

### spec.encryption.kmsKeyArn

`string | valueFrom`

Customer-managed KMS key (ARN) for encrypting the collection. Omit for
the AWS-owned key. ForceNew at AWS - the key choice is fixed when the
collection is created.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.network

`AwsOpenSearchServerlessCollectionNetwork`

Network reachability of the collection's OpenSearch endpoint and its
Dashboards endpoint. Omitting the block renders a PUBLIC-access
network policy for both (the AWS console's easy-create posture -
reachable over the internet, but every request still needs SigV4
auth plus a data-access rule). Set `allow_from_public: false` with
`vpc_endpoint_ids` to restrict access to OpenSearch Serverless VPC
endpoints. Updates in place.

- rule: when allow_from_public is false, provide at least one entry in vpc_endpoint_ids
- rule: vpc_endpoint_ids only take effect when allow_from_public is false

### spec.network.allowFromPublic

`bool` · optional (explicit presence)

Allow access from the public internet (SigV4-authenticated; data
access still requires a data_access rule). Default true - the AWS
easy-create posture. Set false to restrict access to the VPC
endpoints listed in `vpc_endpoint_ids`.

- default: `true`

### spec.network.vpcEndpointIds

`[]string`

OpenSearch Serverless VPC endpoint IDs (vpce-...) allowed to reach the
collection when `allow_from_public` is false. These are the service's
OWN VPC endpoints (created through the OpenSearch Serverless API, not
ordinary Interface Endpoints); create them outside this component and
reference their IDs here.

- rule: {"repeated":{"items":{"string":{"pattern":"^vpce-[0-9a-z]+$"}}}}

### spec.network.includeDashboards

`bool` · optional (explicit presence)

Whether the network rules above also cover the OpenSearch Dashboards
endpoint (default true - one posture for both endpoints). Set false
to render network rules for the collection endpoint only, leaving
Dashboards unreachable (no Dashboards rule at all).

- default: `true`

### spec.dataAccess

`[]AwsOpenSearchServerlessCollectionAccessRule`

Data-plane permission rules for this collection. OpenSearch Serverless
authorizes every data operation through data access policies - IAM
identity permissions alone grant NOTHING, so a collection without at
least one rule is write-proof and read-proof. Each rule grants a set
of principals (IAM roles/users) permissions on the collection and/or
its indexes. The modules render one data access policy scoped to
exactly this collection. Updates in place.

- rule: grant at least one of collection_permissions or index_permissions
- rule: index_patterns only take effect together with index_permissions

### spec.dataAccess[].principals

`[]string | valueFrom` · required

IAM principals (role or user ARNs) the rule grants. For most
applications this is the workload's IAM role.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.dataAccess[].collectionPermissions

`[]string`

Collection-level permissions to grant. Values: "aoss:CreateCollectionItems",
"aoss:DeleteCollectionItems", "aoss:UpdateCollectionItems",
"aoss:DescribeCollectionItems", or "aoss:*" (all of them). Leave empty
to grant index-level permissions only.

- rule: {"repeated":{"items":{"string":{"in":["aoss:CreateCollectionItems","aoss:DeleteCollectionItems","aoss:UpdateCollectionItems","aoss:DescribeCollectionItems","aoss:*"]}}}}

### spec.dataAccess[].indexPermissions

`[]string`

Index-level permissions to grant on indexes matching `index_patterns`.
Values: "aoss:ReadDocument", "aoss:WriteDocument", "aoss:CreateIndex",
"aoss:DeleteIndex", "aoss:UpdateIndex", "aoss:DescribeIndex", or
"aoss:*". Leave empty to grant collection-level permissions only.

- rule: {"repeated":{"items":{"string":{"in":["aoss:ReadDocument","aoss:WriteDocument","aoss:CreateIndex","aoss:DeleteIndex","aoss:UpdateIndex","aoss:DescribeIndex","aoss:*"]}}}}

### spec.dataAccess[].indexPatterns

`[]string`

Index name patterns (within this collection) the index permissions
apply to. Default: all indexes ("*"). Patterns support a trailing
wildcard, e.g. "logs-*".

### spec.retentionRules

`[]AwsOpenSearchServerlessCollectionRetentionRule`

Index retention rules for this collection (lifecycle policy of type
"retention"). Each rule sets how long documents in matching indexes
are retained; without rules, data is retained indefinitely. Mostly
used with TIMESERIES collections. Updates in place; removing all
rules deletes the policy (retention becomes indefinite again).

- rule: set exactly one of min_index_retention (e.g. "24h", "30d") or unlimited: true

### spec.retentionRules[].indexPatterns

`[]string` · required

Index name patterns (within this collection) the rule applies to,
e.g. "*" (all indexes) or "logs-*". A trailing wildcard is the only
wildcard AWS supports.

- rule: {"repeated":{"minItems":"1"}}

### spec.retentionRules[].minIndexRetention

`string`

Minimum retention period for matching indexes - a number followed by
"h" (hours) or "d" (days), e.g. "24h", "30d". Mutually exclusive with
`unlimited`.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]{1,4}(h|d)$"}}

### spec.retentionRules[].unlimited

`bool`

Retain matching indexes indefinitely (explicit no-expiry rule -
useful to exempt specific indexes from a broader retention rule).
Mutually exclusive with `min_index_retention`.

## Validation Rules

- `vector_acceleration_requires_vectorsearch`: serverless_vector_acceleration only applies when type is VECTORSEARCH

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsOpenSearchServerlessCollection, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.collection_id` | `string` | The unique ID of the collection (the API's own identifier, also the leading label of the collection endpoints). |
| `status.outputs.collection_arn` | `string` | The Amazon Resource Name of the collection. Used in IAM policies and as the vector-store ARN when a Bedrock knowledge base consumes the collection. |
| `status.outputs.collection_name` | `string` | The name of the collection. Matches metadata.name. Data access and lifecycle rules key off this name. |
| `status.outputs.collection_endpoint` | `string` | Collection-specific endpoint for OpenSearch API operations (HTTPS). Applications index and query through this endpoint with SigV4 auth. |
| `status.outputs.dashboard_endpoint` | `string` | Collection-specific endpoint for OpenSearch Dashboards. |
| `status.outputs.kms_key_arn` | `string` | The ARN of the KMS key encrypting the collection - the AWS-owned key or the customer-managed key chosen in spec.encryption. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.encryption.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.dataAccess[].principals` | AwsIamRole | `status.outputs.role_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBedrockKnowledgeBase | `spec.storage.opensearchServerless.collectionArn` | `status.outputs.collection_arn` |

## See Also

- [Overview](../README.md)
