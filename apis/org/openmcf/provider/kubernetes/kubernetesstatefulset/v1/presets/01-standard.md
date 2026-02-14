# Standard StatefulSet

This preset deploys a single-replica StatefulSet without persistent volumes. Suitable for stateful applications that need stable pod identities and network names but do not require persistent storage.

## When to Use

- Applications that need stable, unique pod identifiers (e.g., `my-statefulset-0`, `my-statefulset-1`)
- Services that need stable network identity (headless service with predictable DNS names)
- Stateful workloads without persistent disk requirements

## Key Configuration Choices

- **Single replica** (`minReplicas: 1`) -- provides a stable pod identity; increase for multi-instance stateful apps
- **No volume claim templates** -- no persistent storage; add `volumeClaimTemplates` for data persistence
- **No ingress** -- exposed only within the cluster via headless service

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |
| `<your-container-registry>/<your-image>` | Container image repository | Your container registry |
| `<your-image-tag>` | Image tag or version | Your CI/CD pipeline output |

## Related Presets

- **02-with-persistent-volumes** -- StatefulSet with PVC templates for data persistence
