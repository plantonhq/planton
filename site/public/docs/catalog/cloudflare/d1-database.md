---
title: "D1 Database"
description: "D1 Database deployment documentation"
icon: "package"
order: 100
componentName: "cloudflared1database"
---

# Cloudflare D1 Database

Deploys a Cloudflare D1 serverless SQL database with an optional primary location hint and read replication configuration. The component provisions a single D1 database resource in the specified Cloudflare account and exports the database identifier for use by Workers bindings.

## What Gets Created

When you deploy a CloudflareD1Database resource, OpenMCF provisions:

- **D1 Database** — a `cloudflare_d1_database` resource in the specified Cloudflare account, with an optional region hint that maps to the Cloudflare `primary_location_hint` property
- **Read Replication (optional)** — when `readReplication` is configured, the database is created with D1 Read Replication (Beta) enabled, placing read-only replicas across multiple regions to reduce global read latency

## Prerequisites

- **Cloudflare credentials** configured via environment variables or OpenMCF provider config
- **A Cloudflare account ID** with D1 access enabled
- **Application-level Sessions API support** if enabling read replication — failing to use the D1 Sessions API with replication enabled will cause data consistency errors

## Quick Start

Create a file `d1-database.yaml`:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareD1Database
metadata:
  name: my-d1-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CloudflareD1Database.my-d1-db
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  databaseName: my-app-db
```

Deploy:

```shell
openmcf apply -f d1-database.yaml
```

This creates a D1 database named `my-app-db` with Cloudflare selecting the default storage region.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `accountId` | `string` | The Cloudflare account ID in which to create the database. | Required |
| `databaseName` | `string` | The unique name for the D1 database within the account. | Required, max 64 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `region` | `enum` | unspecified | The geographical region for the database's primary instance. Maps to Cloudflare's `primary_location_hint`. If omitted, Cloudflare selects a default location based on account settings. One of: `weur` (Western Europe), `eeur` (Eastern Europe), `apac` (Asia Pacific), `oc` (Oceania), `wnam` (Western North America), `enam` (Eastern North America). |
| `readReplication.mode` | `string` | — | The replication mode. Set to `auto` to enable automatic read replication across regions, or `disabled` to turn it off. Required when the `readReplication` object is present. Enabling replication requires application-level code changes to use the D1 Sessions API. |

## Examples

### Minimal Database with Default Region

A D1 database where Cloudflare selects the optimal storage location:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareD1Database
metadata:
  name: analytics-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CloudflareD1Database.analytics-db
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  databaseName: analytics
```

### Database in a Specific Region

A D1 database pinned to Western Europe, useful for GDPR compliance or when your Workers are deployed in that region:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareD1Database
metadata:
  name: eu-users-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CloudflareD1Database.eu-users-db
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  databaseName: eu-users
  region: weur
```

### Database with Read Replication

A production D1 database in Eastern North America with automatic read replication enabled for lower global read latency. Your application code must use the D1 Sessions API to maintain consistency:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareD1Database
metadata:
  name: global-app-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CloudflareD1Database.global-app-db
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  databaseName: global-app
  region: enam
  readReplication:
    mode: auto
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `databaseId` | `string` | The unique identifier (UUID) of the created D1 database |
| `databaseName` | `string` | The name of the database as confirmed by Cloudflare |
| `connectionString` | `string` | Reserved for future use. The Pulumi Cloudflare provider does not currently expose a connection string for D1, so this field is empty. |

## Related Components

- [CloudflareWorker](/docs/catalog/cloudflare/worker) — Workers bind to D1 databases for query execution; deploy a Worker alongside your D1 database to serve your application
- [CloudflareR2Bucket](/docs/catalog/cloudflare/r2-bucket) — object storage commonly paired with D1 for storing large blobs while keeping metadata in D1
- [CloudflareKvNamespace](/docs/catalog/cloudflare/kv-namespace) — key-value storage useful as a caching layer in front of D1 queries
