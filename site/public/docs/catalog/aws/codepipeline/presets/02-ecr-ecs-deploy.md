---
title: "ECR Source + ECS Deploy (Container Deployment Pipeline)"
description: "This preset creates a V2 pipeline that triggers when a new Docker image is pushed to an ECR repository, runs a CodeBuild step to prepare the deployment bundle (generating `imagedefinitions.json`),..."
type: "preset"
rank: "02"
presetSlug: "02-ecr-ecs-deploy"
componentSlug: "codepipeline"
componentTitle: "CodePipeline"
provider: "aws"
icon: "package"
order: 2
---

# ECR Source + ECS Deploy (Container Deployment Pipeline)

This preset creates a V2 pipeline that triggers when a new Docker image is pushed to an ECR repository, runs a CodeBuild step to prepare the deployment bundle (generating `imagedefinitions.json`), and then deploys to an ECS service. This is the standard container deployment pipeline for teams that build images externally (e.g., via GitHub Actions or a separate CodeBuild project) and want automated deployments to ECS.

## When to Use

- Container-based applications deployed to ECS Fargate or EC2
- Teams that push Docker images to ECR from an external CI system
- Decoupled build/deploy pipelines where the build pipeline pushes to ECR and the deploy pipeline picks up new images
- Production deployments that need a consistent, auditable deployment path from ECR to ECS

## Key Configuration Choices

- **ECR source** (`provider: ECR`) — Pipeline triggers automatically when a new image with the specified tag is pushed to the repository
- **CodeBuild preparation stage** — Generates the `imagedefinitions.json` file that the ECS deploy action consumes. The buildspec should output a JSON array like `[{"name":"container-name","imageUri":"..."}]`
- **ECS deploy** (`provider: ECS`) — Performs a rolling update of the ECS service with the new image. ECS creates a new task definition revision and updates the service
- **QUEUED execution** — Ensures deployments are processed in order, preventing a newer image from being superseded by an older one

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<codepipeline-service-role-arn>` | IAM role ARN with permissions for S3, ECR, CodeBuild, and ECS | AWS IAM console or `AwsIamRole` status outputs |
| `<pipeline-artifacts-bucket-name>` | S3 bucket name for pipeline artifact storage | AWS S3 console or `AwsS3Bucket` status outputs |
| `<ecr-repository-name>` | ECR repository name (not URI) to watch for new images | AWS ECR console or `AwsEcrRepo` status outputs |
| `<codebuild-deploy-prep-project>` | CodeBuild project that generates `imagedefinitions.json` | AWS CodeBuild console or `AwsCodeBuildProject` status outputs |
| `<ecs-cluster-name>` | Name of the target ECS cluster | AWS ECS console or `AwsEcsCluster` status outputs |
| `<ecs-service-name>` | Name of the target ECS service to update | AWS ECS console or `AwsEcsService` status outputs |

## Related Presets

- **01-github-source-codebuild** — Use when source is a GitHub repo (not ECR)
- **03-s3-lambda-deploy** — Use for serverless Lambda deployments instead of ECS
