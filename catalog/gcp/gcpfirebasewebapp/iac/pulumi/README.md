# GCP Firebase Web App - Pulumi Module

## Overview

This directory contains the Pulumi implementation for registering a web app in a Firebase-enabled Google Cloud project using Planton's `GcpFirebaseWebApp` API. The module is written in Go and creates `firebase.WebApp` — the registration — plus the composed `firebase.AppCheckRecaptchaV3Config` and `firebase.AppCheckRecaptchaEnterpriseConfig` (when configured) and `firebase.AppCheckDebugToken` (per spec entry), with `projects.Service` for the App Check API, then reads the app's `firebaseConfig` through `firebase.GetWebAppConfigOutput`.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: `roles/firebase.admin` (or broader) on the target project; `roles/firebaseappcheck.admin` when App Check is configured
5. **Firebase enabled on the project** — a `GcpFirebaseProject` deployed first (the registration exists only inside the enablement)

## Directory Structure

```
iac/pulumi/
├── main.go                # Pulumi program entry point
├── Pulumi.yaml            # Pulumi project configuration
├── README.md              # This file
└── module/
    ├── main.go            # Module coordinator (provider with user_project_override and billing_project)
    ├── web_app.go         # Registration, App Check, config lookup
    ├── locals.go          # Resolved resource
    └── outputs.go         # Stack output constants
```

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --stack org/project/stack
planton pulumi destroy --manifest ../../e2e/manifest.yaml --stack org/project/stack
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`.

## What the module does

- Builds the Google provider with `user_project_override` set and `billing_project` naming the resource's project (the data-source reads need the quota project named under a user credential) — the Firebase Management API attributes quota to the caller's project on user-credential calls, and a deploy under plain ADC fails with "requires a quota project" otherwise.
- Registers the app with its display name, the API key UID (sent only when set — Firebase associates or provisions one otherwise), and the spec's `deletion_policy`.
- Enables `firebaseappcheck.googleapis.com` exactly when the spec composes an App Check resource (never disabled on destroy), then creates the reCAPTCHA v3 configuration (its site secret marked secret in state), the reCAPTCHA Enterprise configuration (its site key is public), and one debug token per entry keyed by display name, each waiting on the registration and the API.
- Reads the app's `firebaseConfig` through the created app's id (the lookup names its input `WebAppId` — the one naming divergence in the family), so the invoke waits on the registration, and exports `app_id`, `name`, `api_key_id`, `app_urls`, and the seven `firebaseConfig` values.

## Parity with Terraform

Both engines create the same resources with the same arguments and produce the same eleven outputs. One packaging asymmetry, not a behavioral one: the Terraform module attaches `provider = google-beta` to the registration and its config lookup (Google publishes them only there, under a recorded admission), while this module needs no beta declaration — the Pulumi GCP SDK is bridged from the beta provider, so one provider instance serves both channels. There is no `PARITY-EXCEPTION` in this module.

## Notes

- **DELETE is permanent and immediate.** A shipped site's registration should carry `PREVENT`.
- **Both reCAPTCHA configurations may coexist** — the v3-to-Enterprise migration shape; a client initialises with one.
- **The attestation configurations are never deleted on Google's side**; removing one from the spec forgets it, and destroying the app removes it with the app.
- **Google's own App Check docs place a 30-second wait between a new app and its first App Check configuration.** This module orders the configurations after the registration and adds no artificial delay; the live proof lane confirms whether Firebase's propagation needs one.
