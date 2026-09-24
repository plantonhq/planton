# Event Stream

## Use Case

A business event stream: 12 partitions for parallel consumers, a copy in each of three zones, and seven days of retention for replay.

## When to Use

- Domain events (orders, payments, clicks)
- Streams several consumer groups read independently
- Topics that must survive a zone outage

## What This Creates

- The `orders` topic on the `events` cluster with 12 partitions, replication 3, and 7-day retention

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `partitionCount` | `12` | Match your peak consumer parallelism; it can grow later, never shrink. |
| `configs.retention.ms` | 7 days | How far back consumers can replay. |
