# GcpFirebaseAppleApp - Terraform Module

This Terraform module registers an iOS / macOS app in a Firebase-enabled Google Cloud project (`google_firebase_apple_app`) and composes the app's App Check surface: App Attest (`google_firebase_app_check_app_attest_config`), DeviceCheck (`google_firebase_app_check_device_check_config`), and debug tokens (`google_firebase_app_check_debug_token`), with the App Check API enabled as plumbing. It then reads the app's `GoogleService-Info.plist` for the build. It is the Terraform-side implementation of the Planton `GcpFirebaseAppleApp` resource kind and has feature parity with the Pulumi module.

## Overview

The registration's `bundle_id` is the app's identity in Firebase: it forces replacement, and a project accepts each bundle id once. `deletion_policy` DELETE posts `:remove` with `immediate=true` — the app is gone PERMANENTLY at once, skipping Firebase's 30-day recoverable window. The spec's `deletion_policy` governs the app and every debug token; the two attestation configurations have no delete on Google's side (per-app singletons the provider only forgets), so they carry none.

Two blocks ride the `google-beta` provider: Google publishes `google_firebase_apple_app` and its `google_firebase_apple_app_config` lookup only there. The resource attaches `provider = google-beta` under a recorded admission in `pkg/providerparity/admissions/google-beta.yaml`; the beta channel is pinned to the same `~> 7.43` line as `google`, and both provider blocks set `user_project_override = true` (the Firebase Management API needs a quota project on user-credential calls). App Check and API enablement stay on the GA provider.

What the module does NOT do, by Google's design: upload the APNs authentication key push needs on Apple platforms. Firebase exposes no API for it; it is a console step.

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
cd catalog/gcp/gcpfirebaseappleapp/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpFirebaseAppleApp spec | — |

The `spec` object includes: `project_id` (the Firebase-enabled project; empty falls back to the provider default project), `display_name`, `bundle_id` (immutable), `app_store_id` and `team_id` (sent only when set; `team_id` is required by validation when App Attest or DeviceCheck is configured), `api_key_id` (the key's UID; empty lets Firebase associate or provision one), `app_check` (`app_attest` with its `enabled` switch and `token_ttl`; `device_check` with `key_id`, `private_key`, `token_ttl`; `debug_tokens` keyed by display name), and `deletion_policy` (DELETE/PREVENT/ABANDON — the app and its debug tokens).

`variables.tf` is generated from the proto contract (`planton tofu generate-variables GcpFirebaseAppleApp`) and formatted with `tofu fmt`; regenerate it when the spec changes rather than editing by hand.

## Outputs

| Name | Description |
|------|-------------|
| `app_id` | The Firebase-assigned app id (`GOOGLE_APP_ID`) |
| `name` | The app's full resource name, `projects/{project}/iosApps/{app_id}` |
| `api_key_id` | The UID of the API key associated with the app |
| `config_filename` | `GoogleService-Info.plist` |
| `config_file_contents` | The configuration file, base64-encoded — a build input that ships in the app bundle, not a secret |

## Resources Created

- `google_firebase_apple_app` (beta) — the registration
- `google_project_service` for `firebaseappcheck.googleapis.com` (`count`-gated on any App Check resource; `disable_on_destroy = false`)
- `google_firebase_app_check_app_attest_config` (`count`-gated on `app_check.app_attest` present and not `enabled: false`)
- `google_firebase_app_check_device_check_config` (`count`-gated on `app_check.device_check`; `private_key` is Sensitive)
- `google_firebase_app_check_debug_token` (`for_each` over `app_check.debug_tokens` by display name; `token` is Sensitive)
- `data.google_firebase_apple_app_config` (beta) — read after the registration (`depends_on`), so offline plans stay credential-free

## Notes

- **Identity is immutable; DELETE is permanent and immediate.** A shipped app's registration should carry `PREVENT`.
- **The DeviceCheck key is not the APNs key.** Apple issues them separately; the APNs key is uploaded in the Firebase console.
- **`api_key_id` is sent only when set**; the provider reads back the key Firebase associated or provisioned otherwise.
- **Google's own App Check docs place a 30-second wait between a new app and its first App Check configuration.** This module orders the configurations after the registration and adds no artificial delay; the live proof lane confirms whether Firebase's propagation needs one.
