# GCP Cloud SQL User - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying database users on Cloud SQL instances using Planton's `GcpCloudSqlUser` API. The module is written in Go and leverages the Pulumi GCP provider to create `sql.User` resources (backed by `google_sql_user`).

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **An existing Cloud SQL instance** in the target project
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set
6. **For IAM users on PostgreSQL**: the instance must set the database flag `cloudsql.iam_authentication = "on"`

## Directory Structure

```
iac/pulumi/
├── main.go           # Pulumi program entry point
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and deployment targets
├── README.md         # This file
└── module/
    ├── main.go       # Module coordinator
    ├── user.go       # User resource creation
    ├── locals.go     # Local values
    └── outputs.go    # Output constants
```

## Quick Start

### 1. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

### 2. Create Input File

Provide a `iac-input.yaml` with the user specification:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpCloudSqlUser
  metadata:
    name: orders-app-user
  spec:
    instance:
      value: orders-db
    userName: orders-app
    password: a-strong-generated-password
```

### 3. Deploy

```bash
make up
```

### 4. Destroy

```bash
make destroy
```

## Secrets Handling

- `password` flows from a `(sensitive)`-annotated spec field and is wrapped with `pulumi.ToSecret` — encrypted in Pulumi state, never exported in outputs.

## Outputs

| Output | Description |
|--------|-------------|
| `user_name` | The user name as stored by Cloud SQL |
