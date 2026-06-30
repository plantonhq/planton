# AwsCodePipeline

Deploy and manage AWS CodePipeline continuous delivery pipelines that orchestrate build, test, and deploy phases through an ordered sequence of stages and actions.

## Overview

AWS CodePipeline is a fully managed CI/CD orchestration service that automates your release process. A pipeline is an ordered sequence of stages — each containing one or more actions that perform tasks such as fetching source code, running builds, executing tests, requiring manual approval, or deploying to production environments.

This component creates a CodePipeline pipeline with full support for:
- **Stages and actions** — source, build, test, deploy, approval, invoke
- **Artifact stores** — S3-backed storage for passing artifacts between stages
- **V2 features** — git-based triggers for automatic execution, pipeline-level variables for parameterization, and advanced execution modes (QUEUED, PARALLEL)
- **Cross-region and cross-account** — actions that execute in different regions or assume roles in different accounts

**Bundled resources:**
- **Pipeline** — the top-level orchestration unit with stages, actions, and artifact stores
- **Triggers** (V2) — automatic execution rules based on git push/PR events
- **Variables** (V2) — pipeline-level parameters referenced in action configurations

**Not included:** Webhooks (legacy V1 mechanism, superseded by triggers), custom action types (account-level resources with independent lifecycles), and stage conditions (before_entry, on_success, on_failure — deferred to v2).

## Prerequisites

- An **IAM role** for CodePipeline with permissions for all action providers used in the pipeline (use `AwsIamRole`)
- An **S3 bucket** for artifact storage between stages (use `AwsS3Bucket`)
- For GitHub/Bitbucket/GitLab sources: a **CodeStar Connection** created and confirmed in the AWS Console
- For CodeBuild actions: an existing **CodeBuild project** (use `AwsCodeBuildProject`)
- For ECS deploy actions: an existing **ECS cluster and service** (use `AwsEcsCluster`, `AwsEcsService`)
- For encrypted artifacts: a **KMS key** (use `AwsKmsKey`)

## Quick Start

### Minimal CI Pipeline (GitHub Source + CodeBuild)

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCodePipeline
metadata:
  name: my-app-ci
spec:
  roleArn:
    value: arn:aws:iam::123456789012:role/codepipeline-service-role
  artifactStores:
    - location:
        value: my-pipeline-artifacts-bucket
  stages:
    - name: Source
      actions:
        - name: GitHubSource
          category: Source
          owner: AWS
          provider: CodeStarSourceConnection
          version: "1"
          configuration:
            ConnectionArn: arn:aws:codeconnections:us-east-1:123456789012:connection/abc12345-def6-7890-ghij-klmnopqrstuv
            FullRepositoryId: my-org/my-app
            BranchName: main
            OutputArtifactFormat: CODE_ZIP
          outputArtifacts:
            - SourceOutput
    - name: Build
      actions:
        - name: CodeBuild
          category: Build
          owner: AWS
          provider: CodeBuild
          version: "1"
          configuration:
            ProjectName: my-app-build
          inputArtifacts:
            - SourceOutput
          outputArtifacts:
            - BuildOutput
```

### V2 Pipeline with Git Triggers

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCodePipeline
metadata:
  name: my-app-ci-triggered
spec:
  pipelineType: V2
  executionMode: QUEUED
  roleArn:
    value: arn:aws:iam::123456789012:role/codepipeline-service-role
  artifactStores:
    - location:
        value: my-pipeline-artifacts-bucket
  stages:
    - name: Source
      actions:
        - name: GitHubSource
          category: Source
          owner: AWS
          provider: CodeStarSourceConnection
          version: "1"
          configuration:
            ConnectionArn: arn:aws:codeconnections:us-east-1:123456789012:connection/abc12345-def6-7890-ghij-klmnopqrstuv
            FullRepositoryId: my-org/my-app
            BranchName: main
          outputArtifacts:
            - SourceOutput
    - name: Build
      actions:
        - name: CodeBuild
          category: Build
          owner: AWS
          provider: CodeBuild
          version: "1"
          configuration:
            ProjectName: my-app-build
          inputArtifacts:
            - SourceOutput
          outputArtifacts:
            - BuildOutput
  triggers:
    - providerType: CodeStarSourceConnection
      gitConfiguration:
        sourceActionName: GitHubSource
        push:
          - branches:
              includes:
                - main
                - "release/*"
            filePaths:
              excludes:
                - "docs/**"
                - "*.md"
```

