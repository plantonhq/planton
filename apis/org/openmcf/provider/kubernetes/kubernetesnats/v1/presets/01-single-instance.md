# Single Instance NATS with JetStream

This preset deploys a single-node NATS server with JetStream enabled and the NATS Box diagnostic tool. JetStream provides persistent messaging, key-value store, and object store capabilities.

## When to Use

- Development or testing environments needing a lightweight message broker
- Applications using NATS core pub/sub or JetStream persistence
- Single-node setups where clustering overhead is unnecessary

## Key Configuration Choices

- **JetStream enabled** (`disableJetStream: false`) -- persistent messaging, streams, and consumers
- **NATS Box** enabled -- deploys a diagnostic pod for testing NATS connectivity
- **1Gi disk** -- JetStream storage for stream data; increase for high-throughput streams

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |

## Related Presets

- **02-clustered** -- 3-node NATS cluster for production high availability
