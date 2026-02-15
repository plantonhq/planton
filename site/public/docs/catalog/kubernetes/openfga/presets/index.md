---
title: "Presets"
description: "Ready-to-deploy configuration presets for OpenFGA"
type: "preset-list"
componentSlug: "openfga"
componentTitle: "OpenFGA"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-with-postgres-datastore"
    rank: "01"
    title: "OpenFGA with PostgreSQL Datastore"
    excerpt: "This preset deploys OpenFGA with a PostgreSQL backend for storing authorization models and relationship tuples. OpenFGA is a high-performance authorization engine based on Google's Zanzibar paper."
---

# OpenFGA Presets

Ready-to-deploy configuration presets for OpenFGA. Each preset is a complete manifest you can copy, customize, and deploy.
