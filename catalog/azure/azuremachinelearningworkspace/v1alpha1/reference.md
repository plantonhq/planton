# AzureMachineLearningWorkspace

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureMachineLearningWorkspaceSpec** defines an Azure Machine
Learning workspace (ARM:
Microsoft.MachineLearningServices/workspaces) -- the central home a
data-science team works in: experiments, models, endpoints,
datastores, and compute all live on a workspace. The workspace
itself is a thin coordination object; it REQUIRES three companion
services at creation -- a storage account (artifacts and the
default datastores), a key vault (workspace secrets), and an
application-insights component (telemetry) -- and optionally a
container registry (environment images).

**The companion services are fixed at creation** (ForceNew):
`application_insights_id`, `key_vault_id`, `storage_account_id`,
and `container_registry_id` cannot be re-pointed later -- neither
can `name`, `region`, `resource_group`, `encryption`,
`high_business_impact`, `service_side_encryption_enabled`, or the
managed network's `provision_on_creation_enabled`. The default
storage account must be a general-purpose account WITHOUT
hierarchical namespace (Data Lake Gen2) -- ARM rejects HNS-enabled
accounts as workspace default storage.

**Deletion is a soft delete.** A deleted workspace becomes a
purgeable ghost that keeps holding the workspace name (the Key
Vault recycle-bin pattern); recreating under the same name fails
until the ghost is purged.

## Example

```yaml
# Offline-plan test manifest. Exercises the deep seams: a user-assigned
# identity with a primary identity, CMK encryption with service-side
# CMK, the managed network at approved-outbound isolation with all
# THREE outbound-rule types (they share one ARM collection -- names are
# unique across the lists), serverless compute pinned to a subnet, the
# container-registry attachment, and the high-business-impact flag.
# storage_account_access_type is deliberately NOT set here: it is the
# kind's one PARITY-EXCEPTION (Terraform-only), and the canonical
# manifest stays deployable by both engines.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMachineLearningWorkspace
metadata:
  name: test-ml-workspace
  org: test-org
  env: dev
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: test-ml-workspace
  applicationInsightsId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Insights/components/test-insights
  keyVaultId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.KeyVault/vaults/test-vault
  storageAccountId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Storage/storageAccounts/testmlstorage
  identity:
    type: SYSTEM_AND_USER_ASSIGNED
    identityIds:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ml-uai
  primaryUserAssignedIdentity:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ml-uai
  containerRegistryId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ContainerRegistry/registries/testmlacr
  description: Offline-plan workspace exercising the deep seams
  friendlyName: Test ML Workspace
  encryption:
    keyVaultId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.KeyVault/vaults/test-vault
    keyId:
      value: https://test-vault.vault.azure.net/keys/ml-cmk
    userAssignedIdentityId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ml-uai
  serviceSideEncryptionEnabled: true
  highBusinessImpact: true
  skuName: Basic
  managedNetwork:
    isolationMode: ALLOW_ONLY_APPROVED_OUTBOUND
    provisionOnCreationEnabled: true
  imageBuildComputeName: image-builder
  serverlessCompute:
    subnetId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/serverless
    publicIpEnabled: false
  fqdnOutboundRules:
    - name: allow-pypi
      destinationFqdn: pypi.org
  privateEndpointOutboundRules:
    - name: to-vault
      serviceResourceId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.KeyVault/vaults/test-vault
      subResourceTarget: vault
      sparkEnabled: true
  serviceTagOutboundRules:
    - name: allow-storage
      serviceTag: Storage
      protocol: TCP
      portRanges: "443"
  tags:
    cost-center: ml-platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.applicationInsightsId` | `string \| valueFrom` | yes |  | AzureApplicationInsights (`status.outputs.application_insights_id`) |
