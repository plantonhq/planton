# AzureEventHubNamespace

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureEventHubNamespaceSpec** defines the configuration for creating an
Azure Event Hubs namespace: the container and billing boundary for
high-throughput event streaming.

Azure Event Hubs is a fully managed, real-time data ingestion service that
receives and processes millions of events per second -- telemetry
collection, log aggregation, IoT ingestion, change-data capture, and the
Kafka-compatible endpoint for existing Kafka producers and consumers. The
namespace is where the pricing tier, throughput capacity, network posture,
and authentication mode are set; the streaming entities are first-class
kinds that reference it: `AzureEventHub` (the event stream, with its
capture-to-storage block), `AzureEventHubConsumerGroup` (independent read
cursors), `AzureEventHubAuthorizationRule` (SAS credentials scoped to the
namespace or a single hub), `AzureEventHubSchemaGroup` (the schema
registry), `AzureEventHubDisasterRecoveryConfig` (the geo-DR alias
pairing two namespaces), and `AzureEventHubNamespaceCustomerManagedKey`
(BYOK encryption for namespaces on a dedicated cluster).

**SKU tiers**:
- **BASIC**: 1-day retention, a single consumer group per hub, no Kafka
  endpoint, no VNet firewall. Simple telemetry ingestion only.
- **STANDARD** (default): the full-featured multi-tenant tier -- Kafka
  endpoint, up to 20 consumer groups per hub, 7-day retention,
  auto-inflate for elastic throughput. The right choice for most
  production workloads.
- **PREMIUM**: reserved processing units with predictable latency,
  dynamic partition scale-up, extended retention (up to 90 days), and
  VNet integration. Choose it for isolation and compliance workloads
  that do not need a whole dedicated cluster.

