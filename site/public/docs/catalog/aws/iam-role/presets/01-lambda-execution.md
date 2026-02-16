---
title: "Lambda Execution Role"
description: "This preset creates an IAM role that Lambda functions can assume for execution. It includes the `AWSLambdaBasicExecutionRole` managed policy, which grants permissions to write logs to CloudWatch..."
type: "preset"
rank: "01"
presetSlug: "01-lambda-execution"
componentSlug: "iam-role"
componentTitle: "IAM Role"
provider: "aws"
icon: "package"
order: 1
---

# Lambda Execution Role

This preset creates an IAM role that Lambda functions can assume for execution. It includes the `AWSLambdaBasicExecutionRole` managed policy, which grants permissions to write logs to CloudWatch Logs. Add additional managed or inline policies for access to other AWS services (DynamoDB, S3, SQS, etc.).

## When to Use

- Any Lambda function that needs to write execution logs to CloudWatch
- Starting point for Lambda roles; add more policies based on what AWS services the function accesses
- Functions deployed via `AwsLambda` that reference this role's ARN

## Key Configuration Choices

- **Lambda service principal** (`lambda.amazonaws.com`) -- Only Lambda can assume this role
- **Basic execution policy** (`AWSLambdaBasicExecutionRole`) -- Grants `logs:CreateLogGroup`, `logs:CreateLogStream`, `logs:PutLogEvents`
- **Root path** (`path: /`) -- Default IAM path; use `/service-roles/` or a custom path for organizational grouping

## Placeholders to Replace

This preset has no placeholders. Deploy as-is for a minimal Lambda execution role. Add `managedPolicyArns` entries or `inlinePolicies` for additional service access.

## Related Presets

- **02-ecs-task-execution** -- Use instead for ECS task execution roles
- **03-ec2-ssm** -- Use instead for EC2 instances needing SSM access
