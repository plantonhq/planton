---
title: "Presets"
description: "Ready-to-deploy configuration presets for Log Group"
type: "preset-list"
componentSlug: "log-group"
componentTitle: "Log Group"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-vcn-flow-logs"
    rank: "01"
    title: "VCN Flow Logs"
    excerpt: "This preset creates a log group with a service log that automatically collects VCN flow log data from a subnet. Flow logs capture metadata about every network packet accepted or rejected by security..."
  - slug: "02-custom-application-logs"
    rank: "02"
    title: "Custom Application Logs"
    excerpt: "This preset creates a log group with a custom log that accepts application-level log entries pushed via the OCI Logging Ingestion API. Unlike service logs that are auto-collected from OCI..."
---

# Log Group Presets

Ready-to-deploy configuration presets for Log Group. Each preset is a complete manifest you can copy, customize, and deploy.