**Capacity semantics differ by tier**: on BASIC/STANDARD, capacity is
throughput units (TUs; 1 MB/s ingress and 2 MB/s egress each); on
PREMIUM it is processing units (PUs) of reserved compute. Auto-inflate
(STANDARD's elastic scaling) grows TUs up to maximum_throughput_units
under load but never shrinks them back -- scale-down is a manual
capacity edit.

**Authentication posture**: every namespace carries a root SAS rule
(`RootManageSharedAccessKey`) whose keys and connection strings surface
as sensitive outputs. Scoped credentials belong in
`AzureEventHubAuthorizationRule`; for a keyless posture, disable
local_authentication_enabled and grant Entra identities data-plane roles
(Azure Event Hubs Data Owner/Sender/Receiver) via `AzureRoleAssignment`.

**ForceNew fields** (changing these replaces the namespace -- and
destroys every entity in it): `namespace_name`, `dedicated_cluster_id`,
and `sku` when moving into or out of PREMIUM (Azure cannot convert a
namespace across the reserved/multi-tenant boundary in place;
BASIC <-> STANDARD updates in place).

## Example

```yaml
# Offline-plan manifest: a STANDARD namespace exercising every deep seam
# at once -- elastic throughput (auto-inflate + ceiling), the combined
# identity model, the DENY firewall with both admitted-source lists, the
# keyless posture, and user tags. (PREMIUM/dedicated placement and CMK
# are exercised by their own kinds' manifests.)
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventHubNamespace
metadata:
  name: test-eh-namespace
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  namespaceName: hack-eventhubs-ns
  sku: STANDARD
  capacity: 2
  autoInflateEnabled: true
  maximumThroughputUnits: 10
  identity:
    type: SYSTEM_AND_USER_ASSIGNED
    userAssignedIdentityIds:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/hack-uai
  localAuthenticationEnabled: false
  publicNetworkAccessEnabled: true
  networkRuleSets:
    defaultAction: DENY
    trustedServiceAccessEnabled: true
    ipRules:
      - 203.0.113.0/24
    virtualNetworkRules:
      - subnetId:
          value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/hack-vnet/subnets/app
        ignoreMissingVirtualNetworkServiceEndpoint: true
  tags:
    team: streaming
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
| `spec.autoInflateEnabled` | `bool` |  |  |  |
| `spec.maximumThroughputUnits` | `int32` |  |  |  |
| `spec.dedicatedClusterId` | `string \| valueFrom` |  |  | AzureEventHubCluster (`status.outputs.cluster_id`) |
| `spec.identity` | `AzureEventHubNamespaceIdentity` |  |  |  |
| `spec.identity.type` | `enum` |  |  |  |
| `spec.identity.userAssignedIdentityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.localAuthenticationEnabled` | `bool` |  | `true` |  |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.networkRuleSets` | `AzureEventHubNamespaceNetworkRuleSets` |  |  |  |
| `spec.networkRuleSets.defaultAction` | `enum` |  |  |  |
| `spec.networkRuleSets.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.networkRuleSets.trustedServiceAccessEnabled` | `bool` |  |  |  |
| `spec.networkRuleSets.ipRules` | `[]string` |  |  |  |
| `spec.networkRuleSets.virtualNetworkRules` | `[]AzureEventHubNamespaceVirtualNetworkRule` |  |  |  |
| `spec.networkRuleSets.virtualNetworkRules[].subnetId` | `string \| valueFrom` | yes |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.networkRuleSets.virtualNetworkRules[].ignoreMissingVirtualNetworkServiceEndpoint` | `bool` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region where the Event Hubs namespace is created.
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
the public endpoint `{name}.servicebus.windows.net` (Event Hubs shares
the Service Bus DNS zone) and the Kafka bootstrap host
`{name}.servicebus.windows.net:9093`.

6-50 characters; starts with a letter; ends with a letter or number;
letters, numbers, and hyphens only.

**ForceNew**: changing the name replaces the namespace and every
entity in it. Treat it as permanent.

- rule: namespace_name must be 6-50 characters of letters, numbers, and hyphens, starting with a letter and ending with a letter or number
- rule: {"required":true,"string":{"minLen":"6","maxLen":"50"}}

### spec.sku

`enum`

The pricing tier. Unspecified deploys STANDARD -- the full-featured
multi-tenant tier with the Kafka endpoint and auto-inflate. Choose
PREMIUM for reserved capacity, VNet-scale isolation, and extended
retention; choose BASIC only for simple single-consumer telemetry.

Moving a namespace into or out of PREMIUM replaces it; BASIC <->
STANDARD updates in place.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_event_hub_namespace_sku_unspecified` -- Not specified -- deploys STANDARD, the full-featured multi-tenant tier.
- `BASIC` -- 1-day retention, single consumer group per hub, no Kafka endpoint. Simple telemetry ingestion only.
- `STANDARD` -- The full-featured multi-tenant tier: Kafka endpoint, 20 consumer groups per hub, 7-day retention, auto-inflate. The right choice for most production workloads.
- `PREMIUM` -- Reserved processing units: predictable latency, dynamic partition scale-up, extended retention, VNet integration. Moving into or out of PREMIUM replaces the namespace.

### spec.capacity

`int32` · optional (explicit presence)

The namespace's capacity. On BASIC/STANDARD: throughput units (TUs,
1-40; each provides 1 MB/s ingress and 2 MB/s egress) -- Azure's
default is 1. On PREMIUM: processing units (PUs) of reserved compute;
Azure sells 1, 2, 4, 8, or 16. Scaling updates in place.

- rule: {"int32":{"gte":1}}

### spec.autoInflateEnabled

`bool` · optional (explicit presence)

Whether auto-inflate is enabled: STANDARD's elastic throughput --
Azure grows the namespace's TUs automatically (up to
maximum_throughput_units) when traffic would otherwise be throttled.
Auto-inflate only scales UP; scale-down is a manual capacity edit.
Not applicable to PREMIUM (reserved PUs) or BASIC.
Default: false

### spec.maximumThroughputUnits

`int32` · optional (explicit presence)

