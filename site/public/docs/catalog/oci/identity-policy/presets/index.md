---
title: "Presets"
description: "Ready-to-deploy configuration presets for Identity Policy"
type: "preset-list"
componentSlug: "identity-policy"
componentTitle: "Identity Policy"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-compartment-admin"
    rank: "01"
    title: "Compartment Admin Policy"
    excerpt: "This preset creates an IAM policy granting a group full administrative access to all resources within a compartment. This is the most common OCI policy pattern -- the first thing every team creates..."
  - slug: "02-service-access"
    rank: "02"
    title: "Dynamic Group Service Access Policy"
    excerpt: "This preset creates an IAM policy granting a dynamic group access to specific OCI services. Dynamic groups are OCI's workload identity mechanism -- they let compute instances, OKE pods, and Functions..."
  - slug: "03-read-only-auditor"
    rank: "03"
    title: "Read-Only Auditor Policy"
    excerpt: "This preset creates a tenancy-level IAM policy granting a group read-only visibility across all compartments. The `inspect` verb allows listing and viewing resource metadata without accessing data..."
---

# Identity Policy Presets

Ready-to-deploy configuration presets for Identity Policy. Each preset is a complete manifest you can copy, customize, and deploy.
