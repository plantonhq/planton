# Azure Event Hub Authorization Rule

Deploys a shared-access authorization rule for Azure Event Hubs -- a named SAS credential with listen/send/manage rights, scoped to exactly one of a namespace (every hub) or a single event hub. The scope is the blast radius: a producer service holds a send-only rule on its one stream, a consumer fleet holds a listen-only rule, and neither can touch anything else. Rights contract (Azure's own): at least one of listen/send/manage must be granted, and manage requires BOTH listen and send. The rule's keys and connection strings surface as sensitive outputs; both the scope and the rule name are fixed at creation.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Authorization Rule** -- namespace-scoped or hub-scoped per which reference the spec carries, with your chosen rights
- **Two keys and their connection strings** -- primary and secondary, surfaced as sensitive outputs for zero-downtime rotation; hub-scoped connection strings carry EntityPath so clients connect straight to the stream

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **The scope's parent**: an AzureEventHub (reference its `event_hub_id` output) for the least-privilege hub scope, or an AzureEventHubNamespace (reference its `namespace_id` output) for operational tooling.
- **A rule name** unique within the scope -- up to 60 characters. `RootManageSharedAccessKey` is the namespace's built-in root rule and is reserved; its keys already surface as AzureEventHubNamespace outputs.

## Deploy

### Console

Open the deployment store, find **Azure Event Hub Authorization Rule**, and click **Deploy**. The creation wizard walks you through the credential scope (with the blast-radius model taught live), the rule name, and the rights trio with Producer / Consumer / Full manage quick-picks. Start from the **Hub-Scoped Producer Credential** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventHubAuthorizationRule
metadata:
  name: producer-credential
  org: acme-corp
  env: prod
spec:
  ruleName: telemetry-producer
  eventHubId:
    valueFrom:
      kind: AzureEventHub
      name: telemetry-stream
      fieldPath: status.outputs.event_hub_id
  send: true
```

```shell
planton apply -f auth-rule.yaml
```

This creates a send-only SAS rule named `telemetry-producer` scoped to the `telemetry-stream` hub, with its keys and connection strings surfaced as sensitive outputs. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the rule to a hub deployed in the same InfraPipeline:

```yaml
spec:
  eventHubId:
    valueFrom:
      kind: AzureEventHub
      name: telemetry-stream
      fieldPath: status.outputs.event_hub_id
```

The InfraPipeline resolves the dependency graph, deploys the namespace and hub first, then provisions the rule with the resolved values.

## Key Configuration

These are the most important decisions when configuring an authorization rule. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The scope** -- hub-scoped rules are the least-privilege choice for application workloads: the connection string is entity-scoped, and a leaked credential exposes one stream. Namespace-scoped rules cover every hub, present and future -- reserve them for entity-management tooling, and note that a namespace-scoped rule is also what geo-DR pairing surfaces through the alias connection strings.

**The rights trio** -- `send` is the producer right (append only), `listen` the consumer right (read through consumer groups only), and `manage` the superset for tooling that creates and deletes entities (Azure requires listen AND send alongside it). Rights edit in place later without regenerating keys -- widening a consumer to a producer-consumer is a day-two edit, not a new credential.

**Rotation** -- primary and secondary keys exist for zero-downtime rotation: move clients to the secondary, regenerate the primary in Azure, move back. Prefer keyless Entra data-plane roles (via AzureRoleAssignment) where clients support them; SAS rules remain the shape for Kafka clients and legacy SDKs.

**Fixed at creation** -- exactly one of `namespaceId` and `eventHubId` must be set, and both the scope and `ruleName` are one-way doors: replacing either regenerates the keys and cuts off every client holding the old connection strings. The rights trio (`listen`/`send`/`manage`) edits in place without touching the keys -- widening a consumer to a producer-consumer is a day-two edit, not a new credential.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureEventHubNamespace** | `namespaceId` | `status.outputs.namespace_id` |
| **AzureEventHub** | `eventHubId` | `status.outputs.event_hub_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `authorization_rule_id` | Azure Resource Manager ID of the rule | Governance and automation addressing the credential as an ARM resource |
| `rule_name` | The SharedAccessKeyName clients carry | Tracing which credential a client holds from its configuration |
| `primary_key` / `secondary_key` | The two SAS keys (sensitive) | Application configuration via Config Manager secret references |
| `primary_connection_string` / `secondary_connection_string` | Ready-to-use connection strings (sensitive; hub-scoped rules append EntityPath) | The value producer and consumer applications hold |
| `primary_connection_string_alias` / `secondary_connection_string_alias` | Failover-stable alias connection strings (sensitive) | Populated only when the namespace carries a geo-DR pairing -- DR-aware clients survive failover without reconfiguration |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Hub producer** -- the everyday least-privilege credential: a send-only rule scoped to one stream. Start from the **Hub-Scoped Producer Credential** preset.

**Namespace operator** -- the full-manage trio over the whole namespace, for entity-management tooling. Start from the **Namespace-Scoped Operator Credential** preset.

## Works With

- [**Azure Event Hub**](/infra-catalog/azure-event-hub) -- the single-stream scope for application credentials
- [**Azure Event Hub Namespace**](/infra-catalog/azure-event-hub-namespace) -- the namespace-wide scope for tooling; its root rule's keys are its own outputs
- [**Azure Event Hub Consumer Group**](/infra-catalog/azure-event-hub-consumer-group) -- the cursor a listen-rights holder reads through
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- the keyless alternative: Entra data-plane roles scoped to the hub or namespace
