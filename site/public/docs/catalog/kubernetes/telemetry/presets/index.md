---
title: "Presets"
description: "Ready-to-deploy configuration presets for Telemetry"
type: "preset-list"
componentSlug: "telemetry"
componentTitle: "Telemetry"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-mesh-tracing-sampling"
    rank: "01"
    title: "Mesh-Wide Trace Sampling"
    excerpt: "The canonical Telemetry resource: turn on distributed-tracing sampling for the whole mesh and (optionally) stamp every span with an operator-supplied tag. This is how you get traces flowing to your..."
  - slug: "02-prometheus-metric-dimensions"
    rank: "02"
    title: "Prometheus Metric Dimensions"
    excerpt: "Customize the dimensions (labels) Istio attaches to its Prometheus metrics for a namespace or workload: add high-value tags (request host, method) and drop noisy/high-cardinality ones (response..."
---

# Telemetry Presets

Ready-to-deploy configuration presets for Telemetry. Each preset is a complete manifest you can copy, customize, and deploy.
