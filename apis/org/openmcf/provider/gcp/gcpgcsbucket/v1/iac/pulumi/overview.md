# GCP Cloud Storage Bucket Pulumi Module - Architecture Overview

## Overview

The **GCP Cloud Storage Bucket Pulumi Module** is engineered to streamline the deployment and management of production-grade Google Cloud Storage buckets within a unified infrastructure framework. Leveraging OpenMCF's API-driven approach, this module models each storage bucket using a Kubernetes-like structure with `apiVersion`, `kind`, `metadata`, `spec`, and `status` fields. The `GcpGcsBucket` resource encapsulates the necessary specifications for provisioning secure, cost-optimized, and compliant storage infrastructure, enabling platform teams to manage cloud storage as code with consistency across environments.

By utilizing this Pulumi module, developers can automate the creation and configuration of GCS buckets with comprehensive feature support including access control, lifecycle management, encryption, versioning, and compliance policies. The module seamlessly integrates with GCP credentials provided in the resource definition, ensuring secure and authenticated interactions with GCP Storage services. Furthermore, the outputs generated from the deployment are captured in the resource's `status.outputs`, facilitating effective integration with other infrastructure components and enabling robust monitoring and management capabilities.

## Design Philosophy

### Security-First Approach

The module implements security best practices by default:

1. **Uniform Bucket-Level Access (UBLA)**: Enabled by default to enforce IAM-only access control, eliminating the complexity and security risks of dual permission systems (IAM + ACLs).

2. **Explicit IAM Bindings**: Replaces the dangerous "is_public" boolean flag pattern with explicit IAM policy bindings, making access control auditable and preventing accidental public exposure.

3. **Public Access Prevention**: Supports organization-level policies to prevent making buckets public, even with explicit IAM bindings.

4. **CMEK Support**: Enables customer-managed encryption keys for regulatory compliance requirements, with full control over key lifecycle.

### 80/20 Configuration Model

The module follows a tiered configuration approach:

- **Tier 1 (Essential)**: Required fields that must be configured for any bucket (project, location, access model)
- **Tier 2 (Common)**: Fields used in most production deployments (storage class, versioning, lifecycle, IAM, encryption)
- **Tier 3 (Advanced)**: Specialized features for specific use cases (CORS, website hosting, retention policies, requester pays)

This tiering ensures simple deployments remain simple while enabling complex configurations when needed.

### Cost Optimization

The module emphasizes cost control through:

1. **Lifecycle Policies**: First-class support for automated object deletion and storage class transitions
2. **Storage Class Management**: Explicit configuration of storage classes with transitions based on object age
3. **Versioning Cleanup**: Automated deletion of old object versions to prevent unbounded storage growth
4. **Regional Placement**: Encourages co-location with compute resources to minimize egress charges

## Architecture

### Module Structure

```
module/
├── main.go      # Core resource creation logic
├── locals.go    # Variable initialization and label management
└── outputs.go   # Output constant definitions
```

### Resource Dependencies

The module creates and manages the following GCP resources:

1. **Primary Resource**: `google.storage.Bucket`
   - Base bucket with location, storage class, and access model
   - Versioning, lifecycle rules, encryption configuration
   - CORS, website, retention policy, logging configuration

2. **Access Control**: `google.storage.BucketIAMBinding` (0..n)
   - One resource per IAM role binding
   - Authoritative for the specified role
   - Supports conditional access via IAM conditions

### Dependency Graph

```
GcpGcsBucket (API Resource)
    │
    ├─> Pulumi GCP Provider (configured from credentials)
    │       │
    │       └─> google.storage.Bucket
    │               │
    │               ├─> Location (immutable)
    │               ├─> Storage Class
    │               ├─> Uniform Bucket Level Access
    │               ├─> Versioning
    │               ├─> Lifecycle Rules
    │               ├─> Encryption (optional CMEK dependency)
    │               ├─> CORS Rules
    │               ├─> Website Configuration
    │               ├─> Retention Policy
    │               ├─> Logging Configuration
    │               └─> Public Access Prevention
    │
    └─> google.storage.BucketIAMBinding (for each iam_bindings entry)
            ├─> Bucket (parent dependency)
            ├─> Role
            ├─> Members
            └─> Condition (optional)
```

## Key Implementation Details

### Access Control Model

The module exclusively uses `BucketIAMBinding` resources instead of legacy `BucketAccessControl` (ACLs):

**Why IAM Bindings?**
- Authoritative for a single role (prevents permission sprawl)
- Supports conditional access via IAM Conditions
- Required when UBLA is enabled (the recommended default)
- Provides clear audit trail of permissions
- Integrates with organization-wide IAM policies

**IAM Binding Behavior:**
- Each binding manages one role completely
- If a role appears in multiple bindings, the last one wins (Pulumi prevents this)
- Members array is atomic (replaces all members for that role)
- Conditions are evaluated at access time for dynamic permissions

