# GCP Monitoring Dashboard - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a Cloud Monitoring dashboard using Planton's `GcpMonitoringDashboard` API. The module is written in Go and creates `monitoring.Dashboard` from the spec's one JSON document (the Monitoring API's own Dashboard format — the provider deliberately models the fast-moving widget schema as a JSON string, and this module honors that judgment).

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
    ├── dashboard.go           # Dashboard creation from the JSON document
    ├── locals.go              # Resolved resource
    └── outputs.go             # Output constants
```

## How the module maps the spec

| Spec field | Provider argument | Notes |
|---|---|---|
| `dashboard_json` | `dashboard_json` | Validated as JSON at plan time; server-added keys (etag, name) are diff-suppressed by the provider, so console-exported dashboards round-trip cleanly |
| `project_id` | `project` | Omitted when empty — the provider's default project applies |
| `deletion_policy` | `deletion_policy` | Omitted when empty (provider default DELETE); PREVENT fails destroys; ABANDON removes from management only |

The module also enables `monitoring.googleapis.com` on the target project
(`disable_on_destroy` false — tearing down one dashboard never disables
monitoring project-wide).

## Outputs

| Output | Description |
|---|---|
| `dashboard_name` | Server-assigned resource name (`projects/{project}/dashboards/{id}`) |

## Local development

`iac-input.yaml` carries a ready smoke manifest. Run the module directly:

```bash
planton apply --manifest ../../e2e/manifest.yaml --module-dir .
```
