# GCP Log Bucket - Terraform Module

## Overview

This directory contains the Terraform/OpenTofu implementation for deploying a Cloud Logging bucket using Planton's `GcpLogBucket` API. The module creates exactly ONE of the four scope resources — `google_logging_{project|folder|organization|billing_account}_bucket_config` — selected by the spec's `scope` message (count-gated), plus the bucket's `google_logging_log_view` fan-out, the `google_logging_linked_dataset`, and the folder/organization `google_logging_{folder|organization}_settings` singleton when armed.

## Prerequisites

1. **OpenTofu** (or Terraform >= 1.5)
2. **GCP Project** with the Cloud Logging API enabled (the module enables it on the project scope)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs

## Module Files

| File | Purpose |
|---|---|
| `variables.tf` | GENERATED from the proto spec (`planton tofu generate-variables GcpLogBucket`) — never hand-edited |
| `locals.tf` | Scope gating, defaults, bucket-name derivation, count-gated client-config fallback |
| `main.tf` | The four count-gated bucket variants + views + linked dataset + settings |
| `outputs.tf` | Stack outputs |
| `provider.tf` | google provider pin (`~> 8.3`) |
| `backend.tf` | Local state backend (the runner injects the real backend) |

## How the module maps the spec

| Spec field | Provider argument | Notes |
|---|---|---|
| `scope.*` | `project` / `folder` / `organization` / `billing_account` | Selects WHICH of the four resources is created; empty scope = project bucket in the provider's default project (the required `project` argument resolves through a count-gated `google_client_config` read) |
| `bucket_id`, `location`, `description`, `retention_days` | same names | location ("global") and retention (30) sent EXPLICITLY with the spec defaults |
| `locked` | `locked` | Project scope only; one-way in GCP |
| `enable_analytics` | `enable_analytics` | Project scope only; sent ONLY when explicitly configured (the provider's own posture — enablement is an atomic, one-way operation) |
| `cmek_kms_key` | `cmek_settings.kms_key_name` | One-way once set; grant the Logging service account on the key FIRST |
| `index_configs[]` | `index_configs` | ≤ 20 |
| `log_views[]` | `google_logging_log_view` per entry | view_id → `name`; bucket derives from the created bucket |
| `linked_bigquery_dataset` | `google_logging_linked_dataset` | Requires analytics; every field create-time-only |
| `scope_settings` | `google_logging_{folder\|organization}_settings` | Folder/org scopes only; adopted singleton, destroy is a state-only no-op |
| `deletion_policy` | `deletion_policy` | Fans out to the bucket, every log view, and the linked dataset |

## Outputs

| Output | Description |
|---|---|
| `bucket_name` | Full resource name — the GcpLoggingSink / GcpLogMetric composition handle |
| `linked_dataset_id` | The linked BigQuery dataset ID (empty when not armed) |

## Offline validation

```bash
tofu init -backend=false
tofu plan   # against a tfvars converted from e2e/manifest.yaml
```
