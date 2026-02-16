---
title: "Presets"
description: "Ready-to-deploy configuration presets for Grafana"
type: "preset-list"
componentSlug: "grafana"
componentTitle: "Grafana"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Grafana"
    excerpt: "This preset deploys Grafana with ingress for external access to dashboards. Grafana provides visualization and alerting for metrics from Prometheus, Elasticsearch, Loki, and many other data sources."
---

# Grafana Presets

Ready-to-deploy configuration presets for Grafana. Each preset is a complete manifest you can copy, customize, and deploy.
