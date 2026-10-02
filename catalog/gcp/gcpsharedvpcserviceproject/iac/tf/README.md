# GcpSharedVpcServiceProject - Terraform Module

This Terraform module attaches one service project to a Shared VPC host (`google_compute_shared_vpc_service_project`). It is the Terraform-side implementation of the Planton `GcpSharedVpcServiceProject` resource kind and has feature parity with the Pulumi module.

## Overview

One resource. The host is the spec's `host_project_id` (a reference to a `GcpSharedVpcHost`'s `host_project_id` output — referencing the host kind is what orders a chart host-enable → attach), the service project is the spec's `service_project_id` (a `GcpProject`), and `deletion_policy` is sent only when set. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

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
cd catalog/gcp/gcpsharedvpcserviceproject/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpSharedVpcServiceProject spec | — |

The `spec` object includes: `host_project_id`, `service_project_id`, and `deletion_policy` (empty = detach on destroy; `ABANDON` = leave attached — the only value this resource accepts).

`variables.tf` is generated from the proto contract (`planton tofu generate-variables GcpSharedVpcServiceProject`) and formatted with `tofu fmt`; regenerate it when the spec changes rather than editing by hand.

## Outputs

| Name | Description |
|------|-------------|
| `service_project_id` | The attached service project's ID |
| `host_project_id` | The Shared VPC host project's ID |

## Resources Created

- `google_compute_shared_vpc_service_project` — the attachment, with `deletion_policy` sent only when set

## Notes

- **Both projects are immutable**: moving to another host is a detach and an attach.
- **Detaching fails while resources in the service project still use a host subnetwork.**
- **Attaching is not enough by itself**: the service project's deployers and service agents still need `roles/compute.networkUser` on the host's subnetworks — declared with `GcpProjectIamMember` or subnetwork-level IAM, not here.
