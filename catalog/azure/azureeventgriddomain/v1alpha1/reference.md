# AzureEventgridDomain

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureEventgridDomainSpec** defines an Azure Event Grid domain --
one publishing endpoint and one pair of access keys serving MANY
event streams (domain topics). The multi-tenant pattern: a SaaS
publishes every customer's events to the same domain endpoint,
naming the domain topic per event; each customer subscribes only to
their own topic. A domain is free at rest; billing is per operation.

Domain topics come from two postures, both fully supported here:
AUTO-MANAGED (the defaults) -- Azure creates a domain topic when its
first subscription arrives and deletes it when the last one goes --
or PINNED, where auto_create/auto_delete are false and each topic is
declared explicitly as an AzureEventgridDomainTopic resource (the
governance posture: topics exist only by decision).

For a single event stream with its own endpoint, use
AzureEventgridTopic instead.

## Example

```yaml
# Deep-shape example for docs and offline validation: a pinned-topics
# CloudEvents domain with an IP allowlist. References are literal
# values so the manifest validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventgridDomain
metadata:
  name: test-eventgrid-domain
  id: test-eventgrid-domain
  org: test-org
  env: test
spec:
  resourceGroup:
    value: test-rg
  # The name becomes a PUBLIC DNS hostname, unique across the region.
  name: test-org-tenant-events
  region: eastus
  inputSchema: CloudEventSchemaV1_0
  # The pinned-topics governance posture: topics exist only as declared
  # AzureEventgridDomainTopic resources.
  autoCreateTopicWithFirstSubscription: false
  autoDeleteTopicWithLastSubscription: false
  publicNetworkAccessEnabled: true
  localAuthEnabled: true
  inboundIpRules:
    - 203.0.113.0/24
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
| `spec.inputMappingFields` | `AzureEventgridDomainInputMappingFields` |  |  |  |
| `spec.inputMappingFields.id` | `string` |  |  |  |
| `spec.inputMappingFields.topic` | `string` |  |  |  |
| `spec.inputMappingFields.eventTime` | `string` |  |  |  |
| `spec.inputMappingFields.eventType` | `string` |  |  |  |
| `spec.inputMappingFields.subject` | `string` |  |  |  |
| `spec.inputMappingFields.dataVersion` | `string` |  |  |  |
| `spec.inputMappingDefaultValues` | `AzureEventgridDomainInputMappingDefaultValues` |  |  |  |
| `spec.inputMappingDefaultValues.eventType` | `string` |  |  |  |
| `spec.inputMappingDefaultValues.subject` | `string` |  |  |  |
| `spec.inputMappingDefaultValues.dataVersion` | `string` |  |  |  |
| `spec.autoCreateTopicWithFirstSubscription` | `bool` |  | `true` |  |
| `spec.autoDeleteTopicWithLastSubscription` | `bool` |  | `true` |  |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.localAuthEnabled` | `bool` |  | `true` |  |
| `spec.inboundIpRules` | `[]string` |  |  |  |
| `spec.identity` | `AzureEventgridDomainIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group the domain lives in. Can be a literal
string or a reference to an AzureResourceGroup output.

**ForceNew**: changing this destroys and recreates the domain.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The domain's name -- 3-50 characters; letters, numbers, and
hyphens. The name becomes the domain's PUBLIC DNS hostname
({name}.{region}.eventgrid.azure.net), so it must be unique across
ALL Azure customers in the region, not just within the resource
group or subscription -- a taken name fails at deploy time with a
conflict.

**ForceNew**: changing this destroys and recreates the domain (and
its endpoint hostname -- publishers must be repointed).

- rule: Domain names must be 3-50 characters of letters, numbers, and hyphens
- rule: {"required":true}

### spec.region

`string` · required

The Azure region the domain is created in, e.g. "eastus". The
region is part of the endpoint hostname; subscribers can deliver
anywhere.

**ForceNew**: changing this destroys and recreates the domain.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.inputSchema

`string` · optional (explicit presence)

The schema incoming events must arrive in -- one schema for every
topic in the domain. "EventGridSchema" (Azure's native envelope)
is the default; "CloudEventSchemaV1_0" is the CNCF standard (pick
it for new integrations that span clouds or vendors);
"CustomEventSchema" accepts your application's own JSON shape,
mapped onto the envelope with input_mapping_fields /
input_mapping_default_values. Defaults to "EventGridSchema" -- the
platform sends the default explicitly.

**ForceNew**: changing the schema destroys and recreates the
domain.

- default: `EventGridSchema`
- rule: {"string":{"in":["EventGridSchema","CloudEventSchemaV1_0","CustomEventSchema"]}}

### spec.inputMappingFields

`AzureEventgridDomainInputMappingFields`

For "CustomEventSchema" domains: which fields of YOUR event JSON
map onto the Event Grid envelope's fields. Leave unset for the
built-in schemas (they need no mapping).

**ForceNew**: changing the mapping destroys and recreates the
domain.

### spec.inputMappingFields.id

`string`

Your payload's field carrying the event's unique identifier.

### spec.inputMappingFields.topic

`string`

Your payload's field carrying the topic identifier -- for domains
this is the field naming WHICH domain topic the event belongs to.

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

`AzureEventgridDomainInputMappingDefaultValues`

For "CustomEventSchema" domains: default values stamped onto the
envelope when the incoming event does not carry the mapped field.
Leave unset for the built-in schemas.

**ForceNew**: changing the defaults destroys and recreates the
domain.

### spec.inputMappingDefaultValues.eventType

`string`

Default event type stamped when the payload carries none.

### spec.inputMappingDefaultValues.subject

`string`

Default subject stamped when the payload carries none.

### spec.inputMappingDefaultValues.dataVersion

`string`

Default data schema version stamped when the payload carries none.

### spec.autoCreateTopicWithFirstSubscription

`bool` · optional (explicit presence)

Whether Azure creates a domain topic automatically when the
topic's FIRST event subscription is created. Set false for the
pinned-topics governance posture: topics then exist only as
explicitly declared AzureEventgridDomainTopic resources, and a
subscription against an undeclared topic fails instead of
materializing one.
Default: true (Azure's default) -- the platform sends the value
explicitly.

- default: `true`

### spec.autoDeleteTopicWithLastSubscription

`bool` · optional (explicit presence)

Whether Azure deletes a domain topic automatically when its LAST
event subscription is deleted. Set false to keep topics (and
anything addressed to them) in place when subscriptions churn.
Default: true (Azure's default) -- the platform sends the value
explicitly.

- default: `true`

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the domain's endpoint accepts publishes from the public
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

`AzureEventgridDomainIdentity`

The domain's managed identity -- needed when event DELIVERY should
authenticate as the domain (for example delivering to an Event Hub
or storage queue with identity-based access, or dead-lettering to
a storage account). Omit when no delivery target requires one.

- rule: identity_ids is required for USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the domain; USER_ASSIGNED brings an identity you manage
(grantable on delivery targets BEFORE the domain exists). The
provider supports exactly one flavor at a time on this resource --
there is no combined mode.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_eventgrid_domain_identity_type_unspecified` -- Not specified: rejected -- an identity block requires a flavor.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the domain.
- `USER_ASSIGNED` -- An identity you create and manage (AzureUserAssignedIdentity).

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED: the user-assigned identities attached to the
domain, by ARM ID. Reference AzureUserAssignedIdentity resources
so delivery-target grants can be composed before the domain is
created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.tags

`map<string, string>`

Tags to apply to the domain, merged over the Planton-derived
metadata tags (user values win on key conflicts).

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureEventgridDomain, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.domain_id` | `string` | The domain's Azure Resource Manager ID. |
| `status.outputs.domain_name` | `string` | The domain's name (also the first label of its endpoint hostname). |
| `status.outputs.endpoint` | `string` | The HTTPS endpoint publishers POST events to (https://{name}.{region}.eventgrid.azure.net/api/events) -- one endpoint for every topic in the domain; the event's topic field names the stream. |
| `status.outputs.primary_access_key` | `string` (sensitive) | The primary access key publishers authenticate with (the aeg-sas-key header). Inert while local_auth_enabled is false. |
| `status.outputs.secondary_access_key` | `string` (sensitive) | The secondary access key -- the rotation partner: move publishers here, regenerate the primary, move back. |
| `status.outputs.identity_principal_id` | `string` | The principal ID of the domain's system-assigned identity (empty when no identity is configured) -- grant this on delivery targets that use identity-based access. |

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
| AzureEventgridDomainTopic | `spec.domainId` | `status.outputs.domain_id` |

## See Also

- [Overview](../README.md)
