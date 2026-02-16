---
title: "Presets"
description: "Ready-to-deploy configuration presets for Authorization Model"
type: "preset-list"
componentSlug: "authorization-model"
componentTitle: "Authorization Model"
provider: "openfga"
icon: "package"
order: 200
presets:
  - slug: "01-rbac-dsl"
    rank: "01"
    title: "RBAC Authorization Model (DSL)"
    excerpt: "This preset creates an OpenFGA authorization model using the DSL format that implements role-based access control (RBAC) with user types, groups, and document permissions. This is the most common..."
  - slug: "02-document-access-dsl"
    rank: "02"
    title: "Hierarchical Document Access Model (DSL)"
    excerpt: "This preset creates an OpenFGA authorization model with hierarchical folder/document permissions where access is inherited through parent relationships (Google Drive-style). Viewers and editors on a..."
---

# Authorization Model Presets

Ready-to-deploy configuration presets for Authorization Model. Each preset is a complete manifest you can copy, customize, and deploy.
