---
title: "Presets"
description: "Ready-to-deploy configuration presets for ECR Repo"
type: "preset-list"
componentSlug: "ecr-repo"
componentTitle: "ECR Repo"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-production-immutable"
    rank: "01"
    title: "Production Immutable ECR Repository"
    excerpt: "This preset creates an ECR repository with immutable image tags, automatic vulnerability scanning, and a lifecycle policy that balances cost control with rollback capability. Immutable tags ensure..."
  - slug: "02-development"
    rank: "02"
    title: "Development ECR Repository"
    excerpt: "This preset creates an ECR repository optimized for development workflows. Mutable tags allow developers to push `latest` or branch-based tags repeatedly without errors. Aggressive lifecycle rules..."
---

# ECR Repo Presets

Ready-to-deploy configuration presets for ECR Repo. Each preset is a complete manifest you can copy, customize, and deploy.
