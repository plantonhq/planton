# AWS CodePipeline

Deploys an AWS CodePipeline continuous delivery pipeline with ordered stages, provider-specific actions, S3 artifact stores, and optional V2 features including git-based triggers and pipeline-level variables.

## What Gets Created

When you deploy an AwsCodePipeline resource, Planton provisions:

- **CodePipeline** — an `aws_codepipeline` resource (V1 or V2) with the specified execution mode, IAM role, and artifact configuration
- **Artifact Stores** — one or more S3 bucket references for storing pipeline artifacts between stages, with optional KMS encryption and cross-region support
- **Stages and Actions** — an ordered sequence of stages, each containing one or more actions that connect to AWS service providers (CodeBuild, S3, ECS, Lambda, CodeDeploy, etc.)
- **Triggers (V2 only)** — git-based execution triggers using CodeStar Connections that listen for push or pull request events with branch, tag, and file path filtering
- **Variables (V2 only)** — pipeline-level parameters referenced in action configurations using `#{variables.Name}` syntax

## Prerequisites

- **AWS credentials** configured via environment variables or Planton provider config
- **An IAM role** with policies granting CodePipeline access to all action providers used in the pipeline (CodeBuild, S3, ECS, Lambda, etc.)
- **An S3 bucket** for artifact storage between pipeline stages
- **At least two stages** — a Source stage and at least one Build, Test, Deploy, Approval, or Invoke stage
- **A CodeStar Connection** if using GitHub, Bitbucket, or GitLab as a source provider
- **A CodeBuild project** if using CodeBuild for build or test actions

## Quick Start

Create a file `pipeline.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCodePipeline
metadata:
  name: my-app-pipeline
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsCodePipeline.my-app-pipeline
spec:
  region: us-west-2
  roleArn: arn:aws:iam::123456789012:role/codepipeline-service-role
  artifactStores:
    - location: my-pipeline-artifacts-bucket
  stages:
    - name: Source
      actions:
        - name: GitHubSource
          category: Source
          owner: AWS
          provider: CodeStarSourceConnection
          version: "1"
          configuration:
            ConnectionArn: arn:aws:codestar-connections:us-east-1:123456789012:connection/abc-12345
            FullRepositoryId: my-org/my-repo
            BranchName: main
            OutputArtifactFormat: CODE_ZIP
          outputArtifacts:
            - SourceOutput
    - name: Build
      actions:
        - name: BuildApp
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

Deploy:

```shell
planton apply -f pipeline.yaml
```

This creates a V2 pipeline with a GitHub source stage and a CodeBuild build stage, using the default SUPERSEDED execution mode.

## Configuration Reference

### Top-Level Fields

| Field | Type | Default | Description | Validation |
|-------|------|---------|-------------|------------|
| `pipelineType` | `string` | `V2` | Pipeline version. `V1` for legacy pipelines; `V2` for modern pipelines with triggers, variables, and advanced execution modes. | `V1` or `V2` |
| `executionMode` | `string` | `SUPERSEDED` | How concurrent executions are handled. `SUPERSEDED`: new execution replaces in-progress. `QUEUED`: executions queue behind current (V2 only). `PARALLEL`: executions run simultaneously (V2 only). | `SUPERSEDED`, `QUEUED`, or `PARALLEL` |
| `roleArn` | `string` | — | IAM role ARN granting CodePipeline permission to access source providers, invoke actions, and manage artifacts. Can reference an AwsIamRole resource via `valueFrom`. | **Required** |
| `artifactStores` | `ArtifactStore[]` | — | S3 buckets for storing pipeline artifacts between stages. One store for single-region; one per region for cross-region pipelines. | **Required**, minimum 1 item |
| `stages` | `Stage[]` | — | Ordered sequence of pipeline stages. Each stage contains one or more actions. | **Required**, minimum 2 stages |
| `triggers` | `Trigger[]` | `[]` | Automatic execution triggers based on git events. V2 pipelines only. | Maximum 50 items |
| `variables` | `Variable[]` | `[]` | Pipeline-level parameters referenced as `#{variables.Name}` in action configurations. V2 pipelines only. | — |

### ArtifactStore Fields

| Field | Type | Default | Description | Validation |
|-------|------|---------|-------------|------------|
| `location` | `string` | — | S3 bucket name for artifact storage. Can reference an AwsS3Bucket resource via `valueFrom`. | **Required** |
| `region` | `string` | — | AWS region for this artifact store. Required only for cross-region pipelines. Omit for single-region. | — |
| `encryptionKeyId` | `string` | — | KMS key ARN or ID for artifact encryption. If omitted, the default AWS-managed S3 encryption key is used. Can reference an AwsKmsKey resource via `valueFrom`. | — |

