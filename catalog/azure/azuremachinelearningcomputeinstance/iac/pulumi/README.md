# AzureMachineLearningComputeInstance Pulumi Module

## Overview

Creates a compute instance on an Azure Machine Learning workspace using the classic `pulumi-azure` (azurerm-bridged) SDK, from the kind's typed IaC input.

## Design Decisions

- **Wire map identical to the Terraform module**: the same identity-type enum map, the same omit-when-unset semantics for the default-true booleans (`local_auth_enabled`, `node_public_ip_enabled`), the optional subnet, and the SSH block's absent-means-disabled contract.
- **NO update path**: every argument is ForceNew on the provider, tags included -- any change replaces the instance (recorded loudly on the spec).
- **Workspace-region only**: the SDK has no location argument -- the instance always runs in its workspace's region.
- **Provider builder**: credentials resolve through the shared `pulumiazureprovider` builder (static client secret, keyless web identity, or ambient chain).

## Inputs

The module consumes `AzureMachineLearningComputeInstanceIacInput`: the target resource (metadata + spec) and the Azure provider configuration. The workspace and subnet references arrive pre-resolved; `GetValue()` returns the literal ARM ID.

## Outputs

- `machine_learning_compute_instance_id`, `machine_learning_compute_instance_name`
- `system_assigned_identity_principal_id` -- for storage / Key Vault grants, when a system identity is enabled
- `ssh_username`, `ssh_port` -- assigned by the service, populated only when the ssh block is configured

## Local Development

```shell
# compile the module
go build ./...
```
