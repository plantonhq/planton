# GCP Logging Sink - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a Cloud Logging sink using Planton's `GcpLoggingSink` API. The module is written in Go and creates exactly one of `logging.ProjectSink`, `logging.FolderSink`, `logging.OrganizationSink`, or `logging.BillingAccountSink` based on the spec's `scope` — one kind, four provider resources.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs. Folder/org/billing sinks require the caller to hold the permissions AT that scope.

## Directory Structure

```
iac/pulumi/
├── main.go                    # Pulumi program entry point
├── Pulumi.yaml                # Pulumi project configuration
├── README.md                  # This file
└── module/
    ├── main.go                # Module coordinator
    ├── logging_sink.go        # Scope dispatch + per-scope sink creation
    ├── locals.go              # Resolved resource + destination URI rendering
    └── outputs.go             # Stack output constants
```

## Quick Start

```bash
cd iac/pulumi
pulumi stack init dev
```

Provide a `stack-input.yaml`:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpLoggingSink
  metadata:
    name: error-archive
  spec:
    destination:
      gcs_bucket:
        value: my-log-archive-bucket
    filter: severity>=ERROR
```

```bash
pulumi preview
pulumi up
```

## Inputs

The module consumes `GcpLoggingSinkStackInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpLoggingSink` spec (scope, destination, filter, exclusions) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `sink_name` | string | The sink name as it exists in GCP |
| `writer_identity` | string | `serviceAccount:{email}` — GRANT THIS write access on the destination or the sink exports nothing |

## Behavior Notes

- **The destination URI is rendered by the module** from whichever arm the spec sets (`storage.googleapis.com/{bucket}`, `bigquery.googleapis.com/projects/{p}/datasets/{d}`, `pubsub.googleapis.com/projects/{p}/topics/{t}`) — manifests reference resources naturally instead of hand-assembling service URIs.
- **The one post-create step every sink needs**: grant `writer_identity` on the destination (`roles/storage.objectCreator`, `roles/bigquery.dataEditor`, or `roles/pubsub.publisher`) through a standalone grant block (`GcpGcsBucketIamMember`, `GcpPubSubTopicIamMember`) whose `member` references `writer_identity` — never the destination kind's own IAM fields, which would form a dependency cycle.
- **Scope differences are modeled, not smoothed over**: writer-identity controls exist only on project sinks (other scopes always mint a unique writer); include/intercept children exist only on folder/org sinks. The spec's validations enforce both.
- **`unique_writer_identity` is sent explicitly** on project sinks so a `true -> false` transition reaches the API.
- **API enablement**: the project-scope path enables `logging.googleapis.com` (with `disable_on_destroy=false`); folder/org/billing sinks are not project resources, so there is no project to enable the API in.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation
