# GCP Cloud SQL - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying Cloud SQL instances using Planton's `GcpCloudSql` API. The module is written in Go and leverages the Pulumi GCP provider to create `sql.DatabaseInstance` resources (backed by `google_sql_database_instance`), enabling the Cloud SQL Admin API first so a fresh project works on the first deploy. The full instance surface is modeled: primaries, read replicas, and read pools (with auto scaling), restore provenance (clone / backup-run restore / Backup-and-DR point-in-time restore), cross-region DR pairing, hyperdisk provisioned performance, final backup on delete, PSC DNS automation, SQL Server Active Directory and Entra ID authentication, and the `deletion_policy` teardown lever — with identical zero-vs-omit semantics to the Terraform module.

One resource is one instance: a primary, or a read replica when `masterInstanceName` is set. Databases and users inside the instance are separate Planton kinds composed by instance name.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with billing enabled
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: `roles/cloudsql.admin` on the target project
6. **For private IP**: the VPC must already carry a service networking connection (GcpGlobalAddress + GcpServiceNetworkingConnection)

## Directory Structure

```
iac/pulumi/
├── main.go           # Pulumi program entry point
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and deployment targets
├── README.md         # This file
└── module/
    ├── main.go       # Module coordinator
    ├── instance.go   # Instance + settings + replica construction
    ├── locals.go     # Local values and labels
    └── outputs.go    # Output constants
```

## Quick Start

### 1. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

### 2. Create Input File

Provide a `iac-input.yaml` with the instance specification:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpCloudSql
  metadata:
    name: my-postgres
  spec:
    instanceName: orders-db
    region: us-central1
    databaseEngine: POSTGRESQL
    databaseVersion: POSTGRES_16
    tier: db-custom-2-7680
    backup:
      enabled: true
      pointInTimeRecoveryEnabled: true
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

- `rootPassword`, `replicaConfiguration.password`, and `replicaConfiguration.clientKey` flow from `(sensitive)`-annotated spec fields and are wrapped with `pulumi.ToSecret` — encrypted in Pulumi state, never exported in outputs.

## Outputs

| Output | Description |
|--------|-------------|
| `instance_name` | The composition key databases, users, and replicas reference |
| `connection_name` | `project:region:instance` for the Auth Proxy and connectors |
| `private_ip` / `public_ip` | Instance addresses (empty when the path is not enabled) |
| `self_link` | GCP resource self link |
| `service_account_email` | The instance's Google-managed service account |
| `dns_name` | DNS name (PSC-enabled instances) |
| `psc_service_attachment_link` | PSC service attachment link (PSC-enabled instances) |
