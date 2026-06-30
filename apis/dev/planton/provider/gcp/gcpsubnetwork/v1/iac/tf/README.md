# Terraform Module for GCP Subnetwork

This Terraform module creates a custom-mode GCP subnetwork in an existing VPC network.

## Overview

The module provisions:
- A GCP subnetwork in the specified region
- Optional secondary IP ranges for GKE pods and services
- Private Google Access configuration
- Required GCP APIs (compute.googleapis.com)

## Usage with Planton CLI

### Initialize Terraform

```shell
planton tofu init --manifest hack/manifest.yaml --backend-type gcs \
  --backend-config="bucket=my-terraform-state-bucket" \
  --backend-config="prefix=planton/gcp-stacks/test-gcp-subnetwork"
```

### Plan Changes

```shell
planton tofu plan --manifest hack/manifest.yaml
```

### Apply Changes

```shell
planton tofu apply --manifest hack/manifest.yaml --auto-approve
```

### Destroy Resources

```shell
planton tofu destroy --manifest hack/manifest.yaml --auto-approve
```

## Module Inputs

| Variable | Description | Type | Required |
|----------|-------------|------|----------|
| metadata | Resource metadata (name, id, org, env, labels) | object | Yes |
| spec | Subnetwork specification (project_id, vpc_self_link, region, ip_cidr_range, etc.) | object | Yes |

## Module Outputs

| Output | Description |
|--------|-------------|
| subnetwork_self_link | Self-link URL of the created subnetwork |
| region | The region where the subnetwork resides |
| ip_cidr_range | Primary IPv4 CIDR of the subnet |
| secondary_ranges | List of secondary ranges with names and CIDRs |

## Prerequisites

- GCP project with compute API access
- Existing custom-mode VPC network
- Appropriate IAM permissions to create subnetworks
- Terraform/OpenTofu ≥ 1.5.0

## Notes

- Subnetwork region and CIDR cannot be changed after creation
- GCP reserves 4 IPs per subnet (network, broadcast, gateway, DNS)
- For GKE clusters, define secondary ranges at creation time
- Private Google Access is recommended for internal-only instances
