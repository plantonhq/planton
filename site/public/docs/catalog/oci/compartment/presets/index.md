---
title: "Presets"
description: "Ready-to-deploy configuration presets for Compartment"
type: "preset-list"
componentSlug: "compartment"
componentTitle: "Compartment"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-project"
    rank: "01"
    title: "Project Compartment"
    excerpt: "This preset creates a long-lived compartment for a project, team, or workload. The compartment is retained even if the IaC resource is destroyed (`enableDelete: false`), which is OCI's safety..."
  - slug: "02-sandbox"
    rank: "02"
    title: "Sandbox Compartment"
    excerpt: "This preset creates an ephemeral compartment for development, testing, or proof-of-concept work. Unlike the project preset, this compartment is destroyed when the IaC resource is destroyed..."
---

# Compartment Presets

Ready-to-deploy configuration presets for Compartment. Each preset is a complete manifest you can copy, customize, and deploy.
