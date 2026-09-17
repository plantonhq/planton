# GCP Eventarc Trigger - Terraform Module

## Overview

This directory contains the Terraform/OpenTofu implementation for deploying an Eventarc trigger using Planton's `GcpEventarcTrigger` API. The module creates `google_eventarc_trigger` plus two count-gated companions: `google_eventarc_channel` when the spec arms `partner_channel`, and the per-project-per-location `google_eventarc_google_channel_config` singleton when `google_channel_crypto_key` is set.

## Prerequisites

1. **OpenTofu** (or Terraform >= 1.5)
2. **GCP Project** with the Eventarc API enabled (the module enables it if needed)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs; the trigger's `service_account` needs `roles/eventarc.eventReceiver` (plus `roles/run.invoker` for authenticated Cloud Run destinations)

## Module Files

| File | Purpose |
|---|---|
| `variables.tf` | GENERATED from the proto spec (`planton tofu generate-variables GcpEventarcTrigger`) — never hand-edited |
| `locals.tf` | Project fallback + name derivations + count guards + label merge |
| `main.tf` | API enablement + channel + google-channel-config + trigger |
| `outputs.tf` | Stack outputs |
| `provider.tf` | google provider pin (`~> 8.3`) |
| `backend.tf` | Local state backend (the runner injects the real backend) |

## How the module maps the spec

| Spec field | Provider argument | Notes |
|---|---|---|
| `trigger_name` | `name` | Defaults to metadata.name; ForceNew |
| `location` | `location` | Required; a region or `global`; ForceNew |
| `matching_criteria[]` | `matching_criteria` | A `type` criterion is required (API rule, spec-enforced) |
| `destination.cloud_run_service.*` | `destination.cloud_run_service.*` | `region` sent only when set (Optional+Computed) |
| `destination.gke.*` | `destination.gke.*` | |
| `destination.workflow` | `destination.workflow` | References a GcpWorkflow's `workflow_id` output |
| `destination.http_endpoint.uri` | `destination.http_endpoint.uri` | HTTPS-only |
| `destination.http_endpoint.network_attachment` | `destination.network_config.network_attachment` | The spec carries it inside the arm; the provider's sibling block shape is restored here |
| `service_account` | `service_account` | References a GcpServiceAccount's `email` output |
| `transport_pubsub_topic` | `transport.pubsub.topic` | messagePublished triggers only |
| `event_data_content_type` | `event_data_content_type` | API default `application/json` when unset |
| `retry_max_attempts` | `retry_policy.max_attempts` | Only value 1 (provider truth); Cloud Run destinations only |
| `partner_channel.*` | `google_eventarc_channel` + trigger `channel` | Count-gated companion; trigger wired to the created channel's full name (assembled from its computed project) |
| `google_channel_crypto_key` | `google_eventarc_google_channel_config.crypto_key_name` | Count-gated singleton; the ambient-project case reads the count-gated `data.google_client_config` |
| `labels` | `labels` | Platform attribution labels win; applied to trigger and channel |
| `project_id` | `project` | `null` when empty — the provider's default project applies |
| `deletion_policy` | `deletion_policy` | One spec lever wired to trigger, channel, and config |

The first trigger in a project provisions Eventarc's service agent — the
first delivery can lag a few minutes behind the apply.

## Outputs

| Output | Description |
|---|---|
| `trigger_name` | The trigger name in GCP |
| `trigger_id` | Full trigger resource name — the canonical API handle |
| `partner_channel_activation_token` | Partner triggers only: the one-time partner handshake token (sensitive); empty otherwise |

## Offline validation

```bash
tofu init -backend=false
tofu plan   # against a tfvars converted from e2e/manifest.yaml
```
