---
title: "Presets"
description: "Ready-to-deploy configuration presets for ALB"
type: "preset-list"
componentSlug: "alb"
componentTitle: "ALB"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-internet-facing-https"
    rank: "01"
    title: "Internet-Facing HTTPS ALB"
    excerpt: "This preset creates an internet-facing Application Load Balancer with HTTPS termination and Route53 DNS management. It enables deletion protection and uses the AWS-recommended 60-second idle timeout...."
---

# ALB Presets

Ready-to-deploy configuration presets for ALB. Each preset is a complete manifest you can copy, customize, and deploy.
