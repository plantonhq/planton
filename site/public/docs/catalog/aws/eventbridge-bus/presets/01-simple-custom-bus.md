---
title: "Simple Custom Bus"
description: "This preset creates a minimal custom EventBridge event bus with a description and AWS-managed encryption. It is the fastest way to get a dedicated event bus running for application event routing."
type: "preset"
rank: "01"
presetSlug: "01-simple-custom-bus"
componentSlug: "eventbridge-bus"
componentTitle: "EventBridge Bus"
provider: "aws"
icon: "package"
order: 1
---

# Simple Custom Bus

This preset creates a minimal custom EventBridge event bus with a description and AWS-managed encryption. It is the fastest way to get a dedicated event bus running for application event routing.

## When to Use

- Isolating application events from the default AWS event bus
- Building event-driven microservices that communicate via custom events
- Setting up a development or staging event bus for testing event routing
- Any scenario where you need a dedicated bus without advanced security requirements

## Key Configuration Choices

- **Description** — provides human-readable context for the bus in the AWS console and API responses
- **AWS-managed encryption** — events are encrypted at rest with an AWS-owned key at no additional cost (this is the default when `kmsKeyIdentifier` is not set)

## Placeholders to Replace

This preset uses a generic `my-events` name. Rename `metadata.name` to match your domain (e.g., `order-events`, `payment-events`, `notification-events`).

## Common Additions

- Add `kmsKeyIdentifier` with a reference to an AwsKmsKey for customer-managed encryption
- Add `deadLetterConfig` with an SQS queue to catch undeliverable events
- Add `logConfig` with level `ERROR` for production observability

## Related Presets

- **02-production-encrypted-bus** — use when you need KMS encryption, DLQ, and logging for production workloads
- **03-partner-event-bus** — use when integrating with a SaaS partner via EventBridge
