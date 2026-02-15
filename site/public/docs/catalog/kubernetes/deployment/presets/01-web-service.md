---
title: "Web Service Deployment"
description: "This preset deploys a single-replica web application with an HTTP port and ingress. It is the most common Kubernetes Deployment pattern: a containerized web service exposed via an ingress hostname."
type: "preset"
rank: "01"
presetSlug: "01-web-service"
componentSlug: "deployment"
componentTitle: "Deployment"
provider: "kubernetes"
icon: "package"
order: 1
---

# Web Service Deployment

This preset deploys a single-replica web application with an HTTP port and ingress. It is the most common Kubernetes Deployment pattern: a containerized web service exposed via an ingress hostname.

## When to Use

- Standard web applications or REST APIs
- Services that need external HTTP access via an ingress controller
- Development or low-traffic production services where a single replica is sufficient

## Key Configuration Choices

- **Single replica** (`minReplicas: 1`) -- no autoscaling; suitable for dev or low-traffic services
- **Ingress enabled** -- exposes the service at the specified hostname via the cluster's ingress controller
- **Port 8080 container / port 80 service** -- standard HTTP port mapping; the service port faces clients, the container port faces the app
- **Default resources** -- 50m-1000m CPU, 100Mi-1Gi memory; proto recommended defaults

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace for the deployment | Your namespace management or `KubernetesNamespace` resource |
| `<your-container-registry>/<your-image>` | Container image repository (e.g., `ghcr.io/org/app`) | Your container registry |
| `<your-image-tag>` | Image tag or version (e.g., `v1.2.3`) | Your CI/CD pipeline output |
| `<your-app.example.com>` | Hostname for ingress access | Your DNS provider |

## Related Presets

- **02-web-service-with-hpa** -- Adds horizontal pod autoscaling and pod disruption budget
- **03-worker** -- Background worker without ingress
