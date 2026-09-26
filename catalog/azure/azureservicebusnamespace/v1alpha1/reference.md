# AzureServiceBusNamespace

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureServiceBusNamespaceSpec** defines the configuration for creating an
Azure Service Bus namespace: the container and billing boundary for
enterprise messaging entities.

Azure Service Bus is a fully managed enterprise message broker providing
queues (point-to-point) and topics with subscriptions (publish-subscribe),
with ordered delivery, sessions, duplicate detection, dead-lettering, and
transactions. The namespace is where the pricing tier, network posture,
encryption, and authentication mode are set; the messaging entities are
first-class kinds that reference it:
`AzureServiceBusQueue`, `AzureServiceBusTopic` (with
`AzureServiceBusSubscription` under a topic), `AzureServiceBusAuthorizationRule`
(SAS credentials scoped to the namespace or a single entity), and
`AzureServiceBusDisasterRecoveryConfig` (the geo-DR alias pairing two
Premium namespaces).

**SKU tiers**:
- **BASIC**: queues only -- no topics, sessions, or duplicate detection.
  Simple fire-and-forget scenarios.
- **STANDARD** (default): full-featured multi-tenant tier with topics,
  subscriptions, sessions, and duplicate detection. The right choice for
  most production workloads.
- **PREMIUM**: dedicated messaging units with predictable latency, VNet
  integration (network_rule_set), customer-managed-key encryption, geo-DR,
  namespace partitioning, and large messages (up to 100 MB). Migrating a
  namespace into or out of PREMIUM replaces it (Azure cannot convert in
  place across the dedicated/multi-tenant boundary).

**Authentication posture**: every namespace carries a root SAS rule
(`RootManageSharedAccessKey`) whose keys and connection strings surface as
sensitive outputs. Scoped credentials belong in
`AzureServiceBusAuthorizationRule`; for a keyless posture, disable
local_auth_enabled and grant Entra identities data-plane roles
(Azure Service Bus Data Owner/Sender/Receiver) via `AzureRoleAssignment`.

**ForceNew fields** (changing these replaces the namespace -- and destroys
every entity in it): `namespace_name`, `premium_messaging_partitions`,
`sku` when moving into or out of PREMIUM, and removing
`customer_managed_key` once set.

## Example

