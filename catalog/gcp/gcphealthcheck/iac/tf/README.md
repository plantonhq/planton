# GcpHealthCheck - Terraform Module

This Terraform module provisions a GCP Compute Engine health check. It is the Terraform-side implementation of the Planton `GcpHealthCheck` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates exactly one of `google_compute_health_check` (global, when `spec.region` is empty) or `google_compute_region_health_check` (regional, when it is set) — GCP models the two scopes as separate API collections with an identical probe surface. Everything, including the gRPC-with-TLS protocol block, is GA on the `hashicorp/google` 8.x line; no beta provider is involved.

`name` and `project` are immutable (ForceNew); all probe knobs (cadence, thresholds, protocol settings) update in place. Ports left unset fall through to the API's protocol defaults (http/tcp 80, https/http2/ssl 443).

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
cd catalog/gcp/gcphealthcheck/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpHealthCheck spec | — |

The `spec` object includes: exactly one protocol block (`http`/`https`/`http2`/`tcp`/`ssl`/`grpc`/`grpc_tls`), `project_id` (empty falls back to the provider default project), `health_check_name` (empty defaults to `metadata.name`), `region` (empty = global), `check_interval_sec`/`timeout_sec`/`healthy_threshold`/`unhealthy_threshold` (defaults 5/5/2/2), `enable_logging`, global-only `source_regions` (exactly 3 regions), and `deletion_policy` (DELETE/PREVENT/ABANDON destroy behavior — applies to whichever scope the check was created in).

## Outputs

| Name | Description |
|------|-------------|
| `self_link` | Self-link URI — the value backend services reference |
| `health_check_name` | Name of the health check in GCP |
| `type` | Probe protocol GCP computed (HTTP, TCP, GRPC, ...) |
| `region` | Region of a regional check; empty for global |
