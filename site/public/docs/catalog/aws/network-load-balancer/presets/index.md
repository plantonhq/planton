---
title: "Presets"
description: "Ready-to-deploy configuration presets for Network Load Balancer"
type: "preset-list"
componentSlug: "network-load-balancer"
componentTitle: "Network Load Balancer"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-tcp-internal"
    rank: "01"
    title: "TCP Internal NLB"
    excerpt: "This preset creates an internal Network Load Balancer for microservice-to-microservice communication within a VPC. It uses plain TCP (no TLS termination), spans two subnets for high availability, and..."
  - slug: "02-tls-internet-facing"
    rank: "02"
    title: "TLS Internet-Facing NLB"
    excerpt: "This preset creates an internet-facing Network Load Balancer with TLS termination on port 443. The NLB decrypts incoming TLS and forwards plaintext TCP to targets on your application port. Includes..."
  - slug: "03-static-ip-production"
    rank: "03"
    title: "Static IP Production NLB"
    excerpt: "This preset creates a production-grade internet-facing Network Load Balancer with Elastic IPs for static public IPs, TLS termination, Route53 DNS, HTTP health checks, cross-zone load balancing,..."
---

# Network Load Balancer Presets

Ready-to-deploy configuration presets for Network Load Balancer. Each preset is a complete manifest you can copy, customize, and deploy.
