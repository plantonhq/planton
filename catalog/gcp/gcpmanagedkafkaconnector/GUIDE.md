# GcpManagedKafkaConnector Guide

The judgment this guide protects: a connector is one pipeline with one owner. Keep its configuration next to the data it moves, give its service identity exactly the access it needs, and make failures recover on their own.

## Configuration

`configs` is the Kafka Connect configuration, keyed by property. `connector.class` picks the plugin -- `com.google.pubsub.kafka.sink.CloudPubSubSinkConnector` and `...source.CloudPubSubSourceConnector` for Pub/Sub, `com.wepay.kafka.connect.bigquery.BigQuerySinkConnector` for BigQuery, `io.aiven.kafka.connect.gcs.GcsSinkConnector` for Cloud Storage, `org.apache.kafka.connect.mirror.MirrorSourceConnector` for MirrorMaker 2 -- and the plugin's own keys follow. `connector.class` is required. `tasks.max` sets parallelism; `key.converter` and `value.converter` decide how bytes become records. Values are plain text in the manifest, in Planton's stored copy, in the module's input, and on the Connect cluster, so keep credentials out of them wherever the plugin can authenticate with Google IAM instead.

## Identity

Connectors act as the Managed Kafka service agent, not as you. Before a connector runs, grant that agent what it reaches: `roles/pubsub.publisher` on a sink's Pub/Sub topic, `roles/pubsub.subscriber` on a source's subscription, `roles/bigquery.dataEditor` on a BigQuery dataset, object creator on a bucket. A missing grant shows up as failed tasks, not as a failed deploy.

## Failure and restart

A task fails when a destination is briefly unavailable or a record is malformed. `taskRestartPolicy` retries failed tasks with exponential backoff between `minimumBackoff` and `maximumBackoff` (durations such as `60s` and `1800s`; the minimum never above the maximum). A declared policy names at least one of them -- Google's default fills the other. Without it, the modules send no policy and Google's own handling of failed tasks applies. Pair it with the plugin's own error handling (`errors.tolerance`, a dead-letter topic) for records that will never succeed.
