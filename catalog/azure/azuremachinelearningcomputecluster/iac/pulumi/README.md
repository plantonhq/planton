# AzureMachineLearningComputeCluster Pulumi Module

## Overview

Creates an auto-scaling compute cluster on an Azure Machine Learning workspace using the classic `pulumi-azure` (azurerm-bridged) SDK, from the kind's typed IaC input.

## Design Decisions

- **Wire map identical to the Terraform module**: the same VM-priority and identity-type enum maps, the same omit-when-unset semantics for the default-true booleans (`local_auth_enabled`, `node_public_ip_enabled`) and the optional subnet.
- **The region split**: `Location` is the NODES' region and may differ from the workspace's; the provider writes the cluster envelope at the workspace's region (recorded on the spec's region field).
- **SSH credentials**: the admin password is sensitive on the SDK schema; `key_value` carries the SSH PUBLIC key. At least one is required when the block is present (spec CEL mirrors the provider's AtLeastOneOf).
- **Provider builder**: credentials resolve through the shared `pulumiazureprovider` builder (static client secret, keyless web identity, or ambient chain).

## Inputs

The module consumes `AzureMachineLearningComputeClusterIacInput`: the target resource (metadata + spec) and the Azure provider configuration. The workspace and subnet references arrive pre-resolved; `GetValue()` returns the literal ARM ID.

## Outputs

- `machine_learning_compute_cluster_id`, `machine_learning_compute_cluster_name`
- `system_assigned_identity_principal_id` -- for storage / Key Vault / ACR grants, when a system identity is enabled

## Local Development

```shell
# compile the module
go build ./...
```
