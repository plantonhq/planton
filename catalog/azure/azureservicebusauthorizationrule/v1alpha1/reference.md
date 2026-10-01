# AzureServiceBusAuthorizationRule

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureServiceBusAuthorizationRuleSpec** defines a SAS
(shared-access-signature) authorization rule for Azure Service Bus: a
named credential with listen/send/manage rights, scoped to exactly one
of a namespace, a queue, or a topic.

Authorization rules are how applications get least-privilege connection
strings -- a sender service holds a send-only rule on its one queue, a
worker holds a listen-only rule, and neither can touch anything else in
the namespace. The rule's keys and connection strings surface as
sensitive outputs; regenerating keys rotates every client using them.

**The scope is the polymorphism**: set exactly one of `namespace_id`
(namespace-wide rights -- every entity), `queue_id` (that queue only),
or `topic_id` (that topic only). Azure models these as three ARM types
with identical shapes; this kind dispatches to the right one from
whichever parent is set. The scope is fixed at creation.

**Rights contract** (Azure's own): at least one of listen/send/manage
must be true, and manage requires BOTH listen and send.

For a keyless posture, skip SAS rules entirely: disable the namespace's
local_auth_enabled and grant Entra identities data-plane roles via
AzureRoleAssignment -- SAS keys stop working namespace-wide.

**ForceNew fields**: `rule_name` and the parent scope.

## Example

```yaml
# Offline-plan manifest: a QUEUE-scoped rule -- exercises the
# non-default dispatch branch (the scope XOR picks
# azurerm_servicebus_queue_authorization_rule) with the full manage
# rights trio.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureServiceBusAuthorizationRule
metadata:
  name: test-sb-auth-rule
spec:
  ruleName: orders-manage
  queueId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ServiceBus/namespaces/hack-servicebus-ns/queues/orders
  listen: true
  send: true
  manage: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.ruleName` | `string` | yes |  |  |
| `spec.namespaceId` | `string \| valueFrom` |  |  | AzureServiceBusNamespace (`status.outputs.namespace_id`) |
| `spec.queueId` | `string \| valueFrom` |  |  | AzureServiceBusQueue (`status.outputs.queue_id`) |
| `spec.topicId` | `string \| valueFrom` |  |  | AzureServiceBusTopic (`status.outputs.topic_id`) |
| `spec.listen` | `bool` |  |  |  |
| `spec.send` | `bool` |  |  |  |
| `spec.manage` | `bool` |  |  |  |

## Field Details

### spec.ruleName

`string` · required

The rule's name -- unique within its scope, up to 50 characters.
Starts and ends with a letter or number; letters, numbers, periods,
hyphens, and underscores in between. "RootManageSharedAccessKey" is
reserved for the namespace's built-in root rule.

**ForceNew**: changing the name replaces the rule and regenerates its
keys.

- rule: rule_name must be up to 50 characters of letters, numbers, periods, hyphens, and underscores, starting and ending with a letter or number
- rule: RootManageSharedAccessKey is the namespace's built-in root rule -- its keys already surface as AzureServiceBusNamespace outputs; mint a differently-named rule for scoped access
- rule: {"required":true,"string":{"minLen":"1","maxLen":"50"}}

### spec.namespaceId

`string | valueFrom`

Namespace-wide scope: rights over EVERY queue and topic in the
namespace, by ARM ID. References an AzureServiceBusNamespace's
namespace_id output. Set exactly one of namespace_id, queue_id, or
topic_id.

- references: AzureServiceBusNamespace (`status.outputs.namespace_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureServiceBusNamespace, name: <that resource's name>, fieldPath: status.outputs.namespace_id}} -- a bare string does not parse

### spec.queueId

`string | valueFrom`

Single-queue scope: rights over one queue only, by ARM ID. References
an AzureServiceBusQueue's queue_id output. The least-privilege choice
for point-to-point workloads.

- references: AzureServiceBusQueue (`status.outputs.queue_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureServiceBusQueue, name: <that resource's name>, fieldPath: status.outputs.queue_id}} -- a bare string does not parse

### spec.topicId

`string | valueFrom`

Single-topic scope: rights over one topic (and receiving through its
subscriptions), by ARM ID. References an AzureServiceBusTopic's
topic_id output.

- references: AzureServiceBusTopic (`status.outputs.topic_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureServiceBusTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.listen

`bool` · optional (explicit presence)

Whether this rule may RECEIVE messages (and browse/peek) within its
scope.
Default: false

### spec.send

`bool` · optional (explicit presence)

Whether this rule may SEND messages within its scope.
Default: false

### spec.manage

`bool` · optional (explicit presence)

Whether this rule may MANAGE entities within its scope (create and
delete queues/topics/subscriptions). Requires listen AND send --
manage is a superset, never a standalone right.
Default: false

## Validation Rules

- `service_bus_auth_rule_exactly_one_scope`: set exactly one scope -- namespace_id for namespace-wide rights, queue_id for one queue, or topic_id for one topic
- `service_bus_auth_rule_at_least_one_right`: grant at least one right -- listen, send, or manage (a rule with no rights would be an unusable credential)
- `service_bus_auth_rule_manage_requires_listen_send`: manage is a superset right -- Azure requires listen and send to both be true alongside it

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureServiceBusAuthorizationRule, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.authorization_rule_id` | `string` | The Azure Resource Manager ID of the rule (under its namespace, queue, or topic parent). AzureServiceBusDisasterRecoveryConfig's alias_authorization_rule_id consumes it with zero translation. |
| `status.outputs.rule_name` | `string` | The rule's name -- the SharedAccessKeyName clients present. |
| `status.outputs.primary_key` | `string` (sensitive) | The primary key. |
| `status.outputs.secondary_key` | `string` (sensitive) | The secondary key -- the rotation partner. |
| `status.outputs.primary_connection_string` | `string` (sensitive) | The ready-to-use primary connection string (endpoint + key name + primary key), scoped to the rule's entity when queue- or topic-scoped. |
| `status.outputs.secondary_connection_string` | `string` (sensitive) | The secondary connection string -- the rotation partner. |
| `status.outputs.primary_connection_string_alias` | `string` (sensitive) | The primary connection string addressing the geo-DR alias instead of the namespace. Empty unless the namespace carries a disaster-recovery pairing. |
| `status.outputs.secondary_connection_string_alias` | `string` (sensitive) | The secondary alias connection string -- the rotation partner. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespaceId` | AzureServiceBusNamespace | `status.outputs.namespace_id` |
| `spec.queueId` | AzureServiceBusQueue | `status.outputs.queue_id` |
| `spec.topicId` | AzureServiceBusTopic | `status.outputs.topic_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureServiceBusDisasterRecoveryConfig | `spec.aliasAuthorizationRuleId` | `status.outputs.authorization_rule_id` |

## See Also

- [Overview](../README.md)
