---
title: "Presets"
description: "Ready-to-deploy configuration presets for RAM Policy"
type: "preset-list"
componentSlug: "ram-policy"
componentTitle: "RAM Policy"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-scoped-oss-access"
    rank: "01"
    title: "Scoped OSS Bucket Access"
    excerpt: "This preset creates a custom RAM policy that grants read/write access to a single OSS bucket and its objects. System policies like `AliyunOSSFullAccess` grant access to every bucket in the account --..."
  - slug: "02-cicd-deploy-pipeline"
    rank: "02"
    title: "CI/CD Deploy Pipeline"
    excerpt: "This preset creates a custom RAM policy combining the minimal permissions a CI/CD pipeline needs to build container images, deploy to ACK clusters, and write build logs to SLS. No single system..."
---

# RAM Policy Presets

Ready-to-deploy configuration presets for RAM Policy. Each preset is a complete manifest you can copy, customize, and deploy.
