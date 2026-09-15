# AzureEventgridNamespace

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureEventgridNamespaceSpec** defines an Azure Event Grid
namespace -- the newer Event Grid resource that hosts CloudEvents
namespace topics (see AzureEventgridNamespaceTopic) and, optionally,
an MQTT broker for IoT-style pub/sub. Where the classic resources
(AzureEventgridTopic, AzureEventgridDomain) each own one endpoint,
a namespace is a capacity-scaled HUB: throughput units set its
ceiling, topics are created inside it as their own resources, and
MQTT clients connect to it directly when the broker is enabled.

The namespace's SKU has exactly one legal value ("Standard"), so it
is not part of this spec -- both engines send it explicitly.

## Example

```yaml
# Deep-shape example for docs and offline validation: an MQTT-enabled
# namespace with capacity, an inbound IP rule, a combined identity,
# and the routing enrichments -- the route topic is included here so
# the offline plan renders the full MQTT block. References are literal
# values so the manifest validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventgridNamespace
metadata:
  name: test-eventgrid-namespace
  id: test-eventgrid-namespace
  org: test-org
  env: test
spec:
  resourceGroup:
    value: test-rg
  name: acme-events-hub
  region: eastus
  capacity: 2
  publicNetworkAccessEnabled: true
  inboundIpRules:
    - 203.0.113.0/24
  identity:
    type: SYSTEM_AND_USER_ASSIGNED
    identityIds:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/test-uai
  topicSpacesConfiguration:
    alternativeAuthenticationNameSources:
      - ClientCertificateSubject
    maximumClientSessionsPerAuthenticationName: 3
    maximumSessionExpiryInHours: 4
    routeTopicId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.EventGrid/topics/test-route-topic
    dynamicRoutingEnrichments:
      - key: clientname
        value: ${client.authenticationName}
    staticRoutingEnrichments:
      - key: source
        value: mqtt-broker
  tags:
    costCenter: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.capacity` | `int32` |  | `1` |  |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.inboundIpRules` | `[]string` |  |  |  |
| `spec.identity` | `AzureEventgridNamespaceIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.topicSpacesConfiguration` | `AzureEventgridNamespaceTopicSpacesConfiguration` |  |  |  |
| `spec.topicSpacesConfiguration.alternativeAuthenticationNameSources` | `[]string` |  |  |  |
| `spec.topicSpacesConfiguration.maximumClientSessionsPerAuthenticationName` | `int32` |  | `1` |  |
| `spec.topicSpacesConfiguration.maximumSessionExpiryInHours` | `int32` |  | `1` |  |
| `spec.topicSpacesConfiguration.routeTopicId` | `string \| valueFrom` |  |  | AzureEventgridTopic (`status.outputs.topic_id`) |
| `spec.topicSpacesConfiguration.dynamicRoutingEnrichments` | `[]AzureEventgridNamespaceRoutingEnrichment` |  |  |  |
| `spec.topicSpacesConfiguration.dynamicRoutingEnrichments[].key` | `string` | yes |  |  |
| `spec.topicSpacesConfiguration.dynamicRoutingEnrichments[].value` | `string` | yes |  |  |
| `spec.topicSpacesConfiguration.staticRoutingEnrichments` | `[]AzureEventgridNamespaceRoutingEnrichment` |  |  |  |
| `spec.topicSpacesConfiguration.staticRoutingEnrichments[].key` | `string` | yes |  |  |
| `spec.topicSpacesConfiguration.staticRoutingEnrichments[].value` | `string` | yes |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group the namespace lives in. Can be a literal
string or a reference to an AzureResourceGroup output.

**ForceNew**: changing this destroys and recreates the namespace
(and every topic inside it).

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The namespace's name -- 3-50 characters (the provider validates
length; Azure additionally requires letters, numbers, and hyphens
and uses the name in the namespace's regional hostnames, so treat
it as a DNS label and prefix it with your org).

**ForceNew**: changing this destroys and recreates the namespace.

- rule: Namespace names must be 3-50 characters
- rule: {"required":true}

### spec.region

`string` · required

The Azure region the namespace is created in, e.g. "eastus".

**ForceNew**: changing this destroys and recreates the namespace.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.capacity

`int32` · optional (explicit presence)

Throughput units (TUs) provisioned for the namespace, 1-40. Each
TU buys a slice of ingress/egress capacity shared by all topics
and MQTT traffic in the namespace; capacity can be changed in
place. Defaults to 1 -- the platform sends the default explicitly.

- default: `1`
- rule: {"int32":{"lte":40,"gte":1}}

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the namespace's endpoints accept traffic from the public
internet. Set false to restrict access to private endpoints only
(inbound_ip_rules are also ignored when disabled). Default: true
(Azure's default) -- the platform sends the value explicitly.

- default: `true`

### spec.inboundIpRules

`[]string`

Inbound IP firewall rules -- IPv4 CIDR ranges allowed to reach the
namespace, e.g. "203.0.113.0/24", up to 128. An empty list means
no IP restriction. Rules only take effect while
public_network_access_enabled is true. Azure's rule action is
"Allow"-only on this resource, so each entry is just the range --
both engines send the action token explicitly; deny rules would
widen this shape if Azure ever ships them.

- rule: {"repeated":{"maxItems":"128","items":{"string":{"minLen":"1"}}}}

### spec.identity

`AzureEventgridNamespaceIdentity`

The namespace's managed identity -- needed when namespace-topic
subscriptions deliver with identity-based access or when MQTT
routing writes to a protected custom topic. Omit when nothing
requires one.

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the namespace; USER_ASSIGNED brings identities you manage
(grantable on delivery targets BEFORE the namespace exists);
SYSTEM_AND_USER_ASSIGNED carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_eventgrid_namespace_identity_type_unspecified` -- Not specified: rejected -- an identity block requires a flavor.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the namespace. Wire value: "SystemAssigned".
- `USER_ASSIGNED` -- Identities you create and manage (AzureUserAssignedIdentity). Wire value: "UserAssigned".
- `SYSTEM_AND_USER_ASSIGNED` -- Both at once. Wire value: "SystemAssigned, UserAssigned".

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the namespace, by ARM ID. Reference
AzureUserAssignedIdentity resources so delivery-target grants can
be composed before the namespace is created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.topicSpacesConfiguration

