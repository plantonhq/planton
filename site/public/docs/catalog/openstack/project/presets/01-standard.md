---
title: "Standard Project"
description: "This preset creates an enabled project in the default domain. A project is the fundamental organizational unit in OpenStack -- all cloud resources belong to a project, and it provides resource..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "project"
componentTitle: "Project"
provider: "openstack"
icon: "package"
order: 1
---

# Standard Project

This preset creates an enabled project in the default domain. A project is the fundamental organizational unit in OpenStack -- all cloud resources belong to a project, and it provides resource isolation and quota boundaries. This is the standard configuration for most deployments.

## When to Use

- Provisioning a new tenant or team workspace
- Creating isolated environments (dev, staging, production) as separate projects
- Automating project creation as part of onboarding workflows

## Key Configuration Choices

- **Enabled** (`enabled: true`, default) -- project is active and users can create resources immediately
- **Default domain** -- no `domainId` specified; project is created in OpenStack's default domain (suitable for single-domain deployments)
- **Top-level project** -- no `parentId` specified; project is not nested in a hierarchy

## Placeholders to Replace

No placeholders -- this preset is deployable as-is after setting `metadata.name`.
