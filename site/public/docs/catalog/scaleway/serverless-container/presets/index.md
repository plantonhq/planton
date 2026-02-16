---
title: "Presets"
description: "Ready-to-deploy configuration presets for Serverless Container"
type: "preset-list"
componentSlug: "serverless-container"
componentTitle: "Serverless Container"
provider: "scaleway"
icon: "package"
order: 200
presets:
  - slug: "01-public-web-service"
    rank: "01"
    title: "Public Web Service Container"
    excerpt: "This preset creates a publicly accessible Scaleway Serverless Container running an image from a Scaleway Container Registry. It auto-scales from zero to 20 instances based on incoming HTTP traffic...."
  - slug: "02-vpc-connected"
    rank: "02"
    title: "VPC-Connected Private Container"
    excerpt: "This preset creates a Scaleway Serverless Container attached to a Private Network with private privacy, a health check, and a minimum scale of 1 to avoid cold starts. This is the standard..."
---

# Serverless Container Presets

Ready-to-deploy configuration presets for Serverless Container. Each preset is a complete manifest you can copy, customize, and deploy.
