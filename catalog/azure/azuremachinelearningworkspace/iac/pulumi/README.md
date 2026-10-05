# AzureMachineLearningWorkspace Pulumi Module

## Overview

Provisions an Azure Machine Learning workspace with its composed network outbound rules using the classic `pulumi-azure` (azurerm-bridged) SDK, from the kind's typed IaC input.

## Design Decisions

- **Wire map identical to the Terraform module**: the same enum-name-to-wire-value maps (`kind`, isolation mode, identity type) and the same omit-when-unset semantics (managed network read back, provider defaults applied on omission).
- **Composed children as parented resources**: each outbound rule is created with the workspace as parent, keyed by the rule's spec name; the three name-keyed ID maps export exactly like the Terraform outputs.
- **PARITY-EXCEPTION -- `storage_account_access_type`**: the classic SDK does not expose azurerm v5's storage-access-type argument, so the module FAILS LOUDLY when the spec sets it (deploy such workspaces with the Terraform engine). Failing beats silently keeping the account-key auth mode the user turned off.
- **Provider builder**: credentials resolve through the shared `pulumiazureprovider` builder (static client secret, keyless web identity, or ambient chain).

## Inputs

The module consumes `AzureMachineLearningWorkspaceIacInput`: the target resource (metadata + spec) and the Azure provider configuration. Every reference field arrives pre-resolved; `GetValue()` returns the literal.

## Outputs

- `machine_learning_workspace_id`, `machine_learning_workspace_name`
- `workspace_guid`, `discovery_url`, `system_assigned_identity_principal_id`
- `fqdn_outbound_rule_ids`, `private_endpoint_outbound_rule_ids`, `service_tag_outbound_rule_ids` -- name-keyed ARM ID maps

## Local Development

```shell
# compile the module
go build ./...
```
