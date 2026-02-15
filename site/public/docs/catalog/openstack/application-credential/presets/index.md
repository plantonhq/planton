---
title: "Presets"
description: "Ready-to-deploy configuration presets for Application Credential"
type: "preset-list"
componentSlug: "application-credential"
componentTitle: "Application Credential"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-restricted-readonly"
    rank: "01"
    title: "Restricted Read-Only Application Credential"
    excerpt: "This preset creates an application credential with read-only access to compute, network, and block-storage APIs. It uses the `reader` role and further restricts access via `accessRules` to GET..."
  - slug: "02-compute-scoped"
    rank: "02"
    title: "Compute-Scoped Application Credential"
    excerpt: "This preset creates an application credential with access limited to compute (Nova) operations. It can list, create, manage, and delete servers but cannot touch networking, storage, or identity..."
---

# Application Credential Presets

Ready-to-deploy configuration presets for Application Credential. Each preset is a complete manifest you can copy, customize, and deploy.
