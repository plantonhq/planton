---
title: "Presets"
description: "Ready-to-deploy configuration presets for EventBridge Rule"
type: "preset-list"
componentSlug: "eventbridge-rule"
componentTitle: "EventBridge Rule"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-schedule-lambda"
    rank: "01"
    title: "Scheduled Lambda"
    excerpt: "This preset creates a scheduled EventBridge rule that triggers a Lambda function on a recurring schedule. It is the serverless replacement for traditional cron jobs — no servers to manage, no crontab..."
  - slug: "02-event-pattern-sqs"
    rank: "02"
    title: "Event Pattern to SQS"
    excerpt: "This preset creates an event-pattern-based EventBridge rule that routes matching AWS events to an SQS queue with a dead letter queue for failed deliveries. It demonstrates the core event-driven..."
  - slug: "03-multi-target-fanout"
    rank: "03"
    title: "Multi-Target Fan-Out"
    excerpt: "This preset creates an event-pattern-based rule on a custom event bus that routes matching events to two targets simultaneously — a Lambda function with input transformation and retry policy, and an..."
---

# EventBridge Rule Presets

Ready-to-deploy configuration presets for EventBridge Rule. Each preset is a complete manifest you can copy, customize, and deploy.
