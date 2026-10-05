# GcpSccBigQueryExport — Pulumi Implementation

This directory contains the Pulumi implementation for a Security Command Center BigQuery export from the Planton spec: one of `gcp.securitycenter.V2ProjectSccBigQueryExport`, `V2FolderSccBigQueryExport`, or `V2OrganizationSccBigQueryExport`, selected by the scope, plus `gcp.projects.Service` for a project export.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and the resource function |
| `module/locals.go` | IaC input holder |
| `module/big_query_export.go` | The scope switch, project resolution and API enablement, the resource, the outputs |
| `module/outputs.go` | Output key constants (`name`, `principal`) |

## Send Posture (parity with Terraform)

- **Scope** -- a switch creates exactly one resource; the project arm or an empty scope (the provider's project from `organizations.GetClientConfig`) is the project resource.
- **`Location`** -- `spec.location`, defaulting to `global`.
- **`Dataset`** -- trimmed to `projects/{project}/datasets/{dataset}`.
- **`Filter`** -- when set.
- **`Name`** (organization resource) -- always composed from the organization, location, and ID.
- **`Description`**, **`DeletionPolicy`** -- only when set.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
