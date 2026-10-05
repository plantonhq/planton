# GCP Monitoring Uptime Check - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a Cloud Monitoring uptime check using Planton's `GcpMonitoringUptimeCheck` API. The module is written in Go and creates `monitoring.UptimeCheckConfig` — the probe Google runs against a URL, monitored resource, resource group, or synthetic-monitor Cloud Function from multiple regions on a fixed cadence.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with the Cloud Monitoring API enabled (the module enables it if needed)
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: `roles/monitoring.editor` (or broader) on the target project

## Directory Structure

```
iac/pulumi/
├── main.go                    # Pulumi program entry point
├── Pulumi.yaml                # Pulumi project configuration
├── README.md                  # This file
└── module/
    ├── main.go                # Module coordinator
    ├── uptime_check.go        # Uptime check creation + HTTP-check expansion
    ├── locals.go              # Resolved resource + derived values
    └── outputs.go             # Output constants
```

## Quick Start

```bash
cd iac/pulumi
pulumi stack init dev
```

Provide a `iac-input.yaml`:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpMonitoringUptimeCheck
  metadata:
    name: website-https
  spec:
    timeout: 10s
    monitored_resource:
      type: uptime_url
      labels:
        host: example.com
    http_check:
      path: /
      use_ssl: true
      validate_ssl: true
```

```bash
pulumi preview
pulumi up
```

## Inputs

The module consumes `GcpMonitoringUptimeCheckIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpMonitoringUptimeCheck` spec (target, check, cadence, content assertions) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `uptime_check_name` | string | `projects/{project}/uptimeCheckConfigs/{id}` |
| `uptime_check_id` | string | The bare check ID — the value alert policies filter on (`metric.label.check_id`) to page on failures |

## Behavior Notes

- **Exactly one target, exactly one check**: enforced by the spec's validations before the module runs; a synthetic monitor carries its own probe logic and takes no check block.
- **A check only measures** — pair it with a `GcpMonitoringAlertPolicy` filtering on `uptime_check_passed` and this check's `uptime_check_id` to actually page.
- **API defaults preserved**: period (300s), checker type, request method, and content type fall through to GCP's defaults when unset.
- **API enablement**: the module enables `monitoring.googleapis.com` (with `disable_on_destroy=false`) so a fresh project works first try.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation
