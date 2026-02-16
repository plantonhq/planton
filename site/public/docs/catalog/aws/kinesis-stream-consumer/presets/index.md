---
title: "Presets"
description: "Ready-to-deploy configuration presets for Kinesis Stream Consumer"
type: "preset-list"
componentSlug: "kinesis-stream-consumer"
componentTitle: "Kinesis Stream Consumer"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-basic-consumer"
    rank: "01"
    title: "Preset: Basic Consumer"
    excerpt: "Register an enhanced fan-out consumer with an existing Kinesis stream using a direct ARN. Suitable for quick setup when the stream ARN is known and not managed by OpenMCF."
  - slug: "02-stream-reference"
    rank: "02"
    title: "Preset: Stream Reference (valueFrom)"
    excerpt: "Register an enhanced fan-out consumer with an OpenMCF-managed Kinesis stream using a `valueFrom` reference. The platform resolves the stream ARN at deployment time and creates a dependency edge in..."
---

# Kinesis Stream Consumer Presets

Ready-to-deploy configuration presets for Kinesis Stream Consumer. Each preset is a complete manifest you can copy, customize, and deploy.
