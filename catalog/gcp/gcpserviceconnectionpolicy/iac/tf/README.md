# GcpServiceConnectionPolicy - Terraform Module

This Terraform module provisions a service connection policy (`google_network_connectivity_service_connection_policy`) — the per-network authorization for Google's service connectivity automation to create PSC endpoints on a producer's behalf. It is the Terraform-side implementation of the Planton `GcpServiceConnectionPolicy` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates one policy per (network, service class, region) triple. PSC-first managed services (Memorystore for Valkey, Redis Cluster) refuse to create instances on a network until a policy for their service class exists in that region — this module is that prerequisite.

`location`, `network`, `service_class`, and the policy name are ForceNew; the `psc_config` contents, description, and labels update in place, so subnet growth and limit raises never recreate the policy. The Service Connectivity API requires relative resource paths — the module normalizes `https://` self-link URLs for `network` and `subnetworks` identically to the Pulumi module. The module enables the Network Connectivity and Compute Engine APIs so a fresh project works first try, and runs on the plain `google` provider — every modeled field is GA on the released 6.x line.

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
cd catalog/gcp/gcpserviceconnectionpolicy/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpServiceConnectionPolicy spec | — |

The `spec` object includes: `location` (region; required, ForceNew), `network` (VPC resource path or self-link; required, ForceNew), `service_class` (the producer's published class; required, ForceNew), `policy_name` (empty falls back to `metadata.name`; ForceNew), `description`, `labels` (merged beneath platform labels), and `psc_config` (subnetworks + optional limit + optional producer allowlist; mutable), plus `project_id` (empty falls back to the provider default project).

## Outputs

| Name | Description |
|------|-------------|
| `policy_id` | Fully qualified policy resource path |
| `name` | Short policy name |
| `infrastructure` | Underlying connectivity mechanism (PSC) |
| `etag` | Server-computed etag |

## Lifecycle Notes

Deploy the policy before the first instance of its service class in the region, and keep it alive as long as any instance depends on it — deleting the policy strands existing endpoints and blocks new ones.
