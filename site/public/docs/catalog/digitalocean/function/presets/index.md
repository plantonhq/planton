---
title: "Presets"
description: "Ready-to-deploy configuration presets for Function"
type: "preset-list"
componentSlug: "function"
componentTitle: "Function"
provider: "digitalocean"
icon: "package"
order: 200
presets:
  - slug: "01-web-api"
    rank: "01"
    title: "Web API Function"
    excerpt: "This preset creates a serverless HTTP function on DigitalOcean, deployed via App Platform from a GitHub repository. The function is exposed as a web endpoint with auto-deploy on push, suitable for..."
  - slug: "02-scheduled-job"
    rank: "02"
    title: "Scheduled Background Job"
    excerpt: "This preset creates a serverless function on DigitalOcean that runs on a cron schedule. It is not exposed as an HTTP endpoint, making it suitable for ETL pipelines, data synchronization, report..."
---

# Function Presets

Ready-to-deploy configuration presets for Function. Each preset is a complete manifest you can copy, customize, and deploy.
