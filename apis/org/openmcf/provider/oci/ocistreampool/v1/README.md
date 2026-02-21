# OciStreamPool

## Overview

OciStreamPool is an OpenMCF component that deploys an OCI Streaming stream pool with bundled streams. It provides a single declarative manifest to create a Kafka-compatible event-streaming endpoint with configurable Kafka settings, encryption, private networking, and stream definitions.

## Purpose

OCI Streaming is a Kafka-compatible managed event-streaming service for real-time data ingestion, processing, and distribution. The stream pool is the top-level grouping that provides shared Kafka settings, encryption, and networking for all streams within it. This component bundles the pool and its streams so that the entire streaming topology is declared in one manifest.

## Key Features

- **Kafka compatibility** — configurable auto-topic creation, log retention, and default partitions for Kafka clients.
- **Bundled streams** — streams are declared inline and created as sub-resources of the pool.
- **Customer-managed encryption** — optional KMS key for encrypting all streams in the pool. Updatable.
- **Private networking** — optional private endpoint in a subnet with NSG bindings for VCN-only access.
- **Foreign key references** — `compartmentId`, `kmsKeyId`, private endpoint `subnetId`, and `nsgIds` support `valueFrom`.

## Constraints

- `privateEndpointSettings` is entirely ForceNew — any change forces pool recreation.
- Stream `name`, `partitions`, and `retentionInHours` are ForceNew — changes force stream recreation.
- `logRetentionHours` and stream `retentionInHours` must be between 24 and 168 hours.
- Stream `partitions` must be >= 1.

## Use Cases

| Scenario | Configuration |
|----------|---------------|
| Development event bus | Single stream with 1 partition, default retention |
| Production event hub | Multiple streams with high partitions and 7-day retention |
| Kafka migration | Auto-topic creation enabled for Kafka producer compatibility |
| Secure data pipeline | Private endpoint + KMS encryption |
| Multi-tenant streaming | Separate streams per tenant within one pool |

## Production Features

- **Freeform tags** — automatically populated from `metadata.labels` on both the pool and all streams.
- **KMS encryption** — customer-managed keys for encryption at rest across all streams in the pool.
- **Private endpoints** — restricts access to streams from within the specified subnet only.
- **Kafka bootstrap servers** — output provides the connection string for Kafka clients.
