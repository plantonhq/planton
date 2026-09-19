# GCP Folder - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a Google Cloud Resource Manager folder using Planton's `GcpFolder` API. The module is written in Go and creates `organizations.Folder` — the hierarchy node projects and sub-folders are placed inside — through the bridged Google provider.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: `roles/resourcemanager.folderAdmin` (or `folderCreator` plus `folderEditor`) on the parent organization or folder

## Directory Structure

```
iac/pulumi/
├── main.go            # Pulumi program entry point
├── Pulumi.yaml        # Pulumi project configuration
├── README.md          # This file
└── module/
    ├── main.go        # Module coordinator
    ├── folder.go      # The folder
    ├── locals.go      # Resolved display name and rendered parent
    └── outputs.go     # Stack output constants
```

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --stack org/project/stack
planton pulumi destroy --manifest ../../e2e/manifest.yaml --stack org/project/stack
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`.

## What the module does

- Renders the parent from the spec's `parent` arm: `organizations/{id}` for a top-level folder, `folders/{id}` for a nested one (a reference to another `GcpFolder`'s `folder_id` output).
- Defaults the display name to `metadata.name` when the spec leaves it empty.
- Always sends `deletion_protection` (true when the spec is silent), so the destroy guard's state is the spec's, never an implicit provider default.
- Sends the create-time `tags` and `deletion_policy` only when set.
- Exports `folder_id`, `name`, `lifecycle_state`, and `create_time`.

## Parity with Terraform

Both engines send the same arguments and produce the same four outputs. There is no `PARITY-EXCEPTION` in this module.

## Notes

- **The destroy guard is on by default** — a destroy fails until `deletion_protection` is set to false and applied first; `ABANDON` bypasses it, `PREVENT` overrides both.
- **Changing the parent moves the folder in place**; changing the create-time tags recreates it, which Google refuses for a non-empty folder — use `GcpTagBinding` for tags on an existing folder.
- **Deletion is soft** — 30-day recovery window, display name reserved under the same parent for that window.
