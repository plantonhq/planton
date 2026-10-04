# GCP GCS Bucket IAM Member - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying additive bucket-scoped IAM grants using Planton's `GcpGcsBucketIamMember` API. The module is written in Go and uses the Pulumi GCP provider to create `storage.BucketIAMMember` resources (backed by `google_storage_bucket_iam_member`).

The grant is ADDITIVE: it merges one (role, member[, condition]) pair into the bucket's IAM policy without touching any other member's bindings, and destroy subtracts only this exact pair.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **An existing bucket** — the grant is policy metadata; no billed infrastructure is created
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
├── README.md         # This file
└── module/
    ├── main.go        # Module coordinator
    ├── iam_member.go  # IAM member grant creation
    └── outputs.go     # Output constants
```

## Quick Start

### 1. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

### 2. Create Input File

Provide a `iac-input.yaml` with the grant specification:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpGcsBucketIamMember
  metadata:
    name: audit-archive-writer
  spec:
    bucket:
      value: acme-audit-archive
    role:
      value: roles/storage.objectCreator
    member:
      value: serviceAccount:service-123456789@gcp-sa-logging.iam.gserviceaccount.com
```

### 3. Deploy

```bash
export IAC_INPUT_YAML_FILE=iac-input.yaml
pulumi up
```

### 4. Destroy

```bash
pulumi destroy
```

Destroy removes exactly this (role, member) pair from the bucket's policy — no other grant is touched, and the bucket and its objects are never affected.

## Outputs

| Output | Description |
|--------|-------------|
| `bucket` | The bucket whose policy received the grant |
| `role` | The granted role |
| `member` | The granted member |
| `etag` | The bucket IAM policy etag after the grant |

## Behavior Notes

- Every argument is immutable: any spec change replaces the grant (the IAM API has no grant update).
- There is no project argument — bucket names are global.
- The condition block is sent only when `condition` is set, and its description only when non-empty. Conditions require uniform bucket-level access on the bucket.
- The member format is validated on literal values; a referenced `writer_identity` or `member` output already arrives in `serviceAccount:<email>` form.