### Stage Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | Stage name. Must be unique within the pipeline. | **Required**, 1–100 characters |
| `actions` | `Action[]` | Operations performed in this stage. Same `runOrder` = parallel; different `runOrder` = sequential. | **Required**, minimum 1 action |

### Action Fields

| Field | Type | Default | Description | Validation |
|-------|------|---------|-------------|------------|
| `name` | `string` | — | Action name. Must be unique within the stage. | **Required**, 1–100 characters |
| `category` | `string` | — | Action type: `Source`, `Build`, `Test`, `Deploy`, `Approval`, `Invoke`, or `Compute`. | **Required** |
| `owner` | `string` | — | Who created the action type: `AWS`, `ThirdParty`, or `Custom`. | **Required** |
| `provider` | `string` | — | Service performing the action. Depends on category/owner (e.g., `CodeStarSourceConnection`, `CodeBuild`, `S3`, `ECS`, `Lambda`, `CodeDeploy`). | **Required**, 1–35 characters |
| `version` | `string` | — | Action type version. Typically `"1"` for all built-in actions. | **Required**, 1–9 characters |
| `configuration` | `map<string,string>` | `{}` | Provider-specific key-value pairs controlling the action's behavior. Keys depend on the provider. | — |
| `inputArtifacts` | `string[]` | `[]` | Artifact names from previous stages/actions that this action consumes. | — |
| `outputArtifacts` | `string[]` | `[]` | Artifact names this action produces for downstream consumption. | — |
| `namespace` | `string` | — | Variable namespace for this action's output variables. Referenced as `#{namespace.VariableName}`. | Maximum 100 characters |
| `region` | `string` | — | AWS region where this action executes. Required for cross-region actions. Defaults to the pipeline's region. | — |
| `roleArn` | `string` | — | IAM role ARN the action assumes instead of the pipeline role. Useful for cross-account deployments. Can reference an AwsIamRole resource via `valueFrom`. | — |
| `runOrder` | `int` | `1` | Execution order within a stage. Same value = parallel; lower values run first. | 1–999 |
| `timeoutInMinutes` | `int` | — | Maximum action runtime before timeout. If omitted, the provider default applies. | 5–86400 |

### Trigger Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `providerType` | `string` | Trigger provider. Must be `CodeStarSourceConnection`. | **Required** |
| `gitConfiguration` | `GitConfiguration` | Git event filters that trigger the pipeline. | **Required** |

### GitConfiguration Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `sourceActionName` | `string` | Must match the name of a Source action in the first stage that uses CodeStarSourceConnection. | **Required**, 1–100 characters |
| `push` | `GitPush[]` | Filters for git push events. Multiple push filters are OR'd. | Maximum 3 items |
| `pullRequest` | `GitPullRequest[]` | Filters for pull request events. Multiple PR filters are OR'd. | Maximum 3 items |

### GitPush Fields

| Field | Type | Description |
|-------|------|-------------|
| `branches` | `GitFilter` | Branch name patterns (glob syntax). |
| `filePaths` | `GitFilter` | Changed file path patterns (glob syntax). |
| `tags` | `GitFilter` | Tag name patterns (glob syntax). |

### GitPullRequest Fields

| Field | Type | Description |
|-------|------|-------------|
| `branches` | `GitFilter` | Target branch name patterns (glob syntax). |
| `filePaths` | `GitFilter` | Changed file path patterns (glob syntax). |
| `events` | `string[]` | PR lifecycle events: `OPEN`, `UPDATE`, `CLOSED`. |

### GitFilter Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `includes` | `string[]` | Glob patterns that must match. At least one must match for the filter to pass. | Maximum 8 items |
| `excludes` | `string[]` | Glob patterns that reject a match. Exclusions take precedence over inclusions. | Maximum 8 items |

### Variable Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | Variable name. Referenced as `#{variables.Name}` in action configurations. | **Required** |
| `defaultValue` | `string` | Value used when none is supplied at execution time. | — |
| `description` | `string` | Human-readable explanation of the variable's purpose. | — |

### Cross-Field Validations

| Rule | Description |
|------|-------------|
| `triggers_require_v2` | Triggers are only supported on V2 pipelines. Set `pipelineType` to `V2` or remove triggers. |
| `variables_require_v2` | Variables are only supported on V2 pipelines. Set `pipelineType` to `V2` or remove variables. |
| `advanced_execution_mode_requires_v2` | Execution modes `QUEUED` and `PARALLEL` are only supported on V2 pipelines. |

