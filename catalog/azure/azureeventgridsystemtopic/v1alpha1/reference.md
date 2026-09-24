# AzureEventgridSystemTopic

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureEventgridSystemTopicSpec** defines an Azure Event Grid system
topic -- the subscription surface for events AZURE ITSELF publishes
about one of your resources. Where a custom topic
(AzureEventgridTopic) receives events your application posts, a
system topic exposes the events an Azure service already emits: a
storage account announcing blob creations, a resource group
announcing resource writes, a Key Vault announcing secret expiries.
Create one system topic per source resource, then attach event
subscriptions (AzureEventgridEventSubscription) to route its events
to handlers. A system topic is free at rest; billing is per
operation.

Azure allows ONE system topic per source resource per topic type --
a second create against the same source fails at deploy time, so
teams sharing a source share its system topic and attach their own
subscriptions.

## Example

```yaml
# Deep-shape example for docs and offline validation: a storage
# account's blob-event stream with a system-assigned identity for
# secured delivery. References are literal values so the manifest
# validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventgridSystemTopic
metadata:
  name: test-eventgrid-system-topic
  id: test-eventgrid-system-topic
  org: test-org
  env: test
spec:
  resourceGroup:
    value: test-rg
  name: appdata-storage-events
  # Must match the SOURCE's region ("Global" for subscription/
  # resource-group sources).
  region: eastus
  sourceResourceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Storage/storageAccounts/appdata
  topicType: Microsoft.Storage.StorageAccounts
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
| `spec.sourceResourceId` | `string \| valueFrom` | yes |  |  |
| `spec.topicType` | `string` | yes |  |  |
| `spec.identity` | `AzureEventgridSystemTopicIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group the system topic lives in. This is where
the TOPIC resource sits -- the source resource may live in a
different group. Can be a literal string or a reference to an
AzureResourceGroup output.

**ForceNew**: changing this destroys and recreates the topic.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The system topic's name -- 3-128 characters; letters, numbers, and
hyphens. Unlike custom topics, a system topic has no public
endpoint hostname, so the name only needs to be unique within the
resource group.

**ForceNew**: changing this destroys and recreates the topic (and
every subscription attached to it).

- rule: System topic names must be 3-128 characters of letters, numbers, and hyphens
- rule: {"required":true}

### spec.region

`string` · required

The Azure region the system topic is created in. It must MATCH the
source resource's region -- Azure rejects a mismatch at deploy
time. For sources that are global objects rather than regional
ones (Azure subscriptions via topic type
"Microsoft.Resources.Subscriptions", resource groups via
"Microsoft.Resources.ResourceGroups"), the region must be
"Global".

**ForceNew**: changing this destroys and recreates the topic.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.sourceResourceId

`string | valueFrom` · required

The ARM ID of the resource whose events this topic surfaces --
e.g. a storage account's ID for blob events, a resource group's
ID for resource-lifecycle events, a Key Vault's ID for
certificate/secret events. Sources span dozens of kinds, so no
single kind dominates: reference the owning kind's id output with
an explicit valueFrom (e.g. an AzureStorageAccount's
storage_account_id) or pass a literal ID.

The source and the topic_type must agree -- the type names the
service the source belongs to.

Containment-exempt: the topic SURFACES its source's events; it is
its own resource in its resource group, not a tenant of the storage
account or vault it listens to. Without the exemption a topic wired
to a container source by reference would be drawn inside the
resource whose events it publishes.

**ForceNew**: changing this destroys and recreates the topic.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.topicType

`string` · required

Which Azure service's event stream the source emits, e.g.
"Microsoft.Storage.StorageAccounts" (blob/queue events),
"Microsoft.Resources.ResourceGroups" (resource lifecycle),
"Microsoft.KeyVault.vaults" (secret/certificate lifecycle),
"Microsoft.EventGrid.Topics", "Microsoft.Devices.IoTHubs",
"Microsoft.ContainerRegistry.Registries",
"Microsoft.MachineLearningServices.Workspaces". Azure validates
the value against its live catalog of topic types (list them with
`az eventgrid topic-type list`); the value must match the source
resource's service.

**ForceNew**: changing this destroys and recreates the topic.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.identity

`AzureEventgridSystemTopicIdentity`

The system topic's managed identity -- needed when its
subscriptions deliver or dead-letter with identity-based access
(the delivery identity a subscription names must exist on the
topic). Omit when no delivery target requires one.

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the topic; USER_ASSIGNED brings identities you manage
(grantable on delivery targets BEFORE the topic exists);
SYSTEM_AND_USER_ASSIGNED carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_eventgrid_system_topic_identity_type_unspecified` -- Not specified: rejected -- an identity block requires a flavor.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the topic. Wire value: "SystemAssigned".
- `USER_ASSIGNED` -- Identities you create and manage (AzureUserAssignedIdentity). Wire value: "UserAssigned".
- `SYSTEM_AND_USER_ASSIGNED` -- Both at once. Wire value: "SystemAssigned, UserAssigned".

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the topic, by ARM ID. Reference
AzureUserAssignedIdentity resources so delivery-target grants can
be composed before the topic is created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.tags

`map<string, string>`

Tags to apply to the system topic, merged over the
Planton-derived metadata tags (user values win on key conflicts).

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureEventgridSystemTopic, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.system_topic_id` | `string` | The system topic's Azure Resource Manager ID -- the target an AzureEventgridEventSubscription's system_topic_id references. |
| `status.outputs.system_topic_name` | `string` | The system topic's name. |
| `status.outputs.metric_resource_id` | `string` | The GUID-style identifier Azure Monitor uses for the topic's metrics (not an ARM ID) -- useful when wiring metric alerts to the topic's delivery/drop counters. |
| `status.outputs.identity_principal_id` | `string` | The principal ID of the topic's system-assigned identity (empty when no system-assigned identity is configured) -- grant this on delivery targets that use identity-based access. |

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
| AzureEventgridEventSubscription | `spec.systemTopicId` | `status.outputs.system_topic_id` |

## See Also

- [Overview](../README.md)
