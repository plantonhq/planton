# Azure Event Hub Namespace - Examples

## Minimal Standard Namespace

The simplest configuration: a Standard-tier namespace with one event hub. Suitable for a single event stream with moderate throughput.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureEventHubNamespace
metadata:
  name: telemetry-events
spec:
  region: eastus
  resourceGroup:
    value: "production-rg"
  name: myapp-telemetry-eh
  eventHubs:
    - name: app-events
      partitionCount: 4
      messageRetention: 1
```

## Standard with Multiple Event Hubs and Consumer Groups

A Standard-tier namespace with two event hubs, each with dedicated consumer groups for independent processing pipelines.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureEventHubNamespace
metadata:
  name: app-streaming
  org: mycompany
  env: production
spec:
  region: westus2
  resourceGroup:
    value: "prod-streaming-rg"
  name: myapp-streaming-eh
  eventHubs:
    - name: order-events
      partitionCount: 8
      messageRetention: 3
      consumerGroups:
        - name: analytics-pipeline
          userMetadata: "Real-time analytics processor"
        - name: search-indexer
          userMetadata: "Elasticsearch indexing consumer"
    - name: audit-trail
      partitionCount: 4
      messageRetention: 7
      consumerGroups:
        - name: compliance-archiver
          userMetadata: "Long-term compliance storage"
```

## Production IoT Ingestion

A high-throughput namespace for IoT telemetry ingestion with maximum partitions and multiple consumer groups for different pipeline stages.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureEventHubNamespace
metadata:
  name: iot-ingestion
  org: manufacturing-corp
  env: production
spec:
  region: eastus
  resourceGroup:
    value: "iot-platform-rg"
  name: iot-telemetry-eh
  capacity: 4
  eventHubs:
    - name: device-telemetry
      partitionCount: 32
      messageRetention: 7
      consumerGroups:
        - name: hot-path
          userMetadata: "Real-time Stream Analytics for alerting"
        - name: warm-path
          userMetadata: "Spark Structured Streaming for aggregation"
        - name: cold-path
          userMetadata: "Batch export to Data Lake for historical analysis"
        - name: device-health
          userMetadata: "Device connectivity and health monitoring"
```

## Premium Enterprise with Zone Redundancy

Enterprise-grade namespace with Premium tier for dedicated capacity, zone redundancy, and private networking. Multiple event hubs with consumer groups for different domains.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureEventHubNamespace
metadata:
  name: enterprise-streaming
  org: enterprise-corp
  env: production
spec:
  region: westeurope
  resourceGroup:
    value: "enterprise-streaming-rg"
  name: enterprise-events-eh
  sku: Premium
  capacity: 2
  zoneRedundant: true
  minimumTlsVersion: "1.2"
  publicNetworkAccessEnabled: false
  eventHubs:
    - name: domain-events
      partitionCount: 16
      messageRetention: 7
      consumerGroups:
        - name: saga-orchestrator
          userMetadata: "Distributed transaction saga processor"
        - name: projection-builder
          userMetadata: "CQRS read-model projections"
    - name: integration-events
      partitionCount: 8
      messageRetention: 3
      consumerGroups:
        - name: partner-sync
          userMetadata: "External partner data synchronization"
```

## Infra Chart valueFrom Reference

When used within an infra chart, the namespace references its resource group from an upstream AzureResourceGroup component via `valueFrom`.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureEventHubNamespace
metadata:
  name: "{{ values.env }}-events"
spec:
  region: "{{ values.region }}"
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: "{{ values.env }}-rg"
      fieldPath: status.outputs.resource_group_name
  name: "{{ values.app_name }}-{{ values.env }}-eh"
  eventHubs:
    - name: "{{ values.event_hub_name }}"
      partitionCount: 8
      messageRetention: 3
      consumerGroups:
        - name: "{{ values.consumer_name }}"
```

## Auto-Inflate Standard

A Standard-tier namespace with auto-inflate enabled for elastic throughput scaling. Starts with 2 throughput units and scales up to 20 based on traffic demand.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureEventHubNamespace
metadata:
  name: elastic-streaming
spec:
  region: centralus
  resourceGroup:
    value: "streaming-rg"
  name: elastic-events-eh
  capacity: 2
  autoInflateEnabled: true
  maximumThroughputUnits: 20
  eventHubs:
    - name: click-stream
      partitionCount: 16
      messageRetention: 3
      consumerGroups:
        - name: realtime-dashboard
          userMetadata: "Live metrics dashboard consumer"
        - name: batch-export
          userMetadata: "Hourly export to blob storage"
    - name: page-views
      partitionCount: 8
      messageRetention: 1
```
