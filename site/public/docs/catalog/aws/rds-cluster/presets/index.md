---
title: "Presets"
description: "Ready-to-deploy configuration presets for RDS Cluster"
type: "preset-list"
componentSlug: "rds-cluster"
componentTitle: "RDS Cluster"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-aurora-postgresql"
    rank: "01"
    title: "Aurora PostgreSQL Cluster"
    excerpt: "This preset creates a production-ready Aurora PostgreSQL cluster with RDS-managed master password (stored in Secrets Manager), encrypted storage, deletion protection, 7-day backup retention, and..."
  - slug: "02-aurora-mysql"
    rank: "02"
    title: "Aurora MySQL Cluster"
    excerpt: "This preset creates a production-ready Aurora MySQL cluster with the same security and resilience defaults as the PostgreSQL preset: managed password, encrypted storage, deletion protection, and..."
  - slug: "03-aurora-serverless-v2"
    rank: "03"
    title: "Aurora Serverless v2"
    excerpt: "This preset creates an Aurora PostgreSQL cluster with Serverless v2 auto-scaling, where compute capacity automatically adjusts between 0.5 and 16 ACUs based on workload demand. The Data API (HTTP..."
---

# RDS Cluster Presets

Ready-to-deploy configuration presets for RDS Cluster. Each preset is a complete manifest you can copy, customize, and deploy.
