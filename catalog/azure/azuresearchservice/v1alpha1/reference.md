# AzureSearchService

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureSearchServiceSpec** defines an Azure AI Search service (ARM:
Microsoft.Search/searchServices) -- the managed search-and-retrieval
engine AI applications use to index and query their own data
(keyword, vector, and semantic search). AI Search is the standard
retrieval companion to Azure OpenAI (the "R" in RAG).

Capacity is sku x partitions x replicas: the SKU fixes the
per-unit size and price class, partitions scale storage and
indexing, replicas scale query throughput and availability.
Partitions and replicas resize in place; the SKU upgrades in place
ONLY along basic -> standard -> standard2 -> standard3 -- every
other SKU change (downgrades, free, storage-optimized) replaces
the service (the provider's own update contract).

## Example

```yaml
# Offline-plan test manifest. Exercises the deep seams: the
# high-density hosting mode on standard3 (the lookup-mapped enum seam
# -- the plan must render "highDensity"), per-SKU-legal counts, the
# RBAC-alongside-API-keys auth mode, semantic ranking, the IP
# firewall with the AzureServices bypass, a system identity, and one
# composed shared private link to a storage blob target (the ARM
# child keyed by name into the shared_private_link_service_ids
# output).
apiVersion: azure.planton.dev/v1alpha1
kind: AzureSearchService
metadata:
  name: test-search-service
  org: test-org
  env: dev
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: test-acme-search
  sku: standard3
  hostingMode: HIGH_DENSITY
  replicaCount: 2
  partitionCount: 3
  localAuthenticationEnabled: true
  authenticationFailureMode: http401WithBearerChallenge
  customerManagedKeyEnforcementEnabled: false
  publicNetworkAccessEnabled: true
  semanticSearchSku: standard
  allowedIps:
    - 203.0.113.7
    - 203.0.113.0/24
  networkRuleBypassOption: AzureServices
  identity:
    type: SYSTEM_ASSIGNED
  sharedPrivateLinkServices:
    - name: to-blob
      subresourceName: blob
      targetResourceId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Storage/storageAccounts/testsearchdata
      requestMessage: Please approve the search indexer's private access
  tags:
    cost-center: ai-platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.sku` | `string` | yes |  |  |
| `spec.replicaCount` | `int32` |  | `1` |  |
| `spec.partitionCount` | `int32` |  | `1` |  |
| `spec.hostingMode` | `enum` |  |  |  |
| `spec.localAuthenticationEnabled` | `bool` |  | `true` |  |
| `spec.authenticationFailureMode` | `string` |  |  |  |
| `spec.customerManagedKeyEnforcementEnabled` | `bool` |  |  |  |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.semanticSearchSku` | `string` |  |  |  |
| `spec.allowedIps` | `[]string` |  |  |  |
| `spec.networkRuleBypassOption` | `string` |  |  |  |
| `spec.identity` | `AzureSearchServiceIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.tags` | `map<string, string>` |  |  |  |
| `spec.sharedPrivateLinkServices` | `[]AzureSearchServiceSharedPrivateLink` |  |  |  |
| `spec.sharedPrivateLinkServices[].name` | `string` | yes |  |  |
| `spec.sharedPrivateLinkServices[].subresourceName` | `string` | yes |  |  |
| `spec.sharedPrivateLinkServices[].targetResourceId` | `string \| valueFrom` | yes |  |  |
| `spec.sharedPrivateLinkServices[].requestMessage` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the service lives in, e.g. "eastus". Higher
SKUs are not available in every region. Changing the region
replaces the service.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the service is created in. Can be a
literal resource-group name or a reference to an
AzureResourceGroup's name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The service's name -- GLOBALLY unique across all of Azure: it
forms the service's endpoint, https://{name}.search.windows.net.
The provider deliberately carries no format rule (ARM enforces
its own naming at create); a name-taken error at deploy means a
genuine global collision. Changing the name replaces the
service.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.sku

`string` · required

The service's pricing tier (the wire values). "free" is one
shared-cluster service per subscription; "basic" and the
"standard" tiers are the paid dedicated clusters; the
"storage_optimized_l1"/"l2" tiers trade query latency for large
indexes. NOTE: "standard2", "standard3" and both
storage-optimized tiers require a quota increase request to
Microsoft before ARM accepts them. In-place upgrade works ONLY
along basic -> standard -> standard2 -> standard3; every other
change replaces the service.

- rule: {"required":true,"string":{"in":["basic","free","standard","standard2","standard3","storage_optimized_l1","storage_optimized_l2"]}}

### spec.replicaCount

`int32` · optional (explicit presence)

How many replicas the service runs -- query throughput and
availability (3+ replicas give a 99.9% read-write SLA).
Unspecified applies 1. Caps: 1 on "free", 3 on "basic", 12
elsewhere. Resizes in place.

- default: `1`
- rule: {"int32":{"lte":12,"gte":1}}

### spec.partitionCount

`int32` · optional (explicit presence)

How many partitions the service runs -- index storage and
indexing throughput. Unspecified applies 1. Legal values are 1,
2, 3, 4, 6 or 12; capped at 1 on "free", 3 on "basic", and 3 in
high-density hosting. Resizes in place.

- default: `1`
- rule: {"int32":{"in":[1,2,3,4,6,12]}}

### spec.hostingMode

`enum`

The hosting mode. HIGH_DENSITY packs up to 3000 small indexes
per service (multi-tenant SaaS pattern) and is only legal on the
"standard3" SKU. Unspecified applies DEFAULT. Fixed at creation.

Allowed values (use exactly as shown):

- `azure_search_service_hosting_mode_unspecified` -- Not specified: the provider applies its default, "default".
- `DEFAULT` -- Normal index density (wire value "default").
- `HIGH_DENSITY` -- Up to 3000 small indexes on one service, standard3 SKU only (wire value "highDensity").

### spec.localAuthenticationEnabled

`bool` · optional (explicit presence)

Whether API-key authentication is enabled on the data plane.
Unspecified applies true. Set false for the RBAC-only posture --
the admin/query keys stop working, and
authentication_failure_mode must stay unset (the provider's own
contract). Updates in place.

- default: `true`

### spec.authenticationFailureMode

`string`

How the service answers data-plane requests that fail Microsoft
Entra authentication when BOTH auth modes are on (the wire
values): "http401WithBearerChallenge" or "http403". Setting this
is what enables the RBAC-alongside-API-keys mode; leave unset
for API keys only. Only legal while
local_authentication_enabled is true.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["http401WithBearerChallenge","http403"]}}

### spec.customerManagedKeyEnforcementEnabled

`bool`

Enforce that every index/indexer uses a customer-managed key --
the service reports non-compliant objects and blocks new ones
without CMK. The keys themselves are configured per index at the
data plane. Updates in place.

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the service's endpoint answers the public internet.
Unspecified applies true. Set false to make the service
reachable only through private endpoints. Updates in place.

- default: `true`

### spec.semanticSearchSku

`string`

The semantic-ranking tier (the wire values): "free" (1000
requests/month) or "standard" (unlimited, billed). Leave unset
to keep semantic ranking disabled. Not available on the "free"
service SKU. Updates in place.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["free","standard"]}}

### spec.allowedIps

`[]string`

IPv4 addresses or CIDR ranges allowed through the service's
firewall (e.g. "203.0.113.7" or "203.0.113.0/24"). An empty list
leaves the endpoint open to the internet while
public_network_access_enabled is true. Updates in place.

- rule: {"repeated":{"items":{"string":{"pattern":"^([0-9]{1,3}\\.){3}[0-9]{1,3}(/([0-9]|[1-2][0-9]|3[0-2]))?$"}}}}

### spec.networkRuleBypassOption

`string`

Whether trusted Azure services may bypass the IP firewall (the
wire values): "AzureServices" or "None". Unspecified applies
"None". Updates in place.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AzureServices","None"]}}

### spec.identity

`AzureSearchServiceIdentity`

The service's managed identity -- what indexers use to reach
data sources without connection-string secrets.

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the service; USER_ASSIGNED brings identities you manage
(grantable data-source access BEFORE the service exists);
SYSTEM_AND_USER_ASSIGNED carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_search_service_identity_type_unspecified` -- Not specified: rejected when the identity block is present.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the service.
- `USER_ASSIGNED` -- Identities you create and manage (AzureUserAssignedIdentity).
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned and user-assigned identities.

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the service, by ARM ID.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.tags

`map<string, string>`

Free-form tags applied to the service, merged over the
Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins.

### spec.sharedPrivateLinkServices

`[]AzureSearchServiceSharedPrivateLink`

Shared private links from THIS search service out to other Azure
resources -- how the service's indexers reach data sources that
sit behind private endpoints. Each entry deploys as its own ARM
child (.../sharedPrivateLinkResources/{name}); ids surface
name-keyed in the shared_private_link_service_ids output. The
target side must APPROVE the connection before traffic flows
(the link sits "Pending" until then).

### spec.sharedPrivateLinkServices[].name

`string` · required

The link's name, unique within the service. The link's ARM id
surfaces in the shared_private_link_service_ids output under
this name. Changing the name replaces the link.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.sharedPrivateLinkServices[].subresourceName

`string` · required

The target's sub-resource (group id) the private link binds --
must match the target resource's type: a storage account takes
"blob"/"table"/"queue"/"file"/"dfs"/"web", a SQL server
"sqlServer", a Key Vault "vault", a Cognitive/OpenAI account
"account", another search-reachable service its own documented
group id. 3-63 characters, alphanumeric at both ends.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9][a-zA-Z0-9_.-]{1,61}[a-zA-Z0-9]$"}}

