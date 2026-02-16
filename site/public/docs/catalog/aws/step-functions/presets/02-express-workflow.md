---
title: "Preset: Express Workflow"
description: "Use this preset for high-volume, short-duration workflows that process events at scale. EXPRESS state machines are ideal for:"
type: "preset"
rank: "02"
presetSlug: "02-express-workflow"
componentSlug: "step-functions"
componentTitle: "Step Functions"
provider: "aws"
icon: "package"
order: 2
---

# Preset: Express Workflow

## When to Use

Use this preset for high-volume, short-duration workflows that process events at scale. EXPRESS state machines are ideal for:

- Event-driven processing pipelines (EventBridge → Step Functions → Lambda)
- IoT data ingestion and routing
- Real-time stream processing with branching logic
- High-throughput API request orchestration

## Key Configuration Choices

- **Type**: `EXPRESS` — supports up to 5 minutes execution, at-most-once semantics
- **Definition**: Event routing pattern with Choice state — a common Express workflow pattern
- **Logging**: `ERROR` level — captures failures without the volume of `ALL` logging
- **No encryption**: Express workflows are typically for transient data; add encryption if processing sensitive data

## What to Customize

1. **`<workflow-name>`** — A descriptive name (e.g., `event-router`)
2. **`<iam-execution-role-arn>`** — IAM role with permissions for all invoked services
3. **`<enrichment-lambda-arn>`**, **`<urgent-lambda-arn>`**, **`<normal-lambda-arn>`** — Lambda function ARNs
4. **`<cloudwatch-log-group-arn>`** — Log group for error logging
5. **Choice conditions** — Adjust routing logic for your event types

## Express vs Standard

Express workflows trade durability for throughput. Key differences:
- No execution history in the console (use CloudWatch Logs)
- At-most-once execution (no built-in deduplication)
- Priced per execution + duration (cheaper for short, frequent executions)
- Maximum 5 minutes per execution
