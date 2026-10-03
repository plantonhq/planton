# Azure Service Bus Authorization Rule

Mints a SAS (shared-access-signature) credential for Azure Service Bus: a named rule with listen/send/manage rights, scoped to exactly one of a namespace, a queue, or a topic. Authorization rules are how applications get least-privilege connection strings -- a sender service holds a send-only rule on its one queue, a worker holds a listen-only rule, and neither can touch anything else. The rule's keys and connection strings surface as sensitive outputs, with a secondary pair for zero-downtime rotation.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A SAS authorization rule** -- at namespace, queue, or topic scope (Azure models these as three ARM types with identical shapes; this kind dispatches to the right one from whichever parent you set)
- **Primary and secondary keys** -- with ready-to-use connection strings for each, surfaced as sensitive outputs; the secondary pair exists for zero-downtime rotation
- **Geo-DR alias connection strings** -- populated automatically when the namespace carries a disaster-recovery pairing

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **Exactly one parent** to scope the credential to: an AzureServiceBusNamespace (`namespace_id` output), an AzureServiceBusQueue (`queue_id` output), or an AzureServiceBusTopic (`topic_id` output) -- referenced via ValueFromRef. The scope is fixed at creation.
- **A rule name** unique within its scope -- up to 50 characters of letters, numbers, periods, hyphens, and underscores. `RootManageSharedAccessKey` is reserved for the namespace's built-in root rule.

## Deploy

### Console

Open the deployment store, find **Azure Service Bus Authorization Rule**, and click **Deploy**. The creation wizard walks you through the credential scope (namespace-wide, one queue, or one topic -- the blast-radius decision), the rule identity, and the rights trio with persona quick-picks (Sender / Listener / Full manage) and Azure's manage-superset contract enforced live. Start from the **Queue Sender Credential** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureServiceBusAuthorizationRule
metadata:
  name: orders-sender-rule
  org: acme-corp
  env: prod
spec:
  ruleName: orders-api-sender
  queueId:
    valueFrom:
      kind: AzureServiceBusQueue
      name: orders-queue
      fieldPath: status.outputs.queue_id
  send: true
```

```shell
planton apply -f rule.yaml
```

This mints `orders-api-sender`, a send-only SAS credential scoped to the `orders-queue` queue -- the tightest credential Service Bus can offer a producer. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the rule to an entity deployed in the same InfraPipeline:

```yaml
spec:
  queueId:
    valueFrom:
      kind: AzureServiceBusQueue
      name: orders-queue
      fieldPath: status.outputs.queue_id
```

The InfraPipeline resolves the dependency graph, deploys the queue first, then provisions the rule with the resolved values.

## Key Configuration

These are the most important decisions when configuring an authorization rule. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The scope IS the blast radius** -- `namespaceId` grants rights over EVERY queue and topic (tooling, never app workloads); `queueId` and `topicId` scope to one entity. A leaked queue-scoped sender credential exposes one queue, not the namespace.

**The rights trio** -- `listen` receives (and browses/peeks), `send` produces, `manage` creates and deletes entities. Azure's contract: at least one right must be true, and `manage` requires BOTH listen and send -- it is a superset, never standalone. One rule per holder beats one shared rule: when the orders API's rule leaks, you rotate it alone.

**Rotation by design** -- primary and secondary keys both work at all times: move clients to the secondary, regenerate the primary in Azure, move back. Rights edits update the rule IN PLACE (the keys keep working); renaming or re-scoping replaces the rule and its keys.

**The keyless alternative** -- for clients that can hold an Entra identity, skip SAS entirely: disable the namespace's `local_auth_enabled` and grant data-plane roles (Azure Service Bus Data Sender / Data Receiver) via AzureRoleAssignment. SAS rules remain the right tool for everything else.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureServiceBusNamespace** | `namespaceId` | `status.outputs.namespace_id` |
| **AzureServiceBusQueue** | `queueId` | `status.outputs.queue_id` |
| **AzureServiceBusTopic** | `topicId` | `status.outputs.topic_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `authorization_rule_id` | Azure Resource Manager ID of the rule | AzureServiceBusDisasterRecoveryConfig's `alias_authorization_rule_id` consumes a NAMESPACE-scoped rule's ID with zero translation |
| `rule_name` | The SharedAccessKeyName clients present | Diagnostics and connection-string composition |
| `primary_key` / `secondary_key` | The SAS keys (sensitive) | Application secrets -- the secondary is the rotation partner |
| `primary_connection_string` / `secondary_connection_string` | Ready-to-use connection strings (sensitive), entity-scoped when queue- or topic-scoped | What applications actually hold |
| `primary_connection_string_alias` / `secondary_connection_string_alias` | Connection strings addressing the geo-DR ALIAS (sensitive) | What DR-aware clients hold -- empty unless the namespace carries a disaster-recovery pairing |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Queue sender** -- the tightest credential Service Bus can mint: send-only on one queue, the shape most producer services should hold. Start from the **Queue Sender Credential** preset.

**Namespace operator** -- the full listen+send+manage trio at namespace scope, for entity-management tooling and deploy pipelines. Start from the **Namespace Operator Credential** preset.

## Works With

- [**Azure Service Bus Namespace**](/infra-catalog/azure-service-bus-namespace) -- the namespace-wide scope, and the root rule's home
- [**Azure Service Bus Queue**](/infra-catalog/azure-service-bus-queue) -- the least-privilege single-queue scope
- [**Azure Service Bus Topic**](/infra-catalog/azure-service-bus-topic) -- the single-topic scope (publishers and subscription consumers)
- [**Azure Service Bus Disaster Recovery Config**](/infra-catalog/azure-service-bus-disaster-recovery-config) -- consumes a namespace-scoped rule's ID for least-privilege alias credentials
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- the keyless Entra alternative to SAS
