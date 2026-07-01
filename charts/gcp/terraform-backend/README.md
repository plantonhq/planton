# GCP Terraform State Backend

Provisions a Google Cloud Storage (GCS) bucket configured as a Terraform remote state backend with production-ready defaults: object versioning for state recovery, uniform bucket-level access for simplified IAM, and enforced public access prevention.

Unlike the [AWS Terraform State Backend](../../aws/terraform-backend/) which requires both an S3 bucket and a DynamoDB table for state locking, the GCP backend needs only a GCS bucket — GCS provides native object locking through its [state locking mechanism](https://developer.hashicorp.com/terraform/language/settings/backends/gcs).

## Architecture

```
┌──────────────────────────────────┐
│         GcpGcsBucket             │
│  ┌────────────────────────────┐  │
│  │  Versioning: enabled       │  │
│  │  Access: uniform IAM       │  │
│  │  Public: enforced prevent  │  │
│  │  Lifecycle: 30-version     │  │
│  │    noncurrent cleanup      │  │
│  └────────────────────────────┘  │
└──────────────────────────────────┘
```

## Included Cloud Resources

| Resource | Kind | Purpose |
|----------|------|---------|
| GCS Bucket | `GcpGcsBucket` | Stores Terraform state files with versioning and locking |

## Parameters

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `gcp_project_id` | GCP project ID where the bucket will be created | `my-gcp-project` | Yes |
| `bucket_name` | Globally unique name for the state bucket | `my-org-terraform-state` | Yes |
| `location` | Bucket location (region, dual-region, or multi-region) | `US` | Yes |

## Bucket Configuration

The chart applies the following production-ready defaults:

- **Versioning**: Enabled — recover from state corruption by restoring previous versions
- **Uniform bucket-level access**: Enabled — all access controlled via IAM policies (no object ACLs)
- **Public access prevention**: Enforced — state files can never be accidentally exposed publicly
- **Lifecycle rule**: Noncurrent versions are deleted after 30 newer versions exist, preventing unbounded storage growth from versioning

## Usage

### Configure Terraform Backend

After deploying this chart, configure your Terraform backend:

```hcl
terraform {
  backend "gcs" {
    bucket = "my-org-terraform-state"
    prefix = "terraform/state"
  }
}
```

### Naming Convention

GCS bucket names are **globally unique** across all GCP projects. Recommended conventions:

- `<org>-<project>-terraform-state` (e.g., `acme-platform-terraform-state`)
- `<org>-terraform-state-<region>` (e.g., `acme-terraform-state-us`)

### Adding CMEK Encryption

For organizations requiring Customer-Managed Encryption Keys, deploy a [GcpKmsKeyRing](https://github.com/plantonhq/planton) and [GcpKmsKey](https://github.com/plantonhq/planton) separately, then add the encryption configuration to the bucket resource directly.

## Important Notes

- This chart provisions the bucket only. Subsequent configuration changes should be made directly to the Cloud Resource.
- Ensure the GCP project has the Cloud Storage API enabled before deploying.
- The lifecycle rule retains the 30 most recent versions of each state file. Adjust the `numNewerVersions` value in the template if your organization requires a different retention policy.
