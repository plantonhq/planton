# GCP Monitoring SLO - Terraform Module

## Overview

This directory contains the Terraform/OpenTofu implementation for deploying a Cloud Monitoring service-level objective using Planton's `GcpMonitoringSlo` API. The module creates `google_monitoring_slo` plus, when the spec's service arm asks for one, the Monitoring service it measures — `google_monitoring_custom_service` or `google_monitoring_service` — count-gated by the arm (one kind, up to two service resources).

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
| `variables.tf` | GENERATED from the proto spec (`planton tofu generate-variables GcpMonitoringSlo`) — never hand-edited |
| `locals.tf` | Service-arm gating, naming, and label-merge derivations |
| `main.tf` | API enablement + count-gated services + the SLO with the full SLI tree |
| `outputs.tf` | Stack outputs (both derived from the SLO's resource name) |
| `provider.tf` | google provider pin (`~> 8.3`) |
| `backend.tf` | Local state backend (the runner injects the real backend) |

## How the module maps the spec

| Spec field | Provider argument | Notes |
|---|---|---|
| `service.service_id` | `service` | Measures an EXISTING service; no service resource created |
| `service.custom_service.*` | `google_monitoring_custom_service` | Created by the module (count-gated); its `service_id` feeds the SLO |
| `service.basic_service.*` | `google_monitoring_service` | Created from type + labels (count-gated) |
| `goal` | `goal` | > 0 and <= 0.9999 (the API's bound) |
| `calendar_period` / `rolling_period_days` | same names | Exactly one (the provider's ExactlyOneOf) |
| `sli.basic_sli` / `sli.request_based_sli` / `sli.windows_based_sli` | same blocks | The full SLI tree, all nested arms |
| `labels` | `user_labels` | Merged with Planton platform labels (platform wins); also applied to any created service |
| `project_id` | `project` | `null` when empty — the provider's default project applies |
| `deletion_policy` | `deletion_policy` | Applied to the SLO AND any service the module created |

## Outputs

| Output | Description |
|---|---|
| `slo_name` | `projects/{p}/services/{s}/serviceLevelObjectives/{id}` — the burn-rate alert handle |
| `service_name` | `projects/{p}/services/{s}` — the measured service |

## Offline validation

```bash
tofu init -backend=false
tofu plan   # against a tfvars converted from e2e/manifest.yaml
```
