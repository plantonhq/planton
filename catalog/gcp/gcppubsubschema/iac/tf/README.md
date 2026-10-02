# GcpPubSubSchema - Terraform Module

This Terraform/OpenTofu module provisions a Pub/Sub schema (`google_pubsub_schema`) — the message contract publishers and subscribers agree on. It is the Terraform-side implementation of the Planton `GcpPubSubSchema` resource kind and has feature parity with the Pulumi module.

## Overview

One schema can validate messages on many topics: each topic attaches it by reference (`schemaSettings.schema`), so the event contract is evolved in one place. Definition changes commit a new schema revision in place (a schema retains up to 20); only renaming replaces the resource.

The module enables the Pub/Sub API with `disable_on_destroy=false` so a fresh project works first try and teardown never disables the API project-wide. An empty `project_id` falls back to the provider's default project, identically to the Pulumi module. The schema API has no labels surface, so no platform attribution labels are stamped.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu plan --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu apply --manifest ../../e2e/manifest.yaml --module-dir . --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --module-dir . --auto-approve
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Module Layout

- `provider.tf` — google provider pin (`~> 8.3`; all fields GA on the released line)
- `variables.tf` — the converter-contract `metadata`/`spec` variables
- `locals.tf` — ambient-project resolution
- `main.tf` — API enablement + the schema resource
- `outputs.tf` — stack outputs (must match `outputs.proto`)

## Inputs

| Variable | Description |
|----------|-------------|
| `spec.project_id` | GCP project (empty = provider default project) |
| `spec.schema_name` | Schema resource name (ForceNew) |
| `spec.type` | `AVRO` or `PROTOCOL_BUFFER` |
| `spec.definition` | The schema text; changes commit revisions in place |

## Outputs

| Name | Description |
|------|-------------|
| `schema_id` | Fully qualified schema path (`projects/{p}/schemas/{name}`) — the exact string a topic's `schemaSettings.schema` reference consumes |
| `schema_name` | The short name of the schema |

## Lifecycle Notes

`schema_name` and `project` are ForceNew; `definition` updates commit a new revision in place — evolving the contract never recreates the schema or touches attached topics. Keep revisions additive (attached topics accept any available revision) and keep the declared `type` stable. Detach or recreate topics before destroying a schema they reference: publishes fail once the topic points at a deleted schema.