`AzureEventgridNamespaceTopicSpacesConfiguration`

MQTT broker configuration (Azure calls it "topic spaces"). Set
this block to ENABLE the namespace's MQTT broker; omit it to run
a pure CloudEvents namespace. All fields inside are optional.

**ForceNew**: the block cannot be added, removed, or changed after
create -- the provider replaces the namespace instead. Decide the
MQTT posture up front.

### spec.topicSpacesConfiguration.alternativeAuthenticationNameSources

`[]string`

Where the broker reads a connecting client's authentication name
from, in addition to the MQTT CONNECT username: fields of the
client certificate. Legal values: "ClientCertificateSubject",
"ClientCertificateDns", "ClientCertificateUri",
"ClientCertificateIp", "ClientCertificateEmail".

- rule: {"repeated":{"items":{"string":{"in":["ClientCertificateSubject","ClientCertificateDns","ClientCertificateUri","ClientCertificateIp","ClientCertificateEmail"]}}}}

### spec.topicSpacesConfiguration.maximumClientSessionsPerAuthenticationName

`int32` · optional (explicit presence)

How many concurrent MQTT sessions one authentication name may
hold, 1-100. Defaults to 1 -- the platform sends the default
explicitly.

- default: `1`
- rule: {"int32":{"lte":100,"gte":1}}

### spec.topicSpacesConfiguration.maximumSessionExpiryInHours

`int32` · optional (explicit presence)

How long a disconnected client's MQTT session (subscriptions and
queued messages) survives, in hours, 1-8. Defaults to 1 -- the
platform sends the default explicitly.

- default: `1`
- rule: {"int32":{"lte":8,"gte":1}}

### spec.topicSpacesConfiguration.routeTopicId

`string | valueFrom`

Route all MQTT messages into an Event Grid CUSTOM topic
(AzureEventgridTopic) for fan-out to non-MQTT subscribers. The
custom topic must live in the same region and use the CloudEvents
schema. Omit to keep MQTT traffic inside the broker.

The namespace lives in its resource group and routes INTO the topic,
so the reference is access, not placement, on a diagram.

- references: AzureEventgridTopic (`status.outputs.topic_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureEventgridTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.topicSpacesConfiguration.dynamicRoutingEnrichments

`[]AzureEventgridNamespaceRoutingEnrichment`

Enrichments stamped onto routed MQTT messages whose VALUE is
resolved per message from client attributes or topic segments
(e.g. "${client.authenticationName}"). Key: 1-20 characters;
value: 1-128 characters.

### spec.topicSpacesConfiguration.dynamicRoutingEnrichments[].key

`string` · required

The enrichment key added to the routed event, 1-20 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"20"}}

### spec.topicSpacesConfiguration.dynamicRoutingEnrichments[].value

`string` · required

The enrichment value, 1-128 characters. For dynamic enrichments
this is the resolution expression (e.g.
"${client.authenticationName}"); for static enrichments it is the
literal string.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.topicSpacesConfiguration.staticRoutingEnrichments

`[]AzureEventgridNamespaceRoutingEnrichment`

Enrichments stamped onto routed MQTT messages with a FIXED string
value. Key: 1-20 characters; value: 1-128 characters. (Azure also
defines non-string static enrichment types; the provider pins the
type to String -- both engines send that token.)

### spec.topicSpacesConfiguration.staticRoutingEnrichments[].key

`string` · required

The enrichment key added to the routed event, 1-20 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"20"}}

### spec.topicSpacesConfiguration.staticRoutingEnrichments[].value

`string` · required

The enrichment value, 1-128 characters. For dynamic enrichments
this is the resolution expression (e.g.
"${client.authenticationName}"); for static enrichments it is the
literal string.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.tags

`map<string, string>`

Tags to apply to the namespace, merged over the Planton-derived
metadata tags (user values win on key conflicts).

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureEventgridNamespace, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace_id` | `string` | The namespace's Azure Resource Manager ID -- the target an AzureEventgridNamespaceTopic's namespace_id references. |
| `status.outputs.namespace_name` | `string` | The namespace's name. |
| `status.outputs.identity_principal_id` | `string` | The principal ID of the namespace's system-assigned identity (empty when no system-assigned identity is configured) -- grant this on delivery targets that use identity-based access. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.topicSpacesConfiguration.routeTopicId` | AzureEventgridTopic | `status.outputs.topic_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureEventgridNamespaceTopic | `spec.namespaceId` | `status.outputs.namespace_id` |

## See Also

- [Overview](../README.md)
