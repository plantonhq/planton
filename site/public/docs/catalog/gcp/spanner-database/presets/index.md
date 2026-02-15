---
title: "Presets"
description: "Ready-to-deploy configuration presets for Spanner Database"
type: "preset-list"
componentSlug: "spanner-database"
componentTitle: "Spanner Database"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-basic-database"
    rank: "01"
    title: "Basic Database"
    excerpt: "This preset creates a minimal Cloud Spanner database with the default GoogleSQL dialect and 1-hour version retention. It is the simplest starting point for any Spanner workload."
  - slug: "02-postgresql-database"
    rank: "02"
    title: "PostgreSQL Database"
    excerpt: "This preset creates a Cloud Spanner database with the PostgreSQL-compatible dialect and a 7-day version retention period for extended point-in-time recovery. Ideal for teams with PostgreSQL expertise..."
  - slug: "03-cmek-encrypted"
    rank: "03"
    title: "CMEK-Encrypted Database"
    excerpt: "This preset creates a Cloud Spanner database with customer-managed encryption (CMEK), GCP API-level drop protection, a 3-day version retention period, and an explicit UTC time zone. Designed for..."
---

# Spanner Database Presets

Ready-to-deploy configuration presets for Spanner Database. Each preset is a complete manifest you can copy, customize, and deploy.
