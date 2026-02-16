---
title: "Presets"
description: "Ready-to-deploy configuration presets for Job"
type: "preset-list"
componentSlug: "job"
componentTitle: "Job"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-batch-processing"
    rank: "01"
    title: "Batch Processing Job"
    excerpt: "This preset creates a one-shot Kubernetes Job for batch processing. The job runs a single pod to completion and automatically cleans up after 1 hour."
  - slug: "02-data-migration"
    rank: "02"
    title: "Data Migration Job"
    excerpt: "This preset creates a Kubernetes Job for database schema migrations or data migrations. Configured with a 1-hour deadline, higher resource limits, and minimal retries to prevent duplicate migration..."
---

# Job Presets

Ready-to-deploy configuration presets for Job. Each preset is a complete manifest you can copy, customize, and deploy.