## Spec Fields

### Top-Level

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `pipelineType` | string | No | `V2` | Pipeline version: `V1` (legacy) or `V2` (modern with triggers/variables) |
| `executionMode` | string | No | `SUPERSEDED` | How concurrent executions are handled: `SUPERSEDED`, `QUEUED` (V2), `PARALLEL` (V2) |
| `roleArn` | StringValueOrRef | **Yes** | — | IAM role ARN granting pipeline permissions for all action providers |
| `artifactStores` | list | **Yes** | — | S3 buckets for artifact storage (min 1; one per region for cross-region) |
| `stages` | list | **Yes** | — | Ordered sequence of pipeline stages (min 2: source + at least one other) |
| `triggers` | list | No | — | Git-based automatic execution triggers (V2 only, max 50) |
| `variables` | list | No | — | Pipeline-level variables referenced as `#{variables.Name}` (V2 only) |

### Artifact Store

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `location` | StringValueOrRef | **Yes** | S3 bucket name for artifact storage |
| `region` | string | No | AWS region (required only for cross-region pipelines) |
| `encryptionKeyId` | StringValueOrRef | No | KMS key ARN/ID for artifact encryption (default: AWS-managed key) |

### Stage

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Stage name, unique within the pipeline (1-100 chars) |
| `actions` | list | **Yes** | Actions to execute in this stage (min 1) |

### Action

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Action name, unique within the stage (1-100 chars) |
| `category` | string | **Yes** | Action type: `Source`, `Build`, `Test`, `Deploy`, `Approval`, `Invoke`, `Compute` |
| `owner` | string | **Yes** | Action creator: `AWS`, `ThirdParty`, `Custom` |
| `provider` | string | **Yes** | Service provider (e.g., `CodeStarSourceConnection`, `CodeBuild`, `S3`, `ECS`) |
| `version` | string | **Yes** | Action type version (typically `"1"`) |
| `configuration` | map | No | Provider-specific key-value pairs controlling action behavior |
| `inputArtifacts` | list | No | Artifact names consumed from previous stages/actions |
| `outputArtifacts` | list | No | Artifact names produced for downstream stages/actions |
| `namespace` | string | No | Variable namespace for output variables (`#{namespace.Var}`) |
| `region` | string | No | AWS region for cross-region actions |
| `roleArn` | StringValueOrRef | No | IAM role to assume (for cross-account or scoped permissions) |
| `runOrder` | int | No | Execution order within stage (1-999; same value = parallel) |
| `timeoutInMinutes` | int | No | Maximum action runtime (5-86400 minutes) |

### Trigger (V2 Only)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `providerType` | string | **Yes** | Must be `CodeStarSourceConnection` |
| `gitConfiguration.sourceActionName` | string | **Yes** | Name of the source action to trigger from |
| `gitConfiguration.push` | list | No | Push event filters (branch, tag, file path patterns) |
| `gitConfiguration.pullRequest` | list | No | Pull request event filters (branch, file path, event type) |

### Variable (V2 Only)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Variable name, referenced as `#{variables.Name}` |
| `defaultValue` | string | No | Default value when not supplied at execution time |
| `description` | string | No | Human-readable explanation of the variable |

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `pipeline_arn` | string | ARN of the CodePipeline pipeline |
| `pipeline_name` | string | Name of the CodePipeline pipeline |

## Presets

| Preset | Description |
|--------|-------------|
| `01-github-source-codebuild` | V2 pipeline: GitHub (CodeStar) source → CodeBuild build. Classic CI pipeline |
| `02-ecr-ecs-deploy` | V2 pipeline: ECR source → CodeBuild → ECS deploy. Container deployment pipeline |
| `03-s3-lambda-deploy` | V2 pipeline: S3 source → Lambda deploy. Serverless deployment pipeline |

## Deferred to v2

- **Stage conditions** — `before_entry`, `on_success`, `on_failure` condition blocks for stage-level gates
- **Webhooks** — Legacy V1 webhook mechanism (superseded by V2 triggers)
- **Custom action types** — Account-level custom action definitions
- **Retry configuration** — Automatic retry on stage/action failure
- **Rollback configuration** — Automatic rollback on stage failure
