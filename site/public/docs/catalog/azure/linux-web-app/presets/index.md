---
title: "Presets"
description: "Ready-to-deploy configuration presets for Linux Web App"
type: "preset-list"
componentSlug: "linux-web-app"
componentTitle: "Linux Web App"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-node-web-api"
    rank: "01"
    title: "Node.js Web API"
    excerpt: "This preset deploys a Node.js 22 LTS web API with health check monitoring, HTTP/2 for improved performance, CORS for cross-origin requests, and Application Insights telemetry. It is the standard..."
  - slug: "02-docker-container"
    rank: "02"
    title: "Docker Container"
    excerpt: "This preset deploys a containerized web application running a custom Docker image from Azure Container Registry. It uses a system-assigned managed identity for ACR authentication (no registry..."
  - slug: "03-enterprise-private-web-app"
    rank: "03"
    title: "Enterprise Private Web App"
    excerpt: "This preset deploys a production-grade Python web application on a Premium App Service Plan with VNet integration, IP restrictions (default deny), Application Insights monitoring, Key Vault secret..."
---

# Linux Web App Presets

Ready-to-deploy configuration presets for Linux Web App. Each preset is a complete manifest you can copy, customize, and deploy.
