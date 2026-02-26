# Standard Messaging Service Bus

This preset creates an Azure Service Bus namespace on the Standard tier with a single queue and topic — the fastest path to reliable messaging for most applications. The Standard tier provides shared infrastructure with pay-per-operation pricing (~$0.05/million operations), 256 KB max message size, 80 GB per-entity storage, and topics/subscriptions support. The `orders` queue is configured with dead-letter-on-expiration to prevent silent message loss, and the `domain-events` topic provides a publish-subscribe channel for event distribution.

## When to Use

- Microservices that need reliable asynchronous communication between components
- Order processing, task queues, or work distribution where at-least-once delivery is sufficient
- Event distribution across multiple consumers via topics and subscriptions
- Applications that don't require private networking, messaging sessions, or message deduplication

## Key Configuration Choices

- **Standard tier** (`sku: Standard`) -- Shared infrastructure with pay-per-operation pricing. Supports queues, topics, subscriptions, and auto-forwarding. No messaging units to manage
- **TLS 1.2** (`minimumTlsVersion: "1.2"`) -- Enforces modern TLS for all AMQP and HTTPS connections
- **Public access** (`publicNetworkAccessEnabled: true`) -- Namespace is accessible from the internet via connection string. Add firewall rules or upgrade to Premium for VNet injection
- **orders queue** -- Dead lettering enabled (`deadLetteringOnMessageExpiration: true`) with default max delivery count of 10. Messages that expire or exceed retries are moved to the dead-letter sub-queue for inspection
- **domain-events topic** -- A minimal topic for pub/sub event distribution. Add subscriptions with SQL filter rules to route events to specific consumers

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., "eastus", "westeurope") | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-namespace-name>` | Globally unique namespace name (6-50 chars, letters/numbers/hyphens) | Choose a name; becomes `{name}.servicebus.windows.net` |

## Related Presets

- **02-premium-enterprise** -- Use instead for private networking, zone redundancy, messaging sessions, and guaranteed throughput
- **03-event-driven-microservices** -- Use instead for event-driven architectures with queue chaining and dead-letter forwarding
