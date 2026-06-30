# GcpGlobalAddress - Terraform Module

This Terraform module provisions a GCP global address (`google_compute_global_address`). It is the Terraform-side implementation of the Planton `GcpGlobalAddress` resource kind and has feature parity with the Pulumi module.

## Overview

The module reserves a static external or internal IP address (or CIDR range) for use with load balancers, VPC peering, and Private Service Connect. It supports EXTERNAL and INTERNAL address types, IPV4 and IPV6, and optional fields such as network, purpose, and prefix length.

## Usage with Planton CLI

```shell
planton tofu init --manifest hack/manifest.yaml
planton tofu plan --manifest hack/manifest.yaml
planton tofu apply --manifest hack/manifest.yaml --auto-approve
planton tofu destroy --manifest hack/manifest.yaml --auto-approve
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../hack/manifest.yaml`.

## Direct Terraform Usage

```bash
cd apis/dev/planton/provider/gcp/gcpglobaladdress/v1/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `spec` | GcpGlobalAddress spec (project_id, address_name, address_type, ip_version, etc.) | — |
| `provider_config` | GCP provider configuration (service_account_key_base64) | `{}` |
| `labels` | Labels to apply to the global address | `{}` |

The `spec` object includes: `project_id` (object with `value`), `address_name`, `address` (optional), `address_type` (default: `EXTERNAL`), `description` (optional), `ip_version` (default: `IPV4`), `network` (optional), `prefix_length` (optional), `purpose` (optional).

## Outputs

| Name | Description |
|------|-------------|
| `address` | The reserved IP address or start of the reserved range |
| `self_link` | Self-link URL of the global address resource |
| `creation_timestamp` | RFC3339 creation timestamp |
