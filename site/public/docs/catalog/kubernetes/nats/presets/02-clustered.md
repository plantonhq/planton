---
title: "Clustered NATS with JetStream"
description: "This preset deploys a 3-node NATS cluster with JetStream enabled for production messaging. The cluster provides high availability with automatic leader election for JetStream streams."
type: "preset"
rank: "02"
presetSlug: "02-clustered"
componentSlug: "nats"
componentTitle: "NATS"
provider: "kubernetes"
icon: "package"
order: 2
---

# Clustered NATS with JetStream

This preset deploys a 3-node NATS cluster with JetStream enabled for production messaging. The cluster provides high availability with automatic leader election for JetStream streams.

## When to Use

- Production messaging workloads requiring high availability
- JetStream persistent streams with replication factor > 1
- Event-driven architectures needing fault-tolerant message delivery

## Key Configuration Choices

- **3 server replicas** -- forms a NATS cluster with Raft-based leader election for JetStream
- **10Gi disk** -- per-node JetStream storage; size based on stream retention policies
- **NATS Box disabled** (`true`) -- diagnostic tool not needed in production
- **JetStream enabled** -- persistent streams, consumers, key-value, and object stores

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |

## Related Presets

- **01-single-instance** -- Minimal single-node NATS for development
