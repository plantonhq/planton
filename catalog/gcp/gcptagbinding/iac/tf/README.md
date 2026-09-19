# GcpTagBinding - Terraform Module

This Terraform module attaches one Resource Manager tag value to one resource (`google_tags_tag_binding`, or `google_tags_location_tag_binding` for a regional or zonal resource). It is the Terraform-side implementation of the Planton `GcpTagBinding` resource kind and has feature parity with the Pulumi module.

## Overview

Two provider resources, exactly one per instance, selected by `spec.location`: empty uses the global binding (organizations, folders, projects, and other global resources); set uses the location-scoped binding. The parent is rendered from the spec's `parent` message as Google's full resource name (`//cloudresourcemanager.googleapis.com/projects/{number}`, `.../folders/{id}`, `.../organizations/{id}`, or a literal `resource_name`). Google requires a project's NUMBER: a `GcpProject` reference resolves to it and a numeric literal is used as is; a project ID, or an empty parent meaning the provider's default project, is resolved through one count-gated `data.google_project` read. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

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
cd catalog/gcp/gcptagbinding/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpTagBinding spec | — |

The `spec` object includes: `tag_value` (`tagValues/{id}` or the namespaced form), `parent` (at most one of `project_id` / `folder_id` / `organization_id` / `resource_name`; empty = provider default project), `location` (region or zone; only with `resource_name`), and `deletion_policy` (DELETE/PREVENT/ABANDON).

`variables.tf` is generated from the proto contract (`planton tofu generate-variables GcpTagBinding`) and formatted with `tofu fmt`; regenerate it when the spec changes rather than editing by hand.

## Outputs

| Name | Description |
|------|-------------|
| `name` | `tagBindings/{encoded parent}/{tagValues/id}` — the binding's resource name |
| `parent` | The full resource name the tag is bound to, as sent (the resolved project number when the manifest gave an ID or no parent) |
| `tag_value` | The bound value's resource name, `tagValues/{id}` |

## Resources Created

- `data.google_project` — read only when the project number is not already known (count-gated)
- `google_tags_tag_binding` — when `location` is empty
- `google_tags_location_tag_binding` — when `location` is set

## Notes

- **A binding is replaced, never edited.** Every input is immutable; any change destroys and recreates the binding. Deletion is immediate, with no soft-delete window.
- **One value per key per resource.** Binding a second value of the same key to a resource is rejected by Google; destroy the old binding first.
- **Tags inherit down the hierarchy** — a binding on a folder or project applies to everything beneath it.
