# GcpSharedVpcHost - Terraform Module

This Terraform module enables one project as a Shared VPC host (`google_compute_shared_vpc_host_project`). It is the Terraform-side implementation of the Planton `GcpSharedVpcHost` resource kind and has feature parity with the Pulumi module.

## Overview

One resource plus one count-gated data source. The host is the spec's `project_id` (a reference to a `GcpProject`'s `project_id` output or a literal); when the spec names no project, the provider's default project is read from `google_client_config`, because this resource — unlike most GCP resources — requires an explicit project argument. The data source is count-gated on that one case, so every plan that names its project runs credential-free. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

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
cd catalog/gcp/gcpsharedvpchost/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpSharedVpcHost spec | — |

The `spec` object includes: `project_id` (empty means the provider's default project) and `deletion_policy` (DELETE/PREVENT/ABANDON).

`variables.tf` is generated from the proto contract (`planton tofu generate-variables GcpSharedVpcHost`) and formatted with `tofu fmt`; regenerate it when the spec changes rather than editing by hand.

## Outputs

| Name | Description |
|------|-------------|
| `host_project_id` | The project ID enabled as the host — the resolved value, what a `GcpSharedVpcServiceProject`'s `host_project_id` references |

## Resources Created

- `google_compute_shared_vpc_host_project` — the host flag, with `deletion_policy` sent only when set
- `data.google_client_config` (count-gated) — the provider's resolved project when the spec names none

## Notes

- **The project is immutable**: a different host is a new resource.
- **Disabling a host fails while service projects are attached** — destroy the `GcpSharedVpcServiceProject` resources first (a chart's dependency order does this when they reference the host).
- **`roles/compute.xpnAdmin` is organization-level.** The deploying identity holds it on the organization or a folder above the project, never just on the project.
