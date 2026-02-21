---
title: "Presets"
description: "Ready-to-deploy configuration presets for Function"
type: "preset-list"
componentSlug: "function"
componentTitle: "Function"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-event-handler"
    rank: "01"
    title: "Event Handler Function"
    excerpt: "This preset creates a lightweight Python 3.12 function for event-driven workloads. The function code is deployed from an OSS bucket, sized for modest compute needs (0.5 vCPU, 256 MB), and logs..."
  - slug: "02-vpc-api-function"
    rank: "02"
    title: "VPC-Connected API Function"
    excerpt: "This preset creates a Node.js 20 function configured as a backend API handler with access to VPC-internal resources (databases, caches, internal services). The function runs inside a VPC with..."
  - slug: "03-custom-container"
    rank: "03"
    title: "Custom Container Function"
    excerpt: "This preset creates a Function Compute v3 function that runs a custom Docker image. The container listens on port 8080 for HTTP requests with a health check endpoint, and includes lifecycle hooks for..."
---

# Function Presets

Ready-to-deploy configuration presets for Function. Each preset is a complete manifest you can copy, customize, and deploy.
