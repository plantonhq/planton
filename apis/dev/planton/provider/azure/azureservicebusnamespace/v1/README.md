# Azure Service Bus Namespace

## Overview

The **AzureServiceBusNamespace** component provisions an Azure Service Bus namespace with optional queues and topics, providing enterprise-grade message brokering for decoupling applications and services.

Azure Service Bus is a fully managed messaging service that supports message queues (point-to-point) and publish-subscribe topics. It offers reliable delivery, temporal decoupling, load leveling, and ordered message processing -- critical capabilities for distributed microservice architectures.

## When to Use This Component

Use AzureServiceBusNamespace when you need:

- **Reliable async messaging** between microservices that cannot afford to lose messages
- **Publish-subscribe patterns** where multiple consumers need independent copies of each message (topics + subscriptions)
- **Ordered processing** with sessions that group related messages for FIFO delivery
- **Dead-letter handling** for messages that fail processing, enabling inspection and retry workflows
- **Enterprise messaging** with Premium tier features like dedicated capacity, zone redundancy, and VNet integration

## SKU Tiers

| Tier | Queues | Topics | Sessions | Duplicate Detection | Max Message Size | SLA |
|------|--------|--------|----------|---------------------|------------------|-----|
| Basic | Yes | No | No | No | 256 KB | 99.9% |
| Standard | Yes | Yes | Yes | Yes | 256 KB | 99.95% |
| Premium | Yes | Yes | Yes | Yes | 100 MB | 99.95% |

**Recommendation**: Use **Standard** for most production workloads. Use **Premium** only when you need dedicated capacity, zone redundancy, VNet integration, or large messages.

## Key Configuration

### Namespace-Level Settings

- **`sku`**: Tier selection (Basic, Standard, Premium). Defaults to Standard.
- **`capacity`**: Premium messaging units (1-16) for dedicated throughput.
- **`zone_redundant`**: Premium-only availability zone redundancy.
- **`minimum_tls_version`**: Defaults to "1.2" for enterprise security.
- **`public_network_access_enabled`**: Control public internet accessibility.

### Queue Settings

- **`lock_duration`**: How long a message is locked during processing (default: 1 minute). Increase for long-running handlers.
- **`max_delivery_count`**: Attempts before dead-lettering (default: 10). Lower for faster poison message detection.
- **`requires_duplicate_detection`**: Prevents duplicate messages based on MessageId.
- **`requires_session`**: Enables ordered, stateful message processing.
- **`forward_to`**: Auto-forward messages to another queue or topic within the namespace.

### Topic Settings

- **`requires_duplicate_detection`**: Prevents duplicate published messages.
- **`support_ordering`**: Preserves message order for session-enabled subscriptions.

## Topic Subscriptions

Topic subscriptions are **deliberately not included** in this component. Subscriptions have a different lifecycle -- they are typically managed by consuming teams, not the infrastructure team that provisions the namespace. This design decision follows DD03 (Composite Bundling Rules).

For subscription management, consumers can use Terraform or the Azure SDK directly, referencing the `topic_ids` output.

## Outputs

| Output | Description |
|--------|-------------|
| `namespace_id` | Azure Resource Manager ID |
| `namespace_name` | Namespace name |
| `endpoint` | Service Bus endpoint URL |
| `primary_connection_string` | SAS connection string for SDK connectivity |
| `primary_key` | SAS key for manual token generation |
| `queue_ids` | Map of queue name to resource ID |
| `topic_ids` | Map of topic name to resource ID |

## Infra Chart Usage

AzureServiceBusNamespace is a **leaf resource** in infra chart DAGs. It is consumed by applications via the `primary_connection_string` injected into app settings or environment variables. No downstream Planton resources reference its outputs.

```yaml
# In a container-apps-environment infra chart:
spec:
  appSettings:
    SERVICE_BUS_CONNECTION_STRING:
      valueFrom:
        kind: AzureServiceBusNamespace
        name: "{{ values.env }}-messaging"
        fieldPath: status.outputs.primary_connection_string
```
