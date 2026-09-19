# GCP Tag Binding - Pulumi Module

## Overview

This directory contains the Pulumi implementation for attaching one Resource Manager tag value to one resource using Planton's `GcpTagBinding` API. The module is written in Go and creates `tags.TagBinding` (global resources) or `tags.LocationTagBinding` (regional and zonal resources, selected by `spec.location`) through the bridged Google provider.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: `roles/resourcemanager.tagUser` on the tag value AND on the resource being tagged

## Directory Structure

```
iac/pulumi/
├── main.go            # Pulumi program entry point
├── Pulumi.yaml        # Pulumi project configuration
├── README.md          # This file
└── module/
    ├── main.go        # Module coordinator
    ├── tag_binding.go # The binding (global or location-scoped) and the project-number lookup
    ├── locals.go      # Rendered parent and the resource selector
    └── outputs.go     # Stack output constants
```

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --stack org/project/stack
planton pulumi destroy --manifest ../../e2e/manifest.yaml --stack org/project/stack
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`.

## What the module does

- Renders the parent as Google's full resource name from the spec's `parent` arm: `//cloudresourcemanager.googleapis.com/projects/{number}`, `.../folders/{id}`, `.../organizations/{id}`, or a literal `resource_name`.
- Resolves a project's NUMBER when it is not already known — a `GcpProject` reference and a numeric literal need no lookup; a project ID, or an empty parent meaning the provider's default project, is read once through `organizations.LookupProject`.
- Creates the location-scoped binding when `spec.location` is set, the global binding otherwise; sends `deletion_policy` only when set.
- Exports `name`, `parent` (as sent), and `tag_value`.

## Parity with Terraform

Both engines send the same arguments and produce the same three outputs, and both gate the project-number lookup the same way (`data.google_project` with `count` on the Terraform side). There is no `PARITY-EXCEPTION` in this module.

## Notes

- **A binding is replaced, never edited** — every input is immutable.
- **One value per key per resource** — Google rejects a second value of the same key; destroy the old binding first.
- **Regional and zonal resources need `location`** — Google serves their bindings from a regional endpoint.
