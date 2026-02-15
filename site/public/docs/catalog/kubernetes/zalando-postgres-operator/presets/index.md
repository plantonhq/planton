---
title: "Presets"
description: "Ready-to-deploy configuration presets for Zalando Postgres Operator"
type: "preset-list"
componentSlug: "zalando-postgres-operator"
componentTitle: "Zalando Postgres Operator"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Zalando Postgres Operator"
    excerpt: "This preset deploys the Zalando Postgres Operator with recommended default resources and no backup configuration. The operator manages PostgreSQL clusters using the `postgresql` custom resource,..."
  - slug: "02-with-r2-backup"
    rank: "02"
    title: "Zalando Postgres Operator with Cloudflare R2 Backup"
    excerpt: "This preset deploys the Zalando Postgres Operator with WAL-G continuous archiving to Cloudflare R2 storage. All PostgreSQL clusters managed by this operator instance will use the configured R2 bucket..."
---

# Zalando Postgres Operator Presets

Ready-to-deploy configuration presets for Zalando Postgres Operator. Each preset is a complete manifest you can copy, customize, and deploy.
