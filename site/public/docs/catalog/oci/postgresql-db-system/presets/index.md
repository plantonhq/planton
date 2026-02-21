---
title: "Presets"
description: "Ready-to-deploy configuration presets for PostgreSQL DB System"
type: "preset-list"
componentSlug: "postgresql-db-system"
componentTitle: "PostgreSQL DB System"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-regionally-durable"
    rank: "01"
    title: "Regionally Durable"
    excerpt: "This preset creates a production PostgreSQL DB System with regionally durable storage (replicated across availability domains), a read replica for scaling read queries, reader endpoint for automatic..."
  - slug: "02-standalone-development"
    rank: "02"
    title: "Standalone Development"
    excerpt: "This preset creates a single-instance PostgreSQL DB System with AD-local storage for development and testing. It uses the smallest flex shape configuration, short backup retention, and a plain-text..."
---

# PostgreSQL DB System Presets

Ready-to-deploy configuration presets for PostgreSQL DB System. Each preset is a complete manifest you can copy, customize, and deploy.
