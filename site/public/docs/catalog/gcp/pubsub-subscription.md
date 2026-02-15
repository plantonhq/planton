---
title: "Pub/Sub Subscription"
description: "Deploy GCP Pub/Sub subscriptions using OpenMCF"
---

# GcpPubSubSubscription

Provision and manage Pub/Sub subscriptions -- the named resources that receive messages from a topic and deliver them to subscribing applications via pull, push, BigQuery, or Cloud Storage.

## Overview

GcpPubSubSubscription creates a subscription attached to a Pub/Sub topic within a GCP
project. Subscriptions define the delivery method and configuration for consuming
messages. This resource supports all four GCP delivery methods: pull (default), push
(HTTP POST to an endpoint), BigQuery (streaming writes), and Cloud Storage (batched
object writes). Only one delivery method can be active per subscription.

## Quick Start

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpPubSubSubscription
metadata:
  name: my-subscription
spec:
  projectId:
    value: "my-gcp-project"
  subscriptionName: order-events-sub
  topic:
    value: "projects/my-gcp-project/topics/order-events"
```

## Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `projectId` | StringValueOrRef | Yes | GCP project ID |
| `subscriptionName` | string | Yes | Subscription name (3-255 chars, immutable) |
| `topic` | StringValueOrRef | Yes | Source topic (fully qualified or short name) |
| `ackDeadlineSeconds` | int32 | No | Ack deadline: 10-600s (default 10) |
| `messageRetentionDuration` | string | No | Backlog retention: 600s-2678400s (default 7d) |
| `retainAckedMessages` | bool | No | Keep acked messages for replay |
| `expirationPolicy` | object | No | Auto-delete inactive subscriptions |
| `filter` | string | No | Attribute filter (max 256 bytes, immutable) |
| `enableMessageOrdering` | bool | No | FIFO by ordering key (immutable) |
| `enableExactlyOnceDelivery` | bool | No | Exactly-once guarantees |
| `deadLetterPolicy` | object | No | Dead-letter topic and max attempts |
| `retryPolicy` | object | No | Backoff between retries |
| `pushConfig` | object | No | Push delivery to HTTPS endpoint |
| `bigqueryConfig` | object | No | BigQuery streaming delivery |
| `cloudStorageConfig` | object | No | Cloud Storage batch delivery |

## Outputs

| Output | Description |
|--------|-------------|
| `subscription_id` | Fully qualified ID (`projects/{project}/subscriptions/{name}`) |
| `subscription_name` | Short subscription name (same as input) |

## Delivery Methods

**Pull** (default): No delivery config needed. Consumers call the API to receive messages.

**Push**: Set `pushConfig` with an HTTPS endpoint. Supports OIDC authentication and unwrapped payloads.

**BigQuery**: Set `bigqueryConfig` with a target table. Messages are streamed as rows.

**Cloud Storage**: Set `cloudStorageConfig` with a target bucket. Messages are batched into objects.

Only one method can be active -- they are mutually exclusive.

## Important

**Subscription name is immutable.** Changing `subscriptionName` destroys and recreates
the subscription.

**Filter is immutable.** Once set, the filter expression cannot be changed.

**Delivery methods are mutually exclusive.** Setting more than one of `pushConfig`,
`bigqueryConfig`, or `cloudStorageConfig` is a validation error.

**Dead-letter requires IAM.** Grant Subscriber on this subscription and Publisher on
the dead-letter topic to the Pub/Sub service account.

**Exactly-once is per-subscription.** Does not prevent publisher-side duplicates.

## Related

- [GcpPubSubTopic](/docs/catalog/gcp/pubsub-topic) -- Source topic
- [GcpGcsBucket](/docs/catalog/gcp/gcs-bucket) -- Cloud Storage delivery target
- [GcpProject](/docs/catalog/gcp/project) -- Parent GCP project
