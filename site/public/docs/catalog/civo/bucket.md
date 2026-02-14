---
title: "Bucket"
description: "Bucket deployment documentation"
icon: "package"
order: 100
componentName: "civobucket"
---

# Civo Bucket

Deploys an S3-compatible object storage bucket on Civo Cloud with auto-generated access credentials. The component provisions both the bucket and an associated Object Store credential, exporting the endpoint URL and key references as stack outputs for use by application workloads.

## What Gets Created

When you deploy a CivoBucket resource, OpenMCF provisions:

- **Object Store Credential** — a `civo_object_store_credential` resource that generates an access key and secret key pair for authenticating against the bucket
- **Object Store Bucket** — a `civo_object_store` resource in the specified region, linked to the generated credential

> **Note on versioning:** The `versioningEnabled` field is accepted in the spec but cannot be applied directly through the Civo control plane. If versioning is requested, the module logs an advisory message. You must enable versioning post-deployment using the AWS S3 CLI or SDK pointed at the Civo endpoint.

> **Note on tags:** The Civo ObjectStore provider resource does not currently support tags. Any values in the `tags` field are recorded in metadata and logged but are not applied to the Civo resource.

## Prerequisites

- **Civo credentials** configured via environment variables or OpenMCF provider config
- **A Civo account** with Object Store access enabled in the target region

## Quick Start

Create a file `civo-bucket.yaml`:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoBucket
metadata:
  name: my-bucket
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CivoBucket.my-bucket
spec:
  bucketName: my-bucket
  region: nyc1
```

Deploy:

```shell
openmcf apply -f civo-bucket.yaml
```

This creates an object storage bucket named `my-bucket` in the New York region with auto-generated credentials.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `bucketName` | `string` | DNS-compatible name for the bucket. Lowercase letters, digits, and hyphens only. Must start and end with a letter or digit. | Required, 3–63 characters, pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` |
| `region` | `enum` | Civo region where the bucket is created. Valid values: `lon1`, `lon2`, `fra1`, `nyc1`, `phx1`, `mum1`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `versioningEnabled` | `bool` | `false` | Flag to request versioning on the bucket. Logged as advisory only; must be enabled post-deployment via the S3 API. |
| `tags` | `string[]` | `[]` | Tags for organizational purposes. Must be unique. Not currently applied to the Civo resource due to provider limitations. |

## Examples

### Basic Bucket

A minimal bucket for development or scratch storage:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoBucket
metadata:
  name: dev-assets
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CivoBucket.dev-assets
spec:
  bucketName: dev-assets
  region: fra1
```

### Bucket with Versioning Intent and Tags

A production bucket that records the intent to enable versioning and applies organizational tags:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoBucket
metadata:
  name: prod-uploads
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CivoBucket.prod-uploads
spec:
  bucketName: prod-uploads
  region: lon1
  versioningEnabled: true
  tags:
    - environment:production
    - team:platform
```

After deployment, enable versioning using the AWS CLI:

```shell
aws s3api put-bucket-versioning \
  --bucket prod-uploads \
  --versioning-configuration Status=Enabled \
  --endpoint-url <endpoint_url from stack outputs>
```

### Multi-Region Backup Buckets

Separate buckets across regions for geo-distributed backups:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoBucket
metadata:
  name: backup-us
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CivoBucket.backup-us
spec:
  bucketName: backup-us
  region: nyc1
  tags:
    - purpose:backup
    - region:us
---
apiVersion: civo.openmcf.org/v1
kind: CivoBucket
metadata:
  name: backup-eu
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CivoBucket.backup-eu
spec:
  bucketName: backup-eu
  region: fra1
  tags:
    - purpose:backup
    - region:eu
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `bucket_id` | `string` | Unique identifier (UUID) of the created bucket |
| `endpoint_url` | `string` | S3-compatible endpoint URL for the bucket (e.g., `https://objectstore.civo.com/<bucket-name>`) |
| `access_key_secret_ref` | `string` | Reference to the access key ID generated for the bucket credential |
| `secret_key_secret_ref` | `string` | Reference to the secret key generated for the bucket credential |

## Related Components

- [CivoKubernetesCluster](/docs/catalog/civo/kubernetes-cluster) — application workloads that read from or write to the bucket
- [CivoVpc](/docs/catalog/civo/vpc) — provides private network connectivity for workloads accessing the bucket
- [CivoFirewall](/docs/catalog/civo/firewall) — controls network access to services that interact with the bucket
