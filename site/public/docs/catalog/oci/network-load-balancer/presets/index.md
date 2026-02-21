---
title: "Presets"
description: "Ready-to-deploy configuration presets for Network Load Balancer"
type: "preset-list"
componentSlug: "network-load-balancer"
componentTitle: "Network Load Balancer"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-public-tcp"
    rank: "01"
    title: "Public TCP Pass-Through NLB"
    excerpt: "This preset creates a public OCI Network Load Balancer that distributes TCP traffic at Layer 4 with source IP preservation enabled. Backends receive the original client IP address in every packet,..."
  - slug: "02-private-internal"
    rank: "02"
    title: "Private Internal NLB"
    excerpt: "This preset creates a private OCI Network Load Balancer for internal service-to-service communication within the VCN. The NLB receives only a private IP address and is not accessible from the public..."
  - slug: "03-development"
    rank: "03"
    title: "Development NLB"
    excerpt: "This preset creates a minimal OCI Network Load Balancer for development, testing, or learning. It deploys a public NLB with a single TCP listener on port 80, a TCP health check, and no backends...."
---

# Network Load Balancer Presets

Ready-to-deploy configuration presets for Network Load Balancer. Each preset is a complete manifest you can copy, customize, and deploy.
