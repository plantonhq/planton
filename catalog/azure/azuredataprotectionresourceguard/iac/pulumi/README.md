# AzureDataProtectionResourceGuard Pulumi Module

## Overview

Creates a Data Protection Resource Guard -- the approval gate behind Multi-User Authorization for Azure Backup vaults -- on the classic Pulumi Azure SDK (`pulumi-azure/sdk/v6`), wire-identical to the Terraform module. The guard is a free governance object; its protection comes entirely from living in a scope a DIFFERENT administrator controls than the vaults it guards.

## Resources Created

- `dataprotection.ResourceGuard` -- the guard

## Outputs

- `resource_guard_id` -- the guard's full ARM ID; what backup vaults reference to enable Multi-User Authorization
- `resource_guard_name` -- the guard's name

## Behavior Notes

- **An empty exclusion list guards everything** -- the strongest posture, and the module omits the argument entirely so the provider's optional-argument semantics hold exactly. (The SDK's bridged argument name is `VaultCriticalOperationExclusionLists` -- the trailing s is the bridge's pluralization of the provider's already-plural `vault_critical_operation_exclusion_list`, not a different argument.)
- **The exclusion list updates in place**; name, resource group, and location replace the guard (the provider's ForceNew).
- **The guard deletes cleanly even while vaults reference it** -- and the approval requirement goes with it. Gate guard deletion with resource locks or pipeline approvals; Azure does not gate the guard with itself.
- **Engine parity**: the classic SDK v6.38.0 carries the FULL azurerm v5 surface for this kind -- zero parity exceptions.

## Required Permissions

Least-privilege runner permissions for this kind are declared in [`../permissions.yaml`](../permissions.yaml).
