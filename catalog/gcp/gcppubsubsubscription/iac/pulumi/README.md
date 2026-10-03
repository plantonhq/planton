# GcpPubSubSubscription - Pulumi Module

This Pulumi (Go) module provisions a Pub/Sub subscription (`pubsub.Subscription`) — the named message stream from one topic to one consumer. It is the Pulumi-side implementation of the Planton `GcpPubSubSubscription` resource kind and has feature parity with the Terraform module.

## Overview

The module covers all four delivery modes — pull (default), push with OIDC/no-wrapper, BigQuery (table resolved from a `GcpBigQueryTable` reference), and Cloud Storage with Avro — plus dead-lettering, retry backoff, expiration policy, exactly-once delivery, ordering, filtering, and the ordered JavaScript UDF transform pipeline.

The module enables the Pub/Sub API with `disable_on_destroy=false` so a fresh project works first try and teardown never disables the API project-wide. An empty `projectId` falls back to the provider's default project, identically to the Terraform module. User labels merge beneath the platform attribution labels (`planton-ai_*`), which win on key conflicts — identical label sets on both engines.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Pulumi Usage

```bash
cd catalog/gcp/gcppubsubsubscription/iac/pulumi
make build
pulumi up --stack dev
```

## Module Layout

- `main.go` — entrypoint; loads the IaC input and calls the module
- `module/main.go` — provider setup and resource orchestration
- `module/locals.go` — metadata-derived values + the merged label map
- `module/subscription.go` — API enablement + the subscription resource + exports
- `module/outputs.go` — output keys (must match `outputs.proto`)

## Outputs

| Name | Description |
|------|-------------|
| `subscription_id` | The fully qualified subscription path (`projects/{project}/subscriptions/{name}`) |
| `subscription_name` | The short subscription name |

## Lifecycle Notes

`subscriptionName`, `topic`, `filter`, `enableMessageOrdering`, and `projectId` are ForceNew — changing any of them replaces the subscription and its backlog. Delivery configs, deadlines, retention, dead-lettering, retries, expiration, transforms, and labels all update in place. For dead-lettering, the Pub/Sub service agent needs Subscriber on this subscription and Publisher on the dead-letter topic.
