---
title: "Presets"
description: "Ready-to-deploy configuration presets for WAF Web ACL"
type: "preset-list"
componentSlug: "waf-web-acl"
componentTitle: "WAF Web ACL"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-managed-rules-basic"
    rank: "01"
    title: "Managed Rules Basic"
    excerpt: "This preset creates a Web ACL with two AWS Managed Rule Groups that cover the most common web application threats. It allows all traffic by default and relies on the managed rules to block known bad..."
  - slug: "02-rate-limiting-with-managed-rules"
    rank: "02"
    title: "Rate Limiting with Managed Rules"
    excerpt: "This preset creates a Web ACL that combines IP-based rate limiting with three AWS Managed Rule Groups. The rate limit is evaluated first (lowest priority number) to block volumetric attacks before..."
  - slug: "03-production-web-app"
    rank: "03"
    title: "Production Web Application"
    excerpt: "This preset creates a comprehensive Web ACL suitable for production web applications. It combines rate limiting, geographic blocking, five AWS Managed Rule Groups, custom response bodies, and logging..."
---

# WAF Web ACL Presets

Ready-to-deploy configuration presets for WAF Web ACL. Each preset is a complete manifest you can copy, customize, and deploy.
