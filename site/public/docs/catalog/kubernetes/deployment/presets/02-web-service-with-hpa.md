---
title: "Production Web Service with HPA"
description: "This preset deploys a production-grade web application with horizontal pod autoscaling, a pod disruption budget, and a zero-downtime rolling update strategy."
type: "preset"
rank: "02"
presetSlug: "02-web-service-with-hpa"
componentSlug: "deployment"
componentTitle: "Deployment"
provider: "kubernetes"
icon: "package"
order: 2
---

# Production Web Service with HPA

This preset deploys a production-grade web application with horizontal pod autoscaling, a pod disruption budget, and a zero-downtime rolling update strategy.

## When to Use

- Production web services that need to scale with traffic
- Services that must remain available during node maintenance and rolling updates
- Applications where zero-downtime deployments are required

## Key Configuration Choices

- **HPA enabled** (`targetCpuUtilizationPercent: 70`) -- scales out when average CPU exceeds 70%; starts with 2 replicas minimum
- **Pod disruption budget** (`minAvailable: 1`) -- at least 1 pod stays running during voluntary disruptions (node drains, upgrades)
- **Rolling update** (`maxUnavailable: 0`, `maxSurge: 1`) -- zero-downtime deployments; a new pod is created before an old one is terminated
- **Higher resource requests** (`100m` CPU, `256Mi` memory) -- production-appropriate baseline that gives HPA meaningful metrics

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace for the deployment | Your namespace management or `KubernetesNamespace` resource |
| `<your-container-registry>/<your-image>` | Container image repository | Your container registry |
| `<your-image-tag>` | Image tag or version | Your CI/CD pipeline output |
| `<your-app.example.com>` | Hostname for ingress access | Your DNS provider |

## Related Presets

- **01-web-service** -- Simpler single-replica deployment without autoscaling
- **03-worker** -- Background worker without ingress
