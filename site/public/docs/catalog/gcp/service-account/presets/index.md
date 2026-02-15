---
title: "Presets"
description: "Ready-to-deploy configuration presets for Service Account"
type: "preset-list"
componentSlug: "service-account"
componentTitle: "Service Account"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-workload-identity"
    rank: "01"
    title: "Workload Identity Service Account"
    excerpt: "This preset creates a GCP service account designed for GKE Workload Identity. No JSON key is generated -- pods authenticate via KSA-to-GSA binding instead. The account is granted logging, monitoring,..."
  - slug: "02-ci-cd-pipeline"
    rank: "02"
    title: "CI/CD Pipeline Service Account"
    excerpt: "This preset creates a GCP service account with a JSON key for CI/CD pipelines (GitHub Actions, GitLab CI, Jenkins). It has permissions to push container images, deploy to GKE, and deploy Cloud Run..."
---

# Service Account Presets

Ready-to-deploy configuration presets for Service Account. Each preset is a complete manifest you can copy, customize, and deploy.
