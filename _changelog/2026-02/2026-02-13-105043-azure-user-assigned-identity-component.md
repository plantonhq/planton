# Azure User-Assigned Managed Identity Deployment Component

**Date**: February 13, 2026
**Type**: Feature
**Components**: API Definitions, Azure Provider, Pulumi CLI Integration, Terraform Module

## Summary

Added `AzureUserAssignedIdentity` as a new Azure deployment component (R03 in the Azure resource expansion queue). This component provisions User-Assigned Managed Identities with bundled RBAC role assignments, following the `AwsIamRole` / `GcpServiceAccount` pattern adapted for Azure's scope-based RBAC model. A key design innovation is using `StringValueOrRef` for role assignment scopes, enabling infra charts to dynamically wire identity permissions to other Azure resources in the same chart.

## Problem Statement / Motivation

Azure enterprise workloads (AKS, Function Apps, Web Apps, Container Apps) need secure, credential-free access to Azure services like Key Vault, Storage, and Container Registry. While Azure provides managed identities natively, there was no Planton component to declaratively define an identity with its role assignments in a composable, infra-chart-ready format.

### Pain Points

- No way to declare an Azure managed identity in an Planton manifest
- Role assignments had to be managed separately from identity creation
- Infra charts couldn't wire identity permissions to dynamically created resources (e.g., "give this identity Key Vault Secrets User on the Key Vault created in this chart")
- The aks-environment, function-app-environment, and web-app-environment infra charts couldn't define identity layers

## Solution / What's New

### AzureUserAssignedIdentity Component

A complete deployment component at `apis/dev/planton/provider/azure/azureuserassignedidentity/v1/` with:

- **4 proto files**: spec, stack_outputs, api, stack_input
- **Validation tests**: 15 test cases covering all valid and invalid input scenarios
- **Pulumi module**: Creates identity + N role assignments via `authorization` package
- **Terraform module**: Creates identity + role assignments via `for_each`
- **Comprehensive documentation**: README, examples, research docs

### StringValueOrRef Scope (Design Innovation)

The role assignment `scope` field uses `StringValueOrRef` without `default_kind` annotation, making it the first polymorphic `StringValueOrRef` field in the codebase. This enables infra charts to wire role assignment scopes to any Azure resource's output:

```yaml
role_assignments:
  - scope:
      valueFrom:
        kind: AzureKeyVault
        name: platform-kv
        fieldPath: status.outputs.key_vault_id
    role_definition_name: Key Vault Secrets User
```

## Implementation Details

### Proto Design

- `region` (required): Azure region
- `resource_group` (required, StringValueOrRef): References AzureResourceGroup
- `name` (required, 3-128 chars): Identity name matching Azure's naming constraints
- `role_assignments` (repeated): Bundled RBAC bindings per DD03 (composite bundling rules)

Each `RoleAssignment` has:
- `scope` (required, StringValueOrRef): Polymorphic Azure resource ID
- `role_definition_name` (required): Azure built-in or custom role name

### Stack Outputs

- `identity_id`: Azure Resource Manager ID (referenced by AKS, Function Apps, Web Apps, Container Apps)
- `principal_id`: Service Principal Object ID (used for access policies)
- `client_id`: Application/Client ID (used by Azure SDK for authentication)
- `tenant_id`: Azure AD Tenant ID (for cross-tenant scenarios)

### Pulumi Module

Uses `authorization.NewUserAssignedIdentity` and `authorization.NewAssignment` from the `pulumi-azure` v6 SDK. Key implementation details:

- `SkipServicePrincipalAadCheck: true` on all role assignments to handle Azure AD replication latency
- Explicit `DependsOn` on the identity resource for proper ordering
- Role assignments named `{identity-name}-ra-{index}` for uniqueness

### Terraform Module

Uses `azurerm_user_assigned_identity` + `azurerm_role_assignment` with `for_each` over a local map keyed by `"{index}-{role_definition_name}"`. Sets `skip_service_principal_aad_check = true`.

### Enum Registration

Registered as `AzureUserAssignedIdentity = 460` in `cloud_resource_kind.proto` with `id_prefix: "azid"`, placed between AzureApplicationInsights (451) and the GCP section (600+).

## Benefits

- **Credential-free security**: Managed identities eliminate secrets, passwords, and key rotation
- **Infra-chart composable**: StringValueOrRef scope enables dynamic permission wiring
- **Cross-cloud consistency**: Follows the same pattern as AwsIamRole and GcpServiceAccount
- **Enterprise-ready**: Supports multiple role assignments, fine-grained scoping, shared identities
- **Zero Azure AD race conditions**: SkipServicePrincipalAadCheck handles replication latency

## Impact

- **4th Azure resource** in the expansion queue (R03 of 24)
- **Unblocks 3 infra charts**: aks-environment (enhanced), function-app-environment, web-app-environment
- **New pattern**: First polymorphic StringValueOrRef (no default_kind) -- may be reused by other resources with multi-type references

## Related Work

- **R00**: AzureResourceGroup (referenced by resource_group field)
- **R01**: AzureLogAnalyticsWorkspace
- **R02**: AzureApplicationInsights
- **DD03**: Composite Bundling Rules (why role assignments are bundled)
- **DD05**: AzureResourceGroup as First-Class Resource (all Azure resources use StringValueOrRef resource_group)

---

**Status**: Production Ready
**Build**: go build passes
**Tests**: 15/15 pass
**Timeline**: Single session
