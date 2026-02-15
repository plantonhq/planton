---
title: "Professional Container Registry"
description: "This preset creates a DigitalOcean Container Registry (DOCR) with the professional tier and garbage collection enabled. The professional tier provides the highest storage limits and is suitable for..."
type: "preset"
rank: "01"
presetSlug: "01-professional"
componentSlug: "container-registry"
componentTitle: "Container Registry"
provider: "digitalocean"
icon: "package"
order: 1
---

# Professional Container Registry

This preset creates a DigitalOcean Container Registry (DOCR) with the professional tier and garbage collection enabled. The professional tier provides the highest storage limits and is suitable for production teams pushing many images. Garbage collection removes untagged images to control storage costs.

## When to Use

- Production teams pushing container images for Kubernetes or App Platform
- CI/CD pipelines building and pushing images frequently
- Need for larger storage than starter/basic tiers
- Desire to automatically clean untagged images

## Key Configuration Choices

- **Professional tier** (`subscriptionTier: professional`) -- highest storage and bandwidth limits; production-ready.
- **Garbage collection** (`garbageCollectionEnabled: true`) -- automatically deletes untagged images; reduces storage costs.
- **Region** (`region: nyc3`) -- registry data location; choose nearest to your DOKS clusters or pipelines.
- **Registry name** (`name`) -- must be unique in your account; used in image paths (`registry.digitalocean.com/<name>/<image>:<tag>`).

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `nyc3` | Target DigitalOcean region slug | [DigitalOcean Regions API](https://docs.digitalocean.com/reference/api/api-reference/#tag/Regions) |
| `my-registry` | Unique registry name (1-63 chars, lowercase, hyphens) | Choose a unique name; used in `docker push` and `imagePullSecret` |

## Related Presets

- None for this component; consider `DigitalOceanKubernetesCluster` with `registryIntegration: true` for seamless image pull
