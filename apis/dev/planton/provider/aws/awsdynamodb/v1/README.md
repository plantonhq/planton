# AwsDynamodb

The **AwsDynamodb** resource provides a standardized way to provision and manage AWS DynamoDB tables through Planton. It simplifies DynamoDB table configuration while supporting essential features like flexible billing modes, indexes, streams, TTL, and encryption.

## Spec Fields (80/20)

### Essential Fields (80% Use Case)

- **billing_mode**: Billing mode for the table. Options are:
  - `BILLING_MODE_PAY_PER_REQUEST`: On-demand billing (recommended for variable or unpredictable workloads)
  - `BILLING_MODE_PROVISIONED`: Specify read and write capacity units (RCUs/WCUs) for predictable workloads
- **attribute_definitions**: List of attribute definitions including name and type (S=String, N=Number, B=Binary). Only attributes used in keys or indexes need to be defined here.
- **key_schema**: Primary key schema for the table. Must include exactly one HASH key (partition key) and optionally one RANGE key (sort key).

### Advanced Fields (20% Use Case)

- **provisioned_throughput**: Required when `billing_mode` is PROVISIONED. Specifies read capacity units (RCUs) and write capacity units (WCUs).
- **global_secondary_indexes**: List of global secondary indexes (GSIs) with their own key schemas, projections, and optional provisioned throughput.
- **local_secondary_indexes**: List of local secondary indexes (LSIs) that share the same partition key as the table but use a different sort key.
- **ttl**: Time-to-live configuration for automatic item expiration. Includes `enabled` flag and `attribute_name` storing epoch time in seconds.
- **stream_enabled**: Boolean to enable DynamoDB Streams for change data capture.
- **stream_view_type**: Type of data to include in stream records when streams are enabled:
  - `STREAM_VIEW_TYPE_KEYS_ONLY`: Only key attributes
  - `STREAM_VIEW_TYPE_NEW_IMAGE`: Entire item after modification
  - `STREAM_VIEW_TYPE_OLD_IMAGE`: Entire item before modification
  - `STREAM_VIEW_TYPE_NEW_AND_OLD_IMAGES`: Both before and after states
- **point_in_time_recovery_enabled**: Enable continuous backups for point-in-time recovery (PITR).
- **server_side_encryption**: Server-side encryption settings with optional customer-managed KMS key ARN.
- **table_class**: Storage class - `TABLE_CLASS_STANDARD` or `TABLE_CLASS_STANDARD_INFREQUENT_ACCESS` (for cost optimization on infrequently accessed data).
- **deletion_protection_enabled**: Prevent accidental deletion of the table.
- **contributor_insights_enabled**: Enable CloudWatch Contributor Insights for monitoring access patterns.

## Stack Outputs

After provisioning, the AwsDynamodb resource provides the following outputs:

- **table_name**: The name of the DynamoDB table.
- **table_arn**: The Amazon Resource Name (ARN) of the table.
- **table_id**: Provider-assigned unique identifier for the table.
- **stream_arn**: ARN of the DynamoDB Stream (when streams are enabled).
- **stream_label**: Label of the stream (when streams are enabled).

## How It Works

When you define an AwsDynamodb resource, Planton:

1. **Creates Table**: Provisions a DynamoDB table with the specified key schema and attribute definitions.
2. **Configures Billing**: Sets up either on-demand or provisioned billing based on your specification.
3. **Adds Indexes**: Creates global and/or local secondary indexes for efficient querying on non-primary-key attributes.
4. **Enables Features**: Activates optional features like TTL, streams, PITR, and encryption as specified.
5. **Applies Protection**: Configures deletion protection and table class for cost and safety optimization.

The resource uses Pulumi or Terraform under the hood (depending on your stack configuration) to provision all necessary AWS resources.

## Use Cases

### Simple Key-Value Store
Use on-demand billing with a single partition key for simple key-value storage with unpredictable access patterns.

### Time-Series Data with TTL
Configure a composite key (partition + sort key) with TTL enabled to automatically expire old time-series data.

### Event Sourcing with Streams
Enable DynamoDB Streams with NEW_AND_OLD_IMAGES to capture all changes for event sourcing, replication, or triggering Lambda functions.

### Multi-Access Pattern with GSIs
Create global secondary indexes to support multiple query patterns on the same data without table scans.

### Cost Optimization
Use PAY_PER_REQUEST billing for variable workloads or STANDARD_INFREQUENT_ACCESS table class for infrequently accessed data.

## Important Notes

### Billing Modes
- **PAY_PER_REQUEST** (on-demand): Best for new applications, variable traffic, or when you can't predict capacity needs. No capacity planning required.
- **PROVISIONED**: Best for predictable workloads. Lower cost when you can accurately forecast usage, but risk throttling if exceeded.

### Key Schema Design
- **Partition Key (HASH)**: Must have high cardinality for even data distribution across partitions.
- **Sort Key (RANGE)**: Optional, enables range queries and allows multiple items with same partition key.
- **Attribute Definitions**: Only define attributes used in keys or indexes, not all table attributes.

### Index Considerations
- **GSIs**: Can have different partition and sort keys than the table. Support eventual consistency. Can be added/removed after table creation.
- **LSIs**: Must share the same partition key as the table. Support strong consistency. Must be defined at table creation time (cannot add later).
- **Projection**: Controls which attributes are copied to the index:
  - `ALL`: All attributes (highest storage cost)
  - `KEYS_ONLY`: Only key attributes (lowest cost)
  - `INCLUDE`: Specific non-key attributes

### Streams and TTL
- **Streams**: Capture item-level changes for up to 24 hours. Commonly used with Lambda for processing changes, replication, or analytics.
- **TTL**: Automatic deletion of expired items within 48 hours. No additional cost. TTL attribute must be a number representing epoch time in seconds.

### Capacity Planning for Provisioned Mode
- 1 RCU = 1 strongly consistent read/sec for items up to 4KB (or 2 eventually consistent reads/sec)
- 1 WCU = 1 write/sec for items up to 1KB
- GSIs consume their own capacity units
- Enable auto-scaling for provisioned mode to handle traffic spikes

## Validation Rules

The API enforces several validations:

1. **Billing Mode Required**: Must be set to either PROVISIONED or PAY_PER_REQUEST.
2. **Key Schema**: Must have exactly one HASH key and at most one RANGE key.
3. **Provisioned Throughput**: Required when billing mode is PROVISIONED (with RCUs and WCUs > 0).
4. **Stream View Type**: Must be set when streams are enabled, must be unspecified when disabled.
5. **TTL Attribute**: Must be set when TTL is enabled, must be empty when disabled.
6. **GSI Keys**: Each GSI must have exactly one HASH key and at most one RANGE key.
7. **LSI Keys**: Each LSI must have exactly one HASH key and exactly one RANGE key.
8. **Projection**: When projection type is INCLUDE, non_key_attributes must be specified.

## References

- [DynamoDB Core Components](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/HowItWorks.CoreComponents.html)
- [DynamoDB Billing Modes](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/HowItWorks.ReadWriteCapacityMode.html)
- [DynamoDB Streams](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Streams.html)
- [DynamoDB TTL](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/TTL.html)
- [Global Secondary Indexes](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/GSI.html)
- [Local Secondary Indexes](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/LSI.html)
- [DynamoDB Best Practices](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/best-practices.html)
