# GCP Service Account IAM Member - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying additive service-account-scoped IAM grants using Planton's `GcpServiceAccountIamMember` API. The module is written in Go and uses the Pulumi GCP provider to create `serviceaccount.IAMMember` resources (backed by `google_service_account_iam_member`).

The grant is ADDITIVE: it merges one (role, member[, condition]) pair into the service account's IAM policy without touching any other member's bindings, and destroy subtracts only this exact pair.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **An existing service account** — the grant is policy metadata; no billed infrastructure is created
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: the deploying principal's permissions are listed in [`../permissions.yaml`](../permissions.yaml)

## Directory Structure

```
iac/pulumi/
├── main.go           # Pulumi program entry point
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and deployment targets
├── README.md         # This file
└── module/
    ├── main.go        # Module coordinator
    ├── iam_member.go  # IAM member grant creation
    ├── locals.go      # Resolved resource holder
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
  kind: GcpServiceAccountIamMember
  metadata:
    name: github-deployer-impersonation
  spec:
    service_account_id:
      value: projects/my-gcp-project-123/serviceAccounts/deployer@my-gcp-project-123.iam.gserviceaccount.com
    role:
      value: roles/iam.workloadIdentityUser
    member:
      value: principalSet://iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/github/attribute.repository/my-org/my-repo
```

### 3. Deploy

```bash
export IAC_INPUT_FILE_PATH=iac-input.yaml
make up
```

### 4. Destroy

```bash
make destroy
```

Destroy removes exactly this (role, member) pair from the account's policy — no other grant is touched.

## Outputs

| Output | Description |
|--------|-------------|
| `service_account_id` | The account whose policy received the grant |
| `role` | The granted role |
| `member` | The granted member |
| `etag` | The account IAM policy etag after the grant |

## Behavior Notes

- Every argument is immutable: any spec change replaces the grant atomically (the IAM API has no grant update).
- The member format is validated at deploy time (deleted principals rejected) because the value usually arrives through a reference resolved only at deploy time.
- There is no project argument — the account's project is embedded in its fully-qualified resource name.
