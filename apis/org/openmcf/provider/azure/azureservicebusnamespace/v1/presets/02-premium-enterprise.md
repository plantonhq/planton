# Premium Enterprise Service Bus

This preset creates an Azure Service Bus namespace on the Premium tier with zone redundancy, private networking, and advanced messaging features. Premium tier provides dedicated resources (1 messaging unit ≈ ~$668/month), guaranteed throughput, 100 MB max message size, and enterprise features including messaging sessions, duplicate detection, and VNet isolation. The `command-processing` queue uses sessions for ordered processing with a 5-minute lock duration, and the `integration-events` topic enables cross-system event distribution with duplicate detection.

## When to Use

- Production workloads requiring guaranteed throughput and latency SLAs (99.95% uptime)
- Enterprise environments mandating private networking with no public internet exposure
- CQRS/command patterns requiring messaging sessions for ordered, exclusive processing per session ID
- Integration scenarios between internal services where duplicate detection prevents reprocessing
- Compliance-driven environments (PCI-DSS, HIPAA, SOC 2) requiring network isolation and zone redundancy

## Key Configuration Choices

- **Premium tier** (`sku: Premium`) -- Dedicated resources, 1 messaging unit (~$668/month). Scale by adding messaging units (2, 4, 8, 16) for higher throughput
- **1 messaging unit** (`capacity: 1`) -- Each MU provides ~1000 messages/sec throughput. Monitor namespace CPU/memory to determine when to scale
- **Zone redundant** (`zoneRedundant: true`) -- Replicates metadata and data across three Azure availability zones for fault tolerance
- **Private networking** (`publicNetworkAccessEnabled: false`) -- Namespace is only accessible via private endpoints or VNet service endpoints. Requires `AzurePrivateEndpoint` + `AzurePrivateDnsZone` for connectivity
- **Session-enabled queue** (`requiresSession: true`) -- Messages are grouped by session ID and processed in order by a single consumer at a time. Essential for CQRS command processing and per-entity workflows
- **5-minute lock** (`lockDuration: PT5M`) -- Consumer has 5 minutes to process and complete a message before it becomes available to other consumers
- **Duplicate detection topic** (`requiresDuplicateDetection: true`) -- Automatically deduplicates messages within a configurable time window (default 10 minutes)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., "eastus", "westeurope") | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-namespace-name>` | Globally unique namespace name (6-50 chars, letters/numbers/hyphens) | Choose a name; becomes `{name}.servicebus.windows.net` |

## Related Presets

- **01-standard-messaging** -- Use instead for cost-effective messaging without private networking or session requirements
- **03-event-driven-microservices** -- Use instead for event-driven architectures with queue chaining and dead-letter forwarding on Standard tier
