# GcpSslPolicy - Terraform Module

This Terraform module provisions a Compute Engine SSL policy (`google_compute_ssl_policy` when `spec.region` is empty, `google_compute_region_ssl_policy` when set). It is the Terraform-side implementation of the Planton `GcpSslPolicy` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates one SSL policy — the control for which TLS versions and cipher suites a load balancer accepts from clients. Target HTTPS (and SSL) proxies reference the policy's `self_link`; without one, GCP applies its permissive default (minimum TLS 1.0, COMPATIBLE ciphers).

`profile`, `min_tls_version`, and `custom_features` update in place and apply fleet-wide to every referencing proxy; name, project, and description are ForceNew. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

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
cd catalog/gcp/gcpsslpolicy/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpSslPolicy spec | — |

The `spec` object includes: `profile` (COMPATIBLE/MODERN/RESTRICTED/CUSTOM; empty falls through to the API default), `min_tls_version` (TLS_1_0/TLS_1_1/TLS_1_2; empty falls through), `custom_features` (required with — and only valid with — CUSTOM), `region` (empty means global), `project_id` (empty falls back to the provider default project), `ssl_policy_name` (empty defaults to `metadata.name`), and optional `description`.

## Outputs

| Name | Description |
|------|-------------|
| `self_link` | Self-link URI — the value a target HTTPS (or SSL) proxy references in `ssl_policy` |
| `ssl_policy_name` | Name of the SSL policy in GCP |
| `enabled_features` | Cipher suites the policy actually enables, as computed by GCP |
| `region` | Region of a regional policy; empty for global |
