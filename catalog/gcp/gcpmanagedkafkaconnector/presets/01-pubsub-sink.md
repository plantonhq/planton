# Pub/Sub Sink

## Use Case

Fan a Kafka topic out to Pub/Sub subscribers with Google's Pub/Sub sink connector: every message on `orders` is published to the `orders-events` topic.

## When to Use

- Bridging Kafka producers to Pub/Sub consumers (Cloud Run, Dataflow, push subscriptions)
- Gradual migrations between the two systems

## What This Creates

- A Pub/Sub sink connector with three tasks and a restart policy

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `configs.topics` | `orders` | The Kafka topics to read. |
| `configs.cps.topic` | `orders-events` | Grant the Managed Kafka service agent `roles/pubsub.publisher` on it. |
