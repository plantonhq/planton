---
title: "Private Versioned"
description: "This preset creates a private Object Storage bucket with versioning enabled, KMS encryption, auto-tiering for cost optimization, and lifecycle rules to archive old object versions and clean up..."
type: "preset"
rank: "01"
presetSlug: "01-private-versioned"
componentSlug: "object-storage-bucket"
componentTitle: "Object Storage Bucket"
provider: "oci"
icon: "package"
order: 1
---

# Private Versioned

This preset creates a private Object Storage bucket with versioning enabled, KMS encryption, auto-tiering for cost optimization, and lifecycle rules to archive old object versions and clean up incomplete multipart uploads. This is the standard configuration for application data where objects need overwrite protection and cost-efficient long-term storage.

## When to Use

- Application data buckets (uploads, documents, media files) requiring overwrite and deletion protection
- Data lake storage where versioning provides an audit trail of data changes
- Terraform state backends where versioning protects against accidental state corruption
- Any production bucket where accidental deletions must be recoverable

## Key Configuration Choices

- **Private access** (`accessType: no_public_access`) -- no anonymous internet access. Objects are accessible only via authenticated OCI API calls, IAM policies, or pre-authenticated requests.
- **Versioning enabled** (`versioning: enabled`) -- every overwrite and delete creates a new version. Previous versions are retained until explicitly purged or managed by lifecycle rules.
- **KMS encryption** (`kmsKeyId`) -- customer-managed encryption provides key rotation control and meets compliance requirements. Use Oracle-managed keys (omit `kmsKeyId`) for simpler setups.
- **Auto-tiering to InfrequentAccess** (`autoTiering: infrequent_access`) -- OCI automatically moves objects that have not been accessed for 30+ days to the InfrequentAccess tier, reducing storage costs without changing access patterns.
- **Archive old versions after 90 days** -- previous object versions are moved to Archive tier after 90 days, significantly reducing storage costs for version history while maintaining compliance.
- **Abort incomplete uploads after 7 days** -- cleans up orphaned multipart upload parts that failed to complete, preventing wasted storage charges.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment for the bucket | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<object-storage-namespace>` | Object Storage namespace for your tenancy | `oci os ns get` CLI command, or OCI Console > Object Storage |
| `<bucket-name>` | Globally unique bucket name within the namespace | Choose a name (e.g., `myapp-prod-data`) |
| `<kms-key-ocid>` | OCID of the KMS encryption key | OCI Console > Security > Vault > Keys, or `OciKmsKey` outputs |

## Related Presets

- **02-archive-storage** -- Use instead for long-term retention data that is rarely accessed (backups, compliance archives)
- **03-public-read** -- Use instead for static assets served directly to end users
