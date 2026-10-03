# GCP Monitoring Notification Channel - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a Cloud Monitoring notification channel using Planton's `GcpMonitoringNotificationChannel` API. The module is written in Go and creates `monitoring.NotificationChannel` — the delivery endpoint (email, Slack, PagerDuty, SMS, webhook, or Pub/Sub) that alert policies notify when incidents open or close.

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
    ├── notification_channel.go # Channel creation
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
  kind: GcpMonitoringNotificationChannel
  metadata:
    name: oncall-email
  spec:
    type: email
    channel_labels:
      email_address: oncall@example.com
```

```bash
pulumi preview
pulumi up
```

## Inputs

The module consumes `GcpMonitoringNotificationChannelIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpMonitoringNotificationChannel` spec (type, per-type config, credentials, enablement) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `channel_name` | string | `projects/{project}/notificationChannels/{id}` — the value alert policies reference |
| `verification_status` | string | Verification state (SMS/email channels require verification before delivering) |

## Behavior Notes

- **Two label surfaces, never conflated**: the provider's `labels` argument is the TYPE-SPECIFIC channel configuration (fed from `spec.channel_labels`); `user_labels` is freeform metadata (fed from `spec.labels` merged with the platform attribution labels).
- **Credentials ride `sensitive_labels`**: the API stores and redacts them server-side; validation refuses credential keys in the plain config map.
- **`enabled` is sent explicitly** on every apply so a `true -> false` transition reaches the API instead of being omitted (the silent-no-op class).
- **`force_delete` semantics**: false (default) fails deletion while alert policies still reference the channel — the safe posture; true removes the channel from those policies in the same operation.
- **API enablement**: the module enables `monitoring.googleapis.com` (with `disable_on_destroy=false`) so a fresh project works first try.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation
