# AzureAiFoundry Pulumi Module

## Overview

Creates an Azure AI Foundry hub using the classic `pulumi-azure`
(azurerm-bridged) SDK, from the kind's typed IaC input.

## Design Decisions

- **Wire map identical to the Terraform module**: the same
  identity-type and isolation-mode enum maps, the same
  send-only-when-true posture for `high_business_impact_enabled`, and
  the same bool-to-"Enabled"/"Disabled" mapping for public network
  access.
- **`aifoundry.Hub` SDK token**: the classic SDK groups the hub under
  its own `aifoundry` package (not `machinelearning`); it creates the
  same ARM workspace-of-kind-Hub object azurerm's `azurerm_ai_foundry`
  does.
- **Encryption key is VERSIONED**: the provider validates the key as
  a versioned Key Vault key URL (the hub's contract -- differs from
  the classic ML workspace's versionless guidance).
- **Provider builder**: credentials resolve through the shared
  `pulumiazureprovider` builder (static client secret, keyless web
  identity, or ambient chain).

## Inputs

The module consumes `AzureAiFoundryIacInput`: the target resource
(metadata + spec) and the Azure provider configuration. All
references arrive pre-resolved; `GetValue()` returns the literal
value.

## Outputs

- `ai_foundry_id`, `ai_foundry_name`, `workspace_guid`,
  `discovery_url`
- `system_assigned_identity_principal_id` -- for key vault / storage
  grants, when a system identity is enabled

## Local Development

```shell
# compile the module
go build ./...
```
