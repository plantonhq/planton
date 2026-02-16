---
title: "Presets"
description: "Ready-to-deploy configuration presets for CronJob"
type: "preset-list"
componentSlug: "cronjob"
componentTitle: "CronJob"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-scheduled-task"
    rank: "01"
    title: "Scheduled Task CronJob"
    excerpt: "This preset creates a CronJob that runs every 6 hours with a Forbid concurrency policy. The most common CronJob pattern: a periodic task that should not overlap with previous runs."
  - slug: "02-database-backup"
    rank: "02"
    title: "Database Backup CronJob"
    excerpt: "This preset creates a daily database backup CronJob that runs at 2:00 AM UTC. Designed for PostgreSQL backup to S3-compatible storage; adapt the command for your database and storage backend."
---

# CronJob Presets

Ready-to-deploy configuration presets for CronJob. Each preset is a complete manifest you can copy, customize, and deploy.
