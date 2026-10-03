# GCP Backend Bucket - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying GCP Compute Engine backend buckets using Planton's `GcpBackendBucket` API. The module is written in Go and creates a `compute.BackendBucket` (backed by `google_compute_backend_bucket`) plus one `compute.BackendBucketSignedUrlKey` per configured signing key.

A backend bucket serves a Cloud Storage bucket's objects through an external HTTP(S) load balancer, optionally cached at Google's edge by Cloud CDN. URL maps route static paths to it by self-link.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **An existing GCS bucket** — the origin whose objects are served
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs

## Directory Structure

```
iac/pulumi/
├── main.go           # Pulumi program entry point
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and deployment targets
├── README.md         # This file
└── module/
    ├── main.go            # Module coordinator
    ├── backend_bucket.go  # Backend bucket + signed-URL key creation
    ├── locals.go          # Resolved resource + derived values
    └── outputs.go         # Output constants
```

## Quick Start

### 1. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

### 2. Create Input File

Provide a `iac-input.yaml` with the backend bucket specification:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpBackendBucket
  metadata:
    name: static-assets
  spec:
    project_id:
      value: my-gcp-project-123
    bucket_name:
      value: my-static-assets-bucket
    enable_cdn: true
    cdn_policy:
      cache_mode: CACHE_ALL_STATIC
      default_ttl: 3600
```

### 3. Build, Preview, and Deploy

```bash
make build
pulumi preview
pulumi up
```

### 4. View Outputs

```bash
pulumi stack output self_link
```

## Inputs

The module consumes `GcpBackendBucketIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpBackendBucket` spec (origin bucket, CDN policy, keys, etc.) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

Spec fields: `bucket_name` (required; the origin), optional `project_id` (falls back to the provider default project when empty), optional `backend_bucket_name` (defaults to `metadata.name`), `enable_cdn` + `cdn_policy`, `compression_mode`, `custom_response_headers`, `edge_security_policy`, `load_balancing_scheme`, and up to 3 `signed_url_keys`.

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `self_link` | string | Self-link URI — the value URL maps reference |
| `backend_bucket_name` | string | Name of the backend bucket in GCP |
| `bucket_name` | string | The origin GCS bucket currently being served |

## Behavior Notes

- **Immutability**: `backend_bucket_name`, `project_id`, and `load_balancing_scheme` are ForceNew; everything else — including the origin `bucket_name` — updates in place.
- **Signed-URL keys** are marked secret in Pulumi state (`pulumi.ToSecret`) and never exported. Keys are immutable in GCP; changing a key's value replaces that key — the rotation semantics signed URLs need.
- **TTL semantics**: a 0 TTL in the spec means "unset — let the GCP API default"; cache-mode/TTL coherence is enforced by the spec before deploy.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation
