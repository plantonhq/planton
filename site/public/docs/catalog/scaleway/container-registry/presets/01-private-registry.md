---
title: "Private Container Registry"
description: "This preset creates a private Scaleway Container Registry namespace. Images pushed to this registry require authentication to pull, making it suitable for proprietary application images. This is the..."
type: "preset"
rank: "01"
presetSlug: "01-private-registry"
componentSlug: "container-registry"
componentTitle: "Container Registry"
provider: "scaleway"
icon: "package"
order: 1
---

# Private Container Registry

This preset creates a private Scaleway Container Registry namespace. Images pushed to this registry require authentication to pull, making it suitable for proprietary application images. This is the standard and only common configuration -- virtually all production registries are private.

## When to Use

- Storing proprietary Docker images for Kapsule clusters, Serverless Containers, or Instances
- CI/CD pipelines that build and push application images
- Any scenario where container images should not be publicly downloadable

## Key Configuration Choices

- **Private** (`isPublic: false`) -- images require Scaleway IAM authentication to pull; only authenticated users and services can access them
- **Paris region** (`region: fr-par`) -- images are stored in the Paris region; use `nl-ams` or `pl-waw` for data residency requirements

## Placeholders to Replace

No placeholders -- this preset is ready to deploy as-is. The registry endpoint is available in `status.outputs.endpoint` after creation (e.g., `rg.fr-par.scw.cloud/my-registry`).
