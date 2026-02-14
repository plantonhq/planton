# Web Service

This preset deploys a publicly accessible web service with HTTP auto-scaling, health probes, and external ingress. It starts with 1 replica and scales up to 10 based on concurrent HTTP requests. This is the most common pattern for web applications and REST APIs on Azure Container Apps.

## When to Use

- Web applications or REST APIs that need to be publicly accessible
- Services that should always have at least one replica running (no cold starts)
- Standard HTTP workloads with request-based auto-scaling

## Key Configuration Choices

- **1 min replica** (`min_replicas: 1`) -- Avoids cold start latency; at least one instance is always warm
- **10 max replicas** (`max_replicas: 10`) -- Reasonable ceiling for most web services; increase for high-traffic APIs
- **0.5 vCPU / 1 GiB memory** -- Good starting point for most web apps; adjust based on profiling
- **HTTP scale rule** (`concurrent_requests: "100"`) -- Scales up when more than 100 concurrent requests per replica
- **Liveness probe** (`/healthz`) -- Restarts the container if the process becomes unresponsive
- **Readiness probe** (`/ready`) -- Removes the container from load balancing until it's ready to serve

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `container_app_environment_id: ""` | ARM ID of the Container App Environment | Azure portal or `AzureContainerAppEnvironment` status outputs |
| `resource_group: my-rg` | Resource group name | Azure portal or `AzureResourceGroup` status outputs |
| `image: mcr.microsoft.com/...` | Your application container image | Your container registry (ACR, GHCR, Docker Hub) |

## Related Presets

- **02-background-worker** -- Use instead for queue-processing workers with no ingress and scale-to-zero
- **03-enterprise-api** -- Use instead for production APIs with identity, IP restrictions, and Key Vault secrets
