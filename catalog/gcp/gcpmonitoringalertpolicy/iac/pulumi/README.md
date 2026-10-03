# GCP Monitoring Alert Policy - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a Cloud Monitoring alert policy using Planton's `GcpMonitoringAlertPolicy` API. The module is written in Go and creates `monitoring.AlertPolicy` — the rule that watches metrics or logs (threshold, absence, log match, MQL, PromQL, or SQL conditions) and notifies the referenced channels when incidents open.

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
    ├── alert_policy.go        # Policy creation + condition/strategy expansion
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
  kind: GcpMonitoringAlertPolicy
  metadata:
    name: cpu-saturation
  spec:
    combiner: OR
    severity: WARNING
    conditions:
      - display_name: cpu above 80%
        condition_threshold:
          filter: metric.type="compute.googleapis.com/instance/cpu/utilization" AND resource.type="gce_instance"
          comparison: COMPARISON_GT
          threshold_value: 0.8
          duration: 300s
          aggregations:
            - alignment_period: 60s
              per_series_aligner: ALIGN_MEAN
    notification_channels:
      - value: projects/my-project/notificationChannels/1234567890
```

```bash
pulumi preview
pulumi up
```

## Inputs

The module consumes `GcpMonitoringAlertPolicyIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpMonitoringAlertPolicy` spec (conditions, combiner, channels, strategy, documentation) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `policy_name` | string | `projects/{project}/alertPolicies/{id}` |

## Behavior Notes

- **Exactly one condition type per condition**: the API models the choice as a oneof; the spec's validations enforce it before the module runs (the Terraform provider leaves it unchecked client-side).
- **`enabled` is sent explicitly** on every apply so disabling a policy actually reaches the API instead of being omitted (a silently still-enabled policy pages people at 3am).
- **Log-based policies** (`condition_matched_log`) require `alert_strategy.notification_rate_limit` — the API's own pairing.
- **`threshold_value` is always sent**, including 0 — "greater than zero" is a legal, common threshold.
- **API enablement**: the module enables `monitoring.googleapis.com` (with `disable_on_destroy=false`) so a fresh project works first try.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation
