# GcpWorkloadIdentityPoolProvider - Terraform Module

This Terraform module provisions a GCP Workload Identity Pool Provider (`google_iam_workload_identity_pool_provider`). It is the Terraform-side implementation of the Planton `GcpWorkloadIdentityPoolProvider` resource kind and has feature parity with the Pulumi module.

## Overview

The module attaches one external issuer (OIDC, AWS, SAML, or X.509 — exactly one) to a Workload Identity Pool, with claim-to-attribute mappings and an optional CEL condition gating which credentials are accepted. `workload_identity_pool_id`, `workload_identity_pool_provider_id`, and `project` are immutable (ForceNew), and the issuer type cannot change on a live provider. GCP soft-deletes providers for ~30 days, during which the provider ID cannot be reused (no undelete-on-create) — prefer `disabled = true` for temporary shutoffs.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Terraform Usage

```bash
cd catalog/gcp/gcpworkloadidentitypoolprovider/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpWorkloadIdentityPoolProvider spec | — |

The `spec` object includes: `workload_identity_pool_id` (plain string — the CLI resolves the pool reference before the module runs), `workload_identity_pool_provider_id` (4-32 chars; lowercase, digits, hyphens; `gcp-` reserved), `project_id` (empty falls back to the provider default project), optional `display_name`, `description`, `disabled` (default `false`), `attribute_mapping` (required for OIDC — must include `google.subject`), `attribute_condition`, and exactly one of `aws` / `oidc` / `saml` / `x509`.

## Outputs

| Name | Description |
|------|-------------|
| `name` | Full provider resource name — the audience string for token exchange |
| `workload_identity_pool_provider_id` | The bare provider ID |
| `state` | `ACTIVE`, or `DELETED` while soft-deleted |
