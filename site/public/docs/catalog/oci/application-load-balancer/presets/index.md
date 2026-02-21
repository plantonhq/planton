---
title: "Presets"
description: "Ready-to-deploy configuration presets for Application Load Balancer"
type: "preset-list"
componentSlug: "application-load-balancer"
componentTitle: "Application Load Balancer"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-internet-facing-https"
    rank: "01"
    title: "Internet-Facing HTTPS Load Balancer"
    excerpt: "This preset creates a public OCI Application Load Balancer with HTTPS termination on port 443 and an automatic HTTP-to-HTTPS redirect on port 80. It uses the flexible shape with 10-100 Mbps..."
  - slug: "02-internal-http"
    rank: "02"
    title: "Internal HTTP Load Balancer"
    excerpt: "This preset creates a private OCI Application Load Balancer for distributing traffic across backend services within the VCN. It uses the flexible shape with 10-50 Mbps bandwidth, deploys in a single..."
  - slug: "03-development"
    rank: "03"
    title: "Development Load Balancer"
    excerpt: "This preset creates a minimal-cost public OCI Application Load Balancer for dev/test environments. It uses the flexible shape locked to the minimum 10 Mbps bandwidth, a single subnet, an HTTP-only..."
---

# Application Load Balancer Presets

Ready-to-deploy configuration presets for Application Load Balancer. Each preset is a complete manifest you can copy, customize, and deploy.
