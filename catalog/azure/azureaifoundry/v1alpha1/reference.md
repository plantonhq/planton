# AzureAiFoundry

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureAiFoundrySpec** defines an Azure AI Foundry hub (ARM:
Microsoft.MachineLearningServices/workspaces with kind "Hub") --
the shared foundation a company sets up ONCE for its AI teams. The
hub owns the security and connectivity posture (identity, key
vault, storage, optional CMK encryption, managed network, public
access) that every AzureAiFoundryProject created inside it
inherits; teams then work in projects, not in the hub itself.

The hub REQUIRES two companion services at creation -- a key vault
(hub secrets and connection credentials) and a storage account
(artifacts) -- and optionally wires an application-insights
component (telemetry) and a container registry (environment
images). The key vault and storage attachments are fixed at
creation (ForceNew); insights and registry can be attached or
re-pointed in place (unlike the classic ML workspace, where the
registry is ForceNew).

**Deletion is a soft delete.** Like the classic ML workspace, a
deleted hub becomes a purgeable ghost that keeps holding the hub
name; recreating under the same name fails until the ghost is
purged.

## Example

```yaml
# Offline-plan test manifest. Exercises the deep seams: a
# system-and-user-assigned identity with a primary identity, CMK
# encryption with a VERSIONED key URL (the hub's own contract --
# versionless is rejected), the managed network at approved-outbound
# isolation, both optional attachments (insights, registry), disabled
# public access, and the high-business-impact flag.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureAiFoundry
metadata:
  name: test-ai-foundry
  org: test-org
  env: dev
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: test-ai-foundry
  keyVaultId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.KeyVault/vaults/test-vault
  storageAccountId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Storage/storageAccounts/testaifstorage
  identity:
    type: SYSTEM_AND_USER_ASSIGNED
    identityIds:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/aif-uai
  applicationInsightsId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Insights/components/test-insights
  containerRegistryId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ContainerRegistry/registries/testaifacr
  primaryUserAssignedIdentity:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/aif-uai
  publicNetworkAccessEnabled: false
  encryption:
    keyVaultId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.KeyVault/vaults/test-vault
    keyId:
      value: https://test-vault.vault.azure.net/keys/aif-cmk/0123456789abcdef0123456789abcdef
    userAssignedIdentityId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/aif-uai
  managedNetwork:
    isolationMode: ALLOW_ONLY_APPROVED_OUTBOUND
  highBusinessImpactEnabled: true
  description: Offline-plan AI Foundry hub exercising the deep seams
  friendlyName: Test AI Foundry Hub
  tags:
    cost-center: ai-platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.keyVaultId` | `string \| valueFrom` | yes |  | AzureKeyVault (`status.outputs.key_vault_id`) |
| `spec.storageAccountId` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_id`) |
| `spec.identity` | `AzureAiFoundryIdentity` | yes |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.applicationInsightsId` | `string \| valueFrom` |  |  | AzureApplicationInsights (`status.outputs.application_insights_id`) |
| `spec.containerRegistryId` | `string \| valueFrom` |  |  | AzureContainerRegistry (`status.outputs.container_registry_id`) |
| `spec.primaryUserAssignedIdentity` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.encryption` | `AzureAiFoundryEncryption` |  |  |  |
| `spec.encryption.keyVaultId` | `string \| valueFrom` | yes |  | AzureKeyVault (`status.outputs.key_vault_id`) |
| `spec.encryption.keyId` | `string \| valueFrom` | yes |  | AzureKeyVaultKey (`status.outputs.key_id`) |
| `spec.encryption.userAssignedIdentityId` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.managedNetwork` | `AzureAiFoundryManagedNetwork` |  |  |  |
| `spec.managedNetwork.isolationMode` | `enum` |  |  |  |
| `spec.highBusinessImpactEnabled` | `bool` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.friendlyName` | `string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the hub lives in, e.g. "eastus". Projects created
in the hub live in their own region field but deploy into the
hub's resource group. Changing the region replaces the hub.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the hub is created in. Can be a literal
resource-group name or a reference to an AzureResourceGroup's
name output. Projects created in this hub are placed in THIS
resource group (the provider derives it from the hub).

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The hub's name, unique within the resource group: 3-33
characters, starting with an alphanumeric character, then
alphanumerics, hyphens or underscores (the provider's own code
regex -- its error text omits the underscore the regex accepts).
Changing the name replaces the hub -- and the OLD name stays held
by the soft-deleted ghost until purged.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9][a-zA-Z0-9_-]{2,32}$"}}

