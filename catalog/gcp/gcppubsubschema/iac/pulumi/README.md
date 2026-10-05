# GcpPubSubSchema - Pulumi Module

This Pulumi (Go) module provisions a Pub/Sub schema (`pubsub.Schema`) — the message contract publishers and subscribers agree on. It is the Pulumi-side implementation of the Planton `GcpPubSubSchema` resource kind and has feature parity with the Terraform module.

## Overview

One schema can validate messages on many topics: each topic attaches it by reference (`schemaSettings.schema`), so the event contract is evolved in one place. Definition changes commit a new schema revision in place (a schema retains up to 20); only renaming replaces the resource.

The module enables the Pub/Sub API with `disable_on_destroy=false` so a fresh project works first try and teardown never disables the API project-wide. An empty `projectId` falls back to the provider's default project, identically to the Terraform module. The schema API has no labels surface, so no platform attribution labels are stamped.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Pulumi Usage

```bash
cd catalog/gcp/gcppubsubschema/iac/pulumi
make build
pulumi up --stack dev
```

## Module Layout

- `main.go` — entrypoint; loads the IaC input and calls the module
- `module/main.go` — provider setup and resource orchestration
- `module/locals.go` — metadata-derived values
- `module/schema.go` — API enablement + the schema resource + exports
- `module/outputs.go` — output keys (must match `outputs.proto`)

## Outputs

| Name | Description |
|------|-------------|
| `schema_id` | Fully qualified schema path (`projects/{p}/schemas/{name}`) — the exact string a topic's `schemaSettings.schema` reference consumes |
| `schema_name` | The short name of the schema |

## Lifecycle Notes

`schemaName` and `projectId` are ForceNew; `definition` updates commit a new revision in place — evolving the contract never recreates the schema or touches attached topics. Keep revisions additive (attached topics accept any available revision) and keep the declared `type` stable. Detach or recreate topics before destroying a schema they reference: publishes fail once the topic points at a deleted schema.
