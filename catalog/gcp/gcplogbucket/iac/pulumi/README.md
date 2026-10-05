# GCP Log Bucket - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a Cloud Logging bucket using Planton's `GcpLogBucket` API. The module is written in Go and creates exactly ONE of the four scope resources — `logging.ProjectBucketConfig`, `logging.FolderBucketConfig`, `logging.OrganizationBucketConfig`, or `logging.BillingAccountBucketConfig` — selected by the spec's `scope` message, plus the bucket's `logging.LogView` fan-out, the `logging.LinkedDataset`, and the folder/organization `logging.{Folder,Organization}Settings` singleton when armed.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with the Cloud Logging API enabled (the module enables it on the project scope)
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs

## Directory Structure

```
iac/pulumi/
├── main.go                    # Pulumi program entry point
├── Pulumi.yaml                # Pulumi project configuration
├── README.md                  # This file
└── module/
    ├── main.go                # Module coordinator
    ├── log_bucket.go          # Scope-gated bucket + views + linked dataset + settings
    ├── locals.go              # Resolved resource + scope/default derivation
    └── outputs.go             # Output constants
```

## How the module maps the spec

| Spec field | Provider argument | Notes |
|---|---|---|
| `scope.*` | `project` / `folder` / `organization` / `billing_account` | Selects WHICH of the four resources is created; empty scope = project bucket in the provider's default project (resolved via client config — the project argument is required by the provider) |
| `bucket_id`, `location`, `description`, `retention_days` | same names | location ("global") and retention (30) sent EXPLICITLY with the spec defaults |
| `locked` | `locked` | Project scope only; one-way in GCP |
| `enable_analytics` | `enable_analytics` | Project scope only; sent ONLY when explicitly configured (the provider's own posture — enablement is an atomic, one-way operation) |
| `cmek_kms_key` | `cmek_settings.kms_key_name` | One-way once set; grant the Logging service account on the key FIRST |
| `index_configs[]` | `index_configs` | ≤ 20 |
| `log_views[]` | `logging.LogView` per entry | view_id → `name`; bucket/location derive from the created bucket |
| `linked_bigquery_dataset` | `logging.LinkedDataset` | Requires analytics; every field create-time-only |
| `scope_settings` | `logging.FolderSettings` / `OrganizationSettings` | Folder/org scopes only; adopted singleton, destroy is a state-only no-op |
| `deletion_policy` | `deletion_policy` | Fans out to the bucket, every log view, and the linked dataset |

## Outputs

| Output | Description |
|---|---|
| `bucket_name` | Full resource name (`projects/{p}/locations/{l}/buckets/{b}` or the scope's form) — the GcpLoggingSink / GcpLogMetric composition handle |
| `linked_dataset_id` | The linked BigQuery dataset ID (empty when not armed) |

## Local development

`iac-input.yaml` carries a ready smoke manifest. Run the module directly:

```bash
planton apply --manifest ../../e2e/manifest.yaml --module-dir .
```
