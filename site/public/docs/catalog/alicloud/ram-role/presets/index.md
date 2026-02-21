---
title: "Presets"
description: "Ready-to-deploy configuration presets for RAM Role"
type: "preset-list"
componentSlug: "ram-role"
componentTitle: "RAM Role"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-ecs-service-role"
    rank: "01"
    title: "ECS Service Role"
    excerpt: "This preset creates a RAM role that ECS instances can assume to access common Alibaba Cloud services: OSS for object storage, CloudMonitor for metrics and alerting, and Log Service (SLS) for..."
  - slug: "02-fc-execution-role"
    rank: "02"
    title: "FC Execution Role"
    excerpt: "This preset creates a RAM role that Alibaba Cloud Function Compute can assume when executing functions. It includes Log Service (SLS) full access so function invocation logs are written to your SLS..."
  - slug: "03-cross-account-audit"
    rank: "03"
    title: "Cross-Account Audit Role"
    excerpt: "This preset creates a RAM role that another Alibaba Cloud account can assume for read-only security auditing. The trusted account gains access to billing data (BSS), centralized logs (SLS), and..."
---

# RAM Role Presets

Ready-to-deploy configuration presets for RAM Role. Each preset is a complete manifest you can copy, customize, and deploy.
