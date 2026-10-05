# AzureCognitiveAccount Pulumi Module

## Overview

Provisions an Azure AI services account and its composed responsible-AI children (blocklists and content-filter policies) using the classic `pulumi-azure` (azurerm-bridged) SDK, from the kind's typed IaC input.

## Design Decisions

- **Wire map identical to the Terraform module**: the same enum-name-to-wire-value maps (identity types, bypass, RAI severity/mode), the same omit-when-unset semantics for optional fields, the same name-keyed child pattern for RAI blocklists and policies.
- **Children are parented to the account** (`pulumi.Parent`) and policies carry `pulumi.DependsOn` on every blocklist so a policy can reference a blocklist defined in the same spec.
- **Sensitive outputs**: `primary_access_key` and `secondary_access_key` are exported through `pulumi.ToSecret` in addition to the provider schema's own masking.
- **Provider builder**: credentials resolve through the shared `pulumiazureprovider` builder (static client secret, keyless web identity, or ambient chain).

## Inputs

The module consumes `AzureCognitiveAccountIacInput`: the target resource (metadata + spec) and the Azure provider configuration. StringValueOrRef fields arrive pre-resolved; `GetValue()` returns the literal.

## Outputs

- `cognitive_account_id`, `cognitive_account_name`, `endpoint`
- `primary_access_key`, `secondary_access_key` (secret; empty when local auth is disabled)
- `system_assigned_identity_principal_id`
- `rai_blocklist_ids`, `rai_policy_ids` (maps keyed by the spec entries' names)

## Local Development

```shell
# compile the module
go build ./...

# run the module's release entrypoint build
go build -o /dev/null ./..
```
