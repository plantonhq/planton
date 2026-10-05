# GcpAddress - Terraform Module

This Terraform module provisions a GCP regional address (`google_compute_address`). It is the Terraform-side implementation of the Planton `GcpAddress` resource kind and has feature parity with the Pulumi module.

## Overview

The module reserves a static external or internal IP address (or CIDR range) at regional scope for use with Cloud NAT, regional load balancers, VM instances, internal LB VIPs, VPC peering, and IPsec interconnect. It supports EXTERNAL and INTERNAL address types, IPV4 and IPV6, and optional fields such as network, subnetwork, purpose, prefix length, and network tier.

## Usage with Planton CLI

```shell
planton tofu init --manifest e2e/manifest.yaml
planton tofu plan --manifest e2e/manifest.yaml
planton tofu apply --manifest e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest e2e/manifest.yaml --auto-approve
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Terraform Usage

```bash
cd catalog/gcp/gcpaddress/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `spec` | GcpAddress spec (project_id, address_name, region, address_type, etc.) | — |
| `metadata` | Resource metadata (name, org, env, labels) | — |

The `spec` object includes: `project_id`, `address_name`, `region` (required), `address` (optional), `address_type` (default: `EXTERNAL`), `description`, `ip_version` (default: `IPV4`), `network`, `subnetwork`, `network_tier`, `prefix_length`, `purpose`, `ipv6_endpoint_type`.

## Outputs

| Name | Description |
|------|-------------|
| `address` | The reserved IP address or start of the reserved range |
| `self_link` | Self-link URL of the regional address resource |
| `name` | Name of the regional address resource in GCP |
| `region` | Plain region name from the spec (e.g. `us-central1`) |
