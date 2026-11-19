# Background Worker

This preset deploys a queue consumer or background processor: no ports, no Service — just the container, its environment, and a secret pulled from an existing Kubernetes Secret. When the app container exposes no ports, the module skips Service creation entirely.

## When to Use

- Queue consumers, schedulers, stream processors
- Any long-running process that receives work over a connection it initiates (never inbound traffic)

## Key Configuration Choices

- **No ports** — no Service is created; the workload is unreachable by design
- **Env split** — plain configuration under `variables`, sensitive values under `secrets` referencing an existing Kubernetes Secret by name and key
- **One pod to start** — `availability.replicas` is unset, so the worker starts with one pod; set it to hold a fixed size, or leave it unset and point an autoscaler at the Deployment (KEDA on queue depth is the usual worker lane; CPU-based HPA works too when work is CPU-bound)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your `KubernetesNamespace` resource |
| `<your-container-registry>/<your-image>` | Container image repository | Your container registry |
| `<your-image-tag>` | Image tag or version | Your CI/CD pipeline output |
| `<your-existing-secret>` / `<your-secret-key>` | Kubernetes Secret name and key holding the credential | Your `KubernetesSecret` resource or cluster |

## Related Presets

- **01-web-service** — HTTP service with a fronting Service
- **02-web-service-with-hpa** — Autoscaled production web service
