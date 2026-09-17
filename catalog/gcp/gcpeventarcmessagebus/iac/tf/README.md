# GCP Eventarc Message Bus - Terraform Module

## Overview

This directory contains the Terraform/OpenTofu implementation for deploying an Eventarc ADVANCED message bus family using Planton's `GcpEventarcMessageBus` API. The module creates `google_eventarc_message_bus` plus its satellites: `google_eventarc_google_api_source` per spec source (auto-wired to THIS bus), `google_eventarc_pipeline` per spec pipeline, and `google_eventarc_enrollment` per spec enrollment (wired to the created pipelines by resource reference).

## Prerequisites

1. **OpenTofu** (or Terraform >= 1.5)
2. **GCP Project** with the Eventarc API enabled (the module enables it if needed)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs
5. **Region**: Eventarc Advanced serves a subset of regions — the API rejects unsupported ones at create time

## Module Files

| File | Purpose |
|---|---|
| `variables.tf` | GENERATED from the proto spec (`planton tofu generate-variables GcpEventarcMessageBus`) — never hand-edited |
| `locals.tf` | Project fallback + bus-id derivation + label merge |
| `main.tf` | API enablement + bus + sources + pipelines + enrollments |
| `outputs.tf` | Stack outputs |
| `provider.tf` | google provider pin (`~> 8.3`) |
| `backend.tf` | Local state backend (the runner injects the real backend) |

## How the module maps the spec

| Spec field | Provider argument | Notes |
|---|---|---|
| `message_bus_id` | `message_bus_id` | Defaults to metadata.name; ForceNew |
| `location` | `location` on every resource | Required; ForceNew |
| `crypto_key` | `crypto_key_name` | References a GcpKmsKey's `key_id` output (same-region key) |
| `log_severity` | `logging_config.log_severity` | Same wiring on source and pipeline |
| `google_api_sources[]` | `google_eventarc_google_api_source` | `destination` auto-wired to the created bus's computed full name |
| `enrollments[]` | `google_eventarc_enrollment` | `message_bus` + `destination` wired by resource reference (sibling-id contract is spec-enforced) |
| `pipelines[].destination.*` | `destinations[0].*` | Exactly one destination per pipeline (API truth) |
| `pipelines[].destination.http_endpoint.network_attachment` | `destinations[0].network_config.network_attachment` | Required for HTTP endpoints, forbidden otherwise |
| `pipelines[].authentication.*` | `destinations[0].authentication_config.*` | google_oidc XOR oauth_token |
| `pipelines[].input_payload_format` | `input_payload_format` | avro XOR json XOR protobuf |
| `pipelines[].output_payload_format` | `destinations[0].output_payload_format` | The provider nests output format inside the destination element |
| `pipelines[].mediation_transformation_template` | `mediations[0].transformation.transformation_template` | At most one mediation (API truth) |
| `pipelines[].retry_policy.*` | `retry_policy.*` | max_attempts 1–100; delays 1–600s in `Ns` form |
| `labels` | `labels` on every resource | Bus-level merge layered over per-satellite labels |
| `project_id` | `project` | `null` when empty — the provider's default project applies |
| `deletion_policy` | `deletion_policy` on every resource | One spec lever, every resource |

## Outputs

| Output | Description |
|---|---|
| `message_bus_name` | Full bus resource name — the cross-bus / external-publisher handle |

## Offline validation

```bash
tofu init -backend=false
tofu plan   # against a tfvars converted from e2e/manifest.yaml
```
