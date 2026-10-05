# GcpRegionNetworkEndpointGroup - Terraform Module

This Terraform module provisions a GCP regional network endpoint group (`google_compute_region_network_endpoint_group`). It is the Terraform-side implementation of the Planton `GcpRegionNetworkEndpointGroup` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates one regional NEG whose target is decided by `network_endpoint_type`. The endpoint type gates which nested block is sent — `cloud_run`, `cloud_function`, or `app_engine` for SERVERLESS; `psc_data` + `psc_target_service` for PSC; `network`/`subnetwork` for PSC/INTERNET/PORTMAP. The spec's CEL rules enforce block coherence before deploy, so the module stays declarative.

The whole resource is immutable (every field is ForceNew). Because an in-use NEG cannot be deleted, a NEG referenced by a backend service should be recreated create-before-destroy. The module runs on the plain `google` provider — every modeled field is GA on the released 6.x line.

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
cd catalog/gcp/gcpregionnetworkendpointgroup/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpRegionNetworkEndpointGroup spec | — |

The `spec` object includes: `region` (required), `network_endpoint_type` (default SERVERLESS), one serverless block (`cloud_run`/`cloud_function`/`app_engine`) or the PSC/internet fields, `project_id` (empty falls back to the provider default project), and `network_endpoint_group_name` (empty defaults to `metadata.name`).

## Outputs

| Name | Description |
|------|-------------|
| `self_link` | Self-link URI — the value a backend service references in `backends[].group` |
| `network_endpoint_group_name` | Name of the NEG in GCP |
| `network_endpoint_type` | The NEG's endpoint type |
| `region` | Region the NEG lives in |
