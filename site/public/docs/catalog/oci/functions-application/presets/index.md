---
title: "Presets"
description: "Ready-to-deploy configuration presets for Functions Application"
type: "preset-list"
componentSlug: "functions-application"
componentTitle: "Functions Application"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-standard-x86"
    rank: "01"
    title: "Standard x86"
    excerpt: "This preset creates an OCI Functions Application with x86 processor architecture, NSG-protected networking, and APM tracing enabled. Functions deployed to this application run on Intel/AMD x86-64..."
  - slug: "02-secure-production"
    rank: "02"
    title: "Secure Production"
    excerpt: "This preset creates an OCI Functions Application with container image signature verification enforced, NSG-protected networking, and APM tracing. Only container images signed by the specified KMS key..."
---

# Functions Application Presets

Ready-to-deploy configuration presets for Functions Application. Each preset is a complete manifest you can copy, customize, and deploy.
