# Pulumi Module Architecture: AWS DynamoDB

## Overview

This Pulumi module provisions AWS DynamoDB tables through a declarative, protobuf-defined specification. The architecture follows OpenMCF's standard pattern: input transformation → resource provisioning → output extraction.

The module is designed to expose DynamoDB's full feature set while maintaining simplicity through careful abstraction of AWS API complexity.

## Module Structure

```
iac/pulumi/
├── main.go              # Pulumi entrypoint
├── Pulumi.yaml          # Pulumi project metadata
├── Makefile             # Build/test helpers
├── debug.sh             # Delve debugging wrapper
└── module/
    ├── main.go          # Orchestration logic (provider setup, output export)
    ├── locals.go        # Input transformation and tag construction
    ├── table.go         # DynamoDB table resource implementation
    └── outputs.go       # Output key constants
```

### File Responsibilities

#### `main.go` (entrypoint)
- Unmarshals `AwsDynamodbStackInput` from Pulumi stack configuration
- Delegates to `module.Resources()` for actual provisioning
- Minimal logic—purely a thin entrypoint wrapper

#### `module/main.go` (orchestrator)
**Key Function:** `Resources(ctx, stackInput)`

Responsibilities:
1. **Locals Initialization**: Transform stack input into typed locals struct
2. **Provider Configuration**: Handle two provider scenarios:
   - **Default**: Create provider with ambient AWS credentials (IAM role, environment variables)
   - **Explicit**: Create provider with credentials from `stackInput.ProviderConfig` (access key, secret key, session token)
3. **Resource Creation**: Invoke `createTable()` with initialized locals and provider
4. **Output Export**: Export five DynamoDB-specific outputs using Pulumi's export mechanism

**Design Decision**: Provider handling is bifurcated to support both:
- **CI/CD environments**: IRSA (IAM Roles for Service Accounts) or instance profiles provide ambient credentials
- **Local/manual deployments**: Explicit credentials passed via stack input

#### `module/locals.go` (input transformer)
**Key Function:** `initializeLocals(ctx, stackInput)`

Transforms the protobuf `AwsDynamodbStackInput` into a strongly-typed `Locals` struct and constructs AWS resource tags.

**Tag Construction**: Every DynamoDB table receives five mandatory tags:
- `resource=true`: Marks this as a OpenMCF managed resource
- `organization=<org>`: Organization ID from metadata
- `environment=<env>`: Environment (dev/staging/prod) from metadata
- `resource-kind=AwsDynamodb`: CloudResourceKind enum string
- `resource-id=<id>`: Unique resource identifier from metadata

**Why Tags Matter**: These tags enable:
- Cost allocation reporting by org/env/resource-kind
- Policy enforcement (e.g., "only production resources can use CMK encryption")
- Resource discovery and inventory management
- Drift detection (manually created resources lack these tags)

#### `module/table.go` (resource implementation)
**Key Function:** `createTable(ctx, locals, provider)`

This is the core implementation file containing all DynamoDB provisioning logic. It translates the protobuf `AwsDynamodbSpec` into Pulumi's `dynamodb.TableArgs`.

**Mapping Strategy**: The implementation follows a deliberate pattern of **explicit field-by-field mapping** rather than automated reflection or struct copying. This provides:
- **Type Safety**: Compile-time verification of field compatibility
- **API Evolution Resilience**: Protobuf and AWS API changes are caught at build time
- **Explicit Defaults**: Clear handling of zero-values vs intentional configuration

#### `module/outputs.go` (constants)
Defines string constants for output keys. This prevents typos in export calls and provides a single source of truth for output names.

**Exported Outputs**:
- `table_name`: The DynamoDB table's resource name
- `table_arn`: Full ARN for IAM policy construction
- `table_id`: Pulumi resource identifier
- `stream_arn`: DynamoDB Streams ARN (only populated if streams are enabled)
- `stream_label`: Stream label for Lambda event source mapping

## Data Flow Diagram