```yaml
# Offline-plan manifest: a PREMIUM namespace exercising every deep seam
# at once -- the capacity/partitions pairing, the combined identity
# model, CMK (key + unwrapping identity), the DENY firewall with both
# admitted-source lists, the keyless posture, and user tags.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureServiceBusNamespace
metadata:
  name: test-sb-namespace
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  namespaceName: hack-servicebus-ns
  sku: PREMIUM
  capacity: 2
  premiumMessagingPartitions: 2
  identity:
    type: SYSTEM_AND_USER_ASSIGNED
    userAssignedIdentityIds:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/hack-uai
  customerManagedKey:
    keyVaultKeyId:
      value: https://hack-vault.vault.azure.net/keys/sb-cmk
    userAssignedIdentityId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/hack-uai
    infrastructureEncryptionEnabled: true
  localAuthEnabled: false
  publicNetworkAccessEnabled: true
  networkRuleSet:
    defaultAction: DENY
    trustedServicesAllowed: true
    ipRules:
      - 203.0.113.0/24
    networkRules:
      - subnetId:
          value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/hack-vnet/subnets/app
        ignoreMissingVnetServiceEndpoint: true
  tags:
    team: messaging
    cost-center: cc-42
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.namespaceName` | `string` | yes |  |  |
| `spec.sku` | `enum` |  |  |  |
| `spec.capacity` | `int32` |  |  |  |
| `spec.premiumMessagingPartitions` | `int32` |  |  |  |
| `spec.identity` | `AzureServiceBusNamespaceIdentity` |  |  |  |
| `spec.identity.type` | `enum` |  |  |  |
| `spec.identity.userAssignedIdentityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.customerManagedKey` | `AzureServiceBusNamespaceCustomerManagedKey` |  |  |  |
| `spec.customerManagedKey.keyVaultKeyId` | `string \| valueFrom` | yes |  | AzureKeyVaultKey (`status.outputs.versionless_id`) |
| `spec.customerManagedKey.userAssignedIdentityId` | `string \| valueFrom` | yes |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.customerManagedKey.infrastructureEncryptionEnabled` | `bool` |  |  |  |
| `spec.localAuthEnabled` | `bool` |  | `true` |  |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.networkRuleSet` | `AzureServiceBusNamespaceNetworkRuleSet` |  |  |  |
| `spec.networkRuleSet.defaultAction` | `enum` |  |  |  |
| `spec.networkRuleSet.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.networkRuleSet.trustedServicesAllowed` | `bool` |  |  |  |
| `spec.networkRuleSet.ipRules` | `[]string` |  |  |  |
| `spec.networkRuleSet.networkRules` | `[]AzureServiceBusNamespaceNetworkRule` |  |  |  |
| `spec.networkRuleSet.networkRules[].subnetId` | `string \| valueFrom` | yes |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.networkRuleSet.networkRules[].ignoreMissingVnetServiceEndpoint` | `bool` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region where the Service Bus namespace is created.
Examples: "eastus", "westus2", "westeurope", "southeastasia".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group the namespace lives in.
Can be a literal string or a reference to an AzureResourceGroup output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.namespaceName

`string` · required

The namespace name -- globally unique across Azure, because it becomes
the public endpoint `{name}.servicebus.windows.net`.

6-50 characters; starts with a letter; ends with a letter or number;
letters, numbers, and hyphens only. Azure additionally reserves the
suffixes "-sb" and "-mgmt".

**ForceNew**: changing the name replaces the namespace and every entity
in it. Treat it as permanent.

- rule: namespace_name must be 6-50 characters of letters, numbers, and hyphens, starting with a letter and ending with a letter or number
- rule: namespace_name cannot end with '-sb' or '-mgmt' -- Azure reserves these suffixes for its own service endpoints
- rule: {"required":true,"string":{"minLen":"6","maxLen":"50"}}

### spec.sku

`enum`

The pricing tier. Unspecified deploys STANDARD -- the full-featured
multi-tenant tier that fits most production workloads. Choose PREMIUM
for dedicated capacity, VNet integration, customer-managed keys, geo-DR,
or messages larger than 256 KB; choose BASIC only for simple queue-only
scenarios (no topics).

Moving a namespace into or out of PREMIUM replaces it; BASIC <->
STANDARD updates in place.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_service_bus_namespace_sku_unspecified` -- Not specified -- deploys STANDARD, the full-featured multi-tenant tier.
- `BASIC` -- Queues only -- no topics, sessions, or duplicate detection. Simple fire-and-forget scenarios.
- `STANDARD` -- Full-featured multi-tenant tier: queues, topics, subscriptions, sessions, duplicate detection. The right choice for most production workloads.
- `PREMIUM` -- Dedicated messaging units: predictable latency, VNet integration, customer-managed keys, geo-DR, partitioning, 100 MB messages. Moving into or out of PREMIUM replaces the namespace.

### spec.capacity

`int32` · optional (explicit presence)

Messaging units for the PREMIUM tier -- the dedicated processing
capacity of the namespace. Required with (and only valid with) PREMIUM.
Allowed values: 1, 2, 4, 8, 16. Each messaging unit provides roughly
1 MB/s ingress and 2 MB/s egress with predictable latency; scale up for
high-throughput workloads (scaling updates in place).

- rule: {"int32":{"in":[1,2,4,8,16]}}

### spec.premiumMessagingPartitions

`int32` · optional (explicit presence)

Namespace partitions for the PREMIUM tier. Partitioning spreads the
namespace across multiple message stores, multiplying throughput beyond
what one store sustains. Allowed values: 1 (not partitioned -- the
standard choice), 2, or 4. Every queue and topic in a partitioned
namespace must set partitioning_enabled to true; in a non-partitioned
namespace they must not.

**ForceNew**: the partition layout is fixed at creation -- changing it
replaces the namespace.

- rule: {"int32":{"in":[1,2,4]}}

### spec.identity

`AzureServiceBusNamespaceIdentity`

Managed identity for the namespace -- required for customer-managed-key
encryption (the identity unwraps the key) and usable anywhere the
namespace itself must authenticate to other Azure services.

- rule: user_assigned_identity_ids is required with USER_ASSIGNED or SYSTEM_AND_USER_ASSIGNED, and must be empty with SYSTEM_ASSIGNED

### spec.identity.type

`enum`

