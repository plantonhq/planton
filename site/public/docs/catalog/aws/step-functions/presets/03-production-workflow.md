---
title: "Preset: Production Workflow"
description: "Use this preset for production-grade workflows that require full observability, encryption, and robust error handling. This configuration represents the recommended setup for any workflow handling..."
type: "preset"
rank: "03"
presetSlug: "03-production-workflow"
componentSlug: "step-functions"
componentTitle: "Step Functions"
provider: "aws"
icon: "package"
order: 3
---

# Preset: Production Workflow

## When to Use

Use this preset for production-grade workflows that require full observability, encryption, and robust error handling. This configuration represents the recommended setup for any workflow handling business-critical data.

## Key Configuration Choices

- **Type**: `STANDARD` — durable, exactly-once execution with full history
- **Logging**: `ALL` with execution data — full visibility into every state transition
- **Tracing**: Enabled — X-Ray traces for end-to-end request visualization
- **Encryption**: Customer-managed KMS key — compliance-ready data encryption
- **Error handling**: Retry with exponential backoff + catch-all error handler
- **Cross-resource references**: `valueFrom` for IAM role, log group, and KMS key

## What to Customize

1. **`<workflow-name>`** — Production workflow name (e.g., `order-processor`)
2. **`<workflow-description>`** — Clear description of the workflow's purpose
3. **`<iam-role-resource-name>`** — Name of the AwsIamRole resource in your infra chart
4. **`<lambda-function-arn>`** and **`<lambda-function-arn-2>`** — Lambda functions for each step
5. **`<error-handler-lambda-arn>`** — Dedicated error handling function
6. **`<log-group-resource-name>`** — Name of the AwsCloudwatchLogGroup resource
7. **`<kms-key-resource-name>`** — Name of the AwsKmsKey resource
8. **Definition** — Replace with your actual multi-step workflow

## Production Checklist

- [ ] IAM execution role follows least-privilege principle
- [ ] All Task states have `Retry` blocks for transient failures
- [ ] Critical paths have `Catch` blocks routing to error handlers
- [ ] Error handler sends alerts (SNS, PagerDuty, etc.)
- [ ] Logging set to `ALL` for full audit trail
- [ ] X-Ray tracing enabled for performance monitoring
- [ ] Customer-managed KMS key for data encryption
- [ ] CloudWatch alarms configured for failed executions
