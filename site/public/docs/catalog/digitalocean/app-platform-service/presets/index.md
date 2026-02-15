---
title: "Presets"
description: "Ready-to-deploy configuration presets for App Platform Service"
type: "preset-list"
componentSlug: "app-platform-service"
componentTitle: "App Platform Service"
provider: "digitalocean"
icon: "package"
order: 200
presets:
  - slug: "01-git-source-web"
    rank: "01"
    title: "Git Source Web Service"
    excerpt: "This preset deploys a web service on DigitalOcean App Platform from a GitHub repository. App Platform automatically builds and deploys the application from source, providing HTTPS, health checks, and..."
  - slug: "02-container-image"
    rank: "02"
    title: "Container Image Service"
    excerpt: "This preset deploys a web service on DigitalOcean App Platform from a pre-built container image stored in DigitalOcean Container Registry (DOCR). It uses a professional-tier instance with..."
---

# App Platform Service Presets

Ready-to-deploy configuration presets for App Platform Service. Each preset is a complete manifest you can copy, customize, and deploy.
