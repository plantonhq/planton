---
title: "Deploy Across Providers"
description: "Deploy object storage on both AWS and GCP using OpenMCF — same workflow, same CLI, different providers"
icon: "tutorial"
order: 50
---

# Deploy Across Providers

In this tutorial, you will deploy an object storage bucket on both AWS (S3) and GCP (GCS) using OpenMCF. The purpose is not to compare cloud providers — it is to demonstrate the pattern that makes OpenMCF useful: the same KRM manifest structure, the same CLI commands, and the same deployment workflow across every provider. Only the `spec` changes.

By the end, you will see how OpenMCF's consistent interface reduces the cognitive overhead of working across multiple cloud providers.

## What You Will Build

Two object storage buckets:

| | AWS S3 | GCP GCS |
|---|--------|---------|
| apiVersion | `aws.openmcf.org/v1` | `gcp.openmcf.org/v1` |
| kind | `AwsS3Bucket` | `GcpGcsBucket` |
| Versioning | Enabled | Enabled |
| Encryption | SSE-S3 (AES-256) | Google-managed (default) |
| Deploy command | `openmcf apply -f aws-bucket.yaml` | `openmcf apply -f gcp-bucket.yaml` |

## Prerequisites

Before starting, ensure you have:

- **OpenMCF CLI** installed (`openmcf version`). See [Getting Started](../getting-started) for installation.
- **AWS credentials** configured. See [AWS Provider Setup](../guides/aws-provider-setup).
- **GCP credentials** configured. See [GCP Provider Setup](../guides/gcp-provider-setup).
- **Pulumi CLI** installed with a backend configured.

If you have only one provider configured, you can still follow along and deploy to just that provider. The tutorial is structured so each deployment is independent.

## Step 1: Write the AWS Manifest

Create a file named `aws-bucket.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3Bucket
metadata:
  name: multi-provider-demo-aws
  labels:
    openmcf.org/provisioner: pulumi
spec:
  awsRegion: us-east-1
  versioningEnabled: true
  encryptionType: ENCRYPTION_TYPE_SSE_S3
  tags:
    project: multi-provider-tutorial
    managed-by: openmcf
```

The `AwsS3Bucket` spec requires only `awsRegion`. Versioning, encryption, and tags are optional but represent production best practices. See [Deploy Your First AWS Resource](./first-aws-resource) for a deeper walkthrough of S3 configuration.

## Step 2: Write the GCP Manifest

Create a file named `gcp-bucket.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGcsBucket
metadata:
  name: multi-provider-demo-gcp
  labels:
    openmcf.org/provisioner: pulumi
spec:
  gcpProjectId:
    value: my-gcp-project
  location: US
  bucketName: openmcf-multi-provider-demo
  versioningEnabled: true
  uniformBucketLevelAccessEnabled: true
  gcpLabels:
    project: multi-provider-tutorial
    managed-by: openmcf
```

Replace `my-gcp-project` with your actual GCP project ID.

The `GcpGcsBucket` spec requires three fields: `gcpProjectId`, `location`, and `bucketName`. The `gcpProjectId` field uses a `StringValueOrRef` type (same pattern as the Kubernetes `namespace` field from the [previous tutorial](./first-kubernetes-resource)) — the literal value goes inside a `value` wrapper.

The `bucketName` must be globally unique across all GCP projects, 3-63 characters, lowercase alphanumeric with hyphens and dots.

## Step 3: Compare the Manifests

Place the two manifests side by side. The structure is identical:

```
apiVersion: <provider>.openmcf.org/v1     # Provider-specific API
kind: <ComponentName>                      # Component type
metadata:                                  # KRM metadata (same structure)
  name: <resource-name>
  labels:
    openmcf.org/provisioner: pulumi        # Same label, same options
spec:                                      # Provider-specific configuration
  ...
```

The `apiVersion`, `kind`, `metadata` structure is the same for every OpenMCF component — AWS, GCP, Azure, Kubernetes, or any of the other providers. The only thing that changes is `spec`, because each cloud provider has different resource configuration.

