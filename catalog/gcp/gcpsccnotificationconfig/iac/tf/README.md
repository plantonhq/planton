# GcpSccNotificationConfig — Terraform Implementation

This directory contains the Terraform implementation for a Security Command Center notification config from the Planton spec: one of `google_scc_v2_project_notification_config`, `google_scc_v2_folder_notification_config`, or `google_scc_v2_organization_notification_config`, selected by the scope, plus `google_project_service` for a project config.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Scope selection (`is_project` / `is_folder` / `is_org`), the bare folder ID, the `global` location default, the topic |
| `main.tf` | `data.google_client_config` (empty scope only), `google_project_service` (project configs), the three scope resources with `count` |
| `outputs.tf` | `name`, `service_account`, `service_account_member` (`serviceAccount:` + the email, composed identically to the Pulumi module) |

## Send Posture

- **Scope** -- exactly one resource is created; the project arm, or an empty scope (the provider's project from `google_client_config`, provider configuration, no API call), is the project resource -- PARITY with the Pulumi module's switch and `GetClientConfig`.
- **`location`** -- `spec.location`, defaulting to `global`.
- **`streaming_config.filter`** -- always sent (Google requires the block; empty streams everything).
- **`pubsub_topic`** -- sent when set.
- **`description`**, **`deletion_policy`** -- sent only when set.
- **API** -- `securitycenter.googleapis.com` on a project config's project; `disable_on_destroy = false`.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
