---
title: "Presets"
description: "Ready-to-deploy configuration presets for Dynamic Group"
type: "preset-list"
componentSlug: "dynamic-group"
componentTitle: "Dynamic Group"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-compute-instance-principal"
    rank: "01"
    title: "Compute Instance Principal"
    excerpt: "This preset creates a dynamic group that matches all compute instances in a specific compartment. Combined with an `OciIdentityPolicy`, this enables instance principal authentication -- OCI's..."
  - slug: "02-functions-workload-identity"
    rank: "02"
    title: "Functions Workload Identity"
    excerpt: "This preset creates a dynamic group that matches all OCI Functions in a specific compartment. Combined with an `OciIdentityPolicy`, this enables serverless workload identity -- letting Functions call..."
---

# Dynamic Group Presets

Ready-to-deploy configuration presets for Dynamic Group. Each preset is a complete manifest you can copy, customize, and deploy.
