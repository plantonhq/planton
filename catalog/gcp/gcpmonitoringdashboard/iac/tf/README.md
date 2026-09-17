# GCP Monitoring Dashboard - Terraform Module

## Overview

This directory contains the Terraform/OpenTofu implementation for deploying a Cloud Monitoring dashboard using Planton's `GcpMonitoringDashboard` API. The module creates `google_monitoring_dashboard` from the spec's one JSON document (the Monitoring API's own Dashboard format — the provider deliberately models the fast-moving widget schema as a JSON string, and this module honors that judgment).

## Prerequisites

1. **OpenTofu** (or Terraform >= 1.5)
2. **GCP Project** with the Cloud Monitoring API enabled (the module enables it if needed)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs

## Module Files

| File | Purpose |
|---|---|
| `variables.tf` | GENERATED from the proto spec (`planton tofu generate-variables GcpMonitoringDashboard`) — never hand-edited |
| `locals.tf` | Project fallback derivation |
| `main.tf` | API enablement + the dashboard resource |
| `outputs.tf` | Stack outputs |
| `provider.tf` | google provider pin (`~> 8.3`) |
| `backend.tf` | Local state backend (the runner injects the real backend) |

## How the module maps the spec

| Spec field | Provider argument | Notes |
|---|---|---|
| `dashboard_json` | `dashboard_json` | Validated as JSON at plan time; server-added keys are diff-suppressed by the provider |
| `project_id` | `project` | `null` when empty — the provider's default project applies |
| `deletion_policy` | `deletion_policy` | `null` when empty (provider default DELETE) |

## Outputs

| Output | Description |
|---|---|
| `dashboard_name` | Server-assigned resource name (`projects/{project}/dashboards/{id}`) |

## Offline validation

```bash
tofu init -backend=false
tofu plan   # against a tfvars converted from e2e/manifest.yaml
```
