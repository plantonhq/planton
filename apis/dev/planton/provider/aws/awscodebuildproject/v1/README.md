# AwsCodeBuildProject

Deploy and manage AWS CodeBuild projects with optional webhook triggers for automated CI/CD builds.

## Overview

AWS CodeBuild is a fully managed build service that compiles source code, runs tests, and produces deployable artifacts. This component creates a CodeBuild project — the build configuration unit — defining where source code comes from, how to build it, and where to put the output. An optional webhook enables automatic build triggers from source providers like GitHub, Bitbucket, or GitLab.

**Supported source types:**

| Source | Description | Use Case |
|--------|-------------|----------|
| `GITHUB` | GitHub.com repository | Most common for open-source and SaaS teams |
| `BITBUCKET` | Bitbucket Cloud repository | Atlassian ecosystem |
| `CODECOMMIT` | AWS CodeCommit repository | Fully AWS-native CI/CD |
| `CODEPIPELINE` | Source provided by CodePipeline | Stage in a multi-step pipeline |
| `S3` | S3 bucket containing source archive | Artifact-based builds |
| `GITLAB` | GitLab.com or self-managed | GitLab ecosystem |
| `NO_SOURCE` | No source — inline buildspec | Utility/maintenance builds |

**Bundled resources:**
- **CodeBuild project** — build configuration (source, environment, artifacts, logs)
- **Webhook** (optional) — automatic build triggers from source providers

**Not included:** Source credentials (account-level), report groups (shared), fleets (shared), and resource policies (niche).

## Prerequisites

- An IAM service role for CodeBuild with appropriate permissions (use `AwsIamRole`)
- For VPC builds: a VPC with private subnets and security groups (use `AwsVpc`, `AwsSecurityGroup`)
- For S3 artifacts/cache: an S3 bucket (use `AwsS3Bucket`)
- For encrypted artifacts: a KMS key (use `AwsKmsKey`)

## Quick Start

### Minimal (GitHub CI)

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCodeBuildProject
metadata:
  name: my-app-ci
spec:
  source:
    type: GITHUB
    location: https://github.com/my-org/my-app.git
    reportBuildStatus: true
  environment:
    type: LINUX_CONTAINER
    computeType: BUILD_GENERAL1_SMALL
    image: aws/codebuild/amazonlinux2-x86_64-standard:5.0
  artifacts:
    type: NO_ARTIFACTS
  serviceRole:
    value: arn:aws:iam::123456789012:role/codebuild-service-role
  webhook:
    filterGroups:
      - filters:
          - type: EVENT
            pattern: PUSH, PULL_REQUEST_CREATED, PULL_REQUEST_UPDATED
          - type: HEAD_REF
            pattern: ^refs/heads/main$
```

### Production (Docker Build with Caching)

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCodeBuildProject
metadata:
  name: docker-build
spec:
  source:
    type: GITHUB
    location: https://github.com/my-org/my-app.git
    gitCloneDepth: 1
    reportBuildStatus: true
  environment:
    type: LINUX_CONTAINER
    computeType: BUILD_GENERAL1_LARGE
    image: aws/codebuild/amazonlinux2-x86_64-standard:5.0
    privilegedMode: true
    environmentVariables:
      - name: ECR_REPO
        value: 123456789012.dkr.ecr.us-east-1.amazonaws.com/my-app
      - name: DOCKER_BUILDKIT
        value: "1"
  artifacts:
    type: NO_ARTIFACTS
  serviceRole:
    value: arn:aws:iam::123456789012:role/codebuild-service-role
  buildTimeout: 30
  cache:
    type: LOCAL
    modes:
      - LOCAL_DOCKER_LAYER_CACHE
      - LOCAL_SOURCE_CACHE
  logsConfig:
    cloudwatchLogs:
      status: ENABLED
      groupName:
        value: /aws/codebuild/docker-build
  webhook:
    filterGroups:
      - filters:
          - type: EVENT
            pattern: PUSH
          - type: HEAD_REF
            pattern: ^refs/heads/(main|release/.*)$
```

