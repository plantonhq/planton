# AwsCodePipeline

Deploy and manage AWS CodePipeline continuous delivery pipelines that orchestrate build, test, and deploy phases through an ordered sequence of stages and actions.

## Overview

AWS CodePipeline is a fully managed CI/CD orchestration service that automates your release process. A pipeline is an ordered sequence of stages — each containing one or more actions that perform tasks such as fetching source code, running builds, executing tests, requiring manual approval, or deploying to production environments.

This component creates a CodePipeline pipeline with full support for:
- **Stages and actions** — source, build, test, deploy, approval, invoke, and compute (inline shell commands without a build project)
- **Artifact stores** — S3-backed storage for passing artifacts between stages
- **V2 features** — git-based triggers for automatic execution, pipeline-level variables for parameterization, and advanced execution modes (QUEUED, PARALLEL)
- **Stage conditions** — entry gates (before_entry), post-success verification (on_success), and failure handling (on_failure) with automatic rollback or retry, driven by managed rules (DeploymentWindow, CloudWatchAlarm, LambdaInvoke, VariableCheck, Commands)
- **Cross-region and cross-account** — actions that execute in different regions or assume roles in different accounts

**Bundled resources:**
- **Pipeline** — the top-level orchestration unit with stages, actions, and artifact stores
- **Triggers** (V2) — automatic execution rules based on git push/PR events
- **Variables** (V2) — pipeline-level parameters referenced in action configurations

**Not included:** Webhooks (legacy V1 mechanism, superseded by triggers) and custom action types (account-level resources with independent lifecycles).

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
apiVersion: aws.planton.dev/v1alpha1
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
apiVersion: aws.planton.dev/v1alpha1
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
| `beforeEntry` | object | No | Entry gate: rules that must pass before the stage starts (V2 only) |
| `onSuccess` | object | No | Post-success verification rules (V2 only) |
| `onFailure` | object | No | Failure handling: ROLLBACK / RETRY / rule-gated (V2 only) |

### Stage Condition (`beforeEntry` / `onSuccess` / `onFailure.condition`)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `result` | string | No | Outcome when rules do NOT pass: `FAIL` (default) or `SKIP` |
| `rules` | list | **Yes** | 1-5 checks, AND'd together |

### Rule

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Rule name, unique within the condition |
| `ruleTypeId.provider` | string | **Yes** | Managed rule: `DeploymentWindow`, `CloudWatchAlarm`, `LambdaInvoke`, `VariableCheck`, `Commands` |
| `ruleTypeId.category` / `.owner` / `.version` | string | No | Default `Rule` / `AWS` / current version |
| `configuration` | map | No | Rule-provider-specific keys (e.g., `Cron` + `TimeZone`, `AlarmName` + `WaitTime`) |
| `commands` | list | No | Shell commands for the `Commands` rule (max 50) |
| `inputArtifacts` | list | No | Artifacts available to the rule |
| `region` / `roleArn` | — | No | Cross-region / scoped-role execution |
| `timeoutInMinutes` | int | No | Rule timeout (5-86400) |

### Failure Handling (`onFailure`)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `result` | string | No | `ROLLBACK` (revert to last successful state), `RETRY`, or `FAIL` |
| `retryConfiguration.retryMode` | string | With RETRY | `FAILED_ACTIONS` (default) or `ALL_ACTIONS` |
| `condition` | object | No | Rules gating the failure handling |

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
| `outputArtifacts` | list | No | Artifact names produced for downstream stages/actions (every category EXCEPT `Compute`) |
| `commands` | list | No | Shell commands run by a `Compute` action (max 50, each 1-1000 chars; bills CodeBuild compute per execution) |
| `outputVariables` | list | No | Variable names a `Compute` action exports (max 15; referenced via `namespace`) |
| `outputArtifactsForComputeAction` | list | No | File-based output artifacts of a `Compute` action (`name` + `files`; mutually exclusive with `outputArtifacts`) |
| `namespace` | string | No | Variable namespace for output variables (`#{namespace.Var}`) |
| `region` | string | No | AWS region for cross-region actions |
| `roleArn` | StringValueOrRef | No | IAM role to assume (for cross-account or scoped permissions) |
| `runOrder` | int | No | Execution order within stage (1-999; same value = parallel) |
| `timeoutInMinutes` | int | No | Timeout override (5-86400 minutes) — AWS supports it ONLY on Manual Approval actions |

### Trigger (V2 Only)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `providerType` | string | **Yes** | Must be `CodeStarSourceConnection` |
| `gitConfiguration.sourceActionName` | string | **Yes** | Name of the source action to trigger from |
| `gitConfiguration.push` | list | Conditional | Push event filters (branch, tag, file path patterns); at least one push or pullRequest filter is required |
| `gitConfiguration.pullRequest` | list | Conditional | Pull request event filters (branch, file path, events `OPEN`/`UPDATED`/`CLOSED`) |

### Variable (V2 Only)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Variable name, referenced as `#{variables.Name}` |
| `defaultValue` | string | No | Default value when not supplied at execution time |
| `description` | string | No | Human-readable explanation of the variable |

## Outputs

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
| `04-gated-deploy-with-rollback` | V2 pipeline: deployment-window entry gate, post-deploy alarm check, automatic rollback |
| `05-inline-compute-checks` | V2 pipeline: Compute action running inline lint/test commands with exported variables and file artifacts — no CodeBuild project |

## Deliberate Exclusions

- **Webhooks (`aws_codepipeline_webhook`)** — the legacy V1 trigger mechanism; V2 native `triggers` supersede it with richer filtering and no shared-secret management.
- **Custom action types** — account-level action-type definitions with independent lifecycles, referenced by any pipeline via `category`/`owner: Custom`/`provider`/`version`.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
