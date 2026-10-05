# GcpBackendBucket - Terraform Module

This Terraform module provisions a GCP Compute Engine backend bucket. It is the Terraform-side implementation of the Planton `GcpBackendBucket` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates a `google_compute_backend_bucket` — serving a Cloud Storage bucket's objects through an external HTTP(S) load balancer, optionally cached by Cloud CDN — plus one `google_compute_backend_bucket_signed_url_key` per configured signing key.

`name`, `project`, and `load_balancing_scheme` are immutable (ForceNew); everything else — including the origin `bucket_name` — updates in place, which makes blue/green static releases an origin swap rather than a rebuild. Signed-URL key values are secret material and never appear in outputs.

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
cd catalog/gcp/gcpbackendbucket/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpBackendBucket spec | — |

The `spec` object includes: `bucket_name` (required; the origin GCS bucket), `project_id` (empty falls back to the provider default project), `backend_bucket_name` (empty defaults to `metadata.name`), `enable_cdn` + `cdn_policy` (cache mode, TTLs, negative caching, cache keys, stale serving, coalescing), `compression_mode`, `custom_response_headers`, `edge_security_policy`, `load_balancing_scheme`, and up to 3 `signed_url_keys`.

## Outputs

| Name | Description |
|------|-------------|
| `self_link` | Self-link URI — the value URL maps reference |
| `backend_bucket_name` | Name of the backend bucket in GCP |
| `bucket_name` | The origin GCS bucket currently being served |
