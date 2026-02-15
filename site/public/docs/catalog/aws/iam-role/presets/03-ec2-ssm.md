---
title: "EC2 SSM and CloudWatch Role"
description: "This preset creates an IAM role for EC2 instances that enables AWS Systems Manager Session Manager for shell access and CloudWatch Agent for metrics and log collection. This is the standard role for..."
type: "preset"
rank: "03"
presetSlug: "03-ec2-ssm"
componentSlug: "iam-role"
componentTitle: "IAM Role"
provider: "aws"
icon: "package"
order: 3
---

# EC2 SSM and CloudWatch Role

This preset creates an IAM role for EC2 instances that enables AWS Systems Manager Session Manager for shell access and CloudWatch Agent for metrics and log collection. This is the standard role for production EC2 instances that use SSM instead of SSH.

## When to Use

- EC2 instances managed via `AwsEc2Instance` with `connectionMethod: SSM`
- Any instance that needs CloudWatch custom metrics or log forwarding
- Production instances where SSH key management should be avoided

## Key Configuration Choices

- **EC2 service principal** (`ec2.amazonaws.com`) -- Only EC2 can assume this role
- **SSM core policy** (`AmazonSSMManagedInstanceCore`) -- Enables Session Manager shell access, patch management, and inventory collection
- **CloudWatch Agent policy** (`CloudWatchAgentServerPolicy`) -- Enables the CloudWatch Agent to push custom metrics and logs from the instance

## Placeholders to Replace

This preset has no placeholders. Deploy as-is and reference the role ARN from `AwsEc2Instance.iamInstanceProfileArn`. Note: You'll need to create an IAM instance profile that uses this role (the IaC module handles this automatically).

## Related Presets

- **01-lambda-execution** -- Use instead for Lambda function execution roles
- **02-ecs-task-execution** -- Use instead for ECS task execution roles
