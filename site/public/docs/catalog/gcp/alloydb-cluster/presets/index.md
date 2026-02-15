---
title: "Presets"
description: "Ready-to-deploy configuration presets for AlloyDB Cluster"
type: "preset-list"
componentSlug: "alloydb-cluster"
componentTitle: "AlloyDB Cluster"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-dev-basic"
    rank: "01"
    title: "Dev Basic"
    excerpt: "Minimal development cluster for local development, CI/CD pipelines, and prototyping. Uses 2 CPUs, ZONAL availability, no CMEK, and deletion protection disabled for easy teardown."
  - slug: "02-ha-production"
    rank: "02"
    title: "HA Production"
    excerpt: "Production-ready AlloyDB cluster with high availability, automated backups with 7-day retention, initial user, deletion protection, and ENCRYPTED_ONLY SSL. Uses 4 CPUs and REGIONAL deployment for..."
  - slug: "03-enterprise-encrypted"
    rank: "03"
    title: "Enterprise Encrypted"
    excerpt: "Enterprise-grade AlloyDB cluster with CMEK on cluster data, automated backups, and continuous backups; query insights enabled; require_connectors true; ENCRYPTED_ONLY SSL; 30-day automated backup..."
---

# AlloyDB Cluster Presets

Ready-to-deploy configuration presets for AlloyDB Cluster. Each preset is a complete manifest you can copy, customize, and deploy.
