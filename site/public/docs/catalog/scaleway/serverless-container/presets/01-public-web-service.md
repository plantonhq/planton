---
title: "Public Web Service Container"
description: "This preset creates a publicly accessible Scaleway Serverless Container running an image from a Scaleway Container Registry. It auto-scales from zero to 20 instances based on incoming HTTP traffic...."
type: "preset"
rank: "01"
presetSlug: "01-public-web-service"
componentSlug: "serverless-container"
componentTitle: "Serverless Container"
provider: "scaleway"
icon: "package"
order: 1
---

# Public Web Service Container

This preset creates a publicly accessible Scaleway Serverless Container running an image from a Scaleway Container Registry. It auto-scales from zero to 20 instances based on incoming HTTP traffic. This is the most common serverless container configuration for APIs and web services.

## When to Use

- Public-facing REST APIs, GraphQL endpoints, or web applications
- Microservices that need to scale to zero when idle (cost optimization)
- Quick deployment of containerized applications without managing infrastructure

## Key Configuration Choices

- **Public privacy** (`privacy: public`) -- the container endpoint is accessible from the internet without authentication; use `private` for internal services
- **Port 8080** (`port: 8080`) -- the port your container listens on; Scaleway routes incoming HTTPS traffic to this port
- **256 MB memory** (`memoryLimitMb: 256`) -- suitable for lightweight services; increase for memory-intensive workloads
- **Auto-scale 0-20** (`maxScale: 20`) -- scales to zero when idle (no cost) and up to 20 instances under load
- **5-minute timeout** (`timeoutSeconds: 300`) -- maximum request processing time; increase for long-running operations
- **Immediate deploy** (`deploy: true`) -- the container is deployed as soon as the resource is created

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-registry-endpoint>` | Scaleway Container Registry endpoint (e.g., `rg.fr-par.scw.cloud/my-namespace`) | Scaleway console or `ScalewayContainerRegistry` status outputs |
| `<your-image-name>` | Container image name in the registry | Your Docker build pipeline |

## Related Presets

- **02-vpc-connected** -- Use instead when the container needs to access Private Network resources (databases, caches)
