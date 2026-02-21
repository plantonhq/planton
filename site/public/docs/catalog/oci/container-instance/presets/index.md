---
title: "Presets"
description: "Ready-to-deploy configuration presets for Container Instance"
type: "preset-list"
componentSlug: "container-instance"
componentTitle: "Container Instance"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-web-service"
    rank: "01"
    title: "Web Service Container Instance"
    excerpt: "This preset creates a single-container OCI Container Instance running an HTTP service with a health check, a public IP for direct access, and an always-restart policy. It uses the CI.Standard.E4.Flex..."
  - slug: "02-private-hardened"
    rank: "02"
    title: "Private Hardened Container Instance"
    excerpt: "This preset creates a production-grade OCI Container Instance in a private subnet with no public IP, full Linux security context hardening, NSG-based network segmentation, and a graceful shutdown..."
  - slug: "03-multi-container-sidecar"
    rank: "03"
    title: "Multi-Container Sidecar"
    excerpt: "This preset creates a multi-container OCI Container Instance with an application container and a log-forwarder sidecar sharing an emptydir volume, plus a configfile volume for injecting configuration..."
---

# Container Instance Presets

Ready-to-deploy configuration presets for Container Instance. Each preset is a complete manifest you can copy, customize, and deploy.