### spec.keyVaultId

`string | valueFrom` · required

The Key Vault the hub stores its secrets in (connection
credentials, project secrets), by ARM ID. Fixed at creation. The hub
WRITES INTO the vault and lives in its own resource group, so the
reference is access, not placement, on a diagram.

- references: AzureKeyVault (`status.outputs.key_vault_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVault, name: <that resource's name>, fieldPath: status.outputs.key_vault_id}} -- a bare string does not parse

### spec.storageAccountId

`string | valueFrom` · required

The storage account backing the hub's artifacts and file shares,
by ARM ID. Use a general-purpose account WITHOUT hierarchical
namespace -- the hub is an ML workspace at ARM, which rejects
Data Lake Gen2 accounts as default storage. Fixed at creation.

The hub READS and WRITES this account as its default workspace storage
and lives in its own resource group, so on a diagram the reference is
access, not placement -- the same rule its key_vault_id already
carries.

- references: AzureStorageAccount (`status.outputs.storage_account_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_id}} -- a bare string does not parse

### spec.identity

`AzureAiFoundryIdentity` · required

The hub's managed identity. REQUIRED -- the hub accesses its
companion services (key vault, storage, insights, registry)
through this identity, and projects inherit the access posture.

- rule: {"required":true}
- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the hub; USER_ASSIGNED brings identities you manage
(grantable key vault / storage access BEFORE the hub exists --
pair with primary_user_assigned_identity);
SYSTEM_AND_USER_ASSIGNED carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_ai_foundry_identity_type_unspecified` -- Not specified: rejected -- the hub requires an identity.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the hub.
- `USER_ASSIGNED` -- Identities you create and manage (AzureUserAssignedIdentity).
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned and user-assigned identities.

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the hub, by ARM ID. Reference
AzureUserAssignedIdentity resources so key vault / storage grants
can be composed before the hub is created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.applicationInsightsId

`string | valueFrom`

The Application Insights component the hub sends telemetry to,
by ARM ID. Attachable and re-pointable in place.

- references: AzureApplicationInsights (`status.outputs.application_insights_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureApplicationInsights, name: <that resource's name>, fieldPath: status.outputs.application_insights_id}} -- a bare string does not parse

### spec.containerRegistryId

`string | valueFrom`

The container registry the hub stores its environment images in,
by ARM ID. Attachable and re-pointable in place (unlike the
classic ML workspace, where the registry attachment is fixed at
creation).

