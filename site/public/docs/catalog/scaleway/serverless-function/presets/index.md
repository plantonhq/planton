---
title: "Presets"
description: "Ready-to-deploy configuration presets for Serverless Function"
type: "preset-list"
componentSlug: "serverless-function"
componentTitle: "Serverless Function"
provider: "scaleway"
icon: "package"
order: 200
presets:
  - slug: "01-http-api"
    rank: "01"
    title: "HTTP API Function"
    excerpt: "This preset creates a public Scaleway Serverless Function using the Node.js 20 runtime. It auto-scales from zero to 20 instances based on incoming HTTP requests. This is the most common serverless..."
  - slug: "02-scheduled-job"
    rank: "02"
    title: "Scheduled Job Function"
    excerpt: "This preset creates a private Scaleway Serverless Function with a CRON trigger that runs daily at 2:00 AM UTC. The function uses the Python 3.11 runtime and scales to a single instance. This is the..."
---

# Serverless Function Presets

Ready-to-deploy configuration presets for Serverless Function. Each preset is a complete manifest you can copy, customize, and deploy.
