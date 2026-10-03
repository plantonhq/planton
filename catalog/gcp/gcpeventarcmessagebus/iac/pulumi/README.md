# GCP Eventarc Message Bus - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying an Eventarc ADVANCED message bus family using Planton's `GcpEventarcMessageBus` API. The module is written in Go and creates `eventarc.MessageBus` plus its satellites: `eventarc.GoogleApiSource` per spec source (auto-wired to THIS bus), `eventarc.Pipeline` per spec pipeline, and `eventarc.Enrollment` per spec enrollment (wired to the created pipelines by resource reference).

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with the Eventarc API enabled (the module enables it if needed)
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs
6. **Region**: Eventarc Advanced serves a subset of regions — the API rejects unsupported ones at create time

## Directory Structure

```
iac/pulumi/
├── main.go                    # Pulumi program entry point
├── Pulumi.yaml                # Pulumi project configuration
├── README.md                  # This file
└── module/
    ├── main.go                # Module coordinator
    ├── message_bus.go         # Bus + sources + pipelines + enrollments wiring
    ├── locals.go              # Resolved resource + derived values + label merges
    └── outputs.go             # Output constants
```

## How the module maps the spec

| Spec field | Provider argument | Notes |
|---|---|---|
| `message_bus_id` | `message_bus_id` | Defaults to metadata.name; ForceNew |
| `location` | `location` on every resource | Required; ForceNew |
| `display_name`, `annotations` | same names | |
| `crypto_key` | `crypto_key_name` | References a GcpKmsKey's `key_id` output (same-region key) |
| `log_severity` | `logging_config.log_severity` | Provider ValidateEnum values; same wiring on source and pipeline |
| `google_api_sources[]` | `google_eventarc_google_api_source` | `destination` auto-wired to the created bus's computed full name — never hand-assembled |
| `enrollments[]` | `google_eventarc_enrollment` | `message_bus` wired to the created bus; `destination` wired to the sibling pipeline's computed full name (the sibling-id contract is proto-CEL-enforced) |
| `pipelines[].destination.*` | `destinations[0].*` | The API supports exactly one destination per pipeline; `topic`/`workflow` reference GcpPubSubTopic `topic_id` / GcpWorkflow `workflow_id` outputs |
| `pipelines[].destination.http_endpoint.network_attachment` | `destinations[0].network_config.network_attachment` | Required for HTTP endpoints, forbidden otherwise (provider rule, spec-enforced) |
| `pipelines[].authentication.*` | `destinations[0].authentication_config.*` | google_oidc XOR oauth_token |
| `pipelines[].input_payload_format` | `input_payload_format` | avro XOR json XOR protobuf |
| `pipelines[].output_payload_format` | `destinations[0].output_payload_format` | The provider nests output format inside the destination element |
| `pipelines[].mediation_transformation_template` | `mediations[0].transformation.transformation_template` | The API allows at most one mediation |
| `pipelines[].retry_policy.*` | `retry_policy.*` | max_attempts 1–100; delays 1–600s in `Ns` form |
| `labels` | `labels` on every resource | Bus-level merge (platform attribution wins) layered over per-satellite labels |
| `project_id` | `project` | Omitted when empty — the provider's default project applies |
| `deletion_policy` | `deletion_policy` on every resource | One spec lever, every resource |

The module also enables `eventarc.googleapis.com` on the target project
(`disable_on_destroy` false).

## Outputs

| Output | Description |
|---|---|
| `message_bus_name` | Full bus resource name (`projects/{p}/locations/{l}/messageBuses/{id}`) — the cross-bus / external-publisher handle |

## Local development

`iac-input.yaml` carries a ready smoke manifest. Run the module directly:

```bash
planton apply --manifest ../../e2e/manifest.yaml --module-dir .
```
