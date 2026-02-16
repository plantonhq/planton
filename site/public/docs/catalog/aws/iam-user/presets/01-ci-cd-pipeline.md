---
title: "CI/CD Pipeline User"
description: "This preset creates an IAM user for CI/CD pipelines with permissions to push container images to ECR, read S3 artifacts, and manage ECS service deployments. Access keys are created by default,..."
type: "preset"
rank: "01"
presetSlug: "01-ci-cd-pipeline"
componentSlug: "iam-user"
componentTitle: "IAM User"
provider: "aws"
icon: "package"
order: 1
---

# CI/CD Pipeline User

This preset creates an IAM user for CI/CD pipelines with permissions to push container images to ECR, read S3 artifacts, and manage ECS service deployments. Access keys are created by default, providing the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` that your CI/CD system needs.

## When to Use

- GitHub Actions, GitLab CI, Jenkins, or other CI/CD systems that need AWS credentials for automated deployments
- Pipelines that build and push Docker images to ECR, then deploy to ECS
- Any automation that requires long-lived AWS credentials (prefer IAM roles where possible)

## Key Configuration Choices

- **ECR power user** (`AmazonEC2ContainerRegistryPowerUser`) -- Push, pull, and manage images in any ECR repository in the account
- **S3 read-only** (`AmazonS3ReadOnlyAccess`) -- Read build artifacts, configuration files, or deployment packages from S3
- **ECS deployment inline policy** -- Scoped permissions for describing and updating ECS services and running tasks
- **Access keys enabled** (`disableAccessKeys` not set) -- Creates an active access key pair for programmatic access

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<ci-cd-user-name>` | IAM user name (e.g., `github-actions-deployer`); must match `[a-zA-Z0-9+=,.@_-]{1,64}` | Your CI/CD naming convention |

## Related Presets

- **02-read-only-service** -- Use instead for monitoring or audit tools that only need read access
