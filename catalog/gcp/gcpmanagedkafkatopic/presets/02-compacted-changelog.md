# Compacted Changelog

## Use Case

A changelog or state topic where only the latest value per key matters: Kafka compacts away older values instead of expiring messages by time, so a new consumer rebuilds current state by reading the topic from the start.

## When to Use

- Entity state (profiles, inventory levels)
- Kafka Streams changelogs and lookup tables
- Change data capture targets keyed by primary key

## What This Creates

- The `customer-profiles.v1` topic with log compaction, keeping each value at least an hour before compacting

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `configs.cleanup.policy` | `compact` | `compact,delete` also expires old keys by retention. |
| `configs.min.compaction.lag.ms` | 1 hour | How long an overwritten value stays readable. |
