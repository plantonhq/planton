# AWS DynamoDB

Deploys an AWS DynamoDB table with configurable key schema, billing mode, secondary indexes, streams, encryption, and point-in-time recovery. The component manages all attribute definitions and index configurations through a single declarative manifest.

## What Gets Created

When you deploy an AwsDynamodb resource, Planton provisions:

- **DynamoDB Table** — a `dynamodb.Table` resource with the specified key schema, billing mode, and attribute definitions
- **Global Secondary Indexes** — created when `globalSecondaryIndexes` entries are defined, each with its own key schema, projection, and optional provisioned throughput
- **Local Secondary Indexes** — created when `localSecondaryIndexes` entries are defined, sharing the table's partition key with an alternate sort key
- **Server-Side Encryption** — configured when `serverSideEncryption.enabled` is `true`, optionally using a customer-managed KMS key
- **Point-in-Time Recovery** — enabled when `pointInTimeRecoveryEnabled` is `true`
- **DynamoDB Streams** — activated when `streamEnabled` is `true` with the specified `streamViewType`

All resources are tagged with Planton metadata (organization, environment, resource kind, resource ID).

## Prerequisites

- **AWS credentials** configured via environment variables or Planton provider config
- **A billing mode decision** — choose between `PROVISIONED` (with explicit read/write capacity) or `PAY_PER_REQUEST` (on-demand)
- **A KMS key ARN** if using customer-managed server-side encryption

## Quick Start

Create a file `dynamodb.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsDynamodb
metadata:
  name: my-table
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsDynamodb.my-table
spec:
  region: us-east-1
  billingMode: PAY_PER_REQUEST
  attributeDefinitions:
    - name: pk
      type: S
  keySchema:
    - attributeName: pk
      keyType: HASH
```

Deploy:

```shell
planton apply -f dynamodb.yaml
```

This creates an on-demand DynamoDB table with a single string partition key.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | The AWS region where the resource will be created. | Must be a valid AWS region |
| `billingMode` | `enum` | Billing mode for the table. Valid values: `PROVISIONED`, `PAY_PER_REQUEST`. | Must be set (not `BILLING_MODE_UNSPECIFIED`) |
| `attributeDefinitions` | `object[]` | Attribute definitions referenced by the primary key and indexes. | Minimum 1 item |
| `attributeDefinitions[].name` | `string` | Attribute name. | Minimum length 1 |
| `attributeDefinitions[].type` | `enum` | Attribute data type. Valid values: `S` (String), `N` (Number), `B` (Binary). | Must be a defined value |
| `keySchema` | `object[]` | Primary key schema for the table. Must have exactly one `HASH` key and at most one `RANGE` key. | Minimum 1 item, maximum 2 items |
| `keySchema[].attributeName` | `string` | Must reference an attribute defined in `attributeDefinitions`. | Minimum length 1 |
| `keySchema[].keyType` | `enum` | Key type. Valid values: `HASH` (partition key), `RANGE` (sort key). | Must be a defined value |
| `provisionedThroughput` | `object` | Provisioned capacity settings. Required when `billingMode` is `PROVISIONED`; must be unset or zero for `PAY_PER_REQUEST`. | Conditional |
| `provisionedThroughput.readCapacityUnits` | `int64` | Read capacity units (RCUs). | >= 0; must be > 0 when `PROVISIONED` |
| `provisionedThroughput.writeCapacityUnits` | `int64` | Write capacity units (WCUs). | >= 0; must be > 0 when `PROVISIONED` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `globalSecondaryIndexes` | `object[]` | `[]` | Global secondary indexes. Each GSI requires its own key schema (one `HASH`, optional `RANGE`), projection, and provisioned throughput when billing mode is `PROVISIONED`. |
| `globalSecondaryIndexes[].name` | `string` | — | Index name. |
| `globalSecondaryIndexes[].keySchema` | `object[]` | — | Key schema for the GSI (same structure as table `keySchema`). |
| `globalSecondaryIndexes[].projection.type` | `enum` | — | Projection type. Valid values: `ALL`, `KEYS_ONLY_PROJECTION`, `INCLUDE`. |
| `globalSecondaryIndexes[].projection.nonKeyAttributes` | `string[]` | `[]` | Non-key attributes to project. Required when projection type is `INCLUDE`, must be empty otherwise. Must be unique. |
| `globalSecondaryIndexes[].provisionedThroughput` | `object` | — | RCU/WCU settings. Required when `billingMode` is `PROVISIONED`. |
| `localSecondaryIndexes` | `object[]` | `[]` | Local secondary indexes. Each LSI must have the same `HASH` key as the table and exactly one `RANGE` key. |
| `localSecondaryIndexes[].name` | `string` | — | Index name. |
| `localSecondaryIndexes[].keySchema` | `object[]` | — | Must have exactly one `HASH` key (same as table) and one `RANGE` key. Minimum 2 items. |
| `localSecondaryIndexes[].projection` | `object` | — | Projection configuration (same structure as GSI projection). |
| `ttl` | `object` | — | Time-to-live (TTL) configuration. |
| `ttl.enabled` | `bool` | `false` | Enable TTL expiration. |
| `ttl.attributeName` | `string` | — | Attribute storing TTL epoch time in seconds. Required when `ttl.enabled` is `true`, must be empty when disabled. |
| `streamEnabled` | `bool` | `false` | Enables DynamoDB Streams on the table. |
| `streamViewType` | `enum` | — | Stream view type. Valid values: `KEYS_ONLY`, `NEW_IMAGE`, `OLD_IMAGE`, `NEW_AND_OLD_IMAGES`. Required when `streamEnabled` is `true`, must be unspecified when disabled. |
| `pointInTimeRecoveryEnabled` | `bool` | `false` | Enables point-in-time recovery (PITR) for continuous backups. |
| `serverSideEncryption` | `object` | — | Server-side encryption settings. |
| `serverSideEncryption.enabled` | `bool` | `false` | Enable server-side encryption. |
| `serverSideEncryption.kmsKeyArn` | `string` | — | Customer-managed KMS key ARN. When omitted, AWS-managed encryption is used. |
| `tableClass` | `enum` | `STANDARD` | Table storage class. Valid values: `STANDARD`, `STANDARD_INFREQUENT_ACCESS`. |
| `deletionProtectionEnabled` | `bool` | `false` | Prevents accidental deletion of the table when enabled. |
| `contributorInsightsEnabled` | `bool` | `false` | Enables CloudWatch Contributor Insights for the table. |

