# GcpApiKey - Terraform Module

This Terraform module provisions a Google Cloud API key (`google_apikeys_key`) with its client restriction and API targets. It is the Terraform-side implementation of the Planton `GcpApiKey` resource kind and has feature parity with the Pulumi module.

## Overview

One resource. The key's identity is the spec's `keyId` (the provider's `name`), its project, and its optional service-account binding — all immutable, so a change to any of them recreates the key and rotates the key string. Restrictions and the display name update in place. The module runs on the plain `google` provider with `user_project_override = true` — every modeled field is GA on the pinned 7.x line, and the API Keys API needs a quota project on user-credential calls.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Terraform Usage

```bash
cd catalog/gcp/gcpapikey/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpApiKey spec | — |

The `spec` object includes: `key_id` (the immutable resource id), `project_id` (empty falls back to the provider default project), `display_name`, `service_account_email` (immutable binding), `restrictions` (at most one of `android_key_restrictions` / `ios_key_restrictions` / `browser_key_restrictions` / `server_key_restrictions`, plus `api_targets`), and `deletion_policy` (DELETE/PREVENT/ABANDON).

`variables.tf` is generated from the proto contract (`planton tofu generate-variables GcpApiKey`) and formatted with `tofu fmt`; regenerate it when the spec changes rather than editing by hand.

## Outputs

| Name | Description |
|------|-------------|
| `name` | `projects/{project}/locations/global/keys/{key_id}` — the key's full resource name (the resource id; the provider's `name` attribute is the short key id) |
| `uid` | The key's unique id — what a Firebase app registration's `api_key_id` references |
| `key_string` | The key string clients present. `sensitive = true`; the Pulumi module exports it as a secret |

## Resources Created

- `google_apikeys_key` — the key, with `restrictions` as dynamic blocks emitted exactly when the spec declares each arm (a hollow arm would be rejected: every arm's list is required by the API)

## Notes

- **Deletion is soft.** `deletion_policy = DELETE` (the provider default) soft-deletes the key: recoverable for 30 days, key id reserved for the window. `PREVENT` fails the destroy; `ABANDON` leaves the key live.
- **Restrictions never rotate the key string** — only the immutable identity fields recreate the key.
- **Quota project.** `user_project_override = true` attributes quota to the key's own project under every credential mode; without it a deploy under plain ADC fails with "requires a quota project" (the Identity Toolkit precedent).
