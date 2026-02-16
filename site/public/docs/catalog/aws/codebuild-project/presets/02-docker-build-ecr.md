---
title: "Docker Build with ECR Push"
description: "This preset creates a CodeBuild project optimized for building Docker images and pushing them to Amazon ECR. Privileged mode enables Docker daemon access inside the build container. Local Docker..."
type: "preset"
rank: "02"
presetSlug: "02-docker-build-ecr"
componentSlug: "codebuild-project"
componentTitle: "CodeBuild Project"
provider: "aws"
icon: "package"
order: 2
---

# Docker Build with ECR Push

This preset creates a CodeBuild project optimized for building Docker images and pushing them to Amazon ECR. Privileged mode enables Docker daemon access inside the build container. Local Docker layer caching speeds up builds by reusing cached layers across consecutive runs.

## When to Use

- Container-based applications that need Docker image builds
- CI/CD pipelines that push images to ECR
- Projects using multi-stage Docker builds that benefit from layer caching
- Teams that want build times under 5 minutes for incremental builds

## Key Configuration Choices

- **privilegedMode: true** — Required for running `docker build` inside the container
- **BUILD_GENERAL1_LARGE** (`computeType`) — 15 GB memory, 8 vCPUs; Docker builds benefit from larger compute
- **LOCAL_DOCKER_LAYER_CACHE** — Caches Docker layers between builds (significant speedup)
- **LOCAL_SOURCE_CACHE** — Caches Git metadata for faster source fetches
- **gitCloneDepth: 1** — Shallow clone since Docker builds only need the current commit
- **buildTimeout: 30** — Docker builds should complete in under 30 minutes
- **DOCKER_BUILDKIT=1** — Enables BuildKit for better caching and parallel stage execution

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<github-repo-https-url>` | GitHub repository HTTPS URL | GitHub repository settings |
| `<ecr-repository-uri>` | ECR repository URI (e.g., `123456789012.dkr.ecr.us-east-1.amazonaws.com/my-app`) | AWS ECR console or `AwsEcrRepo` status outputs |
| `<codebuild-service-role-arn>` | IAM role ARN with ECR push and CloudWatch permissions | AWS IAM console or `AwsIamRole` status outputs |

## Related Presets

- **01-github-ci-linux** — Use instead for CI-only (lint/test) without Docker
- **03-codepipeline-stage** — Use instead when Docker builds are orchestrated by CodePipeline