## Spec Fields

### Top-Level

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `source` | object | **Yes** | — | Source code location and retrieval settings |
| `environment` | object | **Yes** | — | Build container configuration |
| `artifacts` | object | **Yes** | — | Build output configuration |
| `serviceRole` | StringValueOrRef | **Yes** | — | IAM service role ARN |
| `description` | string | No | — | Project description (max 255 chars) |
| `encryptionKey` | StringValueOrRef | No | AWS-managed key | KMS key for artifact encryption |
| `buildTimeout` | int | No | 60 | Build timeout in minutes (5-2160) |
| `queuedTimeout` | int | No | 480 | Queue timeout in minutes (5-480) |
| `concurrentBuildLimit` | int | No | Unlimited | Max concurrent builds (min 1) |
| `sourceVersion` | string | No | — | Default branch/tag/commit to build |
| `cache` | object | No | NO_CACHE | Build caching configuration |
| `logsConfig` | object | No | CloudWatch ENABLED | Logging configuration |
| `vpcConfig` | object | No | — | VPC placement for private resource access |
| `webhook` | object | No | — | Automatic build trigger configuration |

### Source

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | **Yes** | Source provider (GITHUB, BITBUCKET, CODECOMMIT, etc.) |
| `location` | string | Conditional | Repository URL or S3 path (required unless CODEPIPELINE/NO_SOURCE) |
| `buildspec` | string | Conditional | Buildspec file or inline YAML (required for NO_SOURCE) |
| `gitCloneDepth` | int | No | Git clone depth (0 = full clone) |
| `reportBuildStatus` | bool | No | Report status back to source provider |
| `fetchSubmodules` | bool | No | Fetch Git submodules |

### Environment

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | **Yes** | Container type (LINUX_CONTAINER, ARM_CONTAINER, etc.) |
| `computeType` | string | **Yes** | Compute capacity (BUILD_GENERAL1_SMALL, etc.) |
| `image` | string | **Yes** | Docker image identifier |
| `privilegedMode` | bool | No | Enable Docker daemon access |
| `imagePullCredentialsType` | string | No | CODEBUILD (default) or SERVICE_ROLE |
| `environmentVariables` | list | No | Build environment variables |
| `registryCredential` | object | No | Private registry credentials |

### Artifacts

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | **Yes** | Output type (NO_ARTIFACTS, S3, CODEPIPELINE) |
| `location` | StringValueOrRef | Conditional | S3 bucket name (required for S3 type) |
| `name` | string | No | Output artifact name |
| `path` | string | No | S3 prefix path |
| `packaging` | string | No | NONE or ZIP |
| `namespaceType` | string | No | NONE or BUILD_ID |
| `encryptionDisabled` | bool | No | Disable artifact encryption |

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `project_arn` | string | ARN of the CodeBuild project |
| `project_name` | string | Name of the CodeBuild project |
| `service_role_arn` | string | IAM service role ARN |
| `webhook_url` | string | Webhook URL (empty if no webhook) |
| `webhook_payload_url` | string | Webhook payload URL (empty if no webhook) |

## Presets

| Preset | Description |
|--------|-------------|
| `01-github-ci-linux` | GitHub source, Linux container, CI-only (no artifacts), webhook |
| `02-docker-build-ecr` | GitHub source, privileged mode, Docker layer caching |
| `03-codepipeline-stage` | CODEPIPELINE source + artifacts, designed as a pipeline stage |

## Deferred to v2

- Secondary sources and secondary artifacts (multi-input/output builds)
- Build batch configuration (parallel test execution)
- EFS file system locations
- Fleet references (reserved compute capacity)
- Project visibility (PUBLIC_READ for open-source badge/logs)
- Build badge support
- Auto-retry configuration
- Webhook scope configuration (GitHub Apps)
- Webhook pull request build policy
