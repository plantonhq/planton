---
title: "Presets"
description: "Ready-to-deploy configuration presets for IAM User"
type: "preset-list"
componentSlug: "iam-user"
componentTitle: "IAM User"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-ci-cd-pipeline"
    rank: "01"
    title: "CI/CD Pipeline User"
    excerpt: "This preset creates an IAM user for CI/CD pipelines with permissions to push container images to ECR, read S3 artifacts, and manage ECS service deployments. Access keys are created by default,..."
  - slug: "02-read-only-service"
    rank: "02"
    title: "Read-Only Service User"
    excerpt: "This preset creates an IAM user with broad read-only access and no access keys. This is suitable for monitoring integrations, audit tools, or identity-only users where credentials are managed..."
---

# IAM User Presets

Ready-to-deploy configuration presets for IAM User. Each preset is a complete manifest you can copy, customize, and deploy.