```
┌─────────────────────────────────────┐
│ AwsDynamodbStackInput (protobuf)    │
│  ├─ target: AwsDynamodb              │
│  │   ├─ metadata                     │
│  │   └─ spec: AwsDynamodbSpec        │
│  └─ provider_config (optional)      │
└──────────────┬──────────────────────┘
               │
               ▼
      ┌────────────────┐
      │ initializeLocals│
      └────────┬────────┘
               │ Creates:
               │ - Locals.Target
               │ - Locals.Spec
               │ - Locals.AwsTags (map[string]string)
               ▼
     ┌──────────────────────┐
     │ AWS Provider Setup   │
     │  ├─ Ambient creds OR │
     │  └─ Explicit creds   │
     └──────────┬───────────┘
                │
                ▼
       ┌─────────────────┐
       │  createTable()  │
       └────────┬────────┘
                │
                ├─ Map billing_mode enum → "PROVISIONED" | "PAY_PER_REQUEST"
                │
                ├─ Transform attribute_definitions:
                │    proto AttributeDefinition → dynamodb.TableAttributeArgs
                │    (name + type: S/N/B)
                │
                ├─ Extract key_schema:
                │    Identify HASH key (partition key)
                │    Identify RANGE key (sort key) if present
                │
                ├─ Build Local Secondary Indexes:
                │    For each LSI spec:
                │      - Extract range key (hash key is implicit—same as table)
                │      - Map projection type: KEYS_ONLY / INCLUDE / ALL
                │      - Collect non_key_attributes if projection type is INCLUDE
                │
                ├─ Build Global Secondary Indexes:
                │    For each GSI spec:
                │      - Extract hash key and optional range key
                │      - Map projection type
                │      - Conditionally set read/write capacity if PROVISIONED mode
                │
                ├─ Configure server-side encryption:
                │    If spec.server_side_encryption.enabled:
                │      - enabled=true
                │      - optional KMS key ARN
                │
                ├─ Conditionally configure streams:
                │    If spec.stream_enabled:
                │      - Set stream_enabled=true
                │      - Map stream_view_type enum → "KEYS_ONLY" | "NEW_IMAGE" | etc.
                │
                ├─ Set table_class:
                │    STANDARD (default) | STANDARD_INFREQUENT_ACCESS
                │
                ├─ Enable point-in-time recovery:
                │    If spec.point_in_time_recovery_enabled: pitr.enabled=true
                │
                └─ Set provisioned throughput:
                     If billing_mode=PROVISIONED:
                       read_capacity, write_capacity from spec
                │
                ▼
        ┌─────────────────────────┐
        │ dynamodb.NewTable()     │
        │  (Pulumi AWS Provider)  │
        └────────┬────────────────┘
                 │
                 ▼ Creates AWS DynamoDB Table
       ┌────────────────────────┐
       │  AWS DynamoDB Service  │
       │   ├─ Table Resource    │
       │   ├─ GSI Resources     │
       │   ├─ LSI Resources     │
       │   └─ Stream (if enabled)│
       └────────┬───────────────┘
                │
                ▼
        ┌───────────────────┐
        │ Output Exports    │
        │  ├─ table_name    │
        │  ├─ table_arn     │
        │  ├─ table_id      │
        │  ├─ stream_arn    │
        │  └─ stream_label  │
        └───────────────────┘
```

## Resource Relationships

