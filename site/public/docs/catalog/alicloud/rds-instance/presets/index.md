---
title: "Presets"
description: "Ready-to-deploy configuration presets for RDS Instance"
type: "preset-list"
componentSlug: "rds-instance"
componentTitle: "RDS Instance"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-mysql-basic"
    rank: "01"
    title: "MySQL Basic Development Instance"
    excerpt: "This preset creates a minimal MySQL 8.0 instance with a single database and account, suitable for development and testing."
  - slug: "02-postgresql-ha"
    rank: "02"
    title: "PostgreSQL HA Production Instance"
    excerpt: "This preset creates a production-grade PostgreSQL 16 instance with high availability, SSL encryption, cross-AZ deployment, and fine-grained monitoring."
  - slug: "03-mysql-production"
    rank: "03"
    title: "MySQL Production with Encryption"
    excerpt: "This preset creates a production MySQL 8.0 instance with high availability, TDE encryption, KMS disk encryption, SSL, monitoring, and performance-tuned parameters."
---

# RDS Instance Presets

Ready-to-deploy configuration presets for RDS Instance. Each preset is a complete manifest you can copy, customize, and deploy.