The identity model: SYSTEM_ASSIGNED (Azure creates and rotates a
service principal bound to the namespace's lifecycle), USER_ASSIGNED
(bring identities from user_assigned_identity_ids, shareable across
resources), or SYSTEM_AND_USER_ASSIGNED (both).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_service_bus_namespace_identity_type_unspecified` -- Not specified -- invalid; choose an explicit identity model.
- `SYSTEM_ASSIGNED` -- Azure creates a service principal bound to the namespace's lifecycle.
- `USER_ASSIGNED` -- Bring your own AzureUserAssignedIdentity entries -- shareable across resources and grantable before the namespace exists.
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned principal and user-assigned identities.

### spec.identity.userAssignedIdentityIds

`[]string | valueFrom`

The user-assigned identities to attach -- required when (and only
meaningful when) type includes USER_ASSIGNED. Each entry references
an AzureUserAssignedIdentity's ARM id.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.customerManagedKey

`AzureServiceBusNamespaceCustomerManagedKey`

Customer-managed-key encryption for messaging data at rest (BYOK).
PREMIUM only, and the unwrapping user-assigned identity must be attached
via the identity block. Once enabled, CMK cannot be removed -- dropping
this block replaces the namespace (Azure's own contract).

### spec.customerManagedKey.keyVaultKeyId

`string | valueFrom` · required

The Key Vault key that encrypts messaging data, by data-plane key ID.
Defaults to referencing an AzureKeyVaultKey's versionless_id output so
key rotations propagate automatically; pin a versioned ID only when a
compliance regime demands an immutable key version. The key's vault
must have purge protection enabled.

- references: AzureKeyVaultKey (`status.outputs.versionless_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVaultKey, name: <that resource's name>, fieldPath: status.outputs.versionless_id}} -- a bare string does not parse

### spec.customerManagedKey.userAssignedIdentityId

`string | valueFrom` · required

The user-assigned identity Azure uses to unwrap the key, by ARM ID.
Must be one of the identities attached via the namespace's identity
block, with wrap/unwrap access on the key's vault (a "Key Vault Crypto
Service Encryption User" role assignment, or the equivalent access
policy).

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.customerManagedKey.infrastructureEncryptionEnabled

`bool` · optional (explicit presence)

Whether Azure applies a second layer of encryption (infrastructure
encryption) beneath the customer-managed key. **ForceNew**: fixed at
the moment CMK is first configured.

### spec.localAuthEnabled

`bool` · optional (explicit presence)

Whether SAS (shared-access-signature) authentication is allowed on the
namespace. Azure's default is true. Set false for a keyless posture:
clients must then authenticate with Microsoft Entra ID identities, and
every SAS rule's keys -- including the root rule surfaced in this kind's
outputs -- stop being usable credentials.
Default: true

- default: `true`

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the namespace accepts traffic from the public internet. When
false, clients reach it only through private endpoints
(AzurePrivateEndpoint) or admitted VNet service endpoints.
Default: true

- default: `true`

### spec.networkRuleSet

`AzureServiceBusNamespaceNetworkRuleSet`

The namespace firewall: which networks may reach the data plane.
PREMIUM only (multi-tenant tiers do not support VNet integration or IP
filtering). Declaring the block with DENY and no admitted networks is
rejected -- Azure requires at least one admitted source before closing
the default.

- rule: default_action DENY needs at least one admitted source -- add ip_rules or network_rules, or the namespace would reject all data-plane traffic (Azure refuses this configuration)

### spec.networkRuleSet.defaultAction

`enum`