```
AwsDynamodbSpec
  │
  ├─ billing_mode ─────────────────┐
  │                                 │
  ├─ attribute_definitions ─────┐  │
  │                              │  │
  ├─ key_schema ───────────────┐│  │
  │                             ││  │
  ├─ global_secondary_indexes ─┼┼──┼─ DynamoDB Table
  │   ├─ key_schema            │││  │   ├─ Primary Key (HASH + optional RANGE)
  │   ├─ projection            │││  │   ├─ Billing Mode (capacity)
  │   └─ provisioned_throughput┘││  │   ├─ Attributes (for keys only)
  │                              ││  │   ├─ Tags (org/env/resource metadata)
  ├─ local_secondary_indexes ───┼┼──┘   │
  │   ├─ key_schema             │││      ├─ GSI (separate capacity, eventual consistency)
  │   └─ projection             │││      │   ├─ Alternate partition key
  │                              │││      │   ├─ Optional sort key
  ├─ server_side_encryption ────┼┼┼──────┤   └─ Projection (which attributes to copy)
  │                              │││      │
  ├─ stream configuration ───────┼┼┼──────├─ LSI (shares table capacity, strong consistency)
  │   ├─ stream_enabled         │││      │   ├─ Same partition key as table
  │   └─ stream_view_type       │││      │   ├─ Alternate sort key
  │                              │││      │   └─ Projection
  ├─ point_in_time_recovery ────┼┼┼──────┤
  │                              │││      ├─ Encryption (SSE with optional CMK)
  ├─ table_class ────────────────┼┼┼──────┤
  │                              │││      ├─ DynamoDB Streams (change data capture)
  └─ provisioned_throughput ────┘┘└──────┤   ├─ Stream ARN
                                          │   └─ View Type (KEYS_ONLY, NEW_IMAGE, etc.)
                                          │
                                          ├─ Point-in-Time Recovery (35-day backups)
                                          │
                                          └─ Table Class (STANDARD or INFREQUENT_ACCESS)
```

### Critical Relationships

**Billing Mode ↔ Capacity**: 
- `PAY_PER_REQUEST`: No capacity configuration needed. Table and GSIs auto-scale.
- `PROVISIONED`: Requires explicit read/write capacity units for table AND each GSI independently.

**Attributes ↔ Keys**: 
- `attribute_definitions` only declares attributes used in key schemas (partition/sort keys for table, GSIs, LSIs)
- **Common Pitfall**: Developers familiar with relational schemas expect to define all item attributes. DynamoDB is schemaless—only key attributes are declared at table creation.

**GSI Capacity Independence**: 
- In `PROVISIONED` mode, each GSI has its own read/write capacity
- The module conditionally sets `gsi.ReadCapacity` and `gsi.WriteCapacity` only when `billing_mode == PROVISIONED`
- In `PAY_PER_REQUEST` mode, these fields must be omitted (not zero—absent)

**Streams ↔ Lambda Integration**: 
- The exported `stream_arn` is used by Lambda event source mappings
- Stream must be enabled before Lambda can consume it
- The `stream_label` identifies a specific stream version (changes if stream is disabled and re-enabled)

## Key Design Decisions

### 1. Explicit Enum Mapping Over Reflection

**Decision**: All protobuf enums are explicitly mapped to AWS string constants.

**Example**:
```go
switch locals.Spec.BillingMode {
case awsdynamodbv1.AwsDynamodbSpec_BILLING_MODE_PROVISIONED:
    billingMode = pulumi.StringPtr("PROVISIONED")
case awsdynamodbv1.AwsDynamodbSpec_BILLING_MODE_PAY_PER_REQUEST:
    billingMode = pulumi.StringPtr("PAY_PER_REQUEST")
}
```

**Rationale**:
- **Type Safety**: Compiler catches missing enum cases
- **API Versioning**: If AWS adds new billing modes, the code fails to compile, forcing explicit handling
- **Clarity**: No hidden mappings—the transformation is visible

### 2. Projection Type Defaults to ALL

**Decision**: If projection type is unspecified, default to `ALL` (copy all attributes to index).

**Code**:
```go
gsi.ProjectionType = pulumi.String("ALL")  // Default
if g.Projection != nil {
    switch g.Projection.Type {
        // Override if specified
    }
}
```

**Rationale**:
- **Simplicity for Prototypes**: New users don't need to understand projection types to create working indexes
- **Production Optimization Path**: Teams can optimize projections later (change `ALL` to `KEYS_ONLY` or `INCLUDE`)
- **AWS Default Alignment**: AWS API also defaults to `ALL` if projection is unspecified

**Trade-off**: This sacrifices cost efficiency for ease of use. Production-ready deployments should explicitly specify `KEYS_ONLY` or `INCLUDE` projections to minimize storage and write amplification costs.

### 3. Conditional Provider Configuration

**Decision**: Support both ambient AWS credentials and explicit credential configuration.

