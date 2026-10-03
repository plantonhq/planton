# Azure Event Hub Disaster Recovery Config

Creates a geo-disaster-recovery pairing between two Event Hubs namespaces: metadata (hubs, consumer groups, authorization rules -- not event data) continuously replicates from the primary to the partner, and a failover-stable ALIAS DNS name (`{alias}.servicebus.windows.net`) fronts whichever namespace is currently primary. Clients connect through the alias instead of either namespace, so a regional failover needs no client reconfiguration. What replicates is structure and credentials, never events -- after a failover the partner has the same hubs, groups, and rules, but starts with empty streams.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Geo-DR pairing** -- under the primary namespace, replicating metadata continuously to the partner
- **The alias** -- the failover-stable DNS name, globally unique in the namespace name scope

The provider manages the pairing's lifecycle choreography: creation waits for the pairing to reach the Succeeded state, changing the partner BREAKS the existing pairing and re-pairs, and destroy breaks the pairing then waits for Azure to release the alias name -- destroys take minutes by the service's own design.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **Two AzureEventHubNamespaces** in DIFFERENT regions, on the same tier (STANDARD or higher -- geo-DR is not available on BASIC), with the partner EMPTY (no hubs) at pairing time -- Azure validates all three when the pairing is created. Reference both namespaces' `namespace_id` outputs via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **Azure Event Hub Disaster Recovery Config**, and click **Deploy**. The creation wizard walks you through the alias identity (with a live failover-stable endpoint preview) and the namespace pairing, with Azure's three apply-time contracts taught inline. Start from the **Geo-DR Pairing** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventHubDisasterRecoveryConfig
metadata:
  name: telemetry-geo-dr
  org: acme-corp
  env: prod
spec:
  aliasName: myorg-telemetry-alias
  primaryNamespaceId:
    valueFrom:
      kind: AzureEventHubNamespace
      name: telemetry-hubs
      fieldPath: status.outputs.namespace_id
  partnerNamespaceId:
    valueFrom:
      kind: AzureEventHubNamespace
      name: telemetry-hubs-dr
      fieldPath: status.outputs.namespace_id
```

```shell
planton apply -f geo-dr.yaml
```

This creates a geo-DR pairing under the `telemetry-hubs` namespace, replicating its metadata to `telemetry-hubs-dr` and fronting both with the `myorg-telemetry-alias` DNS name. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the pairing to namespaces deployed in the same InfraPipeline:

```yaml
spec:
  primaryNamespaceId:
    valueFrom:
      kind: AzureEventHubNamespace
      name: telemetry-hubs
      fieldPath: status.outputs.namespace_id
```

The InfraPipeline resolves the dependency graph, deploys both namespaces first, then creates the pairing.

## Key Configuration

These are the most important decisions when configuring a geo-DR pairing. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The alias** -- `aliasName` becomes the DNS name DR-aware clients hold. It shares the namespace name uniqueness scope (globally unique; it cannot collide with any existing namespace name either). Move applications onto the alias-addressed connection strings BEFORE an incident -- the alias protects nothing if clients still address the primary directly.

**The pairing** -- `primaryNamespaceId` is the active side the pairing lives under; `partnerNamespaceId` is the standby metadata replicates to. Understand what replicates: STRUCTURE and CREDENTIALS, not events -- after a failover the partner has the same hubs, groups, and rules, but starts with empty streams.

**Failover** -- an operational action, not a config change: triggered from the SECONDARY side (portal/CLI/SDK) during a regional incident, because the primary's region may be unreachable. After failover the alias points at the former partner. Deleting this resource breaks the pairing gracefully -- both namespaces keep running independently.

**One-way doors** -- `aliasName` and `primaryNamespaceId` are fixed at creation: changing either replaces the pairing. Changing `partnerNamespaceId` breaks the current pairing and re-pairs to the new partner -- a live operation, not a rebuild, but the new partner must again be empty and cross-region.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureEventHubNamespace** | `primaryNamespaceId` | `status.outputs.namespace_id` |
| **AzureEventHubNamespace** | `partnerNamespaceId` | `status.outputs.namespace_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `disaster_recovery_config_id` | Azure Resource Manager ID of the pairing | Governance tooling and ARM-level references |
| `alias_name` | The failover-stable alias | Operational runbooks and DNS-level documentation |

The alias-addressed CONNECTION STRINGS surface on the paired kinds, not here: once the pairing exists, the namespace's and each authorization rule's `*_alias` outputs populate -- hand applications those.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Cross-region continuity** -- the everyday shape: an alias fronting the active namespace, paired to an empty partner in the DR region, with applications on alias-addressed credentials from day one. Start from the **Geo-DR Pairing** preset.

## Works With

- [**Azure Event Hub Namespace**](/infra-catalog/azure-event-hub-namespace) -- both sides of the pairing; their `*_alias` connection-string outputs populate once paired
- [**Azure Event Hub Authorization Rule**](/infra-catalog/azure-event-hub-authorization-rule) -- rules replicate to the partner, and their alias-addressed connection strings are what DR-aware clients hold
- [**Azure Event Hub**](/infra-catalog/azure-event-hub) -- hub structure replicates; event data does not
- [**Azure Event Hub Consumer Group**](/infra-catalog/azure-event-hub-consumer-group) -- group structure replicates; offsets are per-namespace
