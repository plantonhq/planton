---
title: "Development Project"
description: "This preset creates a lightweight GCP project for development and testing. It enables the `addSuffix` flag to append a random suffix to the project ID, preventing collisions when multiple developers..."
type: "preset"
rank: "02"
presetSlug: "02-development"
componentSlug: "project"
componentTitle: "Project"
provider: "gcp"
icon: "package"
order: 2
---

# Development Project

This preset creates a lightweight GCP project for development and testing. It enables the `addSuffix` flag to append a random suffix to the project ID, preventing collisions when multiple developers create dev projects from the same template. Only the essential compute, container, and IAM APIs are enabled.

## When to Use

- Development or staging environments with shorter lifecycles
- Individual developer sandboxes where project IDs may collide
- Cost-conscious environments that don't need the full API surface

## Key Configuration Choices

- **Random suffix** (`addSuffix: true`) -- ensures uniqueness across multiple dev deployments
- **No deletion protection** -- dev projects should be easy to tear down
- **No owner member** -- IAM managed separately or inherited from folder
- **3 core APIs only** -- compute, container, and IAM (add more as needed)
- **Default network disabled** -- security hardening applies to all environments

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-project-id>` | Base project ID (suffix will be appended) | Choose a descriptive base name |
| `<your-folder-id>` | Numeric folder ID for dev projects | GCP Resource Manager console |
| `<AAAAAA-BBBBBB-CCCCCC>` | Billing account ID | GCP Billing console |

## Related Presets

- **01-standard-production** -- Use for production projects with full API surface and deletion protection
