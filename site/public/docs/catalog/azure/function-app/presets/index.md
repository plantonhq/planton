---
title: "Presets"
description: "Ready-to-deploy configuration presets for Function App"
type: "preset-list"
componentSlug: "function-app"
componentTitle: "Function App"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-python-http-api"
    rank: "01"
    title: "Python HTTP API"
    excerpt: "This preset deploys a Python 3.12 Function App configured for HTTP-triggered APIs with Application Insights monitoring, a health check endpoint, CORS, and secure defaults. It is the most common..."
  - slug: "02-docker-container"
    rank: "02"
    title: "Docker Container"
    excerpt: "This preset deploys a containerized Function App running a custom Docker image from Azure Container Registry. It uses a system-assigned managed identity for ACR authentication (no registry password),..."
  - slug: "03-enterprise-elastic-premium"
    rank: "03"
    title: "Enterprise Elastic Premium"
    excerpt: "This preset deploys a production-grade Function App on an Elastic Premium plan with VNet integration, managed identity for storage (no access keys), Key Vault secret references, pre-warmed instances..."
---

# Function App Presets

Ready-to-deploy configuration presets for Function App. Each preset is a complete manifest you can copy, customize, and deploy.
