# GCP Monitoring SLO - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a Cloud Monitoring service-level objective using Planton's `GcpMonitoringSlo` API. The module is written in Go and creates `monitoring.Slo` plus, when the spec's service arm asks for one, the Monitoring service it measures — `monitoring.CustomService` or `monitoring.GenericService` — count-gated by the arm (one kind, up to two service resources).

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with the Cloud Monitoring API enabled (the module enables it if needed)
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
    ├── slo.go                 # SLO + count-gated service creation, SLI expansion
    ├── locals.go              # Resolved resource + derived values
    └── outputs.go             # Output constants
```

## How the module maps the spec

| Spec field | Provider argument | Notes |
|---|---|---|
| `service.service_id` | `service` | Measures an EXISTING service; no service resource created |
| `service.custom_service.*` | `monitoring.CustomService` | Created by the module; its `service_id` feeds the SLO |
| `service.basic_service.*` | `monitoring.GenericService` | Created by the module from type + labels |
| `goal` | `goal` | > 0 and <= 0.9999 (the API's bound) |
| `calendar_period` / `rolling_period_days` | same names | Exactly one (the provider's ExactlyOneOf) |
| `sli.basic_sli` / `sli.request_based_sli` / `sli.windows_based_sli` | same blocks | The full SLI tree, all nested arms |
| `labels` | `user_labels` | Merged with Planton platform labels (platform wins); also applied to any created service |
| `project_id` | `project` | Omitted when empty — the provider's default project applies |
| `deletion_policy` | `deletion_policy` | Applied to the SLO AND any service the module created |

Both outputs derive from the SLO's server-assigned resource name, so
they are correct on every service arm — including an existing service in
the provider's ambient project.

## Outputs

| Output | Description |
|---|---|
| `slo_name` | `projects/{p}/services/{s}/serviceLevelObjectives/{id}` — the burn-rate alert handle |
| `service_name` | `projects/{p}/services/{s}` — the measured service |

## Local development

`iac-input.yaml` carries a ready smoke manifest. Run the module directly:

```bash
planton apply --manifest ../../e2e/manifest.yaml --module-dir .
```
