# GCP Event Pipeline

Provisions an event-driven data ingestion pipeline that streams events from Pub/Sub directly into BigQuery using native BigQuery subscription delivery. Includes a dead-letter topic for unprocessable messages and a GCS bucket for overflow storage.

This is a pure infrastructure pipeline -- no application code required. Publishers send events to the Pub/Sub topic, and Pub/Sub's built-in BigQuery delivery writes them directly to a BigQuery table. For events that need transformation before landing in BigQuery, add a Cloud Function or Dataflow job after deploying this chart.

## Architecture

```
  Publishers
      │
      ▼
┌──────────────┐
│ GcpPubSubTopic│
│ (events-ingest)│
└──────┬───────┘
       │
       ▼
┌──────────────────────────┐     ┌──────────────────┐
│ GcpPubSubSubscription    │────▶│ GcpBigQueryDataset│
│ (BigQuery delivery)      │     │ (events)         │
│                          │     └──────────────────┘
│ Dead-letter policy ──────│──┐
└──────────────────────────┘  │
                              ▼
                    ┌──────────────────┐
                    │ GcpPubSubTopic   │
                    │ (dead-letter)    │
                    └──────────────────┘

┌──────────────────┐  ┌──────────────────────┐
│ GcpGcsBucket     │  │ GcpServiceAccount    │
│ (dead-letter/    │  │ (bigquery.dataEditor │
│  overflow)       │  │  + pubsub.subscriber)│
└──────────────────┘  └──────────────────────┘
```

## Dependency Graph

```
Layer 0 (parallel):  GcpPubSubTopic (main), GcpPubSubTopic (dead-letter),
                     GcpBigQueryDataset, GcpGcsBucket, GcpServiceAccount
Layer 1 (dep all):   GcpPubSubSubscription
```

## Included Cloud Resources

| Resource | Kind | Group | Purpose |
|----------|------|-------|---------|
| Pub/Sub Topic | `GcpPubSubTopic` | messaging | Event ingestion endpoint |
| Dead-Letter Topic | `GcpPubSubTopic` | messaging | Receives unprocessable messages (optional) |
| Pub/Sub Subscription | `GcpPubSubSubscription` | messaging | Streams events to BigQuery |
| BigQuery Dataset | `GcpBigQueryDataset` | storage | Event data warehouse |
| GCS Bucket | `GcpGcsBucket` | storage | Dead-letter and overflow storage |
| Service Account | `GcpServiceAccount` | identity | Pub/Sub delivery identity |

## Parameters

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `gcp_project_id` | GCP project ID | `my-gcp-project` | Yes |
| `region` | GCP region | `us-central1` | Yes |
| `topic_name` | Pub/Sub topic for events | `events-ingest` | Yes |
| `subscription_name` | Subscription name | `events-to-bigquery` | Yes |
| `bigquery_table` | Target BigQuery table (`project.dataset.table`) | `my-gcp-project.events.raw_events` | Yes |
| `write_metadata` | Include Pub/Sub metadata in BigQuery | `true` | No |
| `dataset_id` | BigQuery dataset ID | `events` | Yes |
| `bucket_name` | GCS bucket for dead-letter storage | `my-project-events-deadletter` | Yes |
| `deadLetterEnabled` | Enable dead-letter topic | `true` | No |
| `dead_letter_topic_name` | Dead-letter topic name | `events-deadletter` | No |
| `max_delivery_attempts` | Attempts before dead-lettering (5-100) | `10` | No |
| `service_account_id` | Service account ID | `events-pipeline-sa` | Yes |

## How It Works

1. **Publish**: Applications publish events to the Pub/Sub topic using client libraries or REST API
2. **Deliver**: Pub/Sub's BigQuery subscription writes events directly to the BigQuery table as rows
3. **Dead-letter**: Messages that fail delivery after `max_delivery_attempts` are forwarded to the dead-letter topic
4. **Monitor**: Use the GCS bucket to archive dead-letter messages or overflow events

## Pre-requisites

The **BigQuery table** specified in `bigquery_table` must exist before deploying this chart. The chart creates the dataset but not individual tables. Create the target table with a schema that matches your event payload:

```sql
CREATE TABLE `my-gcp-project.events.raw_events` (
  data STRING,
  subscription_name STRING,
  message_id STRING,
  publish_time TIMESTAMP,
  attributes JSON,
  ordering_key STRING
);
```

When `write_metadata` is `true`, the subscription writes `subscription_name`, `message_id`, `publish_time`, `attributes`, and `ordering_key` as additional columns. Include these in your table schema.

## Adding Transformation Logic

This chart provides the raw ingestion pipeline. For event transformation, add one of these after deployment:

- **Cloud Function**: Deploy a `GcpCloudFunction` triggered by the Pub/Sub topic for lightweight transformations
- **Dataflow**: Use a Dataflow streaming job for complex transformations, windowing, and enrichment
- **Cloud Run**: Deploy a push subscription to a Cloud Run service for container-based processing

## Important Notes

- The `subscription_name` and BigQuery delivery configuration are **immutable** after creation.
- The BigQuery table must exist before the subscription can deliver messages. The chart creates the **dataset** (container) only.
- Pub/Sub's BigQuery delivery uses streaming inserts, which have [pricing implications](https://cloud.google.com/bigquery/pricing#streaming_pricing) separate from standard BigQuery storage.
- The dead-letter topic requires the Pub/Sub service agent to have `roles/pubsub.publisher` on the dead-letter topic. GCP typically handles this automatically.
