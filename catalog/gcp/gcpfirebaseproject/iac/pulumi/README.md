# GCP Firebase Project - Pulumi Module

## Overview

This directory contains the Pulumi implementation for enabling Firebase on a Google Cloud project using Planton's `GcpFirebaseProject` API. The module is written in Go and creates `firebase.Project` — the one-way project enablement — plus the composed `firebase.StorageDefaultBucket` (when a location is set), `firebase.AppCheckServiceConfig` and `firebase.AppCheckResourcePolicy` (per spec entry), and `projects.Service` for the APIs each needs, then reads the Admin SDK configuration for the project-level outputs.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: `roles/firebase.admin` (or broader) on the target project; `roles/firebaseappcheck.admin` when App Check is configured
5. **Pay-as-you-go plan** (a linked Cloud Billing account) when `default_storage_location` is set — the enablement itself needs no billing

## Directory Structure

```
iac/pulumi/
├── main.go                # Pulumi program entry point
├── Pulumi.yaml            # Pulumi project configuration
├── README.md              # This file
└── module/
    ├── main.go            # Module coordinator (provider with user_project_override)
    ├── firebase_project.go # Enablement, default bucket, App Check, Admin SDK read
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

- Builds the Google provider with `user_project_override` set — the Firebase Management API attributes quota to the caller's project on user-credential calls, and a deploy under plain ADC fails with "requires a quota project" otherwise.
- Enables `firebase.googleapis.com` and `fcm.googleapis.com` (always), then `firebasestorage.googleapis.com` / `firebaseappcheck.googleapis.com` exactly when the spec composes the default bucket / App Check. Nothing is disabled on destroy.
- Creates the enablement (adopting an already-enabled project), the default bucket when a location is set, and one App Check resource per spec entry, each carrying the spec's `deletion_policy`.
- Reads the Admin SDK configuration through the created resource's project output, so the invoke waits on the enablement, and exports `project_id`, `project_number` (the FCM sender id), `display_name`, `database_url`, `storage_bucket`, `location_id`.

## Parity with Terraform

Both engines create the same resources with the same arguments and produce the same six outputs. One packaging asymmetry, not a behavioral one: the Terraform module attaches `provider = google-beta` to the enablement and the default bucket (Google publishes them only there, under a recorded admission), while this module needs no beta declaration — the Pulumi GCP SDK is bridged from the beta provider, so one provider instance serves both channels. There is no `PARITY-EXCEPTION` in this module.

## Notes

- **Enablement is permanent; destroy detaches.** The spec's `deletion_policy` reaches only the composed resources.
- **An empty `enforcement_mode` is OFF** and is left unset so the API records Google's default state exactly.
