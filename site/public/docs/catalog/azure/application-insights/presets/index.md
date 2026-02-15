---
title: "Presets"
description: "Ready-to-deploy configuration presets for Application Insights"
type: "preset-list"
componentSlug: "application-insights"
componentTitle: "Application Insights"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Application Insights"
    excerpt: "This preset creates an Azure Application Insights resource with full telemetry collection (100% sampling), 90-day retention, and 100 GB daily cap. This is the standard configuration for development..."
  - slug: "02-production-sampled"
    rank: "02"
    title: "Production Application Insights with Sampling"
    excerpt: "This preset creates an Azure Application Insights resource with 25% adaptive sampling and a 10 GB daily ingestion cap. This is the cost-optimized configuration for high-traffic production workloads..."
---

# Application Insights Presets

Ready-to-deploy configuration presets for Application Insights. Each preset is a complete manifest you can copy, customize, and deploy.
