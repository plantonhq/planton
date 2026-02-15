---
title: "Presets"
description: "Ready-to-deploy configuration presets for Step Functions"
type: "preset-list"
componentSlug: "step-functions"
componentTitle: "Step Functions"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-standard-workflow"
    rank: "01"
    title: "Preset: Standard Workflow"
    excerpt: "Use this preset for long-running, durable workflows that require exactly-once execution semantics and full execution history. STANDARD state machines are ideal for:"
  - slug: "02-express-workflow"
    rank: "02"
    title: "Preset: Express Workflow"
    excerpt: "Use this preset for high-volume, short-duration workflows that process events at scale. EXPRESS state machines are ideal for:"
  - slug: "03-production-workflow"
    rank: "03"
    title: "Preset: Production Workflow"
    excerpt: "Use this preset for production-grade workflows that require full observability, encryption, and robust error handling. This configuration represents the recommended setup for any workflow handling..."
---

# Step Functions Presets

Ready-to-deploy configuration presets for Step Functions. Each preset is a complete manifest you can copy, customize, and deploy.
