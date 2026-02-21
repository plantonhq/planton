---
title: "Public Kafka-Compatible"
description: "This preset creates an OCI Stream Pool with Kafka compatibility settings and two pre-defined streams: an `events` stream for application event sourcing and a `commands` stream for command-driven..."
type: "preset"
rank: "01"
presetSlug: "01-public-kafka-compatible"
componentSlug: "stream-pool"
componentTitle: "Stream Pool"
provider: "oci"
icon: "package"
order: 1
---

# Public Kafka-Compatible

This preset creates an OCI Stream Pool with Kafka compatibility settings and two pre-defined streams: an `events` stream for application event sourcing and a `commands` stream for command-driven architectures. The pool is publicly accessible (via OCI service endpoints) and uses Oracle-managed encryption, making it the fastest path to a production-ready Kafka-compatible event bus on OCI.

## When to Use

- Application event streaming where producers and consumers use the Apache Kafka protocol
- Event-driven architectures using CQRS/event sourcing patterns (the two streams map to the events and commands channels)
- Migrating from self-managed Kafka to a managed service without changing application code
- Development and staging environments where private networking is not required

## Key Configuration Choices

- **Public access** (no `privateEndpointSettings`) -- the stream pool is accessible via OCI's public Streaming service endpoints. Kafka clients connect using OCI credentials (SASL/PLAIN with auth token). This is the simplest configuration. For production workloads that require VCN-level network isolation, use the private encrypted preset instead.
- **Kafka auto-create topics** (`autoCreateTopicsEnable: true`) -- when a Kafka producer publishes to a topic that does not exist as a named stream, OCI auto-creates it with the default partition count. This is convenient during development but should be disabled in production to prevent accidental topic proliferation.
- **48-hour default retention** (`logRetentionHours: 48`) -- auto-created topics retain messages for 48 hours. This gives consumers a 2-day window to process messages, accommodating weekend and maintenance gaps. The valid range is 24-168 hours.
- **3 default partitions** (`numPartitions: 3`) -- auto-created topics get 3 partitions. This balances consumer parallelism with partition overhead. Named streams below override this default.
- **Events stream** (5 partitions, 48h retention) -- higher partition count for the events stream enables greater consumer parallelism for event-sourced reads. 48-hour retention matches the pool default.
- **Commands stream** (3 partitions, 24h retention) -- fewer partitions since command throughput is typically lower than event throughput. 24-hour retention is sufficient since commands should be processed promptly.
- **Oracle-managed encryption** (no `kmsKeyId`) -- uses OCI's default encryption at rest. Add a `kmsKeyId` reference to an `OciKmsKey` if customer-managed encryption is required.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the stream pool will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |

## Related Presets

- **02-private-encrypted** -- Use instead for production workloads requiring VCN-level network isolation and customer-managed KMS encryption
