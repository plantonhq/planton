---
title: "Presets"
description: "Ready-to-deploy configuration presets for CloudWatch Log Group"
type: "preset-list"
componentSlug: "cloudwatch-log-group"
componentTitle: "CloudWatch Log Group"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-standard-retention-30d"
    rank: "01"
    title: "Preset: Standard 30-Day Retention"
    excerpt: "**Use case:** General-purpose application logging with a sensible retention period."
  - slug: "02-encrypted-retention-90d"
    rank: "02"
    title: "Preset: Encrypted 90-Day Retention"
    excerpt: "**Use case:** Production application logging with KMS encryption and 90-day retention for compliance."
  - slug: "03-infrequent-access-long-retention"
    rank: "03"
    title: "Preset: Infrequent Access Long Retention"
    excerpt: "**Use case:** High-volume logs with long retention at reduced cost."
---

# CloudWatch Log Group Presets

Ready-to-deploy configuration presets for CloudWatch Log Group. Each preset is a complete manifest you can copy, customize, and deploy.
