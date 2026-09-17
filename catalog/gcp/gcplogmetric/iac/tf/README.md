# GCP Log Metric - Terraform Module

## Overview

This directory contains the Terraform/OpenTofu implementation for deploying a Cloud Logging log-based metric using Planton's `GcpLogMetric` API. The module creates `google_logging_metric` — the bridge from log entries matching a filter to a chartable Cloud Monitoring metric (a counter, or a distribution extracted from entry fields).

## Prerequisites

1. **OpenTofu** (or Terraform >= 1.5)
2. **GCP Project** with the Cloud Logging API enabled (the module enables it if needed)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs

## Module Files

| File | Purpose |
|---|---|
| `variables.tf` | GENERATED from the proto spec (`planton tofu generate-variables GcpLogMetric`) — never hand-edited |
| `locals.tf` | Project fallback + metric-name derivation |
| `main.tf` | API enablement + the metric resource |
| `outputs.tf` | Stack outputs |
| `provider.tf` | google provider pin (`~> 8.3`) |
| `backend.tf` | Local state backend (the runner injects the real backend) |

## How the module maps the spec

| Spec field | Provider argument | Notes |
|---|---|---|
| `metric_name` | `name` | Defaults to metadata.name |
| `filter` | `filter` | Required |
| `bucket_name` | `bucket_name` | Bucket-scoped metrics; references a GcpLogBucket's `bucket_name` output |
| `disabled` | `disabled` | Sent EXPLICITLY on every apply — a true→false transition must reach the API |
| `metric_descriptor.*` | `metric_descriptor.*` | Kind/value-type/unit/display-name/label schema |
| `value_extractor`, `label_extractors` | same names | The DISTRIBUTION extraction surface |
| `bucket_options.*` | `bucket_options.*` | Explicit/exponential/linear histogram layouts |
| `project_id` | `project` | `null` when empty — the provider's default project applies |
| `deletion_policy` | `deletion_policy` | `null` when empty (provider default DELETE) |

## Outputs

| Output | Description |
|---|---|
| `metric_name` | The metric name — `logging.googleapis.com/user/{metric_name}` in Monitoring |

## Offline validation

```bash
tofu init -backend=false
tofu plan   # against a tfvars converted from e2e/manifest.yaml
```
