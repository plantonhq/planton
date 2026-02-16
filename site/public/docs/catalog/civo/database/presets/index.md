---
title: "Presets"
description: "Ready-to-deploy configuration presets for Database"
type: "preset-list"
componentSlug: "database"
componentTitle: "Database"
provider: "civo"
icon: "package"
order: 200
presets:
  - slug: "01-postgresql-production"
    rank: "01"
    title: "Production PostgreSQL Database"
    excerpt: "This preset creates a production-grade managed PostgreSQL 16 database with 2 read replicas (3 nodes total), VPC networking, and firewall protection. This is the most common configuration for..."
  - slug: "02-mysql-development"
    rank: "02"
    title: "Development MySQL Database"
    excerpt: "This preset creates a minimal MySQL 8.0 database for development and testing. Single node (no replicas), smallest instance size, no firewall. Keeps cost low while providing a fully managed MySQL..."
---

# Database Presets

Ready-to-deploy configuration presets for Database. Each preset is a complete manifest you can copy, customize, and deploy.