## Examples

### GitHub Source with CodeBuild

A V2 pipeline that pulls source from GitHub via CodeStar Connection, builds with CodeBuild, and triggers automatically on pushes to `main`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCodePipeline
metadata:
  name: my-app-ci
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsCodePipeline.my-app-ci
spec:
  region: us-west-2
  pipelineType: V2
  executionMode: QUEUED
  roleArn: arn:aws:iam::123456789012:role/codepipeline-service-role
  artifactStores:
    - location: my-pipeline-artifacts-bucket
  stages:
    - name: Source
      actions:
        - name: GitHubSource
          category: Source
          owner: AWS
          provider: CodeStarSourceConnection
          version: "1"
          configuration:
            ConnectionArn: arn:aws:codestar-connections:us-east-1:123456789012:connection/abc-12345
            FullRepositoryId: my-org/my-repo
            BranchName: main
            OutputArtifactFormat: CODE_ZIP
          outputArtifacts:
            - SourceOutput
          namespace: SourceVariables
    - name: Build
      actions:
        - name: BuildAndTest
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
```

### ECR Source with ECS Deployment

A pipeline triggered by new ECR image pushes that prepares a deployment bundle and deploys to ECS:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCodePipeline
metadata:
  name: my-service-deploy
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsCodePipeline.my-service-deploy
spec:
  region: us-west-2
  pipelineType: V2
  executionMode: QUEUED
  roleArn: arn:aws:iam::123456789012:role/codepipeline-service-role
  artifactStores:
    - location: my-pipeline-artifacts-bucket
  stages:
    - name: Source
      actions:
        - name: ECRSource
          category: Source
          owner: AWS
          provider: ECR
          version: "1"
          configuration:
            RepositoryName: my-service
            ImageTag: latest
          outputArtifacts:
            - ECROutput
    - name: Build
      actions:
        - name: PrepareDeployment
          category: Build
          owner: AWS
          provider: CodeBuild
          version: "1"
          configuration:
            ProjectName: my-deploy-prep
          inputArtifacts:
            - ECROutput
          outputArtifacts:
            - DeploymentBundle
    - name: Deploy
      actions:
        - name: ECSDeployProd
          category: Deploy
          owner: AWS
          provider: ECS
          version: "1"
          configuration:
            ClusterName: prod-cluster
            ServiceName: my-service
            FileName: imagedefinitions.json
          inputArtifacts:
            - DeploymentBundle
```

### S3 Source with Lambda Deployment

A pipeline that fetches a Lambda deployment package from S3 and invokes a deployer function:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCodePipeline
metadata:
  name: my-lambda-deploy
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsCodePipeline.my-lambda-deploy
spec:
  region: us-west-2
  pipelineType: V2
  executionMode: SUPERSEDED
  roleArn: arn:aws:iam::123456789012:role/codepipeline-service-role
  artifactStores:
    - location: my-pipeline-artifacts-bucket
  stages:
    - name: Source
      actions:
        - name: S3Source
          category: Source
          owner: AWS
          provider: S3
          version: "1"
          configuration:
            S3Bucket: lambda-packages-bucket
            S3ObjectKey: functions/my-function/package.zip
            PollForSourceChanges: "false"
          outputArtifacts:
            - LambdaPackage
    - name: Deploy
      actions:
        - name: UpdateLambda
          category: Invoke
          owner: AWS
          provider: Lambda
          version: "1"
          configuration:
            FunctionName: lambda-deployer
            UserParameters: '{"targetFunction":"my-target-function","alias":"live"}'
          inputArtifacts:
            - LambdaPackage
