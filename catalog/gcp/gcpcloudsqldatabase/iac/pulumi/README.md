# GCP Cloud SQL Database - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying logical databases inside Cloud SQL instances using Planton's `GcpCloudSqlDatabase` API. The module is written in Go and leverages the Pulumi GCP provider to create `sql.Database` resources (backed by `google_sql_database`).

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **An existing Cloud SQL instance** in the target project
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set

## Directory Structure

```
iac/pulumi/
├── main.go           # Pulumi program entry point
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and deployment targets
├── README.md         # This file
└── module/
    ├── main.go       # Module coordinator
    ├── database.go   # Database resource creation
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

Provide a `iac-input.yaml` with the database specification:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpCloudSqlDatabase
  metadata:
    name: orders-database
  spec:
    instance:
      value: orders-db
    databaseName: orders
    charset: utf8mb4
    collation: utf8mb4_0900_ai_ci
```

### 3. Deploy

```bash
make up
```

### 4. Destroy

```bash
make destroy
```

## Outputs

| Output | Description |
|--------|-------------|
| `database_name` | Name of the database inside the instance |
| `self_link` | GCP resource self link |
