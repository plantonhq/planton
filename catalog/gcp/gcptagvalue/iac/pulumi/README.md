# GCP Tag Value - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying one Resource Manager tag value using Planton's `GcpTagValue` API. The module is written in Go and creates `tags.TagValue` — the value half of a tag, under its key — through the bridged Google provider.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: `roles/resourcemanager.tagAdmin` on the key's owner (organization or project)

## Directory Structure

```
iac/pulumi/
├── main.go            # Pulumi program entry point
├── Pulumi.yaml        # Pulumi project configuration
├── README.md          # This file
└── module/
    ├── main.go        # Module coordinator
    ├── tag_value.go   # The value
    ├── locals.go      # Resolved short name
    └── outputs.go     # Stack output constants
```

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --stack org/project/stack
planton pulumi destroy --manifest ../../e2e/manifest.yaml --stack org/project/stack
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`.

## What the module does

- Creates the value under the spec's `tag_key` (a `GcpTagKey`'s `name` output, `tagKeys/{id}`), with the short name defaulting to `metadata.name`.
- Sends the description and deletion policy only when set.
- Exports `name` (`tagValues/{id}`), `namespaced_name`, `tag_value_id` (derived from the name), and `create_time`.

## Parity with Terraform

Both engines send the same arguments and produce the same four outputs. There is no `PARITY-EXCEPTION` in this module.

## Notes

- **The key and the short name are immutable**; only the description updates in place.
- **A value with bindings cannot be deleted** — destroy the `GcpTagBinding`s first.
