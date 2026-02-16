---
title: "Presets"
description: "Ready-to-deploy configuration presets for Global Accelerator"
type: "preset-list"
componentSlug: "global-accelerator"
componentTitle: "Global Accelerator"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-basic-tcp-accelerator"
    rank: "01"
    title: "Basic TCP Accelerator"
    excerpt: "This preset creates a minimal Global Accelerator that accepts TCP traffic on port 443 and routes it to a single ALB endpoint. It uses all default settings: AWS-allocated static anycast IPs, no flow..."
  - slug: "02-multi-region-production"
    rank: "02"
    title: "Multi-Region Production Accelerator"
    excerpt: "This preset creates a production-grade Global Accelerator that distributes TCP traffic across two AWS regions — `us-east-1` (60% traffic) and `eu-west-1` (40% traffic). It configures HTTP health..."
  - slug: "03-gaming-udp-accelerator"
    rank: "03"
    title: "Gaming UDP Accelerator"
    excerpt: "This preset creates a Global Accelerator optimized for real-time gaming workloads. It uses UDP protocol across a port range of 7000–8000, enables `SOURCE_IP` client affinity to pin each player to the..."
---

# Global Accelerator Presets

Ready-to-deploy configuration presets for Global Accelerator. Each preset is a complete manifest you can copy, customize, and deploy.