**Implementation**:
```go
if awsProviderConfig == nil {
    provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
        Region: pulumi.String(locals.Spec.Region),
    })
} else {
    provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
        AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
        SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
        Region:    pulumi.String(locals.Spec.Region),
        Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
    })
}
```

**Rationale**:
- **CI/CD Environments**: Kubernetes clusters with IRSA (IAM Roles for Service Accounts) or EC2 instances with instance profiles provide ambient credentials. Explicit credentials are unnecessary and create security risks (storing secrets).
- **Local Development**: Developers may use explicit credentials from AWS SSO, temporary credentials, or IAM users for local testing.

**Security Note**: The explicit credential path should be avoided in production. It exists primarily for local testing and legacy deployment patterns.

### 4. Tag Injection at Resource Level

**Decision**: Tags are constructed in `locals.go` and attached to the DynamoDB table resource (not passed as separate parameters).

**Rationale**:
- **Consistency**: Every OpenMCF resource has identical tagging structure
- **Immutability**: Tags reflect metadata from the original manifest—they are derived, not configured
- **Cost Allocation**: AWS Cost Explorer can aggregate costs by `organization`, `environment`, and `resource-kind` tags without per-resource configuration

### 5. Outputs Are Constant Strings, Not Computed

**Decision**: Output keys are predefined constants in `outputs.go`, not derived from user input.

**Rationale**:
- **Predictability**: Downstream systems (application code, CI/CD pipelines) can reference outputs by stable keys
- **Documentation**: The constants serve as documentation of what outputs are available
- **Pulumi Stack Compatibility**: Pulumi's stack output references require exact key matches

## DynamoDB-Specific Implementation Details

### Attribute Definitions: Keys Only

The module only includes attributes used in key schemas. This is a common source of confusion.

**Correct Understanding**: DynamoDB is schemaless. Items can have arbitrary attributes. The table schema only defines key attributes because:
1. Keys determine physical data distribution (partition key hash)
2. Keys enable range queries (sort key comparisons)
3. Keys must be declared at table creation (immutable decision)

**Implementation**:
```go
for _, a := range locals.Spec.AttributeDefinitions {
    // Only includes attributes referenced in:
    // - Table key_schema
    // - GSI key_schema
    // - LSI key_schema
}
```

**Common Mistake**: Users coming from relational databases expect to define all columns/attributes here. Attempting to include non-key attributes causes AWS API errors.

### Key Schema Element Extraction

The module extracts `HASH` and `RANGE` keys from the repeated `key_schema` field.

**Implementation Pattern**:
```go
var tableHashKey, tableRangeKey string
for _, k := range locals.Spec.KeySchema {
    if k.KeyType == awsdynamodbv1.AwsDynamodbSpec_KeySchemaElement_KEY_TYPE_HASH {
        tableHashKey = k.AttributeName
    }
    if k.KeyType == awsdynamodbv1.AwsDynamodbSpec_KeySchemaElement_KEY_TYPE_RANGE {
        tableRangeKey = k.AttributeName
    }
}
```

**Why Not Array Indexing?**: The protobuf repeated field doesn't guarantee ordering. Explicit type-based extraction is more robust than assuming `key_schema[0]` is always `HASH`.

**Validation Guarantee**: The protobuf validation rules (in `spec.proto`) enforce:
- Exactly one `HASH` key
- At most one `RANGE` key
- These validations occur at manifest submission, so by the time the Pulumi module runs, the schema is guaranteed to be valid

### Local Secondary Index Constraints

LSIs have unique constraints that affect the implementation:

**Hash Key Inheritance**: LSIs always use the same partition key as the table. The module only extracts the `RANGE` key from the LSI's `key_schema`:

```go
var lsiRangeKey string
for _, lk := range l.KeySchema {
    if lk.KeyType == awsdynamodbv1.AwsDynamodbSpec_KeySchemaElement_KEY_TYPE_RANGE {
        lsiRangeKey = lk.AttributeName
    }
}
```

**Why No Hash Key?**: Pulumi's `TableLocalSecondaryIndexArgs` only has a `RangeKey` field. The hash key is implicit (inherited from the table's hash key).