## Examples

### On-Demand Table with Composite Key

A table using pay-per-request billing with a partition key and sort key:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsDynamodb
metadata:
  name: orders-table
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsDynamodb.orders-table
spec:
  region: us-east-1
  billingMode: PAY_PER_REQUEST
  attributeDefinitions:
    - name: customerId
      type: S
    - name: orderId
      type: S
  keySchema:
    - attributeName: customerId
      keyType: HASH
    - attributeName: orderId
      keyType: RANGE
```

### Provisioned Table with TTL

A table with explicit read/write capacity and TTL expiration:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsDynamodb
metadata:
  name: sessions-table
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsDynamodb.sessions-table
spec:
  region: us-east-1
  billingMode: PROVISIONED
  provisionedThroughput:
    readCapacityUnits: 10
    writeCapacityUnits: 5
  attributeDefinitions:
    - name: sessionId
      type: S
  keySchema:
    - attributeName: sessionId
      keyType: HASH
  ttl:
    enabled: true
    attributeName: expiresAt
```

### Table with Global Secondary Index

An on-demand table with a GSI for querying by an alternate key:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsDynamodb
metadata:
  name: users-table
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsDynamodb.users-table
spec:
  region: us-east-1
  billingMode: PAY_PER_REQUEST
  attributeDefinitions:
    - name: userId
      type: S
    - name: email
      type: S
  keySchema:
    - attributeName: userId
      keyType: HASH
  globalSecondaryIndexes:
    - name: email-index
      keySchema:
        - attributeName: email
          keyType: HASH
      projection:
        type: ALL
```

### Production Table with Streams and Encryption

A fully configured production table with DynamoDB Streams, customer-managed encryption, point-in-time recovery, and deletion protection:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsDynamodb
metadata:
  name: audit-log
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsDynamodb.audit-log
spec:
  region: us-east-1
  billingMode: PAY_PER_REQUEST
  attributeDefinitions:
    - name: pk
      type: S
    - name: sk
      type: S
    - name: eventType
      type: S
    - name: timestamp
      type: N
  keySchema:
    - attributeName: pk
      keyType: HASH
    - attributeName: sk
      keyType: RANGE
  globalSecondaryIndexes:
    - name: event-type-index
      keySchema:
        - attributeName: eventType
          keyType: HASH
        - attributeName: timestamp
          keyType: RANGE
      projection:
        type: INCLUDE
        nonKeyAttributes:
          - userId
          - action
  streamEnabled: true
  streamViewType: NEW_AND_OLD_IMAGES
  pointInTimeRecoveryEnabled: true
  serverSideEncryption:
    enabled: true
    kmsKeyArn: arn:aws:kms:us-east-1:123456789012:key/abcd-1234-efgh-5678
  tableClass: STANDARD
  deletionProtectionEnabled: true
  contributorInsightsEnabled: true
```

### Provisioned Table with GSI Throughput

When using `PROVISIONED` billing, each GSI must also specify its own throughput:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsDynamodb
metadata:
  name: products-table
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsDynamodb.products-table
spec:
  region: us-east-1
  billingMode: PROVISIONED
  provisionedThroughput:
    readCapacityUnits: 50
    writeCapacityUnits: 25
  attributeDefinitions:
    - name: productId
      type: S
    - name: category
      type: S
    - name: price
      type: N
  keySchema:
    - attributeName: productId
      keyType: HASH
  globalSecondaryIndexes:
    - name: category-price-index
      keySchema:
        - attributeName: category
          keyType: HASH
        - attributeName: price
          keyType: RANGE
      projection:
        type: ALL
      provisionedThroughput:
        readCapacityUnits: 20
        writeCapacityUnits: 10
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `table_name` | `string` | Name of the created DynamoDB table |
| `table_arn` | `string` | ARN of the DynamoDB table |
| `table_id` | `string` | Provider-assigned table ID |
| `stream_arn` | `string` | Stream ARN, populated when `streamEnabled` is `true` |
| `stream_label` | `string` | Stream label, populated when `streamEnabled` is `true` |

## Related Components

- [AwsKmsKey](/docs/catalog/aws/awskmskey) — provides a customer-managed KMS key for server-side encryption
- [AwsIamRole](/docs/catalog/aws/awsiamrole) — creates IAM roles with policies for DynamoDB table access
- [AwsLambda](/docs/catalog/aws/awslambda) — can be triggered by DynamoDB Streams events
