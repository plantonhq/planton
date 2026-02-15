---
title: "Presets"
description: "Ready-to-deploy configuration presets for IAM Role"
type: "preset-list"
componentSlug: "iam-role"
componentTitle: "IAM Role"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-lambda-execution"
    rank: "01"
    title: "Lambda Execution Role"
    excerpt: "This preset creates an IAM role that Lambda functions can assume for execution. It includes the `AWSLambdaBasicExecutionRole` managed policy, which grants permissions to write logs to CloudWatch..."
  - slug: "02-ecs-task-execution"
    rank: "02"
    title: "ECS Task Execution Role"
    excerpt: "This preset creates an IAM role that the ECS agent assumes to pull container images from ECR and write logs to CloudWatch on behalf of your tasks. Every ECS Fargate service needs a task execution..."
  - slug: "03-ec2-ssm"
    rank: "03"
    title: "EC2 SSM and CloudWatch Role"
    excerpt: "This preset creates an IAM role for EC2 instances that enables AWS Systems Manager Session Manager for shell access and CloudWatch Agent for metrics and log collection. This is the standard role for..."
---

# IAM Role Presets

Ready-to-deploy configuration presets for IAM Role. Each preset is a complete manifest you can copy, customize, and deploy.
