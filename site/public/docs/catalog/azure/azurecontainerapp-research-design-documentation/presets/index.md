---
title: "Presets"
description: "Ready-to-deploy configuration presets for AzureContainerApp: Research & Design Documentation"
type: "preset-list"
componentSlug: "azurecontainerapp-research-design-documentation"
componentTitle: "AzureContainerApp: Research & Design Documentation"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-web-service"
    rank: "01"
    title: "Web Service"
    excerpt: "This preset deploys a publicly accessible web service with HTTP auto-scaling, health probes, and external ingress. It starts with 1 replica and scales up to 10 based on concurrent HTTP requests. This..."
  - slug: "02-background-worker"
    rank: "02"
    title: "Background Worker"
    excerpt: "This preset deploys a background worker that processes messages from an Azure Service Bus queue. It has no ingress (not accessible via HTTP), scales to zero when the queue is empty, and scales up to..."
  - slug: "03-enterprise-api"
    rank: "03"
    title: "Enterprise API"
    excerpt: "This preset deploys a production-grade API with User Assigned managed identity, Key Vault secrets, ACR authentication via identity, IP security restrictions, full health probe coverage (liveness,..."
---

# AzureContainerApp: Research & Design Documentation Presets

Ready-to-deploy configuration presets for AzureContainerApp: Research & Design Documentation. Each preset is a complete manifest you can copy, customize, and deploy.
