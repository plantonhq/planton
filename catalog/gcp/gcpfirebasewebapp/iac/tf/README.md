# GcpFirebaseWebApp - Terraform Module

This Terraform module registers a web app in a Firebase-enabled Google Cloud project (`google_firebase_web_app`) and composes the app's App Check surface: reCAPTCHA v3 (`google_firebase_app_check_recaptcha_v3_config`), reCAPTCHA Enterprise (`google_firebase_app_check_recaptcha_enterprise_config`), and debug tokens (`google_firebase_app_check_debug_token`), with the App Check API enabled as plumbing. It then reads the app's `firebaseConfig` for the front-end build. It is the Terraform-side implementation of the Planton `GcpFirebaseWebApp` resource kind and has feature parity with the Pulumi module.

## Overview

A web app has no identity beyond its display name, so nothing on the registration forces replacement. `deletion_policy` DELETE posts `:remove` with `immediate=true` — the app is gone PERMANENTLY at once, skipping Firebase's 30-day recoverable window. The spec's `deletion_policy` governs the app and every debug token; the two reCAPTCHA configurations have no delete on Google's side (per-app singletons the provider only forgets), so they carry none. Both reCAPTCHA configurations may be set at once — that is how a site migrates from v3 to Enterprise without a gap.

Two blocks ride the `google-beta` provider: Google publishes `google_firebase_web_app` and its `google_firebase_web_app_config` lookup only there. The resource attaches `provider = google-beta` under a recorded admission in `pkg/providerparity/admissions/google-beta.yaml`; the beta channel is pinned to the same `~> 7.43` line as `google`, and both provider blocks set `user_project_override = true` (the Firebase Management API needs a quota project on user-credential calls). App Check and API enablement stay on the GA provider.

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
cd catalog/gcp/gcpfirebasewebapp/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpFirebaseWebApp spec | — |

The `spec` object includes: `project_id` (the Firebase-enabled project; empty falls back to the provider default project), `display_name`, `api_key_id` (the key's UID; empty lets Firebase associate or provision one), `app_check` (`recaptcha_v3` with `site_secret` and `token_ttl`; `recaptcha_enterprise` with `site_key` and `token_ttl`; `debug_tokens` keyed by display name), and `deletion_policy` (DELETE/PREVENT/ABANDON — the app and its debug tokens).

`variables.tf` is generated from the proto contract (`planton tofu generate-variables GcpFirebaseWebApp`) and formatted with `tofu fmt`; regenerate it when the spec changes rather than editing by hand.

## Outputs

| Name | Description |
|------|-------------|
| `app_id` | The Firebase-assigned app id (`firebaseConfig.appId`) |
| `name` | The app's full resource name, `projects/{project}/webApps/{app_id}` |
| `api_key_id` | The UID of the API key associated with the app |
| `app_urls` | The URLs Firebase records the app as hosted at |
| `api_key` | `firebaseConfig.apiKey` — the key string the page presents (a client identifier) |
| `auth_domain` | `firebaseConfig.authDomain` |
| `database_url` | `firebaseConfig.databaseURL` (empty without a Realtime Database) |
| `storage_bucket` | `firebaseConfig.storageBucket` (empty without a default bucket) |
| `location_id` | `firebaseConfig.locationId` (empty until finalized) |
| `messaging_sender_id` | `firebaseConfig.messagingSenderId` — the project number a browser client registers with for push |
| `measurement_id` | `firebaseConfig.measurementId` (empty without a linked Google Analytics property) |

## Resources Created

- `google_firebase_web_app` (beta) — the registration
- `google_project_service` for `firebaseappcheck.googleapis.com` (`count`-gated on any App Check resource; `disable_on_destroy = false`)
- `google_firebase_app_check_recaptcha_v3_config` (`count`-gated on `app_check.recaptcha_v3`; `site_secret` is Sensitive)
- `google_firebase_app_check_recaptcha_enterprise_config` (`count`-gated on `app_check.recaptcha_enterprise`)
- `google_firebase_app_check_debug_token` (`for_each` over `app_check.debug_tokens` by display name; `token` is Sensitive)
- `data.google_firebase_web_app_config` (beta; its input is `web_app_id`) — read after the registration (`depends_on`), so offline plans stay credential-free and the conditionally present outputs degrade to `""`

## Notes

- **DELETE is permanent and immediate.** A shipped site's registration should carry `PREVENT`.
- **Every `firebaseConfig` value, the API key included, ships in the page by design**; they are client identifiers, not secrets.
- **`api_key_id` is sent only when set**; the provider reads back the key Firebase associated or provisioned otherwise.
- **Google's own App Check docs place a 30-second wait between a new app and its first App Check configuration.** This module orders the configurations after the registration and adds no artificial delay; the live proof lane confirms whether Firebase's propagation needs one.
