---
title: "Private Docker Repository"
description: "This preset creates a private Artifact Registry repository for Docker container images. It is the most common repository type in cloud-native workflows, used for storing application images built by..."
type: "preset"
rank: "01"
presetSlug: "01-docker-private"
componentSlug: "artifact-registry-repo"
componentTitle: "Artifact Registry Repo"
provider: "gcp"
icon: "package"
order: 1
---

# Private Docker Repository

This preset creates a private Artifact Registry repository for Docker container images. It is the most common repository type in cloud-native workflows, used for storing application images built by CI/CD pipelines and pulled by GKE or Cloud Run.

## When to Use

- Storing container images for GKE, Cloud Run, or Compute Engine workloads
- CI/CD pipelines that build and push Docker images
- Any container-based deployment that needs a private image registry

## Key Configuration Choices

- **Docker format** (`repoFormat: DOCKER`) -- standard OCI container image storage
- **Private** (`enablePublicAccess: false`) -- requires IAM authentication to pull/push
- **Regional** -- place the registry close to your compute for faster image pulls

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<gcp-region>` | GCP region (e.g., `us-central1`) | Match your GKE/Cloud Run region for fastest pulls |

## Related Presets

- **02-npm-private** -- Use for hosting private NPM packages
