---
title: "Presets"
description: "Ready-to-deploy configuration presets for Cloud Composer Environment"
type: "preset-list"
componentSlug: "cloud-composer-environment"
componentTitle: "Cloud Composer Environment"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-dev-small"
    rank: "01"
    title: "Dev Small"
    excerpt: "A minimal Cloud Composer environment for development and testing workloads. Uses small resource allocations and basic configuration without private networking or advanced features."
  - slug: "02-production-private"
    rank: "02"
    title: "Production Private"
    excerpt: "A production-grade Cloud Composer environment with private networking, high resilience, and scaled workloads. Designed for production Airflow workloads requiring network isolation and high..."
  - slug: "03-enterprise-encrypted"
    rank: "03"
    title: "Enterprise Encrypted"
    excerpt: "An enterprise-grade Cloud Composer environment with full security features including CMEK encryption, private networking, web server access control, and disaster recovery. Designed for organizations..."
---

# Cloud Composer Environment Presets

Ready-to-deploy configuration presets for Cloud Composer Environment. Each preset is a complete manifest you can copy, customize, and deploy.
