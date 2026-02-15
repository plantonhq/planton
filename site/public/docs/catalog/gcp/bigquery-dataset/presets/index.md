---
title: "Presets"
description: "Ready-to-deploy configuration presets for BigQuery Dataset"
type: "preset-list"
componentSlug: "bigquery-dataset"
componentTitle: "BigQuery Dataset"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-basic-analytics"
    rank: "01"
    title: "Preset: Basic Analytics Dataset"
    excerpt: "Use this preset when you need a straightforward BigQuery dataset for analytics workloads with default settings. This is the simplest configuration, suitable for development, prototyping, or workloads..."
  - slug: "02-cmek-encrypted"
    rank: "02"
    title: "Preset: CMEK-Encrypted Dataset"
    excerpt: "Use this preset when your dataset contains sensitive or regulated data that requires customer-managed encryption keys (CMEK). Common compliance scenarios:"
  - slug: "03-team-shared"
    rank: "03"
    title: "Preset: Team-Shared Dataset"
    excerpt: "Use this preset when a dataset needs explicit, team-level access control -- separating data engineers (who create and modify tables) from data analysts (who only read). This is the standard pattern..."
---

# BigQuery Dataset Presets

Ready-to-deploy configuration presets for BigQuery Dataset. Each preset is a complete manifest you can copy, customize, and deploy.