### Global Secondary Index Capacity Handling

GSIs have independent capacity in `PROVISIONED` mode but no capacity configuration in `PAY_PER_REQUEST` mode.

**Implementation**:
```go
if g.ProvisionedThroughput != nil && locals.Spec.BillingMode == awsdynamodbv1.AwsDynamodbSpec_BILLING_MODE_PROVISIONED {
    gsi.ReadCapacity = pulumi.IntPtr(int(g.ProvisionedThroughput.ReadCapacityUnits))
    gsi.WriteCapacity = pulumi.IntPtr(int(g.ProvisionedThroughput.WriteCapacityUnits))
}
```

**Critical Detail**: The capacity fields must be **omitted** (not set to zero) in `PAY_PER_REQUEST` mode. Setting them to zero causes AWS API errors. This is why the fields are `IntPtr` (nullable) rather than `Int`.

### Server-Side Encryption: Always Enabled

DynamoDB encryption at rest is **always enabled** and cannot be disabled. The only configuration is the key type.

**Implementation**:
```go
if locals.Spec.ServerSideEncryption != nil && locals.Spec.ServerSideEncryption.Enabled {
    sseArgs = &dynamodb.TableServerSideEncryptionArgs{
        Enabled:   pulumi.Bool(true),
        KmsKeyArn: pulumi.StringPtr(locals.Spec.ServerSideEncryption.KmsKeyArn),
    }
}
```

**Key Type Behavior**:
- If `kms_key_arn` is empty: AWS-owned key (default, no cost)
- If `kms_key_arn` is set: Customer-managed key (CMK) with KMS API costs

**Why the `Enabled` Field Exists**: Historical artifact from when encryption was optional. In modern AWS, the field is always true, but the API still requires it.

### Streams Configuration: Conditional View Type

If streams are enabled, a view type must be specified.

**Implementation**:
```go
if locals.Spec.StreamEnabled {
    switch locals.Spec.StreamViewType {
    case awsdynamodbv1.AwsDynamodbSpec_STREAM_VIEW_TYPE_KEYS_ONLY:
        args.StreamViewType = pulumi.StringPtr("KEYS_ONLY")
    // ... other cases
    }
}
```

**Why Conditional?**: If streams are disabled, setting `stream_view_type` causes an AWS API error. The view type is only valid when `stream_enabled=true`.

**Output Behavior**: The exported `stream_arn` is only populated if streams are enabled. Lambda event source mappings should check for a non-empty ARN before attempting to consume the stream.

## Error Handling Philosophy

The module follows a **fail-fast** approach:

