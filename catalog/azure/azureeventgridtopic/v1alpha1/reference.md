# AzureEventgridTopic

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureEventgridTopicSpec** defines an Azure Event Grid custom topic --
the endpoint an application publishes its own events to. Publishers
POST events to the topic's HTTPS endpoint (authenticated with an
access key or Microsoft Entra ID), and any number of event
subscriptions fan those events out to handlers (Functions, webhooks,
queues, Event Hubs). A topic is free at rest; billing is per
operation.

A topic is a SINGLE stream of events under one endpoint and one pair
of access keys. When many logical streams should share one endpoint
and one set of keys -- the multi-tenant pattern, one stream per
customer -- use AzureEventgridDomain instead: a domain holds many
domain topics behind a single endpoint.

## Example

```yaml
# Deep-shape example for docs and offline validation: a CloudEvents
# topic with an IP allowlist and a system-assigned identity. References
# are literal values so the manifest validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventgridTopic
metadata:
  name: test-eventgrid-topic
  id: test-eventgrid-topic
  org: test-org
  env: test
spec:
  resourceGroup:
    value: test-rg
  # The name becomes a PUBLIC DNS hostname, unique across the region.
  name: test-org-app-events
  region: eastus
  inputSchema: CloudEventSchemaV1_0
  publicNetworkAccessEnabled: true
  localAuthEnabled: false
  inboundIpRules:
    - 203.0.113.0/24
    - 198.51.100.7
  identity:
    type: SYSTEM_ASSIGNED
  tags:
    costCenter: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.inputSchema` | `string` |  | `EventGridSchema` |  |
| `spec.inputMappingFields` | `AzureEventgridTopicInputMappingFields` |  |  |  |
| `spec.inputMappingFields.id` | `string` |  |  |  |
| `spec.inputMappingFields.topic` | `string` |  |  |  |
| `spec.inputMappingFields.eventTime` | `string` |  |  |  |
| `spec.inputMappingFields.eventType` | `string` |  |  |  |
| `spec.inputMappingFields.subject` | `string` |  |  |  |
| `spec.inputMappingFields.dataVersion` | `string` |  |  |  |
| `spec.inputMappingDefaultValues` | `AzureEventgridTopicInputMappingDefaultValues` |  |  |  |
| `spec.inputMappingDefaultValues.eventType` | `string` |  |  |  |
| `spec.inputMappingDefaultValues.subject` | `string` |  |  |  |
| `spec.inputMappingDefaultValues.dataVersion` | `string` |  |  |  |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.localAuthEnabled` | `bool` |  | `true` |  |
| `spec.inboundIpRules` | `[]string` |  |  |  |
| `spec.identity` | `AzureEventgridTopicIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group the topic lives in. Can be a literal
string or a reference to an AzureResourceGroup output.

**ForceNew**: changing this destroys and recreates the topic.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The topic's name -- 3-50 characters; letters, numbers, and hyphens.
The name becomes the topic's PUBLIC DNS hostname
({name}.{region}.eventgrid.azure.net), so it must be unique across
ALL Azure customers in the region, not just within the resource
group or subscription -- a taken name fails at deploy time with a
conflict.

**ForceNew**: changing this destroys and recreates the topic (and
its endpoint hostname -- publishers must be repointed).

- rule: Topic names must be 3-50 characters of letters, numbers, and hyphens
- rule: {"required":true}

### spec.region

`string` · required

The Azure region the topic is created in, e.g. "eastus". The
region is part of the endpoint hostname; subscribers can deliver
anywhere.

**ForceNew**: changing this destroys and recreates the topic.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.inputSchema

`string` · optional (explicit presence)

The schema incoming events must arrive in. "EventGridSchema"
(Azure's native envelope) is the default; "CloudEventSchemaV1_0"
is the CNCF standard (pick it for new integrations that span
clouds or vendors); "CustomEventSchema" accepts your application's
own JSON shape, mapped onto the envelope with input_mapping_fields
/ input_mapping_default_values. Defaults to "EventGridSchema" --
the platform sends the default explicitly.

**ForceNew**: changing the schema destroys and recreates the topic.

- default: `EventGridSchema`
- rule: {"string":{"in":["EventGridSchema","CloudEventSchemaV1_0","CustomEventSchema"]}}

### spec.inputMappingFields

`AzureEventgridTopicInputMappingFields`

For "CustomEventSchema" topics: which fields of YOUR event JSON
map onto the Event Grid envelope's fields. Leave unset for the
built-in schemas (they need no mapping).

**ForceNew**: changing the mapping destroys and recreates the
topic.

### spec.inputMappingFields.id

`string`

Your payload's field carrying the event's unique identifier.

### spec.inputMappingFields.topic

`string`

Your payload's field carrying the topic identifier.

### spec.inputMappingFields.eventTime

`string`

Your payload's field carrying the event timestamp.

### spec.inputMappingFields.eventType

`string`

Your payload's field carrying the event type (what happened).

### spec.inputMappingFields.subject

`string`

Your payload's field carrying the subject (what the event is
about).

