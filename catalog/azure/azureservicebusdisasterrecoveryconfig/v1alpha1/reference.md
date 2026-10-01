# AzureServiceBusDisasterRecoveryConfig

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureServiceBusDisasterRecoveryConfigSpec** defines a geo-disaster-
recovery pairing between two PREMIUM Service Bus namespaces: metadata
(queues, topics, subscriptions, rules, SAS rules -- not message data)
continuously replicates from the primary to the partner, and a
failover-stable ALIAS DNS name fronts whichever namespace is currently
primary.

Clients connect through the alias (`{alias_name}.servicebus.windows.net`)
instead of either namespace, so a failover needs no client
reconfiguration. The alias's connection strings come from the paired
authorization rule -- see alias_authorization_rule_id and the
`*_connection_string_alias` outputs on AzureServiceBusAuthorizationRule.

**Azure's pairing contracts** (apply-time -- they involve both live
namespaces): both namespaces must be PREMIUM, in DIFFERENT regions, and
the partner must be empty (no entities) when pairing is created.
Changing partner_namespace_id breaks the existing pairing and re-pairs.

**Failover is an operational action, not a config change**: triggered
from the SECONDARY side (portal/CLI/SDK) during a regional incident.
Deleting this resource breaks the pairing gracefully (both namespaces
keep running independently; the alias name is released after deletion).

**ForceNew fields**: `alias_name`, `primary_namespace_id`.

## Example

```yaml
# Offline-plan manifest: a geo-DR pairing with a scoped alias rule --
# exercises both namespace references and the least-privilege alias
# credential seam.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureServiceBusDisasterRecoveryConfig
metadata:
  name: test-sb-geo-dr
spec:
  aliasName: hack-servicebus-alias
  primaryNamespaceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ServiceBus/namespaces/hack-sb-eastus
  partnerNamespaceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ServiceBus/namespaces/hack-sb-westus
  aliasAuthorizationRuleId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ServiceBus/namespaces/hack-sb-eastus/authorizationRules/dr-clients
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.aliasName` | `string` | yes |  |  |
| `spec.primaryNamespaceId` | `string \| valueFrom` | yes |  | AzureServiceBusNamespace (`status.outputs.namespace_id`) |
| `spec.partnerNamespaceId` | `string \| valueFrom` | yes |  | AzureServiceBusNamespace (`status.outputs.namespace_id`) |
| `spec.aliasAuthorizationRuleId` | `string \| valueFrom` |  |  | AzureServiceBusAuthorizationRule (`status.outputs.authorization_rule_id`) |

## Field Details

### spec.aliasName

`string` · required

The alias name -- becomes the failover-stable DNS name
`{alias_name}.servicebus.windows.net`, so it shares the namespace
name rules and uniqueness scope (globally unique; it cannot collide
with any existing namespace name either): 6-50 characters of
letters, numbers, and hyphens, starting with a letter and ending
with a letter or number, and never ending in Azure's reserved
"-sb" / "-mgmt" suffixes. The rules below transcribe the namespace
kind's own name contract so a bad alias fails at authoring time,
not at Azure apply time.

**ForceNew**: changing the alias replaces the pairing.

- rule: alias_name must be 6-50 characters of letters, numbers, and hyphens, starting with a letter and ending with a letter or number (it shares the namespace name rules)
- rule: alias_name cannot end with '-sb' or '-mgmt' -- Azure reserves these suffixes for its own service endpoints
- rule: {"required":true,"string":{"minLen":"6","maxLen":"50"}}

### spec.primaryNamespaceId

`string | valueFrom` · required

The PRIMARY namespace -- the one clients actively use, by ARM ID.
References an AzureServiceBusNamespace's namespace_id output. Must be
PREMIUM. Fixed at creation.

- references: AzureServiceBusNamespace (`status.outputs.namespace_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureServiceBusNamespace, name: <that resource's name>, fieldPath: status.outputs.namespace_id}} -- a bare string does not parse

### spec.partnerNamespaceId

`string | valueFrom` · required

The PARTNER (secondary) namespace metadata replicates to, by ARM ID.
Must be PREMIUM, in a DIFFERENT region than the primary, and empty
(no queues/topics) at pairing time. Changing it breaks the current
pairing and re-pairs to the new partner.

- references: AzureServiceBusNamespace (`status.outputs.namespace_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureServiceBusNamespace, name: <that resource's name>, fieldPath: status.outputs.namespace_id}} -- a bare string does not parse

### spec.aliasAuthorizationRuleId

`string | valueFrom`

The authorization rule whose keys the alias connection strings carry,
by ARM ID. References an AzureServiceBusAuthorizationRule's
authorization_rule_id output (a NAMESPACE-scoped rule on the
primary). Unset defaults to the namespace's built-in root rule
(RootManageSharedAccessKey) -- prefer a scoped rule for
least-privilege alias credentials.

- references: AzureServiceBusAuthorizationRule (`status.outputs.authorization_rule_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureServiceBusAuthorizationRule, name: <that resource's name>, fieldPath: status.outputs.authorization_rule_id}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureServiceBusDisasterRecoveryConfig, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.disaster_recovery_config_id` | `string` | The Azure Resource Manager ID of the disaster-recovery config (under the primary namespace). Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.ServiceBus/namespaces/{ns}/disasterRecoveryConfigs/{alias} |
| `status.outputs.alias_name` | `string` | The alias name -- the failover-stable DNS identity `{alias_name}.servicebus.windows.net`. |
| `status.outputs.primary_connection_string_alias` | `string` (sensitive) | The primary connection string addressing the ALIAS -- what DR-aware clients hold. |
| `status.outputs.secondary_connection_string_alias` | `string` (sensitive) | The secondary alias connection string -- the rotation partner. |
| `status.outputs.default_primary_key` | `string` (sensitive) | The paired rule's primary key (the same key the alias connection string embeds). |
| `status.outputs.default_secondary_key` | `string` (sensitive) | The paired rule's secondary key -- the rotation partner. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.primaryNamespaceId` | AzureServiceBusNamespace | `status.outputs.namespace_id` |
| `spec.partnerNamespaceId` | AzureServiceBusNamespace | `status.outputs.namespace_id` |
| `spec.aliasAuthorizationRuleId` | AzureServiceBusAuthorizationRule | `status.outputs.authorization_rule_id` |

## See Also

- [Overview](../README.md)