1. **Validation at Protobuf Level**: The `spec.proto` validation rules catch configuration errors before Pulumi runs
2. **AWS API Errors Propagate**: If AWS rejects a configuration, the error propagates immediately (no silent fallbacks)
3. **No Default Mutations**: The module does not silently change user configuration (e.g., it won't auto-enable PITR if not specified)

**Rationale**: Infrastructure as code demands predictability. Silent defaults or error recovery can mask misconfiguration and create surprise behavior.

## Testing and Debugging

### Unit Testing (Not Present)

The module has no Go unit tests. This is intentional:
- **Integration Testing**: DynamoDB provisioning requires actual AWS API calls
- **Validation Testing**: The protobuf validation tests (in `spec_test.go` at the parent directory) verify configuration correctness

### Debugging with Delve

The `debug.sh` script enables step-through debugging:

1. Uncomment the binary option in `Pulumi.yaml`:
   ```yaml
   runtime:
     options:
       binary: ./debug.sh
   ```
2. Run Pulumi CLI commands normally
3. The debug script launches Delve, allowing breakpoints in any module file

**Use Case**: Debugging complex enum mappings or investigating AWS API errors.

## Common Pitfalls and Gotchas

### Pitfall 1: Defining Non-Key Attributes
**Symptom**: AWS API error: `One or more parameter values were invalid: Number of attributes in AttributeDefinitions does not match number of attributes in KeySchema`

**Cause**: Including attributes in `attribute_definitions` that aren't used in any key schema.

**Solution**: Only define attributes that are partition keys or sort keys for the table, GSIs, or LSIs.

### Pitfall 2: Zero Capacity in Provisioned Mode
**Symptom**: AWS API error about invalid capacity values.

**Cause**: Setting `read_capacity_units=0` or `write_capacity_units=0` in `PROVISIONED` mode.

**Solution**: Capacity values must be ≥1. The protobuf validation enforces this, but if validations are bypassed, the error manifests at AWS API level.

### Pitfall 3: GSI Capacity in On-Demand Mode
**Symptom**: AWS API error: `Provisioned throughput cannot be specified for on-demand billing mode`

**Cause**: Setting GSI `provisioned_throughput` when table billing mode is `PAY_PER_REQUEST`.

**Solution**: The module conditionally omits GSI capacity in on-demand mode. Ensure the stack input doesn't include throughput for GSIs when using on-demand billing.

### Pitfall 4: Missing Stream View Type
**Symptom**: AWS API error when `stream_enabled=true` but `stream_view_type` is unspecified.

**Cause**: Forgetting to specify what data should be included in stream records.

**Solution**: The protobuf validation enforces this: if `stream_enabled=true`, `stream_view_type` must be set.

### Pitfall 5: LSI Created After Table Creation
**Symptom**: Cannot add LSI to existing table.

**Cause**: LSIs are immutable—they can only be created at table creation time.

**Solution**: Design LSIs carefully before initial deployment. If an LSI is needed later, the only path is data migration to a new table.

## Performance Considerations

### Pulumi Resource Creation Time

DynamoDB table creation is typically fast (10-30 seconds for empty tables). However:
- **Large GSI creation**: Creating a GSI on an existing table with millions of items is a long-running operation (minutes to hours)
- **LSI at creation**: Adding LSIs during initial table creation has negligible overhead

### Pulumi Diff Calculation

The module uses Pulumi's automatic diff calculation. Changes to most fields trigger table updates:
- **Immutable Fields**: Changing `hash_key`, `range_key`, or LSI definitions requires table replacement (destroys and recreates)
- **Mutable Fields**: Changing GSIs, billing mode, PITR, or encryption triggers in-place updates

**Recommendation**: Use `pulumi preview` before `pulumi up` to understand whether changes will cause table replacement.

## Future Enhancements

Potential improvements that are not currently implemented:

### Application Auto Scaling
**What**: Automatic capacity scaling for provisioned billing mode.

**Why Not Implemented**: Adds significant complexity. On-demand billing is simpler and now cost-competitive for most workloads.

**Path Forward**: If provisioned+autoscaling is needed, the module could be extended to create:
- `aws.ApplicationAutoScaling.Target` for table and each GSI
- `aws.ApplicationAutoScaling.Policy` for scale-up/scale-down rules

### CloudWatch Alarms
**What**: Automatic alarms for throttled requests.

**Why Not Implemented**: Alarm configuration is opinionated (thresholds, SNS topics). Better managed as separate alarm resources.

**Path Forward**: Separate Pulumi module or Terraform configuration that references the table ARN output.

### Contributor Insights
**What**: Automated hot key detection.

**Why Not Implemented**: It's a per-table feature with additional cost, not universally needed.

**Path Forward**: Add `contributor_insights_enabled` field to spec if demand exists.

## Conclusion

This Pulumi module demonstrates OpenMCF's philosophy: **expose the full power of the underlying cloud service while providing guardrails through validation**.

The architecture is deliberately explicit:
- No hidden defaults or magic mappings
- Clear separation of concerns (locals, provider, resource creation, outputs)
- Validation at the protobuf layer prevents misconfiguration
- Resource creation follows DynamoDB's API semantics closely

For teams familiar with Terraform or CloudFormation, the mapping should feel natural. For teams new to DynamoDB, the explicit field names and comprehensive research documentation (`docs/README.md`) provide the context needed to make informed design decisions.

The result is a production-ready module that can provision simple prototype tables (partition key + on-demand billing) or complex production tables (partition+sort keys, multiple GSIs, provisioned capacity, CMK encryption, streams, PITR) with equal ease.

