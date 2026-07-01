# GCP Pulumi State Backend

Provisions a Google Cloud Storage (GCS) bucket configured as a Pulumi self-managed state backend with production-ready defaults: object versioning for state recovery, uniform bucket-level access for simplified IAM, and enforced public access prevention.

This chart is the GCP equivalent of the [AWS Pulumi State Backend](../../aws/pulumi-backend/). While Pulumi Cloud is the recommended state backend for most teams, organizations that require self-hosted state storage can use this chart to provision a GCS bucket with appropriate security and lifecycle settings.

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
| GCS Bucket | `GcpGcsBucket` | Stores Pulumi stack state files with versioning |

## Parameters

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `gcp_project_id` | GCP project ID where the bucket will be created | `my-gcp-project` | Yes |
| `bucket_name` | Globally unique name for the state bucket | `my-org-pulumi-state` | Yes |
| `location` | Bucket location (region, dual-region, or multi-region) | `US` | Yes |

## Bucket Configuration

The chart applies the following production-ready defaults:

- **Versioning**: Enabled — recover from state corruption by restoring previous versions
- **Uniform bucket-level access**: Enabled — all access controlled via IAM policies (no object ACLs)
- **Public access prevention**: Enforced — state files can never be accidentally exposed publicly
- **Lifecycle rule**: Noncurrent versions are deleted after 30 newer versions exist, preventing unbounded storage growth from versioning

## Usage

### Configure Pulumi Backend

After deploying this chart, log in to the GCS backend:

```bash
pulumi login gs://my-org-pulumi-state
```

Then initialize or select a stack as usual:

```bash
pulumi stack init dev
pulumi up
```

### Naming Convention

GCS bucket names are **globally unique** across all GCP projects. Recommended conventions:

- `<org>-<project>-pulumi-state` (e.g., `acme-platform-pulumi-state`)
- `<org>-pulumi-state-<region>` (e.g., `acme-pulumi-state-us`)

### Adding CMEK Encryption

For organizations requiring Customer-Managed Encryption Keys, deploy a [GcpKmsKeyRing](https://github.com/plantonhq/planton) and [GcpKmsKey](https://github.com/plantonhq/planton) separately, then add the encryption configuration to the bucket resource directly.

## Important Notes

- This chart provisions the bucket only. Subsequent configuration changes should be made directly to the Cloud Resource.
- Ensure the GCP project has the Cloud Storage API enabled before deploying.
- The lifecycle rule retains the 30 most recent versions of each state file. Adjust the `numNewerVersions` value in the template if your organization requires a different retention policy.
- For team environments, ensure all members have `roles/storage.objectAdmin` on the bucket for read/write state access.
