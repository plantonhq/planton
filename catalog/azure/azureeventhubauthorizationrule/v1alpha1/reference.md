# AzureEventHubAuthorizationRule

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureEventHubAuthorizationRuleSpec** defines a SAS
(shared-access-signature) authorization rule for Azure Event Hubs: a
named credential with listen/send/manage rights, scoped to exactly one
of a namespace or a single event hub.

Authorization rules are how applications get least-privilege connection
strings -- a producer holds a send-only rule on its one hub, a consumer
holds a listen-only rule, and neither can touch anything else in the
namespace. The rule's keys and connection strings surface as sensitive
outputs; regenerating keys rotates every client using them.

**The scope is the polymorphism**: set exactly one of `namespace_id`
(namespace-wide rights -- every hub) or `event_hub_id` (that hub only).
Azure models these as two ARM types with identical shapes; this kind
dispatches to the right one from whichever parent is set. The scope is
fixed at creation.

**Rights contract** (Azure's own): at least one of listen/send/manage
must be true, and manage requires BOTH listen and send.

For a keyless posture, skip SAS rules entirely: disable the namespace's
local_authentication_enabled and grant Entra identities data-plane roles
(Azure Event Hubs Data Owner/Sender/Receiver) via AzureRoleAssignment --
SAS keys stop working namespace-wide.

**ForceNew fields**: `rule_name` and the parent scope.

## Example

```yaml
# Offline-plan manifest: a hub-scoped manage rule -- exercises the
# event-hub dispatch path (parent names parsed from the hub's ARM id)
# and the full rights trio. The namespace-scope path is exercised by the
# E2E scenarios.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventHubAuthorizationRule
metadata:
  name: test-eh-auth-rule
spec:
  ruleName: hub-manage
  eventHubId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.EventHub/namespaces/hack-eventhubs-ns/eventhubs/telemetry
  listen: true
  send: true
  manage: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.ruleName` | `string` | yes |  |  |
| `spec.namespaceId` | `string \| valueFrom` |  |  | AzureEventHubNamespace (`status.outputs.namespace_id`) |
| `spec.eventHubId` | `string \| valueFrom` |  |  | AzureEventHub (`status.outputs.event_hub_id`) |
| `spec.listen` | `bool` |  |  |  |
| `spec.send` | `bool` |  |  |  |
| `spec.manage` | `bool` |  |  |  |

## Field Details

### spec.ruleName

`string` · required

The rule's name -- unique within its scope, up to 60 characters.
Starts and ends with a letter or number; letters, numbers, periods,
hyphens, and underscores in between. "RootManageSharedAccessKey" is
reserved for the namespace's built-in root rule.

**ForceNew**: changing the name replaces the rule and regenerates its
keys.

- rule: rule_name must be up to 60 characters of letters, numbers, periods, hyphens, and underscores, starting and ending with a letter or number
- rule: RootManageSharedAccessKey is the namespace's built-in root rule -- its keys already surface as AzureEventHubNamespace outputs; mint a differently-named rule for scoped access
- rule: {"required":true,"string":{"minLen":"1","maxLen":"60"}}

### spec.namespaceId

`string | valueFrom`

Namespace-wide scope: rights over EVERY hub in the namespace, by ARM
ID. References an AzureEventHubNamespace's namespace_id output. Set
exactly one of namespace_id or event_hub_id.

- references: AzureEventHubNamespace (`status.outputs.namespace_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureEventHubNamespace, name: <that resource's name>, fieldPath: status.outputs.namespace_id}} -- a bare string does not parse

### spec.eventHubId

`string | valueFrom`

Single-hub scope: rights over one event hub only, by ARM ID.
References an AzureEventHub's event_hub_id output. The
least-privilege choice for per-stream producers and consumers.

- references: AzureEventHub (`status.outputs.event_hub_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureEventHub, name: <that resource's name>, fieldPath: status.outputs.event_hub_id}} -- a bare string does not parse

### spec.listen

`bool` · optional (explicit presence)

Whether this rule may RECEIVE events within its scope.
Default: false

### spec.send

`bool` · optional (explicit presence)

Whether this rule may SEND events within its scope.
Default: false

### spec.manage

`bool` · optional (explicit presence)

Whether this rule may MANAGE entities within its scope (create and
delete hubs/consumer groups). Requires listen AND send -- manage is
a superset, never a standalone right.
Default: false

## Validation Rules

- `event_hub_auth_rule_exactly_one_scope`: set exactly one scope -- namespace_id for namespace-wide rights, or event_hub_id for one hub
- `event_hub_auth_rule_at_least_one_right`: grant at least one right -- listen, send, or manage (a rule with no rights would be an unusable credential)
- `event_hub_auth_rule_manage_requires_listen_send`: manage is a superset right -- Azure requires listen and send to both be true alongside it

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureEventHubAuthorizationRule, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.authorization_rule_id` | `string` | The Azure Resource Manager ID of the authorization rule. Namespace scope: .../namespaces/{ns}/authorizationRules/{name} Hub scope: .../namespaces/{ns}/eventhubs/{hub}/authorizationRules/{name} |
| `status.outputs.rule_name` | `string` | The rule's name (the SharedAccessKeyName clients present). |
| `status.outputs.primary_key` | `string` (sensitive) | The primary key. |
| `status.outputs.secondary_key` | `string` (sensitive) | The secondary key -- the rotation partner: move clients here, regenerate the primary, move back. |
| `status.outputs.primary_connection_string` | `string` (sensitive) | The ready-to-use primary connection string. Format: Endpoint=sb://{ns}.servicebus.windows.net/;SharedAccessKeyName={rule};SharedAccessKey={key} (hub-scoped rules append ;EntityPath={hub}) |
| `status.outputs.secondary_connection_string` | `string` (sensitive) | The secondary connection string (rotation partner). |
| `status.outputs.primary_connection_string_alias` | `string` (sensitive) | The primary connection string addressing the geo-DR alias hostname. Only populated when the namespace carries an AzureEventHubDisasterRecoveryConfig pairing; empty otherwise. |
| `status.outputs.secondary_connection_string_alias` | `string` (sensitive) | The secondary alias connection string (rotation partner). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespaceId` | AzureEventHubNamespace | `status.outputs.namespace_id` |
| `spec.eventHubId` | AzureEventHub | `status.outputs.event_hub_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureMonitorDiagnosticSetting | `spec.eventhubAuthorizationRuleId` | `status.outputs.authorization_rule_id` |

## See Also

- [Overview](../README.md)
