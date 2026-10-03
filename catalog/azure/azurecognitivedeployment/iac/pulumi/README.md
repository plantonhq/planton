# AzureCognitiveDeployment Pulumi Module

## Overview

Provisions a model deployment on an Azure AI services account using the classic `pulumi-azure` (azurerm-bridged) SDK, from the kind's typed IaC input.

## Design Decisions

- **Wire map identical to the Terraform module**: the same enum-name-to-wire-value maps (`sku.tier`, `version_upgrade_option`) and the same omit-when-unset semantics for `rai_policy_name` and `model.version` (Optional+Computed on the provider -- omitted so ARM defaults apply without read drift).
- **No tag map**: the deployment is an ARM child of its account; the provider's schema carries no tags, location, or resource group.
- **Provider builder**: credentials resolve through the shared `pulumiazureprovider` builder (static client secret, keyless web identity, or ambient chain).

## Inputs

The module consumes `AzureCognitiveDeploymentIacInput`: the target resource (metadata + spec) and the Azure provider configuration. `cognitive_account_id` arrives pre-resolved; `GetValue()` returns the literal ARM ID.

## Outputs

- `deployment_id`, `deployment_name`
- `model_version` -- the version ARM resolved

## Local Development

```shell
# compile the module
go build ./...
```