| `spec.keyVaultId` | `string \| valueFrom` | yes |  | AzureKeyVault (`status.outputs.key_vault_id`) |
| `spec.storageAccountId` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_id`) |
| `spec.identity` | `AzureMachineLearningWorkspaceIdentity` | yes |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.kind` | `enum` |  |  |  |
| `spec.featureStore` | `AzureMachineLearningWorkspaceFeatureStore` |  |  |  |
| `spec.featureStore.computerSparkRuntimeVersion` | `string` |  |  |  |
| `spec.featureStore.offlineConnectionName` | `string` |  |  |  |
| `spec.featureStore.onlineConnectionName` | `string` |  |  |  |
| `spec.featureStore.enabled` | `bool` |  | `true` |  |
| `spec.primaryUserAssignedIdentity` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.containerRegistryId` | `string \| valueFrom` |  |  | AzureContainerRegistry (`status.outputs.container_registry_id`) |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.imageBuildComputeName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.friendlyName` | `string` |  |  |  |
| `spec.encryption` | `AzureMachineLearningWorkspaceEncryption` |  |  |  |
| `spec.encryption.keyVaultId` | `string \| valueFrom` | yes |  | AzureKeyVault (`status.outputs.key_vault_id`) |
| `spec.encryption.keyId` | `string \| valueFrom` | yes |  | AzureKeyVaultKey (`status.outputs.versionless_id`) |
| `spec.encryption.userAssignedIdentityId` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.managedNetwork` | `AzureMachineLearningWorkspaceManagedNetwork` |  |  |  |
| `spec.managedNetwork.isolationMode` | `enum` |  |  |  |
| `spec.managedNetwork.provisionOnCreationEnabled` | `bool` |  |  |  |
| `spec.highBusinessImpact` | `bool` |  |  |  |
| `spec.skuName` | `string` |  |  |  |
| `spec.serviceSideEncryptionEnabled` | `bool` |  |  |  |
| `spec.v1LegacyModeEnabled` | `bool` |  |  |  |
| `spec.storageAccountAccessType` | `enum` |  |  |  |
| `spec.serverlessCompute` | `AzureMachineLearningWorkspaceServerlessCompute` |  |  |  |
| `spec.serverlessCompute.subnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.serverlessCompute.publicIpEnabled` | `bool` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |
| `spec.fqdnOutboundRules` | `[]AzureMachineLearningWorkspaceFqdnOutboundRule` |  |  |  |
| `spec.fqdnOutboundRules[].name` | `string` | yes |  |  |
| `spec.fqdnOutboundRules[].destinationFqdn` | `string` | yes |  |  |
| `spec.privateEndpointOutboundRules` | `[]AzureMachineLearningWorkspacePrivateEndpointOutboundRule` |  |  |  |
| `spec.privateEndpointOutboundRules[].name` | `string` | yes |  |  |
| `spec.privateEndpointOutboundRules[].serviceResourceId` | `string \| valueFrom` | yes |  |  |
| `spec.privateEndpointOutboundRules[].subResourceTarget` | `string` | yes |  |  |
| `spec.privateEndpointOutboundRules[].sparkEnabled` | `bool` |  |  |  |
| `spec.serviceTagOutboundRules` | `[]AzureMachineLearningWorkspaceServiceTagOutboundRule` |  |  |  |
| `spec.serviceTagOutboundRules[].name` | `string` | yes |  |  |
| `spec.serviceTagOutboundRules[].serviceTag` | `string` | yes |  |  |
| `spec.serviceTagOutboundRules[].protocol` | `string` | yes |  |  |
| `spec.serviceTagOutboundRules[].portRanges` | `string` | yes |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the workspace lives in, e.g. "eastus". Serverless
compute quota and some VM families differ per region. Changing the
region replaces the workspace.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the workspace is created in. Can be a
literal resource-group name or a reference to an
AzureResourceGroup's name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The workspace's name, unique within the resource group: 3-33
characters, starting with an alphanumeric character, then
alphanumerics, hyphens or underscores (the provider's own rule).
Changing the name replaces the workspace -- and the OLD name stays
held by the soft-deleted ghost until purged.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9][a-zA-Z0-9_-]{2,32}$"}}

### spec.applicationInsightsId

`string | valueFrom` · required

The Application Insights component the workspace sends telemetry
to, by ARM ID. Fixed at creation.

- references: AzureApplicationInsights (`status.outputs.application_insights_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureApplicationInsights, name: <that resource's name>, fieldPath: status.outputs.application_insights_id}} -- a bare string does not parse

### spec.keyVaultId

`string | valueFrom` · required