### Lifecycle Rule Implementation

The module implements lifecycle rules with full support for:

**Actions:**
- `Delete`: Permanently delete objects meeting conditions
- `SetStorageClass`: Transition to cheaper storage class

**Conditions:**
- `age_days`: Days since object creation
- `created_before`: RFC 3339 date for absolute cutoff
- `is_live`: Target current vs noncurrent versions
- `num_newer_versions`: For version-based cleanup
- `matches_storage_class`: Apply only to specific storage classes

**Common Patterns:**

1. **Version Cleanup**: Delete noncurrent versions after 30 days
2. **Storage Tiering**: STANDARD → NEARLINE (30d) → COLDLINE (90d) → ARCHIVE (365d)
3. **Compliance Deletion**: Delete all objects after 7 years (2,555 days)
4. **Development Cleanup**: Delete everything older than 30 days in dev buckets

### Encryption Configuration

The module supports two encryption models:

**Google-Managed (Default):**
- No configuration required
- Keys automatically managed by Google
- Transparent encryption/decryption
- Suitable for most workloads

**Customer-Managed (CMEK):**
- Requires Cloud KMS key ARN
- Provides audit trail of key usage
- Enables key rotation control
- Required for certain compliance frameworks
- Module does NOT create the KMS key (must exist beforehand)

**Key Points:**
- All data is encrypted at rest (always)
- CMEK adds management overhead but provides control
- KMS key must be in same location as bucket (or multi-region)
- GCS service account needs `cloudkms.cryptoKeyEncrypterDecrypter` role

### Label Management

The module implements a two-tier label system:

**Standard Labels (Auto-Generated):**
```
planton-resource: "true"
planton-resource-name: <bucket-name>
planton-resource-kind: "gcpgcsbucket"
planton-resource-id: <ulid>        # if set
planton-resource-org: <org>        # if set
planton-resource-env: <env>        # if set
```

**User Labels (From Spec):**
- Merged with standard labels
- User labels override standard labels on conflict
- Enables custom cost allocation and filtering
- Supports compliance tagging requirements

### CORS Configuration

The module supports CORS for buckets serving content to browsers:

**Use Cases:**
- Web applications accessing bucket from different origin
- Font hosting for external websites
- API responses with bucket-based content

**Configuration:**
- Multiple CORS rules supported
- Each rule specifies allowed origins, methods, headers
- MaxAgeSeconds controls preflight cache duration

**Important:** For production websites, prefer Cloud CDN + Load Balancer instead of direct bucket access.

### Website Configuration

The module supports static website hosting:

**Features:**
- Index page suffix (e.g., `index.html`)
- Custom 404 page
- Directory index behavior

**Limitations:**
- No HTTPS support (use Load Balancer for HTTPS)
- No custom domain support (use Load Balancer)
- No access logging granularity (use Load Balancer)

**Recommendation:** Only use for internal/dev sites. For production, use the "Private Bucket + Cloud CDN" pattern.

### Retention Policies

The module implements WORM (Write Once, Read Many) compliance:

**Features:**
- Minimum retention period in seconds
- Optional locking (irreversible operation)
- Applies to all objects in bucket

**Use Cases:**
- Financial records (FINRA: 7 years)
- Healthcare records (HIPAA: 6 years)
- Legal holds and e-discovery
- Regulatory compliance requirements

**Critical:** Locking a retention policy is permanent. Objects cannot be deleted even by project owners until retention expires.

## Integration Patterns

### Pattern 1: Private Bucket with Application Access

**Scenario:** Backend application storing/retrieving data

```yaml
spec:
  location: "us-east1"
  uniform_bucket_level_access_enabled: true
  versioning_enabled: true
  iam_bindings:
    - role: "roles/storage.objectAdmin"
      members:
        - "serviceAccount:app@project.iam.gserviceaccount.com"
```

### Pattern 2: Public Read Bucket for Static Assets

**Scenario:** Public website assets, open datasets

```yaml
spec:
  location: "us-east1"
  uniform_bucket_level_access_enabled: true
  iam_bindings:
    - role: "roles/storage.objectViewer"
      members:
        - "allUsers"
```

### Pattern 3: Compliant Archive Bucket

**Scenario:** Long-term retention with lifecycle management

```yaml
spec:
  location: "us-east1"
  storage_class: ARCHIVE
  versioning_enabled: true
  retention_policy:
    retention_period_seconds: 220752000  # 7 years
    is_locked: false  # Lock after initial testing
  lifecycle_rules:
    - action:
        type: "Delete"
      condition:
        age_days: 2555  # 7 years + buffer
```

### Pattern 4: Cost-Optimized Bucket

**Scenario:** Development data with aggressive cleanup

