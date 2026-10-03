# GcpPubSubTopic - Pulumi Module

This Pulumi (Go) module provisions a Pub/Sub topic (`pubsub.Topic`) — the named channel publishers send to. It is the Pulumi-side implementation of the Planton `GcpPubSubTopic` resource kind and has feature parity with the Terraform module.

## Overview

The module covers the full released topic surface: CMEK encryption, topic-level message retention, regional storage policy with in-transit enforcement, schema validation (via a `GcpPubSubSchema` reference), all five ingestion data sources (Kinesis, MSK, Azure Event Hubs, Cloud Storage, Confluent Cloud) with platform logging, and the ordered JavaScript UDF transform pipeline.

The module enables the Pub/Sub API with `disable_on_destroy=false` so a fresh project works first try and teardown never disables the API project-wide. An empty `projectId` falls back to the provider's default project, identically to the Terraform module. User labels merge beneath the platform attribution labels (`planton-ai_*`), which win on key conflicts — identical label sets on both engines.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Pulumi Usage

```bash
cd catalog/gcp/gcppubsubtopic/iac/pulumi
make build
pulumi up --stack dev
```

## Module Layout

- `main.go` — entrypoint; loads the IaC input and calls the module
- `module/main.go` — provider setup and resource orchestration
- `module/locals.go` — metadata-derived values + the merged label map
- `module/topic.go` — API enablement + the topic resource + exports
- `module/outputs.go` — output keys (must match `outputs.proto`)

## Outputs

| Name | Description |
|------|-------------|
| `topic_id` | The fully qualified topic path (`projects/{project}/topics/{name}`) — what subscriptions, event triggers, and Scheduler targets consume |
| `topic_name` | The short topic name |

## Lifecycle Notes

`topicName` and `projectId` are ForceNew — renaming replaces the topic and orphans its subscriptions, so treat names as permanent. Everything else (CMEK key, retention, storage policy, schema settings, ingestion, transforms, labels) updates in place. For CMEK, grant the Pub/Sub service agent `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key before deploying, or publishes fail.