The Key Vault the workspace stores its secrets in (connection
credentials, compute SSH keys), by ARM ID. Fixed at creation. The
workspace WRITES INTO the vault and lives in its own resource group,
so the reference is access, not placement, on a diagram.

- references: AzureKeyVault (`status.outputs.key_vault_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVault, name: <that resource's name>, fieldPath: status.outputs.key_vault_id}} -- a bare string does not parse

### spec.storageAccountId

`string | valueFrom` · required

The storage account backing the workspace's artifacts and its two
built-in datastores, by ARM ID. Must be a general-purpose account
WITHOUT hierarchical namespace (ARM rejects Data Lake Gen2
accounts as default workspace storage). Fixed at creation.

The workspace READS and WRITES this account as its default storage and
lives in its own resource group, so on a diagram the reference is
access, not placement -- the same rule its key_vault_id already
carries.

- references: AzureStorageAccount (`status.outputs.storage_account_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_id}} -- a bare string does not parse

### spec.identity

`AzureMachineLearningWorkspaceIdentity` · required

The workspace's managed identity. REQUIRED -- the workspace
accesses its companion services (storage, key vault, insights)
through this identity.

- rule: {"required":true}
- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the workspace; USER_ASSIGNED brings identities you manage
(grantable storage / Key Vault access BEFORE the workspace
exists -- pair with primary_user_assigned_identity);
SYSTEM_AND_USER_ASSIGNED carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_machine_learning_workspace_identity_type_unspecified` -- Not specified: rejected -- the workspace requires an identity.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the workspace.
- `USER_ASSIGNED` -- Identities you create and manage (AzureUserAssignedIdentity).
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned and user-assigned identities.

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the workspace, by ARM ID. Reference
AzureUserAssignedIdentity resources so storage / Key Vault grants
can be composed before the workspace is created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.kind

`enum`

What flavor of workspace this is. Unspecified applies the
provider default, DEFAULT (a regular ML workspace).
FEATURE_STORE creates a feature-store workspace and requires the
feature_store block. (AI Foundry hubs and projects are separate
kinds, not workspace flavors.)

Allowed values (use exactly as shown):

- `azure_machine_learning_workspace_kind_unspecified` -- Not specified: the provider applies "Default".
- `DEFAULT` -- A regular ML workspace (wire value "Default").
- `FEATURE_STORE` -- A feature-store workspace (wire value "FeatureStore") -- requires the feature_store block.

### spec.featureStore

`AzureMachineLearningWorkspaceFeatureStore`

Feature-store settings. Only legal -- and REQUIRED -- when kind
is FEATURE_STORE (the provider's own contract).

### spec.featureStore.computerSparkRuntimeVersion

`string`

The Spark runtime version of the feature store's materialization
compute. Leave unset for the service default.

### spec.featureStore.offlineConnectionName

`string`

The name of the workspace connection pointing at the OFFLINE
store (the historical feature data).

### spec.featureStore.onlineConnectionName

`string`

The name of the workspace connection pointing at the ONLINE
store (the low-latency serving data).

### spec.featureStore.enabled

`bool` · optional (explicit presence)

Whether the feature store is on. Unset means on: declaring the block has
always meant a feature store, and this switch lets a manifest say the
opposite out loud (a FEATURE_STORE workspace still requires it on).

- default: `true`

### spec.primaryUserAssignedIdentity

`string | valueFrom`

For workspaces with a user-assigned identity: which of the
attached identities the workspace uses as its primary, by ARM ID.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.containerRegistryId

`string | valueFrom`

The container registry the workspace stores its environment
images in, by ARM ID. The registry's admin account must be
enabled for ML image builds. Fixed at creation -- attaching a
registry later replaces the workspace.

- references: AzureContainerRegistry (`status.outputs.container_registry_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureContainerRegistry, name: <that resource's name>, fieldPath: status.outputs.container_registry_id}} -- a bare string does not parse

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the workspace's endpoint answers the public internet.
Unspecified applies true (ARM's default). Set false to make the
workspace reachable only through private endpoints.

- default: `true`

### spec.imageBuildComputeName

`string`

The name of the workspace compute cluster used to build Docker
environment images -- needed when the container registry sits
behind a VNet (image builds cannot run serverless there).

### spec.description

`string`

What the workspace is for.

### spec.friendlyName

`string`

The display name shown in Azure ML studio.

### spec.encryption

`AzureMachineLearningWorkspaceEncryption`

Encrypt the workspace's data with your own Key Vault key instead
of Microsoft-managed keys. The whole block is fixed at creation.
The workspace's identity needs wrap/unwrap access on the key
BEFORE this is configured.

### spec.encryption.keyVaultId

`string | valueFrom` · required

The Key Vault holding the encryption key, by ARM ID. The workspace
UNWRAPS its key through the vault and never lives in it, so the
reference is access, not placement, on a diagram.

- references: AzureKeyVault (`status.outputs.key_vault_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVault, name: <that resource's name>, fieldPath: status.outputs.key_vault_id}} -- a bare string does not parse

### spec.encryption.keyId

`string | valueFrom` · required

The Key Vault key (data-plane URL, e.g.
"https://{vault}.vault.azure.net/keys/{name}"). Reference an
AzureKeyVaultKey's versionless_id so rotation propagates without
intervention; pin a versioned URL only under a compliance regime
that demands a frozen version.

- references: AzureKeyVaultKey (`status.outputs.versionless_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVaultKey, name: <that resource's name>, fieldPath: status.outputs.versionless_id}} -- a bare string does not parse

### spec.encryption.userAssignedIdentityId

`string | valueFrom`

The USER-ASSIGNED identity that unwraps the key, by ARM ID.
Leave unset to use the workspace's system-assigned identity.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.managedNetwork

`AzureMachineLearningWorkspaceManagedNetwork`

The workspace's managed virtual network -- Azure-managed network
isolation for the workspace's compute and endpoints. Unspecified
leaves the mode to ARM (isolation disabled); the isolation mode
is then read back rather than defaulted here. Outbound rules
(the *_outbound_rules lists) only take effect under the
ALLOW_ONLY_APPROVED_OUTBOUND mode.

Live-caught: provision_on_creation with ALLOW_ONLY_APPROVED_OUTBOUND
is a fragile create. ARM can roll the workspace back mid-poll
("workspace identity has been deleted" / workspace NotFound) after
several minutes; a PE outbound rule on top of that then 409-loops
destroy. The smoke lane therefore leaves this block unset and
proves the workspace object; this arm stays offline-proven.

### spec.managedNetwork.isolationMode

`enum`

The isolation mode. Unspecified omits the property (ARM leaves
isolation disabled and the value is read back).
ALLOW_ONLY_APPROVED_OUTBOUND is the mode the outbound-rule lists
exist for.

Allowed values (use exactly as shown):

- `azure_machine_learning_workspace_isolation_mode_unspecified` -- Not specified: the property is omitted and ARM applies its default (isolation disabled).
- `DISABLED` -- No network isolation (wire value "Disabled").
- `ALLOW_INTERNET_OUTBOUND` -- The managed network allows all outbound traffic (wire value "AllowInternetOutbound").
- `ALLOW_ONLY_APPROVED_OUTBOUND` -- The managed network allows only approved outbound traffic -- the outbound-rule lists define what is approved (wire value "AllowOnlyApprovedOutbound").

### spec.managedNetwork.provisionOnCreationEnabled

`bool`

Provision the managed network WITH the workspace instead of
lazily on first compute (spares the first job the provisioning
wait). Fixed at creation.

### spec.highBusinessImpact

`bool`

Marks the workspace as handling sensitive business data ("high
business impact") -- Azure reduces the diagnostic data it
collects. Fixed at creation.

### spec.skuName

`string`

The workspace SKU (the wire value). "Basic" is the only value the
provider accepts at azurerm v5 -- unspecified applies it.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Basic"]}}

### spec.serviceSideEncryptionEnabled

`bool`

Encrypt the workspace's service-side data with the
customer-managed key as well (service-side CMK). Requires the
encryption block. Fixed at creation.

### spec.v1LegacyModeEnabled

`bool`

Run the workspace in v1 legacy API mode (blocks the v2 API
surface). Only for estates pinned to the legacy v1 CLI/SDK.

### spec.storageAccountAccessType

`enum`

How the workspace's built-in datastores authenticate to the
default storage account. Unspecified applies the provider
default, ACCESS_KEY. IDENTITY uses the workspace identity instead
of account keys (the hardened posture; the identity needs Blob
Data Contributor on the storage account).

PARITY-EXCEPTION: the Pulumi engine's classic SDK does not expose
this argument, so workspaces that set it deploy via the Terraform
engine only -- the Pulumi module fails loudly when it is set.

Allowed values (use exactly as shown):

- `azure_machine_learning_workspace_storage_account_access_type_unspecified` -- Not specified: the provider applies "AccessKey".
- `ACCESS_KEY` -- Storage account keys (wire value "AccessKey") -- the provider's default.
- `IDENTITY` -- The workspace's managed identity (wire value "Identity") -- the hardened posture; the identity needs Storage Blob Data Contributor on the account.

### spec.serverlessCompute

`AzureMachineLearningWorkspaceServerlessCompute`

Settings for the workspace's serverless compute (the managed
on-demand compute for jobs without a named cluster).

### spec.serverlessCompute.subnetId

`string | valueFrom`

The subnet serverless compute nodes are placed in, by ARM ID.
Required when public_ip_enabled is false and the workspace's
public network access is disabled.
On a diagram the workspace lives in its resource group and its
serverless nodes attach to this subnet, so the reference is
access, not placement -- a workspace is never drawn inside the
subnet it injects compute into (the Azure-SSIS runtime's rule).

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.serverlessCompute.publicIpEnabled

`bool`

Whether serverless compute nodes get public IPs. The provider's
default is false (no public IPs). NOTE (update behavior): the
provider rejects flipping this from true back to false while no
subnet_id is set.

### spec.tags

`map<string, string>`

Free-form tags applied to the workspace, merged over the
Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins.

### spec.fqdnOutboundRules

`[]AzureMachineLearningWorkspaceFqdnOutboundRule`

FQDN outbound rules on the workspace's managed network: named
allowlist entries for outbound destinations by domain name
(e.g. "pypi.org"). Each deploys as its own ARM child; ids surface
name-keyed in the fqdn_outbound_rule_ids output. Only effective
under ALLOW_ONLY_APPROVED_OUTBOUND isolation.

### spec.fqdnOutboundRules[].name

`string` · required

The rule's name, unique across ALL outbound rules on the
workspace (the three rule lists share one ARM collection). The
rule's ARM id surfaces in the fqdn_outbound_rule_ids output under
this name. Changing the name replaces the rule.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.fqdnOutboundRules[].destinationFqdn

`string` · required

The allowed outbound destination, by domain name (e.g.
"pypi.org", "*.anaconda.com"). Updates in place.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.privateEndpointOutboundRules

`[]AzureMachineLearningWorkspacePrivateEndpointOutboundRule`

Private-endpoint outbound rules on the workspace's managed
network: the managed VNet creates a private endpoint to the named
Azure resource. Each deploys as its own ARM child; ids surface
name-keyed in the private_endpoint_outbound_rule_ids output.

Destroy of a workspace that still has a PE outbound rule races
ARM: workspace delete returns 409 InternalServerError
(privateEndpointConnectionProxies/validate fails in a loop) for
tens of minutes and the provider never finishes. The smoke
scenario therefore proves the workspace object on the default
network and leaves all three outbound-rule children
offline-proven. Delete PE rules and wait for the target-side
connection to drop before deleting the workspace.

- rule: a Microsoft.KeyVault target requires sub_resource_target 'vault'
- rule: a Microsoft.Cache target requires sub_resource_target 'redisCache'
- rule: a Microsoft.MachineLearningServices target requires sub_resource_target 'amlworkspace'
- rule: a Microsoft.Storage target requires sub_resource_target 'blob', 'table', 'queue', 'file', 'web' or 'dfs'

### spec.privateEndpointOutboundRules[].name

`string` · required

The rule's name, unique across ALL outbound rules on the
workspace (the three rule lists share one ARM collection). The
rule's ARM id surfaces in the private_endpoint_outbound_rule_ids
output under this name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.privateEndpointOutboundRules[].serviceResourceId

`string | valueFrom` · required

The Azure resource the managed VNet creates a private endpoint
to, by ARM ID. Supported targets: Key Vault, Storage, Redis
Cache, and other ML workspaces. No default reference kind -- name
the kind explicitly in valueFrom when referencing (the target can
be any of several kinds).

Containment-exempt: an outbound rule is the managed network's door
to a resource the workspace reaches; the workspace is not deployed
into a Key Vault, a storage account, or a cache it merely calls.
Without the exemption a workspace wired to one by reference would
be drawn inside it.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.privateEndpointOutboundRules[].subResourceTarget

`string` · required

Which sub-resource of the target the private endpoint binds
(the wire values). Must match the target's service: Key Vault ->
"vault"; Redis -> "redisCache"; ML workspace -> "amlworkspace";
Storage -> "blob", "table", "queue", "file", "web" or "dfs".

- rule: {"required":true,"string":{"in":["amlworkspace","blob","dfs","file","queue","redisCache","table","vault","web"]}}

### spec.privateEndpointOutboundRules[].sparkEnabled

`bool`

Also wire the private endpoint into the workspace's managed Spark
compute network.

### spec.serviceTagOutboundRules

`[]AzureMachineLearningWorkspaceServiceTagOutboundRule`

Service-tag outbound rules on the workspace's managed network:
allow outbound traffic to an Azure service tag on given protocol
and ports. Each deploys as its own ARM child; ids surface
name-keyed in the service_tag_outbound_rule_ids output. Only
effective under ALLOW_ONLY_APPROVED_OUTBOUND isolation.

### spec.serviceTagOutboundRules[].name

`string` · required

The rule's name, unique across ALL outbound rules on the
workspace (the three rule lists share one ARM collection). The
rule's ARM id surfaces in the service_tag_outbound_rule_ids
output under this name. Changing the name replaces the rule.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.serviceTagOutboundRules[].serviceTag

`string` · required

The Azure service tag traffic is allowed to (the wire values --
the provider's own allowlist).

- rule: {"required":true,"string":{"in":["AppConfiguration","AppService","AppServiceManagement","AutonomousDevelopmentPlatform","AzureActiveDirectory","AzureActiveDirectoryDomainServices","AzureAdvancedThreatProtection","AzureArcInfrastructure","AzureAttestation","AzureBackup","AzureBotService","AzureCloud","AzureConnectors","AzureContainerAppsService","AzureContainerRegistry","AzureCosmosDB","AzureDataLake","AzureDatabricks","AzureDevSpaces","AzureDeviceUpdate","AzureEventGrid","AzureFrontDoor.Backend","AzureFrontDoor.FirstParty","AzureFrontDoor.Frontend","AzureHealthcareAPIs","AzureInformationProtection","AzureIoTHub","AzureKeyVault","AzureLoadBalancer","AzureMachineLearning","AzureManagedGrafana","AzureMonitor","AzureOpenDatasets","AzurePlatformDNS","AzurePlatformIMDS","AzurePlatformLKM","AzureResourceManager","AzureSignalR","AzureSiteRecovery","AzureSphere","AzureSpringCloud","AzureStack","AzureUpdateDelivery","AzureWebPubSub","BatchNodeManagement","ChaosStudio","CognitiveServicesFrontend","CognitiveServicesManagement","DataFactory","DataFactoryManagement","Dynamics365BusinessCentral","Dynamics365ForMarketingEmail","EOPExternalPublishedIPs","EventHub","GuestAndHybridManagement","Internet","LogicApps","M365ManagementActivityApi","Marketplace","MicrosoftAzureFluidRelay","MicrosoftCloudAppSecurity","MicrosoftContainerRegistry","MicrosoftDefenderForEndpoint","PowerBI","PowerPlatformInfra","PowerQueryOnline","ServiceBus","ServiceFabric","Sql","SqlManagement","Storage","StorageSyncService","VirtualNetwork","WindowsAdminCenter","WindowsVirtualDesktop"]}}

### spec.serviceTagOutboundRules[].protocol

`string` · required

The protocol the rule allows (the wire values). "*" allows all.

- rule: {"required":true,"string":{"in":["*","ICMP","TCP","UDP"]}}

### spec.serviceTagOutboundRules[].portRanges

`string` · required

The allowed destination ports: a port ("443"), a range
("1024-65535"), or a comma-separated mix.

- rule: {"required":true,"string":{"minLen":"1"}}

## Validation Rules

- `feature_store_block_only_for_feature_store_kind`: feature_store can only be on when kind is FEATURE_STORE -- set enabled: false or drop the block
- `feature_store_kind_requires_block`: kind FEATURE_STORE requires the feature_store block with its switch on
- `service_side_encryption_requires_encryption`: service_side_encryption_enabled requires the encryption block
- `serverless_no_public_ip_needs_subnet_or_public_workspace`: serverless_compute with public_ip_enabled false requires a subnet_id when public_network_access_enabled is false
- `outbound_rule_names_unique_across_types`: outbound rule names must be unique across the fqdn, private-endpoint and service-tag lists together -- they share one ARM collection

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureMachineLearningWorkspace, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.machine_learning_workspace_id` | `string` | The Azure Resource Manager ID of the workspace -- what datastores, compute, and outbound rules reference as their workspace_id. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.MachineLearningServices/workspaces/{name} |
| `status.outputs.machine_learning_workspace_name` | `string` | The name of the workspace. ARM addresses datastores, compute and outbound rules as children of this name. |
| `status.outputs.workspace_guid` | `string` | The workspace's immutable GUID (distinct from the ARM ID) -- what some data-plane SDKs and diagnostic settings identify the workspace by. |
| `status.outputs.discovery_url` | `string` | The workspace's regional discovery URL -- where SDKs resolve the workspace's data-plane service endpoints from. |
| `status.outputs.system_assigned_identity_principal_id` | `string` | The principal (object) ID of the workspace's system-assigned identity, when one is enabled -- what storage / Key Vault grants bind to. |
| `status.outputs.fqdn_outbound_rule_ids` | `map<string, string>` | The ARM ID of each FQDN outbound rule on the workspace, keyed by the rule's name from the spec. Example valueFrom fieldPath: status.outputs.fqdn_outbound_rule_ids.allow-pypi |
| `status.outputs.private_endpoint_outbound_rule_ids` | `map<string, string>` | The ARM ID of each private-endpoint outbound rule on the workspace, keyed by the rule's name from the spec. Example valueFrom fieldPath: status.outputs.private_endpoint_outbound_rule_ids.to-keyvault |
| `status.outputs.service_tag_outbound_rule_ids` | `map<string, string>` | The ARM ID of each service-tag outbound rule on the workspace, keyed by the rule's name from the spec. Example valueFrom fieldPath: status.outputs.service_tag_outbound_rule_ids.allow-storage |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.applicationInsightsId` | AzureApplicationInsights | `status.outputs.application_insights_id` |
| `spec.keyVaultId` | AzureKeyVault | `status.outputs.key_vault_id` |
| `spec.storageAccountId` | AzureStorageAccount | `status.outputs.storage_account_id` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.primaryUserAssignedIdentity` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.containerRegistryId` | AzureContainerRegistry | `status.outputs.container_registry_id` |
| `spec.encryption.keyVaultId` | AzureKeyVault | `status.outputs.key_vault_id` |
| `spec.encryption.keyId` | AzureKeyVaultKey | `status.outputs.versionless_id` |
| `spec.encryption.userAssignedIdentityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.serverlessCompute.subnetId` | AzureSubnet | `status.outputs.subnet_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureMachineLearningBatchEndpoint | `spec.workspaceId` | `status.outputs.machine_learning_workspace_id` |
| AzureMachineLearningComputeCluster | `spec.workspaceId` | `status.outputs.machine_learning_workspace_id` |
| AzureMachineLearningComputeInstance | `spec.workspaceId` | `status.outputs.machine_learning_workspace_id` |
| AzureMachineLearningDatastore | `spec.workspaceId` | `status.outputs.machine_learning_workspace_id` |
| AzureMachineLearningOnlineEndpoint | `spec.workspaceId` | `status.outputs.machine_learning_workspace_id` |

## See Also

- [Overview](../README.md)
