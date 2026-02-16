---
title: "GitHub Source + CodeBuild (CI Pipeline)"
description: "This preset creates a V2 pipeline that fetches source code from a GitHub repository via CodeStar Connection and runs a CodeBuild project. A git push trigger automatically executes the pipeline when..."
type: "preset"
rank: "01"
presetSlug: "01-github-source-codebuild"
componentSlug: "codepipeline"
componentTitle: "CodePipeline"
provider: "aws"
icon: "package"
order: 1
---

# GitHub Source + CodeBuild (CI Pipeline)

This preset creates a V2 pipeline that fetches source code from a GitHub repository via CodeStar Connection and runs a CodeBuild project. A git push trigger automatically executes the pipeline when commits land on the main branch. This is the classic CI pipeline for running tests, linting, and producing build artifacts.

## When to Use

- Standard CI pipeline for GitHub-hosted repositories
- Teams that want automatic pipeline execution on every push to main
- Projects that use CodeBuild for compilation, testing, and artifact generation
- Starting point for more complex pipelines (add approval and deploy stages later)

## Key Configuration Choices

- **V2 pipeline** (`pipelineType`) — Enables git-based triggers and variable namespaces
- **QUEUED execution** (`executionMode`) — Prevents concurrent runs from superseding each other; pushes queue in order
- **CodeStarSourceConnection** (`provider`) — Modern, OAuth-based GitHub integration (replaces legacy GitHub webhooks)
- **CODE_ZIP** (`OutputArtifactFormat`) — Source is downloaded as a zip. Use `CODEBUILD_CLONE_REF` instead if your buildspec needs full git history
- **Git push trigger** — Pipeline executes automatically on push to main; set `DetectChanges: false` in the action configuration if using triggers exclusively
- **SourceVariables namespace** — Exposes `#{SourceVariables.CommitId}`, `#{SourceVariables.BranchName}`, etc. for downstream actions

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<codepipeline-service-role-arn>` | IAM role ARN granting CodePipeline permissions for S3, CodeStar, and CodeBuild | AWS IAM console or `AwsIamRole` status outputs |
| `<pipeline-artifacts-bucket-name>` | S3 bucket name for pipeline artifact storage | AWS S3 console or `AwsS3Bucket` status outputs |
| `<codestar-connection-arn>` | CodeStar Connection ARN for GitHub access | AWS CodePipeline → Settings → Connections |
| `<github-org>/<github-repo>` | GitHub repository in `owner/repo` format (e.g., `my-org/my-app`) | GitHub repository URL |
| `<codebuild-project-name>` | Name of the CodeBuild project to run | AWS CodeBuild console or `AwsCodeBuildProject` status outputs |

## Related Presets

- **02-ecr-ecs-deploy** — Use when deploying container images to ECS after build
- **03-s3-lambda-deploy** — Use for serverless Lambda deployments from S3 source
