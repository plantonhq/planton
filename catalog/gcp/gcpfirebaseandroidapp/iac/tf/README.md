# GcpFirebaseAndroidApp - Terraform Module

This Terraform module registers an Android app in a Firebase-enabled Google Cloud project (`google_firebase_android_app`) and composes the app's App Check surface: Play Integrity attestation (`google_firebase_app_check_play_integrity_config`) and debug tokens (`google_firebase_app_check_debug_token`), with the App Check API enabled as plumbing. It then reads the app's `google-services.json` for the build. It is the Terraform-side implementation of the Planton `GcpFirebaseAndroidApp` resource kind and has feature parity with the Pulumi module.

## Overview

The registration's `package_name` is the app's identity in Firebase: it forces replacement, and a project accepts each package name once. `deletion_policy` DELETE posts `:remove` with `immediate=true` — the app is gone PERMANENTLY at once, skipping Firebase's 30-day recoverable window. The spec's `deletion_policy` governs the app and every debug token; the Play Integrity configuration has no delete on Google's side (a per-app singleton the provider only forgets), so it carries none.

Two blocks ride the `google-beta` provider: Google publishes `google_firebase_android_app` and its `google_firebase_android_app_config` lookup only there. The resource attaches `provider = google-beta` under a recorded admission in `pkg/providerparity/admissions/google-beta.yaml`; the beta channel is pinned to the same `~> 8.3` line as `google`, and both provider blocks set `user_project_override = true` (the Firebase Management API needs a quota project on user-credential calls). App Check and API enablement stay on the GA provider.

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
cd catalog/gcp/gcpfirebaseandroidapp/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpFirebaseAndroidApp spec | — |

The `spec` object includes: `project_id` (the Firebase-enabled project; empty falls back to the provider default project), `display_name`, `package_name` (immutable), `sha1_hashes` / `sha256_hashes` (sent only when non-empty), `api_key_id` (the key's UID; empty lets Firebase associate or provision one), `app_check` (`play_integrity` with its `enabled` switch and `token_ttl`; `debug_tokens` keyed by display name), and `deletion_policy` (DELETE/PREVENT/ABANDON — the app and its debug tokens).

`variables.tf` is generated from the proto contract (`planton tofu generate-variables GcpFirebaseAndroidApp`) and formatted with `tofu fmt`; regenerate it when the spec changes rather than editing by hand.

## Outputs

| Name | Description |
|------|-------------|
| `app_id` | The Firebase-assigned app id (`mobilesdk_app_id`) |
| `name` | The app's full resource name, `projects/{project}/androidApps/{app_id}` |
| `api_key_id` | The UID of the API key associated with the app |
| `config_filename` | `google-services.json` |
| `config_file_contents` | The configuration file, base64-encoded — a build input that ships in the APK, not a secret |

## Resources Created

- `google_firebase_android_app` (beta) — the registration
- `google_project_service` for `firebaseappcheck.googleapis.com` (`count`-gated on any App Check resource; `disable_on_destroy = false`)
- `google_firebase_app_check_play_integrity_config` (`count`-gated on `app_check.play_integrity` present and not `enabled: false`)
- `google_firebase_app_check_debug_token` (`for_each` over `app_check.debug_tokens` by display name; `token` is Sensitive)
- `data.google_firebase_android_app_config` (beta) — read after the registration (`depends_on`), so offline plans stay credential-free

## Notes

- **Identity is immutable; DELETE is permanent and immediate.** A shipped app's registration should carry `PREVENT`.
- **Play Integrity needs a SHA-256 fingerprint** in `sha256_hashes`; the API accepts the configuration without one and attestation fails at runtime, so the module does not refuse it.
- **`api_key_id` is sent only when set**; the provider reads back the key Firebase associated or provisioned otherwise.
- **Google's own App Check docs place a 30-second wait between a new app and its first App Check configuration.** This module orders the configurations after the registration and adds no artificial delay; the live proof lane confirms whether Firebase's propagation needs one.
