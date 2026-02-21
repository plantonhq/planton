---
title: "Presets"
description: "Ready-to-deploy configuration presets for ALB Load Balancer"
type: "preset-list"
componentSlug: "alb-load-balancer"
componentTitle: "ALB Load Balancer"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-internet-http"
    rank: "01"
    title: "Internet-Facing HTTP ALB"
    excerpt: "This preset creates a public-facing ALB with a single HTTP listener and one server group. This is the quickest way to get an L7 load balancer running."
  - slug: "02-https-production"
    rank: "02"
    title: "Production HTTPS ALB"
    excerpt: "This preset creates a production-grade ALB with HTTPS, WAF integration, strict TLS policy, access logging, and session stickiness."
  - slug: "03-internal-grpc"
    rank: "03"
    title: "Internal GRPC ALB"
    excerpt: "This preset creates a VPC-internal ALB for service-to-service GRPC communication. The ALB is not accessible from the internet."
---

# ALB Load Balancer Presets

Ready-to-deploy configuration presets for ALB Load Balancer. Each preset is a complete manifest you can copy, customize, and deploy.
