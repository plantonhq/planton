# GCP Tag Key - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying one Resource Manager tag key using Planton's `GcpTagKey` API. The module is written in Go and creates `tags.TagKey` — the name half of a tag, owned by the organization or one project — through the bridged Google provider.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: `roles/resourcemanager.tagAdmin` on the owner (organization or project)

## Directory Structure

```
iac/pulumi/
├── main.go            # Pulumi program entry point
├── Pulumi.yaml        # Pulumi project configuration
├── README.md          # This file
└── module/
    ├── main.go        # Module coordinator
    ├── tag_key.go     # The key
    ├── locals.go      # Resolved short name and rendered parent
    └── outputs.go     # Stack output constants
```

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --stack org/project/stack
planton pulumi destroy --manifest ../../e2e/manifest.yaml --stack org/project/stack
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`.

## What the module does

- Renders the owner from the spec's `parent` arm: `organizations/{id}` or `projects/{id}` (a reference to a `GcpProject`).
- Defaults the short name to `metadata.name` when the spec leaves it empty.
- Sends the description, purpose, purpose data, allowed-values regex, and deletion policy only when set.
- Exports `name` (`tagKeys/{id}`), `namespaced_name`, `tag_key_id` (derived from the name), and `create_time`.

## Parity with Terraform

Both engines send the same arguments and produce the same four outputs. There is no `PARITY-EXCEPTION` in this module.

## Notes

- **Almost everything is immutable** — owner, short name, purpose, purpose data; only the description and the regex update in place.
- **A key with values cannot be deleted** — destroy the `GcpTagValue`s first (a chart's dependency order does this when they reference the key).
