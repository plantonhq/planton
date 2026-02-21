---
title: "Presets"
description: "Ready-to-deploy configuration presets for Autonomous Database"
type: "preset-list"
componentSlug: "autonomous-database"
componentTitle: "Autonomous Database"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-serverless-oltp"
    rank: "01"
    title: "Serverless OLTP (Autonomous Transaction Processing)"
    excerpt: "This preset creates a serverless Autonomous Transaction Processing (ATP) database with the ECPU compute model, private endpoint networking, and auto-scaling for both compute and storage. ATP is..."
  - slug: "02-free-tier-development"
    rank: "02"
    title: "Free Tier Development"
    excerpt: "This preset creates an Always Free Autonomous Database for development and experimentation at zero cost. The database is limited to 2 ECPUs and 20 GB of usable storage but provides the full..."
  - slug: "03-serverless-data-warehouse"
    rank: "03"
    title: "Serverless Data Warehouse (Autonomous Data Warehouse)"
    excerpt: "This preset creates a serverless Autonomous Data Warehouse (ADW) with the ECPU compute model, Enterprise Edition for advanced analytic features, private endpoint networking, and auto-scaling. ADW is..."
---

# Autonomous Database Presets

Ready-to-deploy configuration presets for Autonomous Database. Each preset is a complete manifest you can copy, customize, and deploy.
