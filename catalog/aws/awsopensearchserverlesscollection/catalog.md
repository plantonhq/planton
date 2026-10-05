# AWS OpenSearch Serverless Collection

Deploys an Amazon OpenSearch Serverless collection -- a fully managed, auto-scaling OpenSearch workspace (capacity billed in OpenSearch Compute Units, no domains or nodes to size) for search, time-series, and vector workloads -- together with the collection-scoped encryption, network, data-access, and retention policies that make a collection usable. One manifest owns one collection and everything attached to it: the modules render each policy scoped to exactly this collection's name. VECTORSEARCH collections are the vector store Bedrock knowledge bases consume.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Encryption Security Policy** -- always created, scoped to exactly this collection (AWS rejects CreateCollection without a matching encryption policy): AWS-owned key by default, or the referenced customer-managed KMS key
- **Collection** -- SEARCH, TIMESERIES (default), or VECTORSEARCH, with standby replicas (default ENABLED; DISABLED halves the OCU floor for dev/test) and optional collection-group membership
- **Network Security Policy** -- public reachability of the collection and Dashboards endpoints by default (SigV4-authenticated reachability only -- not public data), or restricted to OpenSearch Serverless VPC endpoints
- **Data Access Policy** -- rendered from `dataAccess` rules granting IAM principals collection- and index-level permissions; without at least one rule nothing can read or write data (IAM permissions alone grant nothing in OpenSearch Serverless)
- **Lifecycle Policy** -- rendered from `retentionRules` (index retention by pattern; indefinite without rules)
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically on the collection

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **IAM principals for data access** -- the roles or users your applications run as; grant them through `dataAccess` rules. Reference AwsIamRole Infra Components via ValueFromRef or provide ARNs directly.
- **A KMS key** (optional) for customer-managed encryption -- the key choice is fixed at create time. Provide the ARN or reference an AwsKmsKey Infra Component.
- **OpenSearch Serverless VPC endpoints** (optional) for private network access -- these are the service's OWN endpoint objects (created through the OpenSearch Serverless API, not ordinary Interface Endpoints); create them outside this component and list their IDs in `network.vpcEndpointIds`.

## Deploy

### Console

Open the deployment store, find **AWS OpenSearch Serverless Collection**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Dev Search Collection** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsOpenSearchServerlessCollection
metadata:
  name: app-search
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  type: SEARCH
  standbyReplicas: DISABLED
  dataAccess:
    - principals:
        - value: arn:aws:iam::123456789012:role/app-role
      indexPermissions:
        - aoss:ReadDocument
        - aoss:WriteDocument
```

```shell
planton apply -f collection.yaml
```

This creates a single-AZ SEARCH collection at the halved OCU floor, publicly reachable with SigV4 auth, with one data-access rule letting the application role read and write documents in every index. An Infra Job tracks the provisioning in real time.

### InfraChart

When the collection deploys alongside its KMS key and application role in one chart, wire both references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  type: VECTORSEARCH
  encryption:
    kmsKeyArn:
      valueFrom:
        kind: AwsKmsKey
        name: search-key
        fieldPath: status.outputs.key_arn
  dataAccess:
    - principals:
        - valueFrom:
            kind: AwsIamRole
            name: app-role
            fieldPath: status.outputs.role_arn
      indexPermissions:
        - aoss:ReadDocument
        - aoss:WriteDocument
```

The InfraPipeline resolves the dependency graph, creates the key and role first, then the collection with its policies wired to them.

## Key Configuration

These are the most important decisions when configuring a collection. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The ForceNew surface is wide** -- the name (`metadata.name`), `type`, `standbyReplicas`, `collectionGroupName`, and the encryption key choice are all fixed at create time; changing any of them destroys and recreates the collection. Network, data-access, and retention rules update in place. Decide workload type, availability posture, and key ownership before data lands.

**Standby replicas set the cost floor** -- the OCU floor bills from collection ACTIVE to delete: 0.5+0.5 OCU with standby DISABLED, 2+2 with ENABLED (the production default). Standby is ForceNew, so a dev collection cannot be upgraded in place to the HA posture -- and dev collections should be destroyed when idle rather than left at the floor.

**Data access is the real gate** -- IAM identity permissions alone grant NOTHING in OpenSearch Serverless; every data operation is authorized through the data access policy. A collection without at least one `dataAccess` rule is read-proof and write-proof -- legal to create (policies always trail the collection at AWS), but inert. Grant the workload role `aoss:ReadDocument`/`aoss:WriteDocument` on its index patterns, and reserve `aoss:*` for admin principals.

**Public means reachable, not open** -- an omitted `network` block renders the public posture for both the collection and Dashboards endpoints: reachable over the internet, but every request still needs SigV4 signing plus a data-access rule. For private access, set `allowFromPublic: false` with `vpcEndpointIds` -- and note these must be OpenSearch Serverless's own VPC endpoint objects; a vpce- ID from an ordinary Interface Endpoint will not authorize.

**Retention is for TIMESERIES discipline** -- without `retentionRules`, indexes retain data indefinitely and OCU-backed storage grows unbounded. Log-analytics collections should pair a broad pattern rule (`logs-*` at `30d`) with `unlimited: true` exemptions for the indexes that must survive. Removing all rules deletes the policy and retention becomes indefinite again.

**VECTORSEARCH is a distinct type, not a setting** -- Bedrock knowledge bases require it, and `serverlessVectorAcceleration` (GPU-backed vector capacity) applies only to it. Since `type` is ForceNew, a SEARCH collection can never become the vector store later.

**Collection groups pool capacity** -- `collectionGroupName` places the collection in an existing group that pools OCU limits across members. The group's own standby setting must match the collection's, and membership is fixed at create.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** | `encryption.kmsKeyArn` | `status.outputs.key_arn` |
| **AwsIamRole** | `dataAccess[].principals[]` | `status.outputs.role_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `collection_endpoint` | HTTPS endpoint for OpenSearch API operations | Application indexing and query configuration (SigV4-signed) |
| `collection_arn` | Amazon Resource Name of the collection | The vector-store ARN a Bedrock knowledge base consumes; IAM policies |
| `dashboard_endpoint` | Endpoint for OpenSearch Dashboards | Operator bookmarks and team portals |
| `collection_id` | The API's own collection identifier, also the leading label of the endpoints | CLI and API operations addressing the collection |

`collection_name` (matching `metadata.name`) and `kms_key_arn` (the resolved key, AWS-owned or customer-managed) are also exported -- input echoes useful for audit, not composition inputs.

## Common Patterns

**Dev search sandbox** -- a SEARCH collection with standby DISABLED for the halved OCU floor, public reachability, and one permissive data-access rule for the developer role. Destroy it when the spike ends; the floor bills until you do. Start from the **Dev Search Collection** preset.

**Production log analytics** -- a TIMESERIES collection with standby ENABLED, retention rules expiring `logs-*` indexes on a schedule, and least-privilege data access separating the ingest role (write) from the query role (read). Start from the **Production Time-Series (Log Analytics)** preset.

**Bedrock vector store** -- a VECTORSEARCH collection whose `collection_arn` a Bedrock knowledge base consumes, with the knowledge base's service role granted index read/write through `dataAccess`. Start from the **Bedrock Vector Store** preset.

## Works With

- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- customer-managed encryption, fixed at create time via `encryption.kmsKeyArn`
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the application and service principals granted data-plane access through `dataAccess` rules
- [**AWS Bedrock Knowledge Base**](/infra-catalog/aws-bedrock-knowledge-base) -- consumes a VECTORSEARCH collection as its vector store via `collection_arn`
