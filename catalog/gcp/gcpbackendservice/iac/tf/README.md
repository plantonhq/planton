# GcpBackendService - Terraform Module

This Terraform module provisions a GCP Compute Engine backend service — global, or regional when `spec.region` is set. It is the Terraform-side implementation of the Planton `GcpBackendService` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates exactly one of `google_compute_backend_service` (global; `spec.region` empty) or `google_compute_region_backend_service` (regional; `spec.region` set) — the hub of the load balancing family, owning backends, health checking, session affinity, Cloud CDN policy, IAP, Cloud Armor attachment, and logging — plus, on the global arm, one `google_compute_backend_service_signed_url_key` per configured signing key. The two resources are count-gated on one `is_regional` local; the regional block mirrors the global one plus the passthrough Network Load Balancer policies (network, failover, connection tracking, HA policy, zonal affinity) and minus the surfaces its API lacks, which the spec's CEL walls keep off a regional manifest; outputs select whichever resource was created.

`name`, `project`, and `region` are immutable (ForceNew); everything else — backends, CDN policy, affinity, IAP — updates in place, which makes this node the operational lever of a running load balancer. IAP client secrets, AWS SigV4 access keys, and signed-URL key values are secret material and never appear in outputs.

An unset `load_balancing_scheme` is sent as `EXTERNAL` (the spec's default, the classic global external ALB) rather than left to the provider, whose own default is `EXTERNAL_MANAGED`: the scheme is immutable, so a provider-chosen default would replace an existing classic backend service on its next apply. The Pulumi module does the same.

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
cd catalog/gcp/gcpbackendservice/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpBackendService spec | — |

The `spec` object includes: `project_id` (empty falls back to the provider default project), `backend_service_name` (empty defaults to `metadata.name`), `protocol` / `load_balancing_scheme` / `port_name` / `timeout_sec` / `connection_draining_timeout_sec`, `health_check` (the single health-check self-link; empty only for internet/serverless NEG backends), `backends` (group + balancing mode + capacity dials + per-backend custom metrics), session affinity (all modes incl. the strong-affinity cookie), `locality_lb_policy` / `locality_lb_policies` / `consistent_hash`, `enable_cdn` + `cdn_policy` (with the full cache-key policy), `security_policy` / `edge_security_policy`, `iap`, `log_config`, custom request/response headers, `compression_mode`, `circuit_breakers` / `outlier_detection` / `max_stream_duration`, `security_settings` (incl. AWS SigV4), `tls_settings`, `ip_address_selection_policy`, the EXTERNAL→EXTERNAL_MANAGED migration controls, service-level `custom_metrics`, `service_lb_policy`, and up to 3 `signed_url_keys`.

## Outputs

| Name | Description |
|------|-------------|
| `self_link` | Self-link URI — the value URL maps and passthrough forwarding rules reference (`regions/{region}` in place of `global` for a regional service) |
| `backend_service_name` | Name of the backend service in GCP |
| `generated_id` | Server-assigned numeric ID |
| `fingerprint` | Optimistic-concurrency fingerprint |
| `region` | Region of a regional backend service; empty for global |
