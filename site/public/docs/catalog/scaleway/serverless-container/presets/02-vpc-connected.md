---
title: "VPC-Connected Private Container"
description: "This preset creates a Scaleway Serverless Container attached to a Private Network with private privacy, a health check, and a minimum scale of 1 to avoid cold starts. This is the standard..."
type: "preset"
rank: "02"
presetSlug: "02-vpc-connected"
componentSlug: "serverless-container"
componentTitle: "Serverless Container"
provider: "scaleway"
icon: "package"
order: 2
---

# VPC-Connected Private Container

This preset creates a Scaleway Serverless Container attached to a Private Network with private privacy, a health check, and a minimum scale of 1 to avoid cold starts. This is the standard configuration for backend services that need to access databases, caches, or other Private Network resources.

## When to Use

- Backend services that connect to RDB, MongoDB, or Redis instances on a Private Network
- Internal APIs that should not be directly internet-accessible
- Services requiring consistent low latency (minimum 1 instance always warm)

## Key Configuration Choices

- **Private privacy** (`privacy: private`) -- the container endpoint requires authentication via Scaleway IAM; not accessible from the public internet
- **VPC connected** (`privateNetworkId`) -- the container can reach Private Network resources (databases, caches) using private IPs
- **512 MB memory** (`memoryLimitMb: 512`) -- increased for services maintaining database connection pools
- **Minimum 1 instance** (`minScale: 1`) -- prevents cold starts; at least one instance is always running
- **Health check** (`healthCheck`) -- Scaleway probes `/health` every 10 seconds; unhealthy instances are replaced
- **Scale up to 10** (`maxScale: 10`) -- controlled maximum to prevent runaway scaling and excessive database connections

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-registry-endpoint>` | Scaleway Container Registry endpoint (e.g., `rg.fr-par.scw.cloud/my-namespace`) | Scaleway console or `ScalewayContainerRegistry` status outputs |
| `<your-image-name>` | Container image name in the registry | Your Docker build pipeline |
| `<your-private-network-id>` | UUID of the Private Network to connect to | Scaleway console or `ScalewayPrivateNetwork` status outputs |

## Related Presets

- **01-public-web-service** -- Use instead for public-facing APIs that do not need Private Network access
