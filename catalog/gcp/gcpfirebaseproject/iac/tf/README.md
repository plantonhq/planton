# GcpFirebaseProject - Terraform Module

This Terraform module enables Firebase on a Google Cloud project (`google_firebase_project`) and provisions the project-level Firebase surface the spec declares: the default Cloud Storage for Firebase bucket (`google_firebase_storage_default_bucket`) and App Check enforcement (`google_firebase_app_check_service_config`, `google_firebase_app_check_resource_policy`), with the Firebase, Cloud Messaging, Storage, and App Check APIs enabled as plumbing. It is the Terraform-side implementation of the Planton `GcpFirebaseProject` resource kind and has feature parity with the Pulumi module.

## Overview

The enablement is a ONE-WAY project singleton: the provider reads the project first and adopts an already-enabled project, and its destroy detaches state and leaves the project enabled — Google has no de-enable. It carries no `deletion_policy` (provider truth); the spec's `deletion_policy` governs the composed default bucket and App Check configurations.

Two resources ride the `google-beta` provider: Google publishes `google_firebase_project` and `google_firebase_storage_default_bucket` only there. Both attach `provider = google-beta` under a recorded admission in `pkg/providerparity/admissions/google-beta.yaml`; the beta channel is pinned to the same `~> 8.3` line as `google`, and both provider blocks set `user_project_override = true` (the Firebase Management API needs a quota project on user-credential calls). App Check and API enablement stay on the GA provider.

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
cd catalog/gcp/gcpfirebaseproject/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpFirebaseProject spec | — |

The `spec` object includes: `project_id` (empty falls back to the provider default project), `default_storage_location` (the default bucket's Cloud Storage location; empty skips the bucket), `app_check` (`service_configs` keyed by service id, `resource_policies` keyed by service id plus target resource), and `deletion_policy` (DELETE/PREVENT/ABANDON — composed resources only).

`variables.tf` is generated from the proto contract (`planton tofu generate-variables GcpFirebaseProject`) and formatted with `tofu fmt`; regenerate it when the spec changes rather than editing by hand.

## Outputs

| Name | Description |
|------|-------------|
| `project_id` | The GCP project Firebase is enabled on |
| `project_number` | The project number — the Firebase Cloud Messaging sender id |
| `display_name` | The project's Firebase display name |
| `database_url` | The default Realtime Database URL, when the project has one (empty otherwise) |
| `storage_bucket` | The default Cloud Storage for Firebase bucket name, when the project has one (empty otherwise) |
| `location_id` | The project's default GCP resource location, once finalized (empty until then) |

## Resources Created

- `google_project_service` × 2 (always: `firebase.googleapis.com`, `fcm.googleapis.com`) plus `firebasestorage.googleapis.com` / `firebaseappcheck.googleapis.com` exactly when the spec composes those resources — `disable_on_destroy = false` on every one
- `google_firebase_project` (beta) — the enablement
- `google_firebase_storage_default_bucket` (beta, `count`-gated on `default_storage_location`)
- `google_firebase_app_check_service_config` (`for_each` over `app_check.service_configs` by service id)
- `google_firebase_app_check_resource_policy` (`for_each` over `app_check.resource_policies`)
- `data.google_firebase_admin_sdk_config` (beta) — read after the enablement (`depends_on`), so offline plans stay credential-free and the three Admin SDK outputs degrade to `""` when the project lacks the corresponding resource

## Notes

- **Enablement is permanent; destroy detaches.** `deletion_policy` reaches only the default bucket (DELETE removes it with its objects) and the App Check configurations (DELETE switches enforcement OFF).
- **An empty `enforcement_mode` is OFF** and is sent as `null`, so the API records Google's unset state exactly.
- **The default bucket needs the pay-as-you-go plan** (a linked Cloud Billing account); the create fails without it.
