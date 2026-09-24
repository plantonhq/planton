# GCP API Key - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a Google Cloud API key using Planton's `GcpApiKey` API. The module is written in Go and creates `projects.ApiKey` — the key with its client restriction arm and API targets — through the bridged Google provider with the quota-project override set.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: `roles/serviceusage.apiKeysAdmin` (or broader) on the target project — it includes `apikeys.keys.getKeyString`, which the module needs to export the key string

## Directory Structure

```
iac/pulumi/
├── main.go            # Pulumi program entry point
├── Pulumi.yaml        # Pulumi project configuration
├── README.md          # This file
└── module/
    ├── main.go        # Module coordinator (provider with user_project_override and billing_project)
    ├── api_key.go     # The key and its restrictions
    ├── locals.go      # Resolved resource
    └── outputs.go     # Stack output constants
```

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --stack org/project/stack
planton pulumi destroy --manifest ../../e2e/manifest.yaml --stack org/project/stack
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`.

## What the module does

- Builds the Google provider with `user_project_override` set and `billing_project` naming the resource's project (the data-source reads need the quota project named under a user credential), so the API Keys API attributes quota to the key's own project under every credential mode (a deploy under plain ADC fails with "requires a quota project" otherwise).
- Enables `apikeys.googleapis.com` on the project first (`projects.Service`, left enabled on destroy so one key's removal never switches the API off for the others) — a first apply on a fresh project needs nothing switched on by hand.
- Creates the key named by `spec.key_id` in the spec's project, with the optional display name, service-account binding, and deletion policy, after the API enablement.
- Sends each restriction arm exactly when the spec declares it — an omitted arm means "no restriction of that class" to the API, and a hollow arm would be rejected.
- Exports `name` (the resource id, `projects/{project}/locations/global/keys/{key_id}`), `uid`, and `key_string` — the last as a Pulumi secret, matching the Terraform module's `sensitive = true`.

## Parity with Terraform

Both engines enable the same API, send the same arguments, and produce the same three outputs. There is no `PARITY-EXCEPTION` in this module.

## Notes

- **Every identity field is immutable** — `key_id`, `project_id`, and `service_account_email` recreate the key (and rotate its string) when changed; restrictions and the display name update in place.
- **Deletion is soft** — `DELETE` keeps the key recoverable for 30 days and reserves its id for that window.
