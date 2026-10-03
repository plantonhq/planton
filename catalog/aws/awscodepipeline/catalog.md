# AWS CodePipeline

Deploys a continuous delivery pipeline on AWS CodePipeline with configurable stages, actions, artifact stores, and optional V2 features including git-based triggers and pipeline-level variables. The service role, artifact buckets, encryption keys, and per-action roles all accept ValueFromRef wiring, so a pipeline composes with its IAM and storage dependencies — and the CodeBuild projects its stages invoke — in one InfraChart.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **CodePipeline Pipeline** -- an ordered sequence of stages, each containing one or more actions that fetch source code, run builds, execute tests, request approvals, or deploy to target environments
- **Artifact Stores** -- one or more S3 bucket bindings for storing pipeline artifacts between stages; supports cross-region artifact stores for multi-region pipelines
- **Git Triggers** -- created only when `triggers` are configured on V2 pipelines; listens for push or pull request events via CodeStar Connections with branch, tag, and file path filtering
- **Pipeline Variables** -- created only when `variables` are configured on V2 pipelines; parameterizes action configurations using `#{variables.Name}` syntax
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An IAM role** with permissions for S3 artifact access, source provider interaction (CodeStar Connections, S3, ECR, CodeCommit), and every action provider used in the pipeline (CodeBuild, ECS, Lambda, CloudFormation, etc.). Provide the ARN directly or reference an AwsIamRole Infra Component via ValueFromRef.
- **An S3 bucket** for pipeline artifact storage. Each artifact store requires a bucket. For cross-region pipelines, provide one bucket per region. Provide the bucket name directly or reference an AwsS3Bucket Infra Component.
- **A KMS key** (optional) -- for encrypting pipeline artifacts with a customer-managed key. Provide the ARN directly or reference an AwsKmsKey Infra Component.
- **A CodeStar Connection** (optional) -- required when using git-based triggers or CodeStarSourceConnection source actions for GitHub, Bitbucket, or GitLab repositories.

## Deploy

### Console

Open the deployment store, find **AWS CodePipeline**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **GitHub Source + CodeBuild (CI Pipeline)** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCodePipeline
metadata:
  name: api-deploy
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  roleArn:
    value: "arn:aws:iam::123456789012:role/CodePipelineServiceRole"
  artifactStores:
    - location:
        value: "my-pipeline-artifacts"
  stages:
    - name: Source
      actions:
        - name: GitHubSource
          category: Source
          owner: AWS
          provider: CodeStarSourceConnection
          version: "1"
          configuration:
            ConnectionArn: "arn:aws:codestar-connections:us-east-1:123456789012:connection/abc"
            FullRepositoryId: "acme-corp/api"
            BranchName: "main"
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
            ProjectName: "api-build"
          inputArtifacts:
            - SourceOutput
```

```shell
planton apply -f codepipeline.yaml
```

This creates a V2 pipeline with a GitHub source stage and a CodeBuild build stage. No triggers, variables, or cross-region artifact stores are configured. An Infra Job tracks the provisioning and streams progress in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the pipeline to an IAM role and S3 bucket deployed in the same InfraPipeline:

```yaml
spec:
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: codepipeline-service-role
      fieldPath: status.outputs.role_arn
  artifactStores:
    - location:
        valueFrom:
          kind: AwsS3Bucket
          name: pipeline-artifacts
          fieldPath: status.outputs.bucket_id
      encryptionKeyId:
        valueFrom:
          kind: AwsKmsKey
          name: pipeline-encryption-key
          fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, deploys the IAM role, S3 bucket, and KMS key first, then provisions the CodePipeline with the resolved values.

## Key Configuration

These are the most important decisions when configuring a CodePipeline. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Pipeline type** -- Default is `V2`, which supports git-based triggers, pipeline variables, and advanced execution modes (QUEUED, PARALLEL). Use `V1` only for legacy compatibility. V2 is recommended for all new pipelines.

**Execution mode** -- `SUPERSEDED` (default) cancels in-progress executions when a new one starts. Use `QUEUED` to process executions in order without superseding. Use `PARALLEL` to run multiple executions simultaneously. QUEUED and PARALLEL require V2 pipelines.

**Stage and action design** -- A pipeline requires at minimum two stages. Actions within a stage execute in parallel by default; use `runOrder` to enforce sequential execution within a stage. Each action connects to a provider (CodeBuild, S3, ECS, Lambda, Manual) that performs a specific task.

**Git triggers** -- V2 pipelines can trigger automatically on git push or pull request events via CodeStar Connections. Configure branch, tag, and file path filters to control which events start the pipeline. Triggers replace legacy polling and webhook mechanisms.

**Inline compute checks** -- A `Compute` action runs shell commands directly in CodePipeline-managed compute (no CodeBuild project to maintain): set `commands`, export values to downstream actions with `outputVariables` + `namespace`, and publish files with `outputArtifactsForComputeAction` (Compute actions use file-based artifacts instead of the plain `outputArtifacts`). Compute time bills through CodeBuild per execution.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `roleArn` | `status.outputs.role_arn` |
| **AwsS3Bucket** | `artifactStores[].location` | `status.outputs.bucket_id` |
| **AwsKmsKey** (optional) | `artifactStores[].encryptionKeyId` | `status.outputs.key_arn` |
| **AwsIamRole** (optional) | `stages[].actions[].roleArn` | `status.outputs.role_arn` |
| **AwsIamRole** (optional) | `stages[].{beforeEntry,onSuccess,onFailure.condition}.rules[].roleArn` | `status.outputs.role_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `pipeline_arn` | CodePipeline ARN | IAM policies, EventBridge targets, cross-resource references |
| `pipeline_name` | CodePipeline name | CLI commands, action configurations in other pipelines |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**GitHub CI pipeline** -- GitHub source via CodeStar Connection with a CodeBuild build stage and git push trigger on main. The standard starting point for automated build-and-test on every commit. Start from the **GitHub Source + CodeBuild (CI Pipeline)** preset.

**Container deployment pipeline** -- ECR source triggered by new image pushes, a CodeBuild stage to generate `imagedefinitions.json`, and an ECS deploy stage for rolling updates. Decouples build and deploy into separate pipelines. Start from the **ECR Source + ECS Deploy (Container Deployment Pipeline)** preset.

**Serverless deployment pipeline** -- S3 source triggered by new deployment package uploads, with a Lambda invoke action to update function code and shift aliases. Lightweight pipeline for Lambda function releases. Start from the **S3 Source + Lambda Deploy (Serverless Deployment Pipeline)** preset.

## Works With

- [**AWS CodeBuild Project**](/infra-catalog/aws-code-build-project) -- provides the build and test stages the pipeline's CodeBuild actions invoke by project name
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- provides the pipeline service role and optional per-action cross-account roles
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) -- provides artifact storage between pipeline stages
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for artifact encryption