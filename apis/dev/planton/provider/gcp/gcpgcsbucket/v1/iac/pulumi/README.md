# GCP Cloud Storage Bucket Pulumi Module

## Overview

This Pulumi module provides automated deployment and management of Google Cloud Storage (GCS) buckets with production-grade configurations. It implements the 80/20 principle, focusing on essential configurations for secure, cost-effective, and compliant storage infrastructure.

## Key Features

### API Resource Features

- **Standardized Structure**: The `GcpGcsBucket` API resource adheres to a consistent schema with `apiVersion`, `kind`, `metadata`, `spec`, and `status` fields, ensuring compatibility and ease of integration within Kubernetes-like environments.

- **Configurable Specifications**:
  - **Location Strategy**: Support for regional, dual-region, and multi-region bucket placement with immutable location configuration.
  - **Access Control**: Uniform Bucket-Level Access (UBLA) enabled by default for simplified security model with explicit IAM bindings.
  - **Storage Classes**: Support for STANDARD, NEARLINE, COLDLINE, and ARCHIVE storage classes with automated lifecycle transitions.
  - **Data Protection**: Object versioning with lifecycle policies for cost optimization and retention policies for compliance.
  - **Encryption**: Configurable encryption with support for Google-managed keys or Customer-Managed Encryption Keys (CMEK) via Cloud KMS.
  - **Advanced Features**: CORS configuration, static website hosting, requester pays, and access logging.

- **Security by Default**: 
  - UBLA enabled by default to prevent accidental public access via object ACLs
  - Explicit IAM bindings required instead of boolean "public" flags
  - Support for public access prevention policy
  - Auditable access control configurations

- **Cost Optimization**: 
  - Lifecycle rules for automatic object deletion or storage class transitions
  - Support for versioning with automated cleanup of old versions
  - Requester pays mode for public dataset distribution

- **Compliance Features**:
  - Retention policies for WORM (Write Once, Read Many) compliance requirements
  - CMEK support for regulatory requirements
  - Comprehensive labeling for cost tracking and governance

### Pulumi Module Features

- **Automated GCP Provider Setup**: Leverages the provided GCP credentials to automatically configure the Pulumi GCP provider, enabling seamless and secure interactions with GCS resources.

- **Comprehensive Bucket Management**: Implements all essential GCS features:
  - Bucket creation with location, storage class, and access control
  - IAM policy management with support for conditional access
  - Lifecycle policy configuration for automated object management
  - Encryption configuration with CMEK support
  - CORS rules for cross-origin browser access
  - Website configuration for static hosting
  - Retention policies for compliance
  - Access logging configuration

- **Label Management**: Automatically generates standard GCP labels for resource tracking:
  - Resource identification labels (name, kind, ID)
  - Organizational labels (org, environment)
  - User-provided custom labels merged with generated labels

- **Exported Stack Outputs**: Captures essential outputs in `status.outputs`:
  - `bucket_id`: The unique identifier of the created bucket for reference in other resources

## Directory Structure

```
iac/pulumi/
├── main.go              # Pulumi program entrypoint
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build and deployment commands
├── debug.sh             # Local debugging script
├── README.md            # This file
├── overview.md          # Architecture and design details
└── module/
    ├── main.go          # Resources function and bucket creation logic
    ├── locals.go        # Local variables and label initialization
    └── outputs.go       # Output constant definitions
```

## Prerequisites

### Required Tools

- **Pulumi CLI**: v3.0+
- **Go**: 1.21+
- **gcloud CLI**: For GCP authentication (if using local credentials)

### GCP Prerequisites

- GCP project with Cloud Storage API enabled
- GCP credentials with appropriate permissions:
  - `storage.buckets.create`
  - `storage.buckets.update`
  - `storage.buckets.setIamPolicy`
  - `cloudkms.cryptoKeys.getIamPolicy` (if using CMEK)

## Module Components

### main.go (Entry Point)

The entry point loads the stack input and calls the module's `Resources` function:

```go
func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        stackInput := &gcpgcsbucketv1.GcpGcsBucketStackInput{}
        if err := stackinput.Load(stackInput); err != nil {
            return err
        }
        return module.Resources(ctx, stackInput)
    })
}
```

### module/main.go (Core Logic)

Implements comprehensive bucket creation and configuration:

- **Bucket Creation**: Creates the GCS bucket with all specified configurations
- **Storage Class Management**: Configures storage class with support for all GCS tiers
- **Versioning**: Enables object versioning when specified
- **Lifecycle Rules**: Implements automated object lifecycle management
- **IAM Bindings**: Configures explicit IAM policy bindings with optional conditional access
- **Encryption**: Configures CMEK when specified
- **CORS Rules**: Sets up cross-origin resource sharing for browser access
- **Website Configuration**: Enables static website hosting
- **Retention Policy**: Implements WORM compliance requirements
- **Additional Features**: Requester pays, access logging, public access prevention

### module/locals.go (Local Variables)

Initializes local variables including:
- Standard GCP labels (resource tracking)
- Organizational labels (org, environment)
- User-provided labels from spec
- Bucket specification reference

### module/outputs.go (Output Constants)

