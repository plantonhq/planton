# Azure Service Bus Namespace - Examples

## Minimal Standard Namespace with a Queue

The simplest configuration: a Standard-tier namespace with one queue. Suitable for point-to-point messaging between two services.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureServiceBusNamespace
metadata:
  name: orders-messaging
spec:
  region: eastus
  resourceGroup:
    value: "production-rg"
  name: myapp-orders-sb
  queues:
    - name: order-processing
```

## Standard Namespace with Queues and Topics

A Standard-tier namespace with both point-to-point queues and publish-subscribe topics. The `events` topic enables fan-out to multiple consumers.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureServiceBusNamespace
metadata:
  name: app-messaging
  org: mycompany
  env: production
spec:
  region: westus2
  resourceGroup:
    value: "prod-messaging-rg"
  name: myapp-messaging-sb
  queues:
    - name: order-processing
      maxDeliveryCount: 5
      deadLetteringOnMessageExpiration: true
      defaultMessageTtl: "P14D"
    - name: email-notifications
      lockDuration: "PT30S"
  topics:
    - name: domain-events
    - name: audit-logs
      defaultMessageTtl: "P90D"
```

## Production Queue with Duplicate Detection and Sessions

A queue configured for idempotent, ordered processing. Sessions group related messages (e.g., all messages for an order) and guarantee FIFO delivery within each session.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureServiceBusNamespace
metadata:
  name: payment-messaging
spec:
  region: eastus
  resourceGroup:
    value: "payments-rg"
  name: payments-sb
  queues:
    - name: payment-processing
      requiresDuplicateDetection: true
      requiresSession: true
      lockDuration: "PT5M"
      maxDeliveryCount: 3
      deadLetteringOnMessageExpiration: true
      defaultMessageTtl: "P7D"
    - name: payment-dlq-monitor
    - name: payment-retries
      forwardDeadLetteredMessagesTo: "payment-dlq-monitor"
```

## Premium Namespace with Zone Redundancy

Enterprise-grade namespace with Premium tier for dedicated capacity, zone redundancy, and private networking.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureServiceBusNamespace
metadata:
  name: enterprise-messaging
  org: enterprise-corp
  env: production
spec:
  region: westeurope
  resourceGroup:
    value: "enterprise-messaging-rg"
  name: enterprise-sb-prod
  sku: Premium
  capacity: 4
  premiumMessagingPartitions: 2
  zoneRedundant: true
  publicNetworkAccessEnabled: false
  queues:
    - name: high-throughput-ingest
      partitioningEnabled: true
    - name: command-processing
      requiresSession: true
      lockDuration: "PT5M"
  topics:
    - name: integration-events
      partitioningEnabled: true
    - name: compliance-audit
      defaultMessageTtl: "P365D"
      requiresDuplicateDetection: true
```

## Infra Chart valueFrom Reference

When used within an infra chart, the namespace references its resource group from an upstream AzureResourceGroup component via `valueFrom`.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureServiceBusNamespace
metadata:
  name: "{{ values.env }}-messaging"
spec:
  region: "{{ values.region }}"
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: "{{ values.env }}-rg"
      fieldPath: status.outputs.resource_group_name
  name: "{{ values.app_name }}-{{ values.env }}-sb"
  queues:
    - name: "{{ values.queue_name }}"
      maxDeliveryCount: 5
      deadLetteringOnMessageExpiration: true
```

## Message Forwarding Pattern

A routing pattern where messages flow through a pipeline: incoming messages are forwarded from an intake queue to a processing queue, and dead-lettered messages are forwarded to a monitoring queue.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureServiceBusNamespace
metadata:
  name: routing-messaging
spec:
  region: eastus
  resourceGroup:
    value: "routing-rg"
  name: message-routing-sb
  queues:
    - name: intake
      forwardTo: "processing"
    - name: processing
      maxDeliveryCount: 3
      forwardDeadLetteredMessagesTo: "dlq-monitor"
    - name: dlq-monitor
      defaultMessageTtl: "P30D"
```