The TU ceiling auto-inflate may grow the namespace to (0-40). Azure
requires auto_inflate_enabled to be true for a non-zero ceiling to
take effect -- Azure validates the pairing at apply time (setting a
ceiling without enabling auto-inflate is rejected by ARM, not by the
provider's schema).

- rule: {"int32":{"lte":40,"gte":0}}

### spec.dedicatedClusterId

`string | valueFrom`

The dedicated Event Hubs cluster to place this namespace on, by ARM
ID. References an AzureEventHubCluster's cluster_id output.
Namespaces on a dedicated cluster get single-tenant capacity, up to
1024 partitions per hub, and 90-day retention.

**ForceNew**: a namespace cannot move on or off a cluster in place.

- references: AzureEventHubCluster (`status.outputs.cluster_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureEventHubCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.identity

`AzureEventHubNamespaceIdentity`

Managed identity for the namespace -- required for capture with
identity-based storage authentication and for customer-managed-key
encryption (the identity unwraps the key), and usable anywhere the
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

- `azure_event_hub_namespace_identity_type_unspecified` -- Not specified -- invalid; choose an explicit identity model.
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

### spec.localAuthenticationEnabled

`bool` · optional (explicit presence)

Whether SAS (shared-access-signature) authentication is allowed on
the namespace. Azure's default is true. Set false for a keyless
posture: clients must then authenticate with Microsoft Entra ID
identities, and every SAS rule's keys -- including the root rule
surfaced in this kind's outputs -- stop being usable credentials.
Default: true

- default: `true`

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the namespace accepts traffic from the public internet. When
false, clients reach it only through private endpoints
(AzurePrivateEndpoint) or admitted VNet service endpoints. Must agree
with the network_rule_sets block's own public-access dial when that
block is declared (Azure validates the pair server-side).
Default: true

- default: `true`

### spec.networkRuleSets

`AzureEventHubNamespaceNetworkRuleSets`

The namespace firewall: which networks may reach the data plane.
Not available on BASIC (multi-tenant entry tier). Azure applies the
rule set as a sub-resource of the namespace; the provider folds it
into the namespace resource, and so does this spec -- it has no
independent lifecycle and nothing references it.

- rule: default_action DENY needs at least one admitted source -- add ip_rules or virtual_network_rules, or the namespace would reject all data-plane traffic

### spec.networkRuleSets.defaultAction

`enum`

What happens to traffic no explicit rule admits. Required by Azure
when the rule set is declared: ALLOW (the firewall only annotates) or
DENY (the production posture -- only admitted sources get through).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_event_hub_network_rule_set_default_action_unspecified` -- Not specified -- invalid; Azure requires an explicit choice when the rule set is declared.
- `ALLOW` -- Admit traffic no rule matches (the firewall only annotates).
- `DENY` -- Reject traffic no rule matches -- the production posture. Requires at least one admitted ip_rule or virtual_network_rule.

### spec.networkRuleSets.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the data plane remains reachable over public IP space for
the sources the rules admit. Must equal the namespace-level
public_network_access_enabled (Azure validates the pair server-side;
the spec front-loads the check).
Default: true

- default: `true`

### spec.networkRuleSets.trustedServiceAccessEnabled

`bool` · optional (explicit presence)

Whether trusted Microsoft services (Azure Monitor diagnostics, Event
Grid delivery, IoT Hub routing, ...) bypass the firewall. Azure's
default is false; enable it when platform services must deliver into
a locked-down namespace.

### spec.networkRuleSets.ipRules

`[]string`

Public IPv4 addresses or CIDR ranges admitted to the data plane,
e.g. "203.0.113.0/24". (Azure's per-rule action accepts exactly one
value, Allow -- so each entry here IS an allow rule.)

### spec.networkRuleSets.virtualNetworkRules

`[]AzureEventHubNamespaceVirtualNetworkRule`

VNet subnets admitted to the data plane via service endpoints. Each
subnet should carry the Microsoft.EventHub service endpoint.

### spec.networkRuleSets.virtualNetworkRules[].subnetId

`string | valueFrom` · required

The admitted subnet, by ARM ID. References an AzureSubnet's subnet_id
output.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.networkRuleSets.virtualNetworkRules[].ignoreMissingVirtualNetworkServiceEndpoint

`bool` · optional (explicit presence)

Whether to admit the subnet even if it does not (yet) carry the
Microsoft.EventHub service endpoint -- useful when the endpoint is
being rolled out separately. Azure's default is false.

### spec.tags

`map<string, string>`

Tags to apply to the namespace, merged over the Planton-derived
metadata tags (user values win on key conflicts). ARM tags are
Azure's first-class governance surface -- Azure Policy enforces them
and Microsoft Cost Management groups by them.

## Validation Rules

- `event_hub_network_rules_not_on_basic`: network_rule_sets is not available on the BASIC tier -- Azure rejects firewall rule sets on basic namespaces; use STANDARD or PREMIUM
- `event_hub_ruleset_public_access_matches_namespace`: the network_rule_sets block's public_network_access_enabled must equal the namespace-level public_network_access_enabled -- Azure rejects a mismatched pair
- `event_hub_premium_capacity_values`: on the PREMIUM tier, capacity is processing units -- Azure sells 1, 2, 4, 8, or 16
- `event_hub_throughput_capacity_ceiling`: on BASIC/STANDARD, capacity is throughput units -- 1 to 40

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureEventHubNamespace, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace_id` | `string` | The Azure Resource Manager ID of the namespace. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.EventHub/namespaces/{name} The parent reference for every Event Hubs child kind, the scope for namespace-wide data-plane RBAC, and the private-endpoint target (subresource name: "namespace"). |
| `status.outputs.namespace_name` | `string` | The namespace name -- what SDKs, connection strings, and the Kafka bootstrap address identify the namespace by. The AMQP/Kafka host is {namespace_name}.servicebus.windows.net (Event Hubs shares the Service Bus DNS zone). |
| `status.outputs.identity_principal_id` | `string` | The system-assigned identity's principal (object) ID -- grant this identity access on other resources (e.g. Storage for capture, Key Vault for CMK). Empty unless the identity block includes SYSTEM_ASSIGNED. |
| `status.outputs.default_primary_connection_string` | `string` (sensitive) | The root SAS rule's primary connection string (full manage rights on the namespace). Secret-bearing: it embeds the primary key. Format: Endpoint=sb://{name}.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey={key} |
| `status.outputs.default_secondary_connection_string` | `string` (sensitive) | The root SAS rule's secondary connection string -- the rotation partner: move clients here, regenerate the primary, move back. |
| `status.outputs.default_primary_key` | `string` (sensitive) | The root SAS rule's primary key, for SDKs that take the key and key name separately or mint their own SAS tokens. |
| `status.outputs.default_secondary_key` | `string` (sensitive) | The root SAS rule's secondary key -- the rotation partner. |
| `status.outputs.default_primary_connection_string_alias` | `string` (sensitive) | The root SAS rule's primary connection string addressing the geo-DR alias hostname instead of the namespace hostname. Only meaningful when the namespace carries an AzureEventHubDisasterRecoveryConfig pairing; empty otherwise. Clients configured with the alias keep working across a failover. |
| `status.outputs.default_secondary_connection_string_alias` | `string` (sensitive) | The secondary alias connection string -- the rotation partner. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.dedicatedClusterId` | AzureEventHubCluster | `status.outputs.cluster_id` |
| `spec.identity.userAssignedIdentityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.networkRuleSets.virtualNetworkRules[].subnetId` | AzureSubnet | `status.outputs.subnet_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureEventHub | `spec.namespaceId` | `status.outputs.namespace_id` |
| AzureEventHubAuthorizationRule | `spec.namespaceId` | `status.outputs.namespace_id` |
| AzureEventHubDisasterRecoveryConfig | `spec.primaryNamespaceId` | `status.outputs.namespace_id` |
| AzureEventHubDisasterRecoveryConfig | `spec.partnerNamespaceId` | `status.outputs.namespace_id` |
| AzureEventHubNamespaceCustomerManagedKey | `spec.eventhubNamespaceId` | `status.outputs.namespace_id` |
| AzureEventHubSchemaGroup | `spec.namespaceId` | `status.outputs.namespace_id` |
| AzureMonitorActionGroup | `spec.eventHubReceivers[].eventHubNamespace` | `status.outputs.namespace_name` |

## See Also

- [Overview](../README.md)