### spec.inputMappingFields.dataVersion

`string`

Your payload's field carrying the data schema version.

### spec.inputMappingDefaultValues

`AzureEventgridTopicInputMappingDefaultValues`

For "CustomEventSchema" topics: default values stamped onto the
envelope when the incoming event does not carry the mapped field.
Leave unset for the built-in schemas.

**ForceNew**: changing the defaults destroys and recreates the
topic.

### spec.inputMappingDefaultValues.eventType

`string`

Default event type stamped when the payload carries none.

### spec.inputMappingDefaultValues.subject

`string`

Default subject stamped when the payload carries none.

### spec.inputMappingDefaultValues.dataVersion

`string`

Default data schema version stamped when the payload carries none.

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the topic's endpoint accepts publishes from the public
internet. Set false to restrict publishing to private endpoints
only (inbound_ip_rules are also ignored when disabled).
Default: true (Azure's default) -- the platform sends the value
explicitly.

- default: `true`

### spec.localAuthEnabled

`bool` · optional (explicit presence)

Whether access-key (SAS) authentication is enabled alongside
Microsoft Entra ID. Set false to force every publisher through
Entra ID -- the keys still exist but stop working.
Default: true (Azure's default) -- the platform sends the value
explicitly.

- default: `true`

### spec.inboundIpRules

`[]string`

Inbound IP firewall rules -- IPv4 CIDR ranges allowed to publish,
e.g. "203.0.113.0/24" (a single address works too), up to 128. An
empty list means no IP restriction. Rules only take effect while
public_network_access_enabled is true. Azure's rule action is
"Allow"-only on this resource (the provider rejects anything
else), so each entry is just the range -- both engines send the
action token explicitly; deny rules would widen this shape if
Azure ever ships them.

- rule: {"repeated":{"maxItems":"128","items":{"string":{"minLen":"1"}}}}

### spec.identity

`AzureEventgridTopicIdentity`

The topic's managed identity -- needed when event DELIVERY should
authenticate as the topic (for example delivering to an Event Hub
or storage queue with identity-based access, or dead-lettering to
a storage account). Omit when no delivery target requires one.

- rule: identity_ids is required for USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the topic; USER_ASSIGNED brings an identity you manage
(grantable on delivery targets BEFORE the topic exists). The
provider supports exactly one flavor at a time on this resource --
there is no combined mode.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_eventgrid_topic_identity_type_unspecified` -- Not specified: rejected -- an identity block requires a flavor.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the topic.
- `USER_ASSIGNED` -- An identity you create and manage (AzureUserAssignedIdentity).

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED: the user-assigned identities attached to the
topic, by ARM ID. Reference AzureUserAssignedIdentity resources so
delivery-target grants can be composed before the topic is
created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.tags

`map<string, string>`

Tags to apply to the topic, merged over the Planton-derived
metadata tags (user values win on key conflicts).

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureEventgridTopic, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.topic_id` | `string` | The topic's Azure Resource Manager ID. |
| `status.outputs.topic_name` | `string` | The topic's name (also the first label of its endpoint hostname). |
| `status.outputs.endpoint` | `string` | The HTTPS endpoint publishers POST events to (https://{name}.{region}.eventgrid.azure.net/api/events). |
| `status.outputs.primary_access_key` | `string` (sensitive) | The primary access key publishers authenticate with (the aeg-sas-key header). Inert while local_auth_enabled is false. |
| `status.outputs.secondary_access_key` | `string` (sensitive) | The secondary access key -- the rotation partner: move publishers here, regenerate the primary, move back. |
| `status.outputs.identity_principal_id` | `string` | The principal ID of the topic's system-assigned identity (empty when no identity is configured) -- grant this on delivery targets that use identity-based access. |

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
| AzureDataFactoryTrigger | `spec.customEvent.eventgridTopicId` | `status.outputs.topic_id` |
| AzureEventgridEventSubscription | `spec.scope` | `status.outputs.topic_id` |
| AzureEventgridNamespace | `spec.topicSpacesConfiguration.routeTopicId` | `status.outputs.topic_id` |

## See Also

- [Overview](../README.md)
