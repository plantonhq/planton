# GcpServiceNetworkingConnection - Terraform Module

This Terraform module provisions a private services access connection (`google_service_networking_connection`) — the VPC peering between one of your networks and a service producer's network. It is the Terraform-side implementation of the Planton `GcpServiceNetworkingConnection` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates one connection per (network, service) pair. For the default producer (`servicenetworking.googleapis.com`), this peering is what lets Cloud SQL, AlloyDB, Memorystore (PRIVATE_SERVICE_ACCESS mode), and Filestore hand out private IPs from `VPC_PEERING` address ranges reserved inside your VPC.

`reserved_peering_ranges` updates in place — appending ranges grows producer capacity without disturbing already-provisioned service subnets. `network` and `service` are ForceNew. The module enables the Service Networking and Compute Engine APIs so a fresh project works first try, and runs on the plain `google` provider — every modeled field is GA on the released 6.x line.

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
cd catalog/gcp/gcpservicenetworkingconnection/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpServiceNetworkingConnection spec | — |

The `spec` object includes: `network` (name or self-link of the VPC to peer; required, ForceNew), `reserved_peering_ranges` (names of INTERNAL `VPC_PEERING` global address ranges; at least one; mutable), `service` (empty falls through to `servicenetworking.googleapis.com`; ForceNew), `update_on_creation_fail` (adopt a pre-existing connection instead of failing), and `project_id` (scopes in-module API enablement; empty falls back to the provider default project).

## Outputs

| Name | Description |
|------|-------------|
| `peering` | Name of the VPC peering created on the network (e.g. `servicenetworking-googleapis-com`) |
| `network` | The peered VPC network as the connection resolved it |

## Teardown Ordering

GCP refuses to delete the connection while the producer still holds subnets in the reserved ranges — destroy the private-IP service instances (Cloud SQL, AlloyDB, Memorystore, ...) before destroying this resource.
