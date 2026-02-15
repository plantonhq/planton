---
title: "Presets"
description: "Ready-to-deploy configuration presets for EventBridge Bus"
type: "preset-list"
componentSlug: "eventbridge-bus"
componentTitle: "EventBridge Bus"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-simple-custom-bus"
    rank: "01"
    title: "Simple Custom Bus"
    excerpt: "This preset creates a minimal custom EventBridge event bus with a description and AWS-managed encryption. It is the fastest way to get a dedicated event bus running for application event routing."
  - slug: "02-production-encrypted-bus"
    rank: "02"
    title: "Production Encrypted Bus"
    excerpt: "This preset creates a production-grade EventBridge custom event bus with customer-managed KMS encryption, a dead letter queue for undeliverable events, and error-level logging. Designed for workloads..."
  - slug: "03-partner-event-bus"
    rank: "03"
    title: "Partner Event Bus"
    excerpt: "This preset creates an EventBridge custom event bus for receiving events from a SaaS partner integration. When a partner event source is configured in your AWS account, you create a bus with the same..."
---

# EventBridge Bus Presets

Ready-to-deploy configuration presets for EventBridge Bus. Each preset is a complete manifest you can copy, customize, and deploy.
