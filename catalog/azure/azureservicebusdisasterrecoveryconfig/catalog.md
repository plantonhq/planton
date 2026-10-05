# Azure Service Bus Disaster Recovery Config

Pairs two PREMIUM Azure Service Bus namespaces for geo-disaster recovery: metadata (queues, topics, subscriptions, rules, SAS rules -- not message data) continuously replicates from the primary to the partner, and a failover-stable ALIAS DNS name fronts whichever namespace is currently primary. Clients connect through the alias instead of either namespace, so a failover needs no client reconfiguration.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **The geo-DR pairing** -- on the primary namespace, with continuous metadata replication to the partner
- **The alias DNS name** -- `{alias}.servicebus.windows.net`, resolving to whichever side is currently primary
- **Alias connection strings** -- surfaced as sensitive outputs, carrying the paired authorization rule's keys (the namespace's root rule when none is referenced)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **Two PREMIUM AzureServiceBusNamespaces in DIFFERENT regions** -- reference both `namespace_id` outputs via ValueFromRef. The partner must be EMPTY (no queues or topics) at pairing time; replication populates it.
- **An alias name** -- globally unique with the same rules as a namespace name (it lives in the same DNS space and cannot collide with one either).
- Optionally, **a NAMESPACE-scoped AzureServiceBusAuthorizationRule on the primary** for least-privilege alias credentials.

## Deploy

### Console

Open the deployment store, find **Azure Service Bus Disaster Recovery Config**, and click **Deploy**. The creation wizard walks you through the alias (with a live DNS preview and the namespace-grammar validation), the pairing (with Azure's three apply-time contracts taught live), and the alias credentials with the root-rule-default honesty. Start from the **Geo-DR Pairing** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureServiceBusDisasterRecoveryConfig
metadata:
  name: app-bus-dr
  org: acme-corp
  env: prod
spec:
  aliasName: myorg-app-bus-alias
  primaryNamespaceId:
    valueFrom:
      kind: AzureServiceBusNamespace
      name: app-bus-eastus
      fieldPath: status.outputs.namespace_id
  partnerNamespaceId:
    valueFrom:
      kind: AzureServiceBusNamespace
      name: app-bus-westus
      fieldPath: status.outputs.namespace_id
  aliasAuthorizationRuleId:
    valueFrom:
      kind: AzureServiceBusAuthorizationRule
      name: dr-clients-rule
      fieldPath: status.outputs.authorization_rule_id
```

```shell
planton apply -f dr-config.yaml
```

This pairs the east and west namespaces under the `myorg-app-bus-alias` DNS name, with alias credentials carrying the scoped DR rule's keys instead of the root rule's. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the pairing to namespaces deployed in the same InfraPipeline:

```yaml
spec:
  primaryNamespaceId:
    valueFrom:
      kind: AzureServiceBusNamespace
      name: app-bus-eastus
      fieldPath: status.outputs.namespace_id
```

The InfraPipeline resolves the dependency graph, deploys both namespaces first, then provisions the pairing with the resolved values.

## Key Configuration

These are the most important decisions when configuring a geo-DR pairing. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The alias is the contract** -- `aliasName` becomes the DNS identity every DR-aware client holds. Name it for the SERVICE the pair serves, never for a region: the whole point is that the name outlives whichever region is currently primary.

**Metadata, not messages** -- replication protects the topology (entities, rules, SAS rules), not in-flight message data. Applications with strict no-message-loss requirements pair this with sender-side persistence.

**Failover is operational** -- it is triggered from the SECONDARY side (portal/CLI/SDK) during a regional incident, never a configuration change here. After a promotion, set up a NEW pairing to a new partner region.

**Least-privilege alias credentials** -- unset `aliasAuthorizationRuleId` defaults to the namespace's root rule (full manage rights for every DR client). Prefer a scoped NAMESPACE-level rule minted for DR clients.

**Lifecycle semantics** -- the alias and the primary are fixed at creation; changing the partner breaks the current pairing and re-pairs to the new one while the alias and primary keep serving. Deleting the resource breaks the pairing gracefully: both namespaces keep running independently, and the alias name is released after deletion.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureServiceBusNamespace** | `primaryNamespaceId` | `status.outputs.namespace_id` |
| **AzureServiceBusNamespace** | `partnerNamespaceId` | `status.outputs.namespace_id` |
| **AzureServiceBusAuthorizationRule** | `aliasAuthorizationRuleId` | `status.outputs.authorization_rule_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `disaster_recovery_config_id` | Azure Resource Manager ID of the pairing (under the primary namespace) | Diagnostics and governance tooling |
| `alias_name` | The failover-stable DNS identity | Composing `{alias}.servicebus.windows.net` endpoints |
| `primary_connection_string_alias` / `secondary_connection_string_alias` | Connection strings addressing the ALIAS (sensitive) | What DR-aware applications hold -- they survive a failover without reconfiguration |
| `default_primary_key` / `default_secondary_key` | The paired rule's keys (sensitive) | The same keys the alias connection strings embed; the secondary is the rotation partner |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Geo-DR pairing** -- the full story: an east/west PREMIUM pair fronted by one alias, with a scoped rule for least-privilege alias credentials. Start from the **Geo-DR Pairing** preset.

## Works With

- [**Azure Service Bus Namespace**](/infra-catalog/azure-service-bus-namespace) -- both sides of the pair (PREMIUM, different regions)
- [**Azure Service Bus Authorization Rule**](/infra-catalog/azure-service-bus-authorization-rule) -- the namespace-scoped rule whose keys the alias connection strings carry
- [**Azure Service Bus Queue**](/infra-catalog/azure-service-bus-queue) -- replicated to the partner as metadata, like every entity in the namespace
- [**Azure Service Bus Topic**](/infra-catalog/azure-service-bus-topic) -- likewise replicated; subscriptions and rules included
