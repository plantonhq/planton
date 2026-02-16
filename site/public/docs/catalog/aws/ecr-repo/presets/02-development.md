---
title: "Development ECR Repository"
description: "This preset creates an ECR repository optimized for development workflows. Mutable tags allow developers to push `latest` or branch-based tags repeatedly without errors. Aggressive lifecycle rules..."
type: "preset"
rank: "02"
presetSlug: "02-development"
componentSlug: "ecr-repo"
componentTitle: "ECR Repo"
provider: "aws"
icon: "package"
order: 2
---

# Development ECR Repository

This preset creates an ECR repository optimized for development workflows. Mutable tags allow developers to push `latest` or branch-based tags repeatedly without errors. Aggressive lifecycle rules keep storage costs low by expiring untagged images after 3 days and retaining only the 20 most recent images.

## When to Use

- Development and staging container registries where rapid iteration is prioritized
- Feature branch builds that push repeatedly to the same tag (e.g., `feature-xyz`)
- Cost-conscious environments where old images have no rollback value

## Key Configuration Choices

- **Mutable tags** (`imageImmutable: false`) -- Tags like `latest` or `dev` can be overwritten on each push
- **Scan on push** (`scanOnPush: true`) -- Vulnerability scanning stays on even for dev to catch issues early
- **3-day untagged expiration** (`expireUntaggedAfterDays: 3`) -- Aggressively cleans up orphaned layers from frequent rebuilds
- **20 image retention** (`maxImageCount: 20`) -- Minimal retention; older dev images are rarely needed
- **Force delete disabled** (`forceDelete: false`) -- Still protects against accidental repository deletion

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<repository-name>` | ECR repository name (e.g., `myorg/api-service-dev` or `team-blue/frontend`) | Your team's container image naming convention |

## Related Presets

- **01-production-immutable** -- Use instead for production registries requiring tag integrity and longer image retention
