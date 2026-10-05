# Azure Event Grid Namespace Topic

Deploys one named CloudEvents stream inside an Azure Event Grid namespace, with a 1-7 day delivery retention window as its only tunable property. A namespace holds many topics with independent lifecycles: publishers and teams create and delete their own topics against the shared namespace without touching it or each other. Azure pins the event schema to CloudEvents v1.0 and the publisher type to Custom on this resource -- there is nothing else to configure.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Event Grid namespace topic** -- one CloudEvents stream inside the referenced namespace, with its delivery retention window; the schema (CloudEvents v1.0) and publisher type (Custom) are fixed by Azure and sent by both engines

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### Azure Subscription

- **An Event Grid namespace** -- reference an AzureEventgridNamespace Infra Component's ID output or pass an existing namespace's ARM ID. The topic shares the namespace's throughput units and adds no cost of its own.

## Deploy

### Console

Open the deployment store, find **Azure Event Grid Namespace Topic**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields: the namespace reference, the topic's name, and retention. Start from the **Service Stream** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventgridNamespaceTopic
metadata:
  name: orders-stream
  org: acme-corp
  env: prod
spec:
  namespaceId:
    valueFrom:
      kind: AzureEventgridNamespace
      name: events-hub
      fieldPath: status.outputs.namespace_id
  name: orders
  eventRetentionInDays: 7
```

```shell
planton apply -f namespace-topic.yaml
```

This creates one CloudEvents stream named `orders` inside the referenced namespace, holding published events for delivery for up to 7 days. An Infra Job tracks the provisioning in real time.

### InfraChart

When the namespace and its topics deploy in the same InfraChart, wire them with ValueFromRef:

```yaml
spec:
  namespaceId:
    valueFrom:
      kind: AzureEventgridNamespace
      name: events-hub
      fieldPath: status.outputs.namespace_id
  name: payments
  eventRetentionInDays: 3
```

The InfraPipeline resolves the dependency graph, deploys the namespace first, then provisions the topic with the resolved ARM ID.

## Key Configuration

These are the most important decisions when configuring a namespace topic. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The name is the stream's identity** -- unique within the namespace only (topics have no public hostname of their own), and ForceNew along with `namespaceId`: changing either replaces the topic and drops everything still buffered in it. Drain or stop publishers before any replace or delete; the operation itself succeeds regardless.

**eventRetentionInDays is a delivery buffer, not an archive** -- the 1-7 day window (default 7) is how long Event Grid holds events for delivery, and it is the topic's ONLY updatable property. If consumers can be down longer than the window, or you need replay beyond it, land the events somewhere durable instead of stretching retention.

**The topic is the tenant boundary, the namespace is the landlord** -- keep the namespace in a platform-owned resource group and pipeline, and each topic with the service that publishes to it. Folding topic creation into the namespace's pipeline recreates the bottleneck this resource model exists to remove. Topics are free; throughput bills on the namespace's shared throughput units.

**Subscriptions on namespace topics live outside the catalog today** -- Azure models namespace-topic event subscriptions as their own ARM resource, and the pinned Terraform provider does not ship it yet. Azure Event Grid Event Subscription addresses the CLASSIC resources only. Until the provider catches up, manage namespace-topic subscriptions with `az eventgrid namespace topic event-subscription`, or route through a classic topic via the namespace's MQTT bridge when that fits the shape.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureEventgridNamespace** | `namespaceId` | `status.outputs.namespace_id` |

### What This Kind Provides

`status.outputs` records identifiers only: `namespace_topic_id` (the ARM ID, `{namespace_id}/topics/{name}`) and `namespace_topic_name`, which echoes the manifest's `name`. No catalog kind consumes them via ValueFromRef today -- namespace-topic subscriptions are not yet a catalog kind, so the ID's consumers are CLI tooling and scripts addressing the stream.

## Common Patterns

**One stream per service** -- the platform team owns the namespace; each publishing service owns its topic inside it, onboarding and retiring streams without touching the shared hub. Start from the **Service Stream** preset.

**Tenant onboarding** -- each tenant gets a named stream inside the shared namespace, created and deleted with the tenant's lifecycle rather than the hub's.

**Short retention for hot streams** -- high-volume streams consumed within minutes do not need the 7-day default; tuning retention down bounds how much undelivered backlog a stalled consumer can accumulate.

## Works With

- [**Azure Event Grid Namespace**](/infra-catalog/azure-eventgrid-namespace) -- the capacity-scaled hub the topic lives in and whose throughput units it shares
- [**Azure Event Grid Topic**](/infra-catalog/azure-eventgrid-topic) -- the classic custom topic that bridges namespace MQTT traffic into the classic delivery machinery when subscriptions are needed today
