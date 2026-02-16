---
title: "Presets"
description: "Ready-to-deploy configuration presets for Cloud Run"
type: "preset-list"
componentSlug: "cloud-run"
componentTitle: "Cloud Run"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-public-service"
    rank: "01"
    title: "Public Cloud Run Service"
    excerpt: "This preset deploys a publicly accessible Cloud Run service with unauthenticated access, scale-to-zero, and Gen 2 execution environment. It uses all recommended defaults from the spec and is the..."
  - slug: "02-private-vpc-connected"
    rank: "02"
    title: "Private VPC-Connected Cloud Run Service"
    excerpt: "This preset deploys a Cloud Run service that is only accessible internally (within the VPC and other GCP services), requires IAM authentication, and has Direct VPC Egress for connecting to private..."
---

# Cloud Run Presets

Ready-to-deploy configuration presets for Cloud Run. Each preset is a complete manifest you can copy, customize, and deploy.
