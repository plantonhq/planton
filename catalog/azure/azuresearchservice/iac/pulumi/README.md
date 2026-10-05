# AzureSearchService Pulumi Module

## Overview

Creates an Azure AI Search service and its composed shared private
links using the classic `pulumi-azure` (azurerm-bridged) SDK, from the
kind's typed IaC input.

## Design Decisions

- **Wire map identical to the Terraform module**: the same
  identity-type and hosting-mode enum maps (the hosting-mode wire is
  camelCase: `default` / `highDensity`), the same presence-guarded
  defaults for the counts and booleans, and the same omit-when-unset
  posture for the string vocabularies.
- **Composed children**: one `search.SharedPrivateLinkService` per
  spec entry, parented to the service, ids surfaced name-keyed in
  `shared_private_link_service_ids`.
- **Sensitive outputs**: the admin keys and the built-in query key
  export through `pulumi.ToSecret` (service-minted credentials with
  no vault indirection -- the narrow outputs exception).
- **The default query key is a single output**: the service creates
  exactly one query key, with an EMPTY name -- a name-keyed map would
  carry the empty string as its only key, so the module exports the
  single key instead.
- **Provider builder**: credentials resolve through the shared
  `pulumiazureprovider` builder (static client secret, keyless web
  identity, or ambient chain).

## Inputs

The module consumes `AzureSearchServiceIacInput`: the target
resource (metadata + spec) and the Azure provider configuration. All
references arrive pre-resolved; `GetValue()` returns the literal
value.

## Outputs

- `search_service_id`, `search_service_name`, `endpoint`
- `primary_key`, `secondary_key`, `default_query_key` (secrets)
- `customer_managed_key_encryption_compliance_status`
- `system_assigned_identity_principal_id`
- `shared_private_link_service_ids` (name-keyed map)

## Local Development

```shell
# compile the module
go build ./...
```
