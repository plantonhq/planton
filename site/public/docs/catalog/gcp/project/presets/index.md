---
title: "Presets"
description: "Ready-to-deploy configuration presets for Project"
type: "preset-list"
componentSlug: "project"
componentTitle: "Project"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-standard-production"
    rank: "01"
    title: "Standard Production Project"
    excerpt: "This preset creates a GCP project under a resource hierarchy folder with essential APIs pre-enabled, the default network disabled, and deletion protection turned on. It covers the core services..."
  - slug: "02-development"
    rank: "02"
    title: "Development Project"
    excerpt: "This preset creates a lightweight GCP project for development and testing. It enables the `addSuffix` flag to append a random suffix to the project ID, preventing collisions when multiple developers..."
---

# Project Presets

Ready-to-deploy configuration presets for Project. Each preset is a complete manifest you can copy, customize, and deploy.