Defines output constants:
- `bucket_id`: The unique identifier of the created bucket

## Configuration Tiers

The module implements a tiered configuration approach based on the 80/20 principle:

### Tier 1: Essential Configuration (Required)

- `gcp_project_id`: The GCP project for bucket creation
- `location`: Regional, dual-region, or multi-region location (immutable)
- `uniform_bucket_level_access_enabled`: Security model selection (defaults to true)

### Tier 2: Common Configuration (Recommended)

- `storage_class`: STANDARD, NEARLINE, COLDLINE, or ARCHIVE
- `versioning_enabled`: Data protection through versioning
- `lifecycle_rules`: Cost optimization through automated management
- `iam_bindings`: Explicit access control
- `encryption`: CMEK configuration for compliance
- `gcp_labels`: Custom labels for cost tracking

### Tier 3: Advanced Configuration (Optional)

- `cors_rules`: Cross-origin browser access
- `website`: Static website hosting
- `retention_policy`: WORM compliance
- `requester_pays`: Public dataset distribution
- `logging`: Legacy access logs
- `public_access_prevention`: Organization-level policy enforcement

## Access Control Model

The module enforces secure access control patterns:

### Uniform Bucket-Level Access (Recommended)

When `uniform_bucket_level_access_enabled` is `true` (default):
- All access controlled via IAM policies
- No object-level ACLs (simplified security model)
- Required for modern GCS features
- Prevents "public access trap"

Example IAM binding for public read access:

```yaml
iam_bindings:
  - role: "roles/storage.objectViewer"
    members:
      - "allUsers"
```

### IAM Binding Types

The module uses `BucketIAMBinding` resources, which are authoritative for a specific role:
- Grants specified role to specified members
- Overwrites any existing bindings for that role
- Supports conditional access via IAM conditions

## Lifecycle Management

The module supports comprehensive lifecycle rules for cost optimization:

### Common Patterns

**Delete old versions after 30 days:**

```yaml
lifecycle_rules:
  - action:
      type: "Delete"
    condition:
      num_newer_versions: 5
```

**Transition to COLDLINE after 90 days:**

```yaml
lifecycle_rules:
  - action:
      type: "SetStorageClass"
      storage_class: COLDLINE
    condition:
      age_days: 90
```

**Delete objects after 7 years (compliance):**

```yaml
lifecycle_rules:
  - action:
      type: "Delete"
    condition:
      age_days: 2555  # 7 years
```

## Encryption

The module supports two encryption models:

### Google-Managed Encryption (Default)

No configuration required - all data encrypted at rest automatically.

### Customer-Managed Encryption Keys (CMEK)

For compliance requirements:

```yaml
encryption:
  kms_key_name: "projects/PROJECT_ID/locations/LOCATION/keyRings/KEY_RING/cryptoKeys/KEY"
```

## Outputs

After successful deployment, the following outputs are available:

- **bucket_id**: The unique identifier of the created bucket
  - Format: `projects/PROJECT_ID/buckets/BUCKET_NAME`
  - Used for referencing the bucket in other resources

## Best Practices

### Security

1. **Enable UBLA**: Always set `uniform_bucket_level_access_enabled: true`
2. **Explicit IAM**: Use explicit IAM bindings instead of legacy ACLs
3. **CMEK for Compliance**: Use customer-managed encryption for sensitive data
4. **Public Access Prevention**: Set to "enforced" for private buckets

### Cost Optimization

1. **Lifecycle Policies**: Always configure lifecycle rules for cost control
2. **Storage Classes**: Use STANDARD by default, transition with lifecycle rules
3. **Versioning Cleanup**: Delete old versions automatically
4. **Regional Placement**: Co-locate with compute resources

### Reliability

1. **Versioning**: Enable for important data protection
2. **Regional Strategy**: Use regional buckets for latency-sensitive workloads
3. **Dual-Region**: For HA requirements with regional failover
4. **Multi-Region**: For global content delivery

## Troubleshooting

### Common Issues

**Bucket name already exists:**
- Bucket names are globally unique across all GCP projects
- Choose a unique name or use the project ID as a prefix

**CMEK permission errors:**
- Ensure the GCS service account has `cloudkms.cryptoKeyEncrypterDecrypter` role on the KMS key
- Service account format: `service-PROJECT_NUMBER@gs-project-accounts.iam.gserviceaccount.com`

**Public access not working with UBLA:**
- Cannot use `allUsers` IAM binding with UBLA disabled
- Enable UBLA and use IAM bindings for public access

**Lifecycle rules not triggering:**
- Rules process once per day (may take up to 24 hours)
- Use `age_days` for age-based conditions, not `created_before`

## Further Reading

- [Component README](../../README.md) - User-facing documentation
- [Architecture Overview](overview.md) - Design decisions and patterns
- [Research Document](../../docs/README.md) - Comprehensive analysis of GCS deployment methods
- [Examples](../../examples.md) - Common configuration patterns

## Support

For issues or questions:
1. Check the [examples.md](../../examples.md) for common patterns
2. Review the [research document](../../docs/README.md) for design rationale
3. Consult the [audit report](../../docs/audit/) for component completeness