Here is how the provider-specific concepts map across the two manifests:

| Concept | AWS S3 | GCP GCS |
|---------|--------|---------|
| Region/location | `spec.awsRegion: us-east-1` | `spec.location: US` |
| Versioning | `spec.versioningEnabled: true` | `spec.versioningEnabled: true` |
| Encryption | `spec.encryptionType: ENCRYPTION_TYPE_SSE_S3` | Google-managed by default |
| Tags/labels | `spec.tags` (map) | `spec.gcpLabels` (map) |
| Access control | `spec.isPublic: false` (default) | `spec.uniformBucketLevelAccessEnabled: true` |
| Bucket naming | Auto-derived from `metadata.name` + stack | `spec.bucketName` (explicit, globally unique) |

Some concepts are shared (versioning), some are named differently (tags vs labels), and some are provider-specific (SSE encryption types, uniform bucket-level access). The protobuf schema for each component captures these differences precisely.

## Step 4: Deploy the AWS Bucket

Preview:

```bash
openmcf plan -f aws-bucket.yaml
```

Deploy:

```bash
openmcf apply -f aws-bucket.yaml
```

The deployment outputs for S3:

| Output | Description |
|--------|-------------|
| `bucket_id` | Name of the S3 bucket created on AWS |
| `bucket_arn` | ARN for IAM policies and cross-account access |
| `region` | AWS region where the bucket was created |
| `bucket_regional_domain_name` | Regional endpoint for accessing the bucket |

Verify with the AWS CLI:

```bash
aws s3 ls | grep multi-provider
```

## Step 5: Deploy the GCP Bucket

Preview:

```bash
openmcf plan -f gcp-bucket.yaml
```

Deploy:

```bash
openmcf apply -f gcp-bucket.yaml
```

The deployment output for GCS:

| Output | Description |
|--------|-------------|
| `bucket_id` | Name of the GCS bucket created on GCP |

Verify with the gcloud CLI:

```bash
gcloud storage ls | grep openmcf-multi-provider
```

## Step 6: The Pattern

Step back and look at what just happened. You deployed resources on two different cloud providers:

- **Same CLI**: `openmcf apply -f <manifest>` for both
- **Same manifest structure**: `apiVersion`, `kind`, `metadata`, `spec` for both
- **Same provisioner label**: `openmcf.org/provisioner: pulumi` for both
- **Same lifecycle**: `plan` -> `apply` -> `destroy` for both
- **Same validation**: protobuf-defined schemas with field-level validation for both

The provider-specific complexity — AWS IAM, GCP project IDs, region naming conventions, encryption defaults, naming constraints — is encapsulated in each component's `spec`. You do not need to learn a different tool, a different command syntax, or a different configuration format for each provider.

This extends to all 198 deployment components across 14 providers in the [Component Catalog](../catalog). Whether you are deploying a Kubernetes PostgreSQL database, an AWS VPC, a GCP Cloud SQL instance, or a Cloudflare DNS zone, the workflow is the same: write a manifest, plan, apply.

## Step 7: Clean Up

Destroy both buckets:

```bash
openmcf destroy -f aws-bucket.yaml
openmcf destroy -f gcp-bucket.yaml
```

Each `destroy` command reads the manifest, identifies the managed resources on the respective provider, and removes them.

## What You Learned

- How OpenMCF provides a consistent KRM-based interface across cloud providers
- How the `spec` section is the only provider-specific part of a manifest
- How the same CLI commands (`plan`, `apply`, `destroy`) work identically across providers
- How each component's protobuf schema defines the exact configuration surface for that provider's resources
- The conceptual mapping between equivalent cloud provider features (regions, tags, versioning, encryption)

## What's Next

- [Cloud Resource Kinds](../concepts/cloud-resource-kinds) — the full taxonomy of 198 components across 14 providers
- [Deployment Components](../concepts/deployment-components) — how components are structured with protobuf APIs and dual IaC modules
- [Component Catalog](../catalog) — browse all available components and their configuration
- [Writing Manifests](../guides/manifests) — practical guide to writing manifests for any component
