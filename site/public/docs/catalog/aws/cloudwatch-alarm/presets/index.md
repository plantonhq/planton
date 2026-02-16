---
title: "Presets"
description: "Ready-to-deploy configuration presets for CloudWatch Alarm"
type: "preset-list"
componentSlug: "cloudwatch-alarm"
componentTitle: "CloudWatch Alarm"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-cpu-utilization-alarm"
    rank: "01"
    title: "Preset: CPU Utilization Alarm"
    excerpt: "**Use case:** Alert when EC2 instance CPU exceeds a threshold, indicating high compute load."
  - slug: "02-error-rate-metric-math"
    rank: "02"
    title: "Preset: Error Rate Metric Math Alarm"
    excerpt: "**Use case:** Alert when the 5xx error rate exceeds a percentage threshold using CloudWatch Metric Math."
  - slug: "03-production-multi-action"
    rank: "03"
    title: "Preset: Production Multi-Action Alarm"
    excerpt: "**Use case:** Production-grade alarm with separate notification channels for ALARM, OK, and INSUFFICIENT_DATA state transitions."
---

# CloudWatch Alarm Presets

Ready-to-deploy configuration presets for CloudWatch Alarm. Each preset is a complete manifest you can copy, customize, and deploy.
