---
title: "Production Immutable ECR Repository"
description: "This preset creates an ECR repository with immutable image tags, automatic vulnerability scanning, and a lifecycle policy that balances cost control with rollback capability. Immutable tags ensure..."
type: "preset"
rank: "01"
presetSlug: "01-production-immutable"
componentSlug: "ecr-repo"
componentTitle: "ECR Repo"
provider: "aws"
icon: "package"
order: 1
---

# Production Immutable ECR Repository

This preset creates an ECR repository with immutable image tags, automatic vulnerability scanning, and a lifecycle policy that balances cost control with rollback capability. Immutable tags ensure that production image references are stable and cannot be accidentally overwritten by a CI/CD pipeline.

## When to Use

- Production container registries where image tag integrity is critical
- Regulated environments requiring immutable artifacts for audit trails
- Any CI/CD pipeline pushing Docker images to ECR for production workloads

## Key Configuration Choices

- **Immutable tags** (`imageImmutable: true`) -- Once pushed, a tag cannot be overwritten; guarantees that `v1.2.3` always refers to the same image
- **Scan on push** (`scanOnPush: true`) -- Automatic vulnerability scanning via Amazon Inspector when images are pushed
- **AES256 encryption** (`encryptionType: AES256`) -- AWS-managed server-side encryption at rest (default, no additional cost)
- **7-day untagged expiration** (`expireUntaggedAfterDays: 7`) -- Removes orphaned layers and failed builds quickly to control costs
- **100 image retention** (`maxImageCount: 100`) -- Keeps recent images available for rollback; older images are automatically expired
- **Force delete disabled** (`forceDelete: false`) -- Repository cannot be deleted while it contains images

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<repository-name>` | ECR repository name (e.g., `myorg/api-service` or `team-blue/frontend`) | Your team's container image naming convention |
