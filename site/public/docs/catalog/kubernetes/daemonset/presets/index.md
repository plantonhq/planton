---
title: "Presets"
description: "Ready-to-deploy configuration presets for DaemonSet"
type: "preset-list"
componentSlug: "daemonset"
componentTitle: "DaemonSet"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-monitoring-agent"
    rank: "01"
    title: "Monitoring Agent DaemonSet"
    excerpt: "This preset deploys a monitoring agent on every node in the cluster, including control-plane nodes. Suitable for node-level metrics collection, log forwarding, or security agents that need to run on..."
  - slug: "02-log-collector"
    rank: "02"
    title: "Log Collector DaemonSet"
    excerpt: "This preset deploys a log collector on every node with host path mounts for `/var/log` and container log directories. Designed for log forwarders like Fluent Bit, Fluentd, or Filebeat that need to..."
---

# DaemonSet Presets

Ready-to-deploy configuration presets for DaemonSet. Each preset is a complete manifest you can copy, customize, and deploy.
