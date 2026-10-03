# AzureCognitiveAccountProject Pulumi Module

## Overview

Provisions an AI Foundry project on an Azure AI services account using the classic `pulumi-azure` (azurerm-bridged) SDK, from the kind's typed IaC input.

## Design Decisions

- **Wire map identical to the Terraform module**: the same identity-type enum map and the same omit-when-empty semantics for `description` / `display_name` (ARM cannot clear either in place -- clearing replaces the project).
- **Identity always emitted**: the provider requires it; user-assigned IDs resolve through their StringValueOrRef values.
- **Provider builder**: credentials resolve through the shared `pulumiazureprovider` builder (static client secret, keyless web identity, or ambient chain).

## Inputs

The module consumes `AzureCognitiveAccountProjectIacInput`: the target resource (metadata + spec) and the Azure provider configuration. `cognitive_account_id` arrives pre-resolved; `GetValue()` returns the literal ARM ID.

## Outputs

- `project_id`, `project_name`
- `endpoints` -- the data-plane endpoints map
- `is_default` -- whether ARM made this the account's default project
- `system_assigned_identity_principal_id`

## Local Development

```shell
# compile the module
go build ./...
```
