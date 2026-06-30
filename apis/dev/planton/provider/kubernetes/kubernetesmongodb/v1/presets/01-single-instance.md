# Single Instance MongoDB

This preset deploys a single-replica MongoDB instance with persistence enabled. Suitable for development, testing, or applications that do not require replica set features.

## When to Use

- Development or staging MongoDB databases
- Applications with light read/write requirements
- Environments where replica set overhead is unnecessary

## Key Configuration Choices

- **Single replica** -- standalone MongoDB without replica set
- **Persistence enabled** with 10Gi disk -- data survives pod restarts
- **Default resources** -- proto recommended defaults

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |

## Related Presets

- **02-replica-set** -- 3-node replica set for production use
