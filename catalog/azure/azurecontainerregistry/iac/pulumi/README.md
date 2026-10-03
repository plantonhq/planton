# AzureContainerRegistry Pulumi Module

## Overview

This Pulumi module provisions an Azure Container Registry using the Azure
Classic provider (`pulumi-azure` v6). It creates a
`containerservice.Registry` -- the managed, private OCI registry the
platform's workloads pull their images from -- covering the full surface:
SKU, admin account, network posture (public access, IP rule set, bypass
option, data endpoints), policies (quarantine, retention, export),
geo-replications, managed identity, and customer-managed-key encryption.
Docker Content Trust is deliberately not modeled: the ACR service retired
it. The module behaves identically to the Terraform module: same inputs,
same defaults, same outputs.

The SKU is the registry's feature gate; the Premium-only fields
(geo-replication, zone redundancy, network rules, policies, CMK) are
enforced at spec validation, mirroring ARM's own gates, so the module
receives only deployable shapes. An unset `sku` deploys the STANDARD
baseline, and the true-default optional bools (`public_network_access_enabled`,
`export_policy_enabled`) are presence-guarded so iac-input paths that
bypass the manifest loader deploy identically on both engines.

Lifecycle notes worth knowing before operating this resource: name and
region are the registry's identity -- changing either replaces it and its
CONTENTS DO NOT MIGRATE; every image would need re-pushing. Zone
redundancy (home replica) and CMK encryption are likewise fixed at
creation. The SKU changes in place, but downgrading requires every
Premium-only feature to be unset first. Geo-replications add/remove in
place; the provider expects them in alphabetical location order, and the
module sorts them internally, keeping manifests order-insensitive.

Grants are composed, never bundled: AcrPull/AcrPush role assignments are
standalone `AzureRoleAssignment` resources scoped to the registry's ARM
ID, and the identity that unwraps a CMK key is a referenced first-class
`AzureUserAssignedIdentity` that must hold get/wrapKey/unwrapKey on the
key's vault before the registry is created.

## Resources Created

- `containerservice.Registry` -- the registry, including inline
  geo-replications, network rule set, identity, and encryption blocks

## Inputs

The module receives an `AzureContainerRegistryIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.registry_name` -- the registry's ARM identity (references resolved to literals by the platform; the name becomes `{name}.azurecr.io`)
- `target.spec.sku` -- BASIC / STANDARD / PREMIUM; unset applies the STANDARD baseline
- `target.spec.admin_user_enabled` -- the built-in admin account (default false; Entra auth is the production path)
- `target.spec.public_network_access_enabled` -- default true; false means private-endpoints-only (Premium)
- `target.spec.zone_redundancy_enabled` -- home replica across availability zones (Premium; fixed at creation)
- `target.spec.anonymous_pull_enabled` -- unauthenticated pulls (Standard/Premium)
- `target.spec.data_endpoint_enabled` -- dedicated regional data endpoints for exact firewall allowlisting (Premium)
- `target.spec.quarantine_policy_enabled` / `target.spec.retention_policy_in_days` -- the quarantine and untagged-manifest retention policies (both Premium)
- `target.spec.export_policy_enabled` -- default true; disabling requires Premium + public access off
- `target.spec.network_rule_bypass_option` -- AZURE_SERVICES (Azure's default) or NONE; only an explicit choice is sent
- `target.spec.network_rule_set` -- default action plus IPv4 CIDR allowlist (Premium)
- `target.spec.georeplications` -- additional regions with per-replica zone redundancy, global endpoint routing, and tags (Premium; must not contain the home region); the classic v6 SDK still names the routing argument `regional_endpoint_enabled` -- same ARM property azurerm 5 calls `global_endpoint_routing_enabled`
- `target.spec.identity` -- SYSTEM_ASSIGNED / USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED plus resolved user-assigned-identity ARM IDs
- `target.spec.encryption` -- CMK: the unwrapping identity's client id and the Key Vault key ID (Premium; fixed at creation)
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials (static client secret, keyless web identity, or ambient chain)

## Outputs

| Output | Description |
|--------|-------------|
| `container_registry_id` | Full ARM ID of the registry -- the join key AKS clusters and AcrPull/AcrPush role assignments reference |
| `container_registry_name` | The registry's name as deployed |
| `login_server` | The hostname images are tagged with and pulled from, e.g. `myregistry.azurecr.io` |
| `admin_username` | Admin account username (empty unless `admin_user_enabled`) |
| `admin_password` | One of the admin account's two rotatable passwords (empty unless enabled) |
| `system_assigned_identity_principal_id` | Principal ID of the system-assigned identity (empty unless the identity type includes it) |
| `data_endpoint_host_names` | Dedicated regional data-endpoint hostnames (empty unless `data_endpoint_enabled`) |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest planton
```
