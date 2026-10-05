# Azure Service Bus Subscription

Deploys a subscription under an Azure Service Bus topic -- an independent, optionally filtered view of the topic's message stream, with its own consumer semantics: lock duration, delivery attempts, sessions, and dead-lettering. Subscriptions are many-per-topic and typically owned by the consuming team rather than the team that provisioned the namespace -- which is why they are a first-class Infra Component referencing the topic.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Service Bus Subscription** -- under the referenced topic, with your chosen delivery, TTL, session, dead-lettering, and lifecycle dials
- **Filter rules** -- when the `rules` list is populated: SQL or correlation rules admitting messages into this subscription's view (additive alongside Azure's auto-created `$Default` catch-all)
- **Auto-forwarding** -- when `forwardTo` or `forwardDeadLetteredMessagesTo` is set: broker-side routing into another queue or topic in the same namespace, by entity name
- **Client scoping** -- when the `clientScopedSubscription` block is configured: the JMS 2.0 durable-subscription binding to a specific client ID

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureServiceBusTopic** this subscription attaches to. Reference its `topic_id` output via ValueFromRef -- the deploy graph creates the topic first.
- **A subscription name** unique within the topic -- up to 50 characters of letters, numbers, periods, hyphens, and underscores (a different rule than queue and topic names).
- **A deliberate delivery tolerance** -- `maxDeliveryCount` is REQUIRED: Azure has no server default for how many failed deliveries a message gets before dead-lettering (10 is the queue-side convention).

## Deploy

### Console

Open the deployment store, find **Azure Service Bus Subscription**, and click **Deploy**. The creation wizard walks you through the topic attachment, the delivery contract (with the required tolerance pre-filled at the conventional 10), dead-lettering, sessions and lifecycle, JMS client scoping, the filter-rules builder (with the `$Default` catch-all trap taught live), auto-forwarding, and the gate. Start from the **Catch-All Consumer** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureServiceBusSubscription
metadata:
  name: emea-consumer-sub
  org: acme-corp
  env: prod
spec:
  topicId:
    valueFrom:
      kind: AzureServiceBusTopic
      name: order-events-topic
      fieldPath: status.outputs.topic_id
  subscriptionName: emea-consumer
  maxDeliveryCount: 10
  rules:
    - ruleName: emea-priority
      filterType: SQL_FILTER
      sqlFilter: "region = 'emea' AND priority > 3"
```

```shell
planton apply -f subscription.yaml
```

This creates the `emea-consumer` subscription on the `order-events-topic` topic, admitting only high-priority EMEA messages via the SQL rule (plus everything else through the `$Default` catch-all until it is removed -- see Key Configuration). An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the subscription to a topic deployed in the same InfraPipeline:

```yaml
spec:
  topicId:
    valueFrom:
      kind: AzureServiceBusTopic
      name: order-events-topic
      fieldPath: status.outputs.topic_id
```

The InfraPipeline resolves the dependency graph, deploys the topic first, then provisions the subscription with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Service Bus subscription. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The required delivery tolerance** -- `maxDeliveryCount` is the poison-message circuit breaker with NO Azure default: after that many failed deliveries a message moves to this subscription's dead-letter sub-queue at `{topic}/subscriptions/{name}/$deadletterqueue`. Fast idempotent workers want 3-5; consumers calling flaky dependencies want 10+ with a re-submit job.

**Filter rules** -- SQL rules (`sys.Label = 'important' AND quantity > 10`) are flexible and evaluated per message; correlation rules (pure equality on correlation properties) are cheaper at high throughput. Rules combine with OR semantics, and the optional SQL `action` annotates matches before delivery. Messages that CRASH a SQL filter are dead-lettered by default (`deadLetteringOnFilterEvaluationError` defaults true) -- evidence for finding the malformed producer.

**The `$Default` catch-all trap** -- Azure auto-creates a catch-all rule named `$Default` on every new subscription, and declared rules are ADDITIVE beside it: the subscription keeps receiving EVERYTHING until the catch-all is removed once, after creation, with `az servicebus topic subscription rule delete --name '$Default'` (it never comes back unless the subscription is recreated). `$Default` is also not a declarable rule name.

**Sessions** -- `requiresSession: true` is the consumer half of ordered pub/sub (the topic's `supportOrdering` is the publisher half): each SessionId delivers in strict order to one session-aware consumer at a time. Fixed at creation, as is the `clientScopedSubscription` identity -- changing either replaces the subscription and RESETS its read position, losing undelivered messages.

**Filter-then-funnel routing** -- `forwardTo` turns the subscription into a routing edge: rules decide WHAT is admitted, forwarding decides WHERE it lands (typically a work queue), by entity NAME in the same namespace. The target must exist first and must not be session-aware.

**The per-consumer gate** -- `status: RECEIVE_DISABLED` buffers matched messages while THIS consumer redeploys; the topic's other subscriptions keep flowing. The backlog counts against the topic's shared size budget.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureServiceBusTopic** | `topicId` | `status.outputs.topic_id` |
| **AzureServiceBusQueue** | `forwardTo` / `forwardDeadLetteredMessagesTo` | `status.outputs.queue_name` |
| **AzureServiceBusTopic** | `forwardTo` / `forwardDeadLetteredMessagesTo` | `status.outputs.topic_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `subscription_name` | The subscription's name within the topic | Consumer SDK configuration (with the topic name) |
| `topic_name` | The parent topic's name, parsed from the resolved reference | The receive pair without a second reference |
| `namespace_name` | The namespace's name, parsed from the resolved reference | The full namespace/topic/subscription receive triple |

Consumers configure the receive triple -- the namespace endpoint, the topic name, and this subscription name. Credentials come from an AzureServiceBusAuthorizationRule or keyless Entra data-plane roles; the subscription itself mints none. The `subscription_id` output carries the ARM ID for audit tooling but is not typically wired into other Infra Components.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Catch-all consumer** -- no declared rules; the `$Default` catch-all delivers everything. The right shape for audit archives and full-stream mirrors. Start from the **Catch-All Consumer** preset.

**Filtered consumer** -- a SQL rule admits only this team's slice of the stream (remove the catch-all once for restrictive delivery). Start from the **Filtered Consumer** preset.

**Fan-out to work queue** -- a correlation rule admits matches and forwarding funnels them into a dedicated queue a processing fleet drains. Start from the **Fan-Out to Work Queue** preset.

## Works With

- [**Azure Service Bus Topic**](/infra-catalog/azure-service-bus-topic) -- the parent topic every subscription references
- [**Azure Service Bus Queue**](/infra-catalog/azure-service-bus-queue) -- the common forward-to target in the filter-then-funnel pattern
- [**Azure Service Bus Namespace**](/infra-catalog/azure-service-bus-namespace) -- the namespace whose endpoint consumers connect to
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- keyless data-plane grants for consumers