```yaml
spec:
  location: "us-east1"
  storage_class: STANDARD
  versioning_enabled: true
  lifecycle_rules:
    - action:
        type: "Delete"
      condition:
        age_days: 30
    - action:
        type: "Delete"
      condition:
        num_newer_versions: 3
```

## Comparison to Legacy Implementation

### Old Implementation (Anti-Patterns)

```go
// Used boolean flag (unclear what it does)
IsPublic: true

// Used legacy ACLs (incompatible with UBLA)
BucketAccessControl with "allUsers"

// Disabled UBLA (security risk)
UniformBucketLevelAccess: false

// Used "GcpRegion" (incorrect naming)
Location: spec.GcpRegion
```

### New Implementation (Best Practices)

```go
// Explicit IAM binding (clear and auditable)
iam_bindings:
  - role: "roles/storage.objectViewer"
    members: ["allUsers"]

// UBLA enabled (secure by default)
uniform_bucket_level_access_enabled: true

// Correct field name
location: "us-east1"

// Comprehensive configuration support
- storage_class, versioning, lifecycle_rules
- encryption, cors_rules, website
- retention_policy, requester_pays, logging
```

## Error Handling and Edge Cases

### Bucket Name Conflicts

**Issue:** Bucket names are globally unique across all GCP projects

**Handling:**
- Module does not auto-generate names
- User must ensure uniqueness (e.g., prefix with project ID)
- Pulumi will error on conflict with clear message

### CMEK Key Permissions

**Issue:** GCS service account lacks `cryptoKeyEncrypterDecrypter` role

**Handling:**
- Module does not manage KMS permissions (separation of concerns)
- User must grant permissions before bucket creation
- Clear error message references service account format

### IAM Condition Errors

**Issue:** Invalid IAM condition expression

**Handling:**
- Pulumi validates condition syntax
- Error includes expression and validation failure
- Module passes through condition as-is (no validation)

### Lifecycle Rule Conflicts

**Issue:** Multiple rules could apply to same object

**Handling:**
- GCS evaluates all matching rules
- If multiple Delete actions match, object deleted (action is idempotent)
- If multiple SetStorageClass actions match, cheapest class wins
- Module does not validate rule conflicts (deferred to GCS)

## Performance Considerations

### Resource Creation Time

- **Bucket Creation**: 2-5 seconds
- **IAM Bindings**: 1-2 seconds per binding
- **Total Deployment**: Typically < 15 seconds for bucket with multiple bindings

### State Management

- Pulumi tracks all resources in state file
- IAM bindings tracked separately (enable incremental updates)
- Bucket configuration changes force update (not recreation)
- Location changes force bucket recreation (destructive operation)

### Drift Detection

- Pulumi detects manual changes via refresh
- IAM binding drift detected per binding
- Lifecycle rules compared as arrays (order matters)
- Labels compared as maps (automatic merge)

## Testing Strategy

### Unit Tests

Located in `spec_test.go`, validates:
- Required field presence (gcp_project_id, location)
- Field format validation (project ID pattern)
- Enum value validation (storage classes)

### Integration Tests

Manual testing via `debug.sh`:
- Creates real GCS bucket in test project
- Validates IAM bindings applied correctly
- Tests lifecycle rule behavior
- Verifies encryption configuration

### Validation

The module relies on:
1. **Proto validation**: `buf.validate` rules enforce constraints
2. **Pulumi validation**: GCP provider validates API parameters
3. **GCP validation**: Cloud Storage API validates final configuration

## Future Enhancements

### Planned Features

1. **Hierarchical Namespace**: Support for folder-like structure (Preview feature)
2. **Object Lifecycle Management**: Direct object creation/management
3. **Soft Delete**: Automatic recovery window for deleted objects
4. **Autoclass**: Automatic storage class optimization based on access patterns

### Not Planned

1. **ACL Management**: Incompatible with UBLA (the recommended model)
2. **Bucket Creation Retry**: Deferred to Pulumi's built-in retry logic
3. **KMS Key Creation**: Separation of concerns (use dedicated KMS module)

## References

- [GCS Bucket Resource (Pulumi)](https://www.pulumi.com/registry/packages/gcp/api-docs/storage/bucket/)
- [GCS IAM Binding (Pulumi)](https://www.pulumi.com/registry/packages/gcp/api-docs/storage/bucketiambinding/)
- [GCS Best Practices (Google Cloud)](https://cloud.google.com/storage/docs/best-practices)
- [Uniform Bucket-Level Access](https://cloud.google.com/storage/docs/uniform-bucket-level-access)
- [Object Lifecycle Management](https://cloud.google.com/storage/docs/lifecycle)
- [Customer-Managed Encryption Keys](https://cloud.google.com/storage/docs/encryption/customer-managed-keys)


