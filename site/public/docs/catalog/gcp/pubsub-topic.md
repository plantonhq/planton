---
title: "Pub/Sub Topic"
description: "Deploy GCP Pub/Sub topics using OpenMCF"
---

# GcpPubSubTopic

Provision and manage Pub/Sub topics -- the named channels that receive messages from publishers and fan them out to subscriptions for event-driven and streaming architectures.

## Overview

GcpPubSubTopic creates a Pub/Sub topic within a GCP project. Topics decouple
message producers from consumers and support one-to-many delivery. This resource
manages the infrastructure boundary: encryption (Google-managed or CMEK),
regional storage constraints, message retention, schema validation, and
ingestion from external sources. Subscriptions, which control delivery to
consumers, are managed separately.

## Quick Start

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpPubSubTopic
metadata:
  name: my-topic
spec:
  projectId:
    value: "my-gcp-project"
  topicName: order-events
```

## Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `projectId` | StringValueOrRef | Yes | GCP project ID |
| `topicName` | string | Yes | Topic name (3-255 chars, starts with letter, immutable) |
| `kmsKeyName` | StringValueOrRef | No | CMEK encryption key (full resource name) |
| `messageRetentionDuration` | string | No | Retain messages for this duration (600s-2678400s) |
| `messageStoragePolicy` | object | No | Regional storage constraints |
| `schemaSettings` | object | No | Schema validation settings |
| `ingestionDataSourceSettings` | object | No | External data source ingestion (Kinesis, MSK, Event Hubs, GCS, Confluent) |

## Outputs

| Output | Description |
|--------|-------------|
| `topic_id` | Fully qualified topic ID (`projects/{project}/topics/{name}`) |
| `topic_name` | Short topic name (same as input) |

## Important

**Topic name is immutable.** Changing `topicName` destroys and recreates the
topic along with all its subscriptions.

**CMEK requires IAM setup.** Grant `roles/cloudkms.cryptoKeyEncrypterDecrypter`
to the Pub/Sub service account on the KMS key before creating the topic.

**Message retention enables replay.** When `messageRetentionDuration` is set,
any subscription can seek to a timestamp within the retention window,
independent of subscription-level retention.

**Ingestion sources are one-per-topic.** Configure at most one external
ingestion source (Kinesis, MSK, Event Hubs, Cloud Storage, or Confluent Cloud)
per topic.

## Related

- [GcpKmsKey](/docs/catalog/gcp/kms-key) -- CMEK encryption key
- [GcpProject](/docs/catalog/gcp/project) -- Parent GCP project
- [GcpGcsBucket](/docs/catalog/gcp/gcs-bucket) -- Source bucket for Cloud Storage ingestion
