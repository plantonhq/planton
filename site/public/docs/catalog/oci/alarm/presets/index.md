---
title: "Presets"
description: "Ready-to-deploy configuration presets for Alarm"
type: "preset-list"
componentSlug: "alarm"
componentTitle: "Alarm"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-cpu-utilization-critical"
    rank: "01"
    title: "Critical CPU Utilization Alarm"
    excerpt: "This preset creates a monitoring alarm that fires when average CPU utilization on compute instances exceeds 80% for 5 consecutive minutes. The alarm evaluates the `CpuUtilization` metric from the..."
  - slug: "02-multi-threshold-escalation"
    rank: "02"
    title: "Multi-Threshold Escalation Alarm"
    excerpt: "This preset creates a single monitoring alarm with tiered alerting using the override mechanism. The base rule fires a WARNING when CPU exceeds 70% for 5 minutes, while the override escalates to..."
---

# Alarm Presets

Ready-to-deploy configuration presets for Alarm. Each preset is a complete manifest you can copy, customize, and deploy.
