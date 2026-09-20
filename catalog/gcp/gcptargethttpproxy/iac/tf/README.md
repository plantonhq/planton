# GcpTargetHttpProxy - Terraform Module

This Terraform module provisions a GCP Compute Engine target HTTP proxy — global, or regional when `spec.region` is set. It is the Terraform-side implementation of the Planton `GcpTargetHttpProxy` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates exactly one of `google_compute_target_http_proxy` (global; `spec.region` empty) or `google_compute_region_target_http_proxy` (regional; `spec.region` set) — the plaintext-HTTP frontend adapter that binds a forwarding rule (the VIP) to a URL map (the routing brain). The two resources are count-gated on one `is_regional` local and mirror each other; outputs select whichever was created. The standard production role is serving the http→https redirect on port 80.

`url_map` is the only mutable field (GCP swaps it in place via `setUrlMap`); everything else, region included, is ForceNew. The regional resource has no `proxy_bind`; the spec keeps it off that arm.

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
cd catalog/gcp/gcptargethttpproxy/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpTargetHttpProxy spec | — |

The `spec` object includes: the required `url_map` (plain string after ref resolution), optional `proxy_name` / `description`, the `EXTERNAL_MANAGED`-only `http_keep_alive_timeout_sec`, and the Traffic Director `proxy_bind`.

## Outputs

| Name | Description |
|------|-------------|
| `self_link` | Self-link URI — the value a forwarding rule references (`regions/{region}` in place of `global` for a regional proxy) |
| `proxy_name` | Name of the proxy in GCP |
| `proxy_id` | Server-assigned numeric ID |
| `fingerprint` | Server-computed fingerprint for concurrency control (empty for a regional proxy) |
| `region` | Region of a regional proxy; empty for global |
