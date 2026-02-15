---
title: "BigQuery Dataset"
description: "Deploy GCP BigQuery datasets using OpenMCF"
---

# GcpBigQueryDataset

Provision and manage BigQuery datasets -- the top-level container for tables,
views, and routines in Google BigQuery, with location control, access management,
encryption, and lifecycle policies.

## Overview

GcpBigQueryDataset creates a BigQuery dataset within a GCP project. Datasets
control where data lives (location), who can access it (access entries), how
it's encrypted (Google-managed or CMEK), and how long tables and partitions
are retained. Tables within a dataset are managed by application code or tools
like dbt -- this resource focuses on the infrastructure boundary.

## Quick Start

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigQueryDataset
metadata:
  name: my-analytics-dataset
spec:
  projectId:
    value: "my-gcp-project"
  datasetId: analytics_prod
  location: US
```

## Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `projectId` | StringValueOrRef | Yes | GCP project ID |
| `datasetId` | string | Yes | Dataset ID (letters, numbers, underscores) |
| `location` | string | Yes | Multi-regional (US, EU) or regional |
| `friendlyName` | string | No | Display name |
| `description` | string | No | Description |
| `defaultTableExpirationMs` | int64 | No | Auto-delete tables after N ms (min 3600000) |
| `defaultPartitionExpirationMs` | int64 | No | Auto-delete partitions after N ms |
| `maxTimeTravelHours` | int32 | No | 48-168 hours (default: 168) |
| `isCaseInsensitive` | bool | No | Case-insensitive names (immutable) |
| `defaultCollation` | string | No | "und:ci" for case-insensitive |
| `storageBillingModel` | string | No | LOGICAL (default) or PHYSICAL |
| `deleteContentsOnDestroy` | bool | No | Delete tables on destroy (default: false) |
| `kmsKeyName` | StringValueOrRef | No | CMEK encryption key |
| `access` | list | No | Access control entries (authoritative) |

## Outputs

| Output | Description |
|--------|-------------|
| `dataset_id` | Short dataset ID for SQL queries |
| `self_link` | Fully qualified dataset URI |
| `project` | GCP project containing the dataset |
| `creation_time` | Creation timestamp (ms since epoch) |

## Important

**Access is authoritative.** When you specify `access`, BigQuery removes entries
not in your spec. Omitting `access` preserves default project-level access.

**Dataset ID cannot contain hyphens.** Only letters, numbers, and underscores.

**Location is immutable.** Choose carefully -- changing requires destroy and recreate.

## Related

- [GcpKmsKey](/docs/catalog/gcp/kms-key) -- CMEK encryption key
- [GcpProject](/docs/catalog/gcp/project) -- Parent GCP project
- [GcpGcsBucket](/docs/catalog/gcp/gcs-bucket) -- Often paired for data lake patterns
