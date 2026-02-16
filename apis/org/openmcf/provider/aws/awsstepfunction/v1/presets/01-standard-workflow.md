# Preset: Standard Workflow

## When to Use

Use this preset for long-running, durable workflows that require exactly-once execution semantics and full execution history. STANDARD state machines are ideal for:

- Business process automation (order processing, approval workflows)
- ETL/ELT pipelines that may run for hours
- Workflows that require visual debugging in the AWS console
- Any workflow where execution auditability is important

## Key Configuration Choices

- **Type**: `STANDARD` — supports workflows up to 1 year, exactly-once execution
- **Definition**: Single Lambda task — replace with your actual ASL workflow
- **Role ARN**: Direct value — replace with your IAM role ARN or use `valueFrom` to reference an AwsIamRole resource

## What to Customize

1. **`<workflow-name>`** — A descriptive name for the state machine (e.g., `order-processor`)
2. **`<iam-execution-role-arn>`** — IAM role ARN with `states.amazonaws.com` trust policy
3. **`<lambda-function-arn>`** — ARN of the Lambda function to invoke
4. **`definition`** — Replace the single-task definition with your actual ASL workflow

## Next Steps

- Add retry logic: `Retry` blocks on Task states for transient failures
- Add error handling: `Catch` blocks to route errors to dedicated states
- Add logging: Include the `logging` block for production observability
- Add tracing: Set `tracingEnabled: true` for X-Ray integration