- references: AzureContainerRegistry (`status.outputs.container_registry_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureContainerRegistry, name: <that resource's name>, fieldPath: status.outputs.container_registry_id}} -- a bare string does not parse

### spec.primaryUserAssignedIdentity

`string | valueFrom`

For hubs with a user-assigned identity: which of the attached
identities the hub uses as its primary, by ARM ID. The provider
requires this when identity.type is USER_ASSIGNED (documented
contract; ARM rejects the create without it).

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the hub's endpoint answers the public internet.
Unspecified applies true (the provider's default, wire value
"Enabled"). Set false to make the hub reachable only through
private endpoints (wire value "Disabled"). Updates in place.

- default: `true`

### spec.encryption

`AzureAiFoundryEncryption`

Encrypt the hub's data with your own Key Vault key instead of
Microsoft-managed keys. The whole block is fixed at creation.
The hub's identity needs wrap/unwrap access on the key BEFORE
this is configured, and enabling encryption may force
high_business_impact_enabled true service-side.

### spec.encryption.keyVaultId

`string | valueFrom` · required

The Key Vault holding the encryption key, by ARM ID. The hub UNWRAPS
its key through the vault and never lives in it, so the reference is
access, not placement, on a diagram.

- references: AzureKeyVault (`status.outputs.key_vault_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVault, name: <that resource's name>, fieldPath: status.outputs.key_vault_id}} -- a bare string does not parse

### spec.encryption.keyId

`string | valueFrom` · required

The Key Vault key as a VERSIONED data-plane URL
("https://{vault}.vault.azure.net/keys/{name}/{version}").
Reference an AzureKeyVaultKey's key_id output (the versioned
URL). NOTE this deliberately DIVERGES from the classic ML
workspace's versionless guidance: the provider validates the
hub's key as a versioned key URI and rejects versionless URLs,
so key rotation does NOT auto-propagate here -- re-point the
field to rotate (the provider's hub contract, not a choice).

- references: AzureKeyVaultKey (`status.outputs.key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVaultKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.encryption.userAssignedIdentityId

`string | valueFrom`

The USER-ASSIGNED identity that unwraps the key, by ARM ID.
Leave unset to use the hub's system-assigned identity. Required
when the hub's identity is USER_ASSIGNED only (documented
provider contract -- checked at apply, since the pairing spans
the identity block).

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.managedNetwork

`AzureAiFoundryManagedNetwork`

The hub's managed virtual network -- Azure-managed network
isolation for the hub's (and its projects') compute and
endpoints. Unspecified leaves the mode to ARM (isolation
disabled); the isolation mode is then read back rather than
defaulted here. Updates in place.

### spec.managedNetwork.isolationMode

`enum`

The isolation mode. Unspecified omits the property (ARM leaves
isolation disabled and the value is read back).

Allowed values (use exactly as shown):

- `azure_ai_foundry_isolation_mode_unspecified` -- Not specified: the property is omitted and ARM applies its default (isolation disabled).
- `DISABLED` -- No network isolation (wire value "Disabled").
- `ALLOW_INTERNET_OUTBOUND` -- The managed network allows all outbound traffic (wire value "AllowInternetOutbound").
- `ALLOW_ONLY_APPROVED_OUTBOUND` -- The managed network allows only approved outbound traffic (wire value "AllowOnlyApprovedOutbound"). Outbound rules are managed from the Azure AI Foundry portal or API -- the azurerm provider models no outbound-rule children for hubs.

### spec.highBusinessImpactEnabled

`bool`

Marks the hub as handling sensitive business data ("high
business impact") -- Azure reduces the diagnostic data it
collects. Fixed at creation. False means "leave it to the
service": the modules omit the property when false because the
service itself flips it true when encryption is enabled, and a
pinned false would fight that (the provider reads the value
back).

### spec.description

`string`

What the hub is for.

### spec.friendlyName

`string`

The display name shown in the Azure AI Foundry portal.

### spec.tags

`map<string, string>`

Free-form tags applied to the hub, merged over the
Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureAiFoundry, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.ai_foundry_id` | `string` | The Azure Resource Manager ID of the hub -- what AzureAiFoundryProject resources reference as their ai_services_hub_id. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.MachineLearningServices/workspaces/{name} |
| `status.outputs.ai_foundry_name` | `string` | The name of the hub. ARM addresses the hub (and, at the API level, its projects' hub linkage) by this name. |
| `status.outputs.workspace_guid` | `string` | The hub's immutable GUID (distinct from the ARM ID) -- what some data-plane SDKs and diagnostic settings identify the hub by. |
| `status.outputs.discovery_url` | `string` | The hub's regional discovery URL -- where SDKs resolve the hub's data-plane service endpoints from. |
| `status.outputs.system_assigned_identity_principal_id` | `string` | The principal (object) ID of the hub's system-assigned identity, when one is enabled -- what key vault / storage grants bind to. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.keyVaultId` | AzureKeyVault | `status.outputs.key_vault_id` |
| `spec.storageAccountId` | AzureStorageAccount | `status.outputs.storage_account_id` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.applicationInsightsId` | AzureApplicationInsights | `status.outputs.application_insights_id` |
| `spec.containerRegistryId` | AzureContainerRegistry | `status.outputs.container_registry_id` |
| `spec.primaryUserAssignedIdentity` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.encryption.keyVaultId` | AzureKeyVault | `status.outputs.key_vault_id` |
| `spec.encryption.keyId` | AzureKeyVaultKey | `status.outputs.key_id` |
| `spec.encryption.userAssignedIdentityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureAiFoundryProject | `spec.aiServicesHubId` | `status.outputs.ai_foundry_id` |

## See Also

- [Overview](../README.md)
