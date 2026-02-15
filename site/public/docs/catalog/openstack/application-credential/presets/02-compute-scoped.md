---
title: "Compute-Scoped Application Credential"
description: "This preset creates an application credential with access limited to compute (Nova) operations. It can list, create, manage, and delete servers but cannot touch networking, storage, or identity..."
type: "preset"
rank: "02"
presetSlug: "02-compute-scoped"
componentSlug: "application-credential"
componentTitle: "Application Credential"
provider: "openstack"
icon: "package"
order: 2
---

# Compute-Scoped Application Credential

This preset creates an application credential with access limited to compute (Nova) operations. It can list, create, manage, and delete servers but cannot touch networking, storage, or identity resources. Useful for automation that provisions and manages VMs without broader infrastructure access.

## When to Use

- CI/CD pipelines that create and destroy test instances
- Autoscaling controllers that manage VM lifecycle
- Application deployment tools that only interact with compute APIs

## Key Configuration Choices

- **Member role** -- provides compute operations within the project
- **Compute-only access rules** -- restricts to Nova API paths; blocks network, storage, and identity calls
- **Full server lifecycle** -- GET (list/show), POST (create, actions like reboot/resize), DELETE (destroy)
- **Restricted** (`unrestricted: false`, default) -- cannot create sub-credentials
- **Auto-generated secret** -- OpenStack generates a random secret at creation time

## Placeholders to Replace

No placeholders -- this preset is deployable as-is after setting `metadata.name`. Adjust `accessRules` to add or remove specific compute actions.

## Related Presets

- **01-restricted-readonly** -- Use instead when the credential should only read infrastructure state, not modify it