### spec.sharedPrivateLinkServices[].targetResourceId

`string | valueFrom` · required

The Azure resource the private link points at, by ARM ID. No
default reference kind -- the target can be any of several kinds
(storage account, SQL server, Key Vault, Cognitive account, ...);
name the kind explicitly in valueFrom when referencing.

Containment-exempt: a shared private link is the search service's
outbound door to a resource it indexes or enriches from; the
service is not deployed into that resource. Every listed target is
a container, and a service linked by reference would otherwise be
drawn inside the store it reads.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.sharedPrivateLinkServices[].requestMessage

`string`

The approval-request message shown to the target resource's
owner (the link sits "Pending" until they approve). Updates in
place -- the only field on the link that does.

## Validation Rules

- `high_density_requires_standard3`: hosting_mode HIGH_DENSITY is only available on the standard3 sku
- `partition_count_free_sku_cap`: the free sku supports a single partition -- leave partition_count unset or 1
- `partition_count_basic_sku_cap`: the basic sku supports at most 3 partitions
- `partition_count_high_density_cap`: standard3 in HIGH_DENSITY hosting supports at most 3 partitions
- `replica_count_free_sku_cap`: the free sku supports a single replica -- leave replica_count unset or 1
- `replica_count_basic_sku_cap`: the basic sku supports at most 3 replicas
- `auth_failure_mode_requires_local_auth`: authentication_failure_mode only applies while local_authentication_enabled is true -- in the RBAC-only posture, remove it
- `semantic_search_not_on_free_sku`: semantic_search_sku cannot be set on the free service sku
- `shared_private_link_names_unique`: shared private link names must be unique within the service

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureSearchService, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.search_service_id` | `string` | The Azure Resource Manager ID of the service -- what shared private links and diagnostic settings reference. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Search/searchServices/{name} |
| `status.outputs.search_service_name` | `string` | The name of the service. ARM addresses shared private links as children of this name, and the data-plane endpoint embeds it. |
| `status.outputs.endpoint` | `string` | The service's data-plane endpoint (https://{name}.search.windows.net) -- what applications and SDKs call. |
| `status.outputs.primary_key` | `string` | The primary admin API key -- full read-write control of the service's data plane. Sensitive: treat as a credential (the service mints it; there is no vault indirection). Empty when local authentication is disabled. |
| `status.outputs.secondary_key` | `string` | The secondary admin API key -- the rotation partner of primary_key. Sensitive: treat as a credential. Empty when local authentication is disabled. |
| `status.outputs.default_query_key` | `string` | The service's built-in query API key -- read-only data-plane access for client applications (the service creates exactly one unnamed query key at provisioning; create more at the data plane). Sensitive: treat as a credential. Empty when local authentication is disabled. |
| `status.outputs.customer_managed_key_encryption_compliance_status` | `string` | Whether the service's objects comply with the customer-managed- key enforcement posture ("Compliant" / "NonCompliant") -- ARM's own assessment, read back at deploy time. |
| `status.outputs.system_assigned_identity_principal_id` | `string` | The principal (object) ID of the service's system-assigned identity, when one is enabled -- what data-source grants for indexers bind to. |
| `status.outputs.shared_private_link_service_ids` | `map<string, string>` | The ARM ID of each shared private link on the service, keyed by the link's name from the spec. Example valueFrom fieldPath: status.outputs.shared_private_link_service_ids.to-blob |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureDataFactoryLinkedService | `spec.azureSearch.url` | `status.outputs.endpoint` |

## See Also

- [Overview](../README.md)
