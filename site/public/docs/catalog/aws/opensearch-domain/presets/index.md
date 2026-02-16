---
title: "Presets"
description: "Ready-to-deploy configuration presets for OpenSearch Domain"
type: "preset-list"
componentSlug: "opensearch-domain"
componentTitle: "OpenSearch Domain"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-single-node-dev"
    rank: "01"
    title: "Single-Node Development Domain"
    excerpt: "This preset creates a minimal single-node OpenSearch domain suitable for development, prototyping, and learning. The domain is publicly accessible (no VPC) with encryption enabled for security best..."
  - slug: "02-production-vpc"
    rank: "02"
    title: "Production VPC Domain"
    excerpt: "This preset creates a production-grade OpenSearch domain with 3 data nodes across 3 Availability Zones, 3 dedicated master nodes, VPC deployment, fine-grained access control with internal user..."
  - slug: "03-analytics-warm-cold"
    rank: "03"
    title: "Analytics Domain with Warm + Cold Storage"
    excerpt: "This preset creates an analytics-optimized OpenSearch domain with 3 data nodes, 3 UltraWarm nodes, and cold storage enabled. Designed for log analytics, time-series data, SIEM, and any workload where..."
---

# OpenSearch Domain Presets

Ready-to-deploy configuration presets for OpenSearch Domain. Each preset is a complete manifest you can copy, customize, and deploy.
