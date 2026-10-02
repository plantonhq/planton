# GcpTargetHttpsProxy - Terraform Module

This Terraform module provisions a GCP Compute Engine target HTTPS proxy — global, or regional when `spec.region` is set. It is the Terraform-side implementation of the Planton `GcpTargetHttpsProxy` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates exactly one of `google_compute_target_https_proxy` (global; `spec.region` empty) or `google_compute_region_target_https_proxy` (regional; `spec.region` set) — the TLS-termination node binding a forwarding rule (the VIP) to a URL map (the routing brain). The two resources are count-gated on one `is_regional` local and mirror each other; outputs select whichever was created; the regional resource has no certificate map, QUIC, early-data, or mesh-bind argument, which the spec keeps off that arm. Certificates attach through exactly one of three mechanisms (classic list, Certificate Manager list, or SNI-scale certificate map); SSL policy, QUIC, and TLS early data live here too.

`url_map`, the certificate wiring, `ssl_policy`, `server_tls_policy`, and `quic_override` update in place; name, description, region, keep-alive, `tls_early_data`, and `proxy_bind` are ForceNew.

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
cd catalog/gcp/gcptargethttpsproxy/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpTargetHttpsProxy spec | — |

The `spec` object includes: the required `url_map`, one certificate mechanism (`ssl_certificates` / `certificate_manager_certificates` / `certificate_map`), TLS behavior (`ssl_policy`, `server_tls_policy`, `quic_override`, `tls_early_data`), and the frontend dials (`http_keep_alive_timeout_sec`, `proxy_bind`). Ref fields arrive as plain strings after CLI-side resolution.

## Outputs

| Name | Description |
|------|-------------|
| `self_link` | Self-link URI — the value a forwarding rule references (`regions/{region}` in place of `global` for a regional proxy) |
| `proxy_name` | Name of the proxy in GCP |
| `proxy_id` | Server-assigned numeric ID |
| `fingerprint` | Server-computed fingerprint for concurrency control (empty for a regional proxy) |
| `region` | Region of a regional proxy; empty for global |