```

### Multi-Stage with Manual Approval and Cross-Region Artifacts

A production pipeline with build, staging deploy, manual approval gate, and production deploy across regions:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCodePipeline
metadata:
  name: prod-release-pipeline
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsCodePipeline.prod-release-pipeline
spec:
  region: us-west-2
  pipelineType: V2
  executionMode: QUEUED
  roleArn: arn:aws:iam::123456789012:role/codepipeline-service-role
  artifactStores:
    - location: artifacts-us-east-1
    - location: artifacts-eu-west-1
      region: eu-west-1
  stages:
    - name: Source
      actions:
        - name: GitHubSource
          category: Source
          owner: AWS
          provider: CodeStarSourceConnection
          version: "1"
          configuration:
            ConnectionArn: arn:aws:codestar-connections:us-east-1:123456789012:connection/abc-12345
            FullRepositoryId: my-org/my-service
            BranchName: release
            OutputArtifactFormat: CODE_ZIP
          outputArtifacts:
            - SourceOutput
          namespace: SourceVariables
    - name: Build
      actions:
        - name: BuildApp
          category: Build
          owner: AWS
          provider: CodeBuild
          version: "1"
          configuration:
            ProjectName: my-service-build
          inputArtifacts:
            - SourceOutput
          outputArtifacts:
            - BuildOutput
    - name: DeployStaging
      actions:
        - name: ECSDeployStaging
          category: Deploy
          owner: AWS
          provider: ECS
          version: "1"
          configuration:
            ClusterName: staging-cluster
            ServiceName: my-service
            FileName: imagedefinitions.json
          inputArtifacts:
            - BuildOutput
    - name: Approval
      actions:
        - name: ManualApproval
          category: Approval
          owner: AWS
          provider: Manual
          version: "1"
          configuration:
            CustomData: "Approve deployment to production?"
            NotificationArn: arn:aws:sns:us-east-1:123456789012:pipeline-approvals
    - name: DeployProduction
      actions:
        - name: ECSDeployProd
          category: Deploy
          owner: AWS
          provider: ECS
          version: "1"
          configuration:
            ClusterName: prod-cluster
            ServiceName: my-service
            FileName: imagedefinitions.json
          inputArtifacts:
            - BuildOutput
  variables:
    - name: ReleaseTag
      defaultValue: latest
      description: Container image tag to deploy
  triggers:
    - providerType: CodeStarSourceConnection
      gitConfiguration:
        sourceActionName: GitHubSource
        push:
          - tags:
              includes:
                - "v*"
```

### Using Foreign Key References

Reference other Planton-managed resources instead of hardcoding ARNs and bucket names:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCodePipeline
metadata:
  name: ref-pipeline
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsCodePipeline.ref-pipeline
spec:
  region: us-west-2
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: pipeline-role
      field: status.outputs.role_arn
  artifactStores:
    - location:
        valueFrom:
          kind: AwsS3Bucket
          name: pipeline-artifacts
          fieldPath: status.outputs.bucket_id
      encryptionKeyId:
        valueFrom:
          kind: AwsKmsKey
          name: pipeline-key
          field: status.outputs.key_arn
  stages:
    - name: Source
      actions:
        - name: GitHubSource
          category: Source
          owner: AWS
          provider: CodeStarSourceConnection
          version: "1"
          configuration:
            ConnectionArn: arn:aws:codestar-connections:us-east-1:123456789012:connection/abc-12345
            FullRepositoryId: my-org/my-repo
            BranchName: main
            OutputArtifactFormat: CODE_ZIP
          outputArtifacts:
            - SourceOutput
    - name: Build
      actions:
        - name: BuildApp
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
          roleArn:
            valueFrom:
              kind: AwsIamRole
              name: build-role
              field: status.outputs.role_arn
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `pipeline_arn` | `string` | ARN of the created CodePipeline. Use for IAM policies, EventBridge targets, and cross-resource references. |
| `pipeline_name` | `string` | Name of the pipeline. Use when referencing the pipeline in CLI commands or other action configurations. |

## Presets

Presets provide ready-to-use pipeline configurations for common patterns. Apply a preset as your starting point and customize as needed.

| Preset | File | Description |
|--------|------|-------------|
| **GitHub Source + CodeBuild** | `presets/01-github-source-codebuild.yaml` | V2 pipeline with GitHub source via CodeStar Connection, CodeBuild build stage, and automatic trigger on pushes to `main`. Ideal starting point for CI pipelines. |
| **ECR + ECS Deploy** | `presets/02-ecr-ecs-deploy.yaml` | V2 pipeline triggered by new ECR image pushes. Prepares a deployment bundle via CodeBuild and deploys to an ECS service. Use for container-based continuous deployment. |
| **S3 + Lambda Deploy** | `presets/03-s3-lambda-deploy.yaml` | V2 pipeline that fetches a Lambda deployment package from S3 and invokes a deployer Lambda function. Use for serverless function deployment workflows. |

## Related Components

- [AwsIamRole](/docs/catalog/aws/awsiamrole) — provides the service role granting CodePipeline access to action providers
- [AwsS3Bucket](/docs/catalog/aws/awss3bucket) — provides the artifact store bucket for pipeline artifacts
- [AwsKmsKey](/docs/catalog/aws/awskmskey) — provides encryption keys for artifact store encryption
- [AwsCodeBuildProject](/docs/catalog/aws/awscodebuildproject) — provides build projects referenced in Build and Test actions
- [AwsEcsService](/docs/catalog/aws/awsecsservice) — target for ECS deploy actions
