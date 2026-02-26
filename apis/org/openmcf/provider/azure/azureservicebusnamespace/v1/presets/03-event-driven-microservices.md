# Event-Driven Microservices Service Bus

This preset creates an Azure Service Bus namespace configured for an event-driven microservices architecture with queue chaining and dead-letter forwarding. The Standard tier keeps costs low (~$0.05/million operations) while the queue topology implements a processing pipeline: `order-intake` forwards messages to `order-processing`, which has duplicate detection and dead-letter forwarding to `dlq-monitor`. Failed messages are automatically routed to a dedicated monitoring queue instead of silently accumulating in individual dead-letter sub-queues. Two topics provide pub/sub channels — `order-events` with ordering guarantees and `audit-trail` with 365-day retention.

## When to Use

- Event-driven microservices needing a multi-stage processing pipeline with automatic forwarding
- Order processing systems where intake, processing, and failure monitoring are separate concerns
- Architectures requiring centralized dead-letter monitoring across multiple queues
- Systems publishing domain events to multiple subscribers with ordering guarantees
- Audit/compliance scenarios requiring long-term event retention (365 days)

## Key Configuration Choices

- **Standard tier** (`sku: Standard`) -- Pay-per-operation pricing with auto-forwarding and topics support. Upgrade to Premium if you need private networking or guaranteed throughput
- **Queue chaining** (`forwardTo: order-processing`) -- Messages arriving in `order-intake` are automatically forwarded to `order-processing` without consumer logic. This decouples intake from processing and enables message buffering
- **Duplicate detection** (`requiresDuplicateDetection: true`) -- The processing queue deduplicates messages to prevent double-processing when upstream services retry
- **Dead-letter forwarding** (`forwardDeadLetteredMessagesTo: dlq-monitor`) -- Failed messages from `order-processing` are automatically routed to `dlq-monitor` for centralized alerting and replay. Without this, dead letters accumulate silently in per-queue sub-queues
- **3 max deliveries** (`maxDeliveryCount: 3`) -- Messages failing 3 times are dead-lettered. Lower than the default of 10 to fail fast in processing pipelines
- **30-day DLQ retention** (`defaultMessageTtl: P30D`) -- The `dlq-monitor` queue retains messages for 30 days, giving operations teams time to investigate and replay
- **Ordered topic** (`supportOrdering: true`) -- The `order-events` topic delivers messages in FIFO order within a session. Critical for event sourcing and read-model projections
- **365-day audit trail** (`defaultMessageTtl: P365D`) -- The `audit-trail` topic retains messages for one year for compliance and forensic analysis

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., "eastus", "westeurope") | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-namespace-name>` | Globally unique namespace name (6-50 chars, letters/numbers/hyphens) | Choose a name; becomes `{name}.servicebus.windows.net` |

## Related Presets

- **01-standard-messaging** -- Use instead for simpler single-queue/single-topic messaging without pipeline topology
- **02-premium-enterprise** -- Use instead for private networking, zone redundancy, and guaranteed throughput with messaging sessions
