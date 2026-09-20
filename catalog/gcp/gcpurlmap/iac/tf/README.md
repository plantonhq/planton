# GcpUrlMap - Terraform Module

This Terraform module provisions a GCP Compute Engine URL map — global, or regional when `spec.region` is set. It is the Terraform-side implementation of the Planton `GcpUrlMap` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates exactly one of `google_compute_url_map` (global; `spec.region` empty) or `google_compute_region_url_map` (regional; `spec.region` set) — the L7 routing brain of an Application Load Balancer. The two resources are count-gated on one `is_regional` local; the regional block mirrors the global one minus the surfaces its API lacks (route-scoped CDN caching, custom error pages, stream-duration limits outside a path matcher's default action, header-driven routing tests), which the spec's CEL walls keep off a regional manifest; outputs select whichever resource was created. Host rules map Host headers to path matchers; path matchers evaluate route rules (priority-ordered) then path rules (longest prefix), then their default; unmatched traffic falls through to the URL map's top-level default.

`name`, `project`, and `region` are immutable (ForceNew); routing tables, header actions, and tests update in place. `route_action` carries the full traffic-management surface at every routing level: weighted splits (with per-backend header actions), URL rewrites, timeout and retry policies, request mirroring, CORS, fault injection, stream-duration limits, and route-scoped CDN cache policies. A client-side `deletion_policy` (DELETE/PREVENT/ABANDON) controls what a destroy may do.

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
cd catalog/gcp/gcpurlmap/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpUrlMap spec | — |

The `spec` object includes: exactly one top-level default target (`default_service`, `default_url_redirect`, or `default_route_action` with weighted backends), optional `host_rules`, `path_matchers`, `header_action`, custom error policies, and routing self-tests.

## Outputs

| Name | Description |
|------|-------------|
| `self_link` | Self-link URI — the value target proxies reference (`regions/{region}` in place of `global` for a regional map) |
| `url_map_name` | Name of the URL map in GCP |
| `map_id` | Server-assigned numeric ID |
| `fingerprint` | Server-computed fingerprint for concurrency control |
| `region` | Region of a regional URL map; empty for global |