What happens to traffic no explicit rule admits. Unspecified keeps
Azure's open default (ALLOW). DENY plus the admitted lists below is
the production posture -- but Azure rejects DENY with no admitted
ip_rules or network_rules.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_service_bus_network_default_action_unspecified` -- Not specified -- keeps Azure's open default (ALLOW).
- `ALLOW` -- Admit traffic no rule matches (the firewall only annotates).
- `DENY` -- Reject traffic no rule matches -- the production posture. Requires at least one admitted ip_rule or network_rule.

### spec.networkRuleSet.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the data plane remains reachable over public IP space for the
sources the rules admit. Set false to force all admitted traffic
through private endpoints and service endpoints only.
Default: true

- default: `true`

### spec.networkRuleSet.trustedServicesAllowed

`bool` · optional (explicit presence)

Whether trusted Microsoft services (Event Grid delivery, Azure Monitor
diagnostics, IoT Hub routing, ...) bypass the firewall. Azure's
default is false; enable it when platform services must deliver into a
locked-down namespace.

### spec.networkRuleSet.ipRules

`[]string`

Public IPv4 addresses or CIDR ranges admitted to the data plane,
e.g. "203.0.113.0/24".

### spec.networkRuleSet.networkRules

`[]AzureServiceBusNamespaceNetworkRule`

VNet subnets admitted to the data plane via service endpoints. Each
subnet should carry the Microsoft.ServiceBus service endpoint.

### spec.networkRuleSet.networkRules[].subnetId

`string | valueFrom` · required

The admitted subnet, by ARM ID. References an AzureSubnet's subnet_id
output.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.networkRuleSet.networkRules[].ignoreMissingVnetServiceEndpoint

`bool` · optional (explicit presence)

Whether to admit the subnet even if it does not (yet) carry the
Microsoft.ServiceBus service endpoint -- useful when the endpoint is
being rolled out separately. Azure's default is false.

### spec.tags

`map<string, string>`

Tags to apply to the namespace, merged over the Planton-derived
metadata tags (user values win on key conflicts). ARM tags are Azure's
first-class governance surface -- Azure Policy enforces them and
Microsoft Cost Management groups by them.

## Validation Rules

- `service_bus_capacity_premium_only`: capacity is the PREMIUM tier's messaging-unit dial -- remove it on BASIC/STANDARD, where Azure fixes capacity at 0
- `service_bus_premium_requires_capacity`: the PREMIUM tier requires capacity -- pick 1, 2, 4, 8, or 16 messaging units (1 fits most workloads to start; scaling up later updates in place)
- `service_bus_partitions_premium_only`: premium_messaging_partitions is a PREMIUM-tier setting -- remove it on BASIC/STANDARD
- `service_bus_premium_requires_partitions`: the PREMIUM tier requires premium_messaging_partitions -- set 1 unless you deliberately want a partitioned namespace (2 or 4); the layout is fixed at creation
- `service_bus_cmk_requires_premium`: customer_managed_key encryption is only available on the PREMIUM tier -- Azure rejects BYOK on multi-tenant namespaces
- `service_bus_cmk_requires_user_assigned_identity`: customer_managed_key requires the identity block with USER_ASSIGNED or SYSTEM_AND_USER_ASSIGNED -- the unwrapping identity must be attached to the namespace
- `service_bus_network_rules_require_premium`: network_rule_set is only available on the PREMIUM tier -- multi-tenant namespaces do not support VNet integration or IP filtering

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureServiceBusNamespace, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace_id` | `string` | The Azure Resource Manager ID of the namespace. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.ServiceBus/namespaces/{name} The parent reference for every Service Bus child kind, the scope for namespace-wide data-plane RBAC, and the private-endpoint target (subresource name: "namespace"). |
| `status.outputs.namespace_name` | `string` | The namespace name -- what SDKs and connection strings identify the namespace by. |
| `status.outputs.endpoint` | `string` | The Service Bus endpoint URL. Format: https://{name}.servicebus.windows.net:443/ |
| `status.outputs.identity_principal_id` | `string` | The system-assigned identity's principal (object) ID -- grant this identity access on other resources (e.g. Key Vault for CMK). Empty unless the identity block includes SYSTEM_ASSIGNED. |
| `status.outputs.default_primary_connection_string` | `string` (sensitive) | The root SAS rule's primary connection string (full manage rights on the namespace). Secret-bearing: it embeds the primary key. Format: Endpoint=sb://{name}.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey={key} |
| `status.outputs.default_secondary_connection_string` | `string` (sensitive) | The root SAS rule's secondary connection string -- the rotation partner: move clients here, regenerate the primary, move back. |
| `status.outputs.default_primary_key` | `string` (sensitive) | The root SAS rule's primary key, for SDKs that take the key and key name separately or mint their own SAS tokens. |
| `status.outputs.default_secondary_key` | `string` (sensitive) | The root SAS rule's secondary key -- the rotation partner. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.identity.userAssignedIdentityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.customerManagedKey.keyVaultKeyId` | AzureKeyVaultKey | `status.outputs.versionless_id` |
| `spec.customerManagedKey.userAssignedIdentityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.networkRuleSet.networkRules[].subnetId` | AzureSubnet | `status.outputs.subnet_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureServiceBusAuthorizationRule | `spec.namespaceId` | `status.outputs.namespace_id` |
| AzureServiceBusDisasterRecoveryConfig | `spec.primaryNamespaceId` | `status.outputs.namespace_id` |
| AzureServiceBusDisasterRecoveryConfig | `spec.partnerNamespaceId` | `status.outputs.namespace_id` |
| AzureServiceBusQueue | `spec.namespaceId` | `status.outputs.namespace_id` |
| AzureServiceBusTopic | `spec.namespaceId` | `status.outputs.namespace_id` |

## See Also

- [Overview](../README.md)
