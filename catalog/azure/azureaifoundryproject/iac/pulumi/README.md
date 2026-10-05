# AzureAiFoundryProject Pulumi Module

## Overview

Creates an Azure AI Foundry project using the classic `pulumi-azure`
(azurerm-bridged) SDK, from the kind's typed IaC input.

## Design Decisions

- **Wire map identical to the Terraform module**: the same
  identity-type enum map and the same send-only-when-true posture for
  `high_business_impact_enabled`.
- **`aifoundry.Project` SDK token**: the classic SDK groups the
  project under its own `aifoundry` package; it creates the same ARM
  workspace-of-kind-Project object azurerm's
  `azurerm_ai_foundry_project` does.
- **No resource group input** -- the provider derives the group from
  the hub reference (mirrored from the provider contract; the spec
  carries no group field).
- **Provider builder**: credentials resolve through the shared
  `pulumiazureprovider` builder (static client secret, keyless web
  identity, or ambient chain).

## Inputs

The module consumes `AzureAiFoundryProjectIacInput`: the target
resource (metadata + spec) and the Azure provider configuration. The
hub and identity references arrive pre-resolved; `GetValue()` returns
the literal ARM ID.

## Outputs

- `ai_foundry_project_id`, `ai_foundry_project_name`, `project_guid`
- `system_assigned_identity_principal_id` -- for per-team data
  grants, when a system identity is enabled

## Local Development

```shell
# compile the module
go build ./...
```
