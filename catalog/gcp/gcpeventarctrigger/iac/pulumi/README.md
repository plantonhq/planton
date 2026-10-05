# GCP Eventarc Trigger - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying an Eventarc trigger using Planton's `GcpEventarcTrigger` API. The module is written in Go and creates `eventarc.Trigger` plus two count-gated companions: an `eventarc.Channel` when the spec arms `partner_channel` (SaaS partner events), and the per-project-per-location `eventarc.GoogleChannelConfig` singleton when `google_channel_crypto_key` is set.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with the Eventarc API enabled (the module enables it if needed)
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs; the trigger's `service_account` needs `roles/eventarc.eventReceiver` (plus `roles/run.invoker` for authenticated Cloud Run destinations)

## Directory Structure

```
iac/pulumi/
├── main.go                    # Pulumi program entry point
├── Pulumi.yaml                # Pulumi project configuration
├── README.md                  # This file
└── module/
    ├── main.go                # Module coordinator
    ├── trigger.go             # Trigger + channel + google-channel-config wiring
    ├── locals.go              # Resolved resource + derived values + label merge
    └── outputs.go             # Output constants
```

## How the module maps the spec

| Spec field | Provider argument | Notes |
|---|---|---|
| `trigger_name` | `name` | Defaults to metadata.name; ForceNew |
| `location` | `location` | Required; a region or `global`; ForceNew |
| `matching_criteria[]` | `matching_criteria` | A `type` criterion is required (API rule, spec-enforced) |
| `destination.cloud_run_service.*` | `destination.cloud_run_service.*` | `service` references a GcpCloudRun's `service_name` output; `region` sent only when set (Optional+Computed) |
| `destination.gke.*` | `destination.gke.*` | `cluster` references a GcpGkeCluster's `name` output |
| `destination.workflow` | `destination.workflow` | References a GcpWorkflow's `workflow_id` output (full resource name) |
| `destination.http_endpoint.uri` | `destination.http_endpoint.uri` | HTTPS-only |
| `destination.http_endpoint.network_attachment` | `destination.network_config.network_attachment` | The spec carries it inside the arm; the provider's sibling block shape is restored here |
| `service_account` | `service_account` | References a GcpServiceAccount's `email` output |
| `transport_pubsub_topic` | `transport.pubsub.topic` | messagePublished triggers only; existing topics are never deleted by Eventarc |
| `event_data_content_type` | `event_data_content_type` | API default `application/json` when unset |
| `retry_max_attempts` | `retry_policy.max_attempts` | Only value 1 (provider truth); Cloud Run destinations only |
| `partner_channel.*` | `google_eventarc_channel` + trigger `channel` | Count-gated companion; the trigger is wired to the created channel's full name |
| `google_channel_crypto_key` | `google_eventarc_google_channel_config.crypto_key_name` | Count-gated singleton — manage from at most ONE trigger per project+location |
| `labels` | `labels` | Spec labels merged with platform attribution labels (platform wins); applied to trigger and channel |
| `project_id` | `project` | Omitted when empty — the provider's default project applies |
| `deletion_policy` | `deletion_policy` | One spec lever wired to trigger, channel, and config |

The module also enables `eventarc.googleapis.com` on the target project
(`disable_on_destroy` false). The first trigger in a project provisions
Eventarc's service agent — the first delivery can lag a few minutes
behind the apply.

## Outputs

| Output | Description |
|---|---|
| `trigger_name` | The trigger name in GCP |
| `trigger_id` | Full trigger resource name — the canonical API handle |
| `partner_channel_activation_token` | Partner triggers only: the one-time token the SaaS partner needs (sensitive); empty otherwise |

## Local development

`iac-input.yaml` carries a ready smoke manifest. Run the module directly:

```bash
planton apply --manifest ../../e2e/manifest.yaml --module-dir .
```
