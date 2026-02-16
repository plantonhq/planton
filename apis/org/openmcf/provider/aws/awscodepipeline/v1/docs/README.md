# AWS CodePipeline: Technical Reference

## Introduction

AWS CodePipeline is a fully managed continuous delivery service that orchestrates build, test, and deploy phases of your release process. Rather than scripting CI/CD flows imperatively, CodePipeline defines them declaratively as an ordered sequence of stages containing actions — creating a visual, auditable, and repeatable release pipeline.

OpenMCF's `AwsCodePipeline` component exposes the full pipeline definition as a single declarative manifest: stages, actions, artifact stores, triggers, and variables. The result is a portable, reviewable YAML document that captures your entire release workflow.

## Architecture: Stages, Actions, and Artifact Flow

A CodePipeline pipeline is structured as a directed graph of stages, where each stage contains one or more actions.

### Pipeline Structure

```
Pipeline
├── Artifact Store(s)     ← S3 buckets for passing data between stages
├── Stage: Source          ← Fetches code/artifacts from a provider
│   └── Action: GitHubSource (CodeStarSourceConnection)
│       └── Output: SourceOutput
├── Stage: Build           ← Compiles, tests, transforms
│   ├── Action: UnitTests (CodeBuild, runOrder=1)
│   │   └── Input: SourceOutput
│   └── Action: DockerBuild (CodeBuild, runOrder=2)
│       ├── Input: SourceOutput
│       └── Output: BuildOutput
├── Stage: Approval        ← Human gate
│   └── Action: ManualApproval (Manual)
├── Stage: Deploy          ← Deploys to target
│   └── Action: ECSDeployProd (ECS)
│       └── Input: BuildOutput
├── Triggers (V2)          ← Automatic execution rules
└── Variables (V2)         ← Parameterization
```

### Artifact Flow

Artifacts are the data that flows between actions. They are stored in S3 and referenced by name.

1. **Source actions** produce output artifacts (e.g., `SourceOutput` containing the repository checkout)
2. **Build actions** consume input artifacts and produce output artifacts (e.g., `BuildOutput` containing compiled code or Docker image definitions)
3. **Deploy actions** consume input artifacts and deploy them to target environments
4. **Approval actions** have no artifacts — they simply gate progression

Actions within the same stage that share the same `runOrder` execute in parallel. Actions with different `runOrder` values execute sequentially (lower values first).

### Single-Region vs. Cross-Region

For **single-region** pipelines, provide one artifact store without a `region` field. All actions execute in the pipeline's region.

For **cross-region** pipelines, provide one artifact store per region (each with a `region` field). Actions that specify a `region` different from the pipeline's home region will use the matching regional artifact store.

## V1 vs V2 Pipelines

AWS CodePipeline has two pipeline versions with significant capability differences.

| Feature | V1 | V2 |
|---------|----|----|
| **Triggers** | None (polling or CloudWatch Events only) | Git-based triggers with branch/tag/file path filtering |
| **Variables** | None | Pipeline-level variables with `#{variables.Name}` syntax |
| **Execution Mode** | SUPERSEDED only | SUPERSEDED, QUEUED, or PARALLEL |
| **Compute category** | Not available | EC2-based compute actions |
| **Pricing** | Per pipeline/month ($1/pipeline) | Per action execution (pay-per-use) |
| **Recommended** | Legacy only | All new pipelines |

### When to Use V1

V1 pipelines should only be used when:
- Migrating existing V1 pipelines that cannot yet be upgraded
- Cost optimization for pipelines that execute very frequently (V1's flat monthly pricing may be cheaper)

### When to Use V2

V2 is recommended for all new pipelines. Key advantages:
- **Git triggers** eliminate the need for CloudWatch Events rules or polling
- **Variables** enable parameterized pipelines (environment names, feature flags)
- **QUEUED execution** prevents race conditions in deployment pipelines
- **PARALLEL execution** enables simultaneous feature branch builds

## Action Providers and Configuration Keys

Each action provider expects specific keys in its `configuration` map. Below is a reference for the most common providers.

### Source Providers

#### CodeStarSourceConnection (GitHub, Bitbucket, GitLab)

The recommended source provider for all git-hosted repositories. Uses AWS CodeStar Connections (also called CodeConnections) for secure, OAuth-based access.

| Key | Required | Description |
|-----|----------|-------------|
| `ConnectionArn` | **Yes** | ARN of the CodeStar Connection |
| `FullRepositoryId` | **Yes** | `owner/repo` format (e.g., `my-org/my-app`) |
| `BranchName` | **Yes** | Branch to track (e.g., `main`) |
| `OutputArtifactFormat` | No | `CODE_ZIP` (default) or `CODEBUILD_CLONE_REF` |
| `DetectChanges` | No | `true` (default) or `false` — whether to auto-detect changes |

#### S3

| Key | Required | Description |
|-----|----------|-------------|
| `S3Bucket` | **Yes** | S3 bucket name |
| `S3ObjectKey` | **Yes** | Object key path (e.g., `releases/app.zip`) |
| `PollForSourceChanges` | No | `true` or `false` (prefer CloudTrail events over polling) |

#### ECR (Elastic Container Registry)

| Key | Required | Description |
|-----|----------|-------------|
| `RepositoryName` | **Yes** | ECR repository name |
| `ImageTag` | No | Tag to track (default: `latest`) |

### Build/Test Providers

#### CodeBuild

| Key | Required | Description |
|-----|----------|-------------|
| `ProjectName` | **Yes** | CodeBuild project name |
| `PrimarySource` | No | Primary source artifact name (for multi-source builds) |
| `EnvironmentVariables` | No | JSON array of `{"name","value","type"}` objects |
| `BatchEnabled` | No | `true` to enable batch builds |

### Deploy Providers

#### ECS (Amazon Elastic Container Service)

| Key | Required | Description |
|-----|----------|-------------|
| `ClusterName` | **Yes** | ECS cluster name |
| `ServiceName` | **Yes** | ECS service name |
| `FileName` | No | Image definitions file in the input artifact (default: `imagedefinitions.json`) |
| `DeploymentTimeout` | No | Timeout in minutes |

#### S3 (Deploy)

| Key | Required | Description |
|-----|----------|-------------|
| `BucketName` | **Yes** | Target S3 bucket name |
| `Extract` | **Yes** | `true` to extract zip, `false` to upload as-is |
| `ObjectKey` | No | Target key path prefix |
| `CannedACL` | No | Canned ACL (e.g., `public-read`) |

#### CloudFormation

| Key | Required | Description |
|-----|----------|-------------|
| `ActionMode` | **Yes** | `CREATE_UPDATE`, `DELETE_ONLY`, `REPLACE_ON_FAILURE`, `CHANGE_SET_REPLACE`, `CHANGE_SET_EXECUTE` |
| `StackName` | **Yes** | CloudFormation stack name |
| `TemplatePath` | Conditional | `ArtifactName::template-file.yaml` |
| `RoleArn` | No | CloudFormation service role ARN |
| `Capabilities` | No | `CAPABILITY_IAM`, `CAPABILITY_NAMED_IAM`, `CAPABILITY_AUTO_EXPAND` |
| `ParameterOverrides` | No | JSON object of parameter key-value pairs |
| `ChangeSetName` | Conditional | Required for change set action modes |

#### Lambda (Invoke)

| Key | Required | Description |
|-----|----------|-------------|
| `FunctionName` | **Yes** | Lambda function name or ARN |
| `UserParameters` | No | String passed to the Lambda function |

#### CodeDeploy

| Key | Required | Description |
|-----|----------|-------------|
| `ApplicationName` | **Yes** | CodeDeploy application name |
| `DeploymentGroupName` | **Yes** | Deployment group name |

### Approval Provider

#### Manual

| Key | Required | Description |
|-----|----------|-------------|
| `CustomData` | No | Message shown to the approver (supports variable substitution) |
| `ExternalEntityLink` | No | URL link shown to the approver |
| `NotificationArn` | No | SNS topic ARN for approval notification |

## Trigger Patterns

Triggers are the V2 mechanism for automatically executing pipelines based on git events. They replace the V1 pattern of CloudWatch Events + webhooks.

### Push Triggers

Push triggers fire when commits are pushed to matching branches or tags.

**Branch filtering:**
```yaml
triggers:
  - providerType: CodeStarSourceConnection
    gitConfiguration:
      sourceActionName: GitHubSource
      push:
        - branches:
            includes:
              - main
              - "release/*"
            excludes:
              - "release/experimental-*"
```

**Tag filtering:**
```yaml
triggers:
  - providerType: CodeStarSourceConnection
    gitConfiguration:
      sourceActionName: GitHubSource
      push:
        - tags:
            includes:
              - "v*"
            excludes:
              - "v*-rc*"
```

**File path filtering:**
```yaml
triggers:
  - providerType: CodeStarSourceConnection
    gitConfiguration:
      sourceActionName: GitHubSource
      push:
        - branches:
            includes:
              - main
          filePaths:
            includes:
              - "src/**"
              - "tests/**"
            excludes:
              - "docs/**"
              - "*.md"
```

### Pull Request Triggers

PR triggers fire on pull request lifecycle events against matching target branches.

```yaml
triggers:
  - providerType: CodeStarSourceConnection
    gitConfiguration:
      sourceActionName: GitHubSource
      pullRequest:
        - branches:
            includes:
              - main
          events:
            - OPEN
            - UPDATE
```

### Filter Logic

- **Within a single push/PR block**: All specified filter types (branches, file_paths, tags) are **AND'd**. A push must match ALL specified filter types to trigger the pipeline.
- **Multiple push/PR blocks**: Multiple blocks are **OR'd**. A push that matches ANY block triggers the pipeline.
- **Include/Exclude**: Within each filter type, at least one `includes` pattern must match AND no `excludes` pattern must match. Exclusions take precedence.
- **Glob patterns**: Use glob syntax — `*` matches any string within a segment, `**` matches any number of path segments.

## Variable Usage

Pipeline variables (V2 only) enable parameterized pipelines. Variables can be set at execution time and referenced in action configurations.

### Defining Variables

```yaml
variables:
  - name: Environment
    defaultValue: staging
    description: Target deployment environment
  - name: ImageTag
    defaultValue: latest
    description: Docker image tag to deploy
```

### Referencing Variables

Variables are referenced in action configurations using `#{variables.Name}` syntax:

```yaml
configuration:
  EnvironmentVariables: '[{"name":"ENV","value":"#{variables.Environment}","type":"PLAINTEXT"}]'
```

### Action Output Variables

Source actions can expose output variables through namespaces. When you set `namespace: SourceVariables` on a source action, downstream actions can reference:

- `#{SourceVariables.CommitId}` — the git commit SHA
- `#{SourceVariables.CommitMessage}` — the commit message
- `#{SourceVariables.BranchName}` — the branch name
- `#{SourceVariables.AuthorDate}` — the commit author date

Reference format: `#{namespace.VariableName}`

## Cross-Region Pipelines

Cross-region pipelines execute actions in different AWS regions. This is useful when deploying to multiple regions or when a service (like a CodeBuild project) exists in a different region.

### Configuration

1. **Artifact stores**: Provide one artifact store per region:

```yaml
artifactStores:
  - location:
      value: pipeline-artifacts-us-east-1
    region: us-east-1
  - location:
      value: pipeline-artifacts-eu-west-1
    region: eu-west-1
```

2. **Action region**: Set the `region` field on actions that should execute in a non-default region:

```yaml
actions:
  - name: DeployEurope
    category: Deploy
    owner: AWS
    provider: ECS
    version: "1"
    region: eu-west-1
    configuration:
      ClusterName: eu-production-cluster
      ServiceName: my-service
    inputArtifacts:
      - BuildOutput
```

### How It Works

CodePipeline automatically copies artifacts from the source region's S3 bucket to the destination region's S3 bucket before executing the cross-region action. The KMS keys in each region must allow the pipeline role to encrypt/decrypt.

## Cross-Account Pipelines

Cross-account pipelines execute actions (typically deploy) using IAM roles in different AWS accounts. This is the standard pattern for production deployments where CI runs in a tools account and deploys to staging/production accounts.

### Configuration

Use the action-level `roleArn` to assume a role in the target account:

```yaml
actions:
  - name: DeployToProduction
    category: Deploy
    owner: AWS
    provider: CloudFormation
    version: "1"
    configuration:
      ActionMode: CREATE_UPDATE
      StackName: my-app
      TemplatePath: PackagedTemplates::template.yaml
      RoleArn: arn:aws:iam::999999999999:role/cloudformation-deploy-role
    inputArtifacts:
      - PackagedTemplates
    roleArn:
      value: arn:aws:iam::999999999999:role/codepipeline-cross-account-role
```

### IAM Requirements

The cross-account pattern requires two roles in the target account:

1. **Pipeline action role** (`roleArn` on the action): The role that CodePipeline assumes to execute the action. Must trust the pipeline account.
2. **Service role** (e.g., CloudFormation `RoleArn` in configuration): The role that the action's service assumes to create/modify resources. Must trust the action service.

Additionally, the artifact store's KMS key policy must grant `kms:Decrypt` and `kms:GenerateDataKey` to the cross-account roles.

## IAM Role Requirements

The pipeline's IAM role (`roleArn` at the top level) needs permissions for all action providers used in the pipeline.

### Minimum Policy for Common Pipelines

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:GetBucketVersioning"
      ],
      "Resource": [
        "arn:aws:s3:::pipeline-artifacts-bucket",
        "arn:aws:s3:::pipeline-artifacts-bucket/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "codestar-connections:UseConnection"
      ],
      "Resource": "arn:aws:codeconnections:*:*:connection/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "codebuild:BatchGetBuilds",
        "codebuild:StartBuild"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "ecs:DescribeServices",
        "ecs:DescribeTaskDefinition",
        "ecs:DescribeTasks",
        "ecs:ListTasks",
        "ecs:RegisterTaskDefinition",
        "ecs:UpdateService"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": "iam:PassRole",
      "Resource": "*",
      "Condition": {
        "StringEqualsIfExists": {
          "iam:PassedToService": [
            "ecs-tasks.amazonaws.com"
          ]
        }
      }
    },
    {
      "Effect": "Allow",
      "Action": [
        "kms:Decrypt",
        "kms:GenerateDataKey"
      ],
      "Resource": "arn:aws:kms:*:*:key/*"
    }
  ]
}
```

### Trust Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "codepipeline.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

### Provider-Specific Permissions

| Provider | Required IAM Actions |
|----------|---------------------|
| CodeStarSourceConnection | `codestar-connections:UseConnection` |
| S3 (source/deploy) | `s3:GetObject`, `s3:PutObject`, `s3:GetBucketVersioning` |
| ECR | `ecr:DescribeImages` |
| CodeBuild | `codebuild:BatchGetBuilds`, `codebuild:StartBuild` |
| ECS | `ecs:DescribeServices`, `ecs:UpdateService`, `ecs:RegisterTaskDefinition`, `iam:PassRole` |
| Lambda | `lambda:InvokeFunction`, `lambda:ListFunctions` |
| CloudFormation | `cloudformation:*`, `iam:PassRole` |
| CodeDeploy | `codedeploy:CreateDeployment`, `codedeploy:GetDeployment`, `codedeploy:RegisterApplicationRevision` |
| Manual (Approval) | `sns:Publish` (for notification) |

## Common Patterns

### CI-Only Pipeline

Source → Build. The simplest pipeline for running tests on every commit.

### CI/CD with Approval

Source → Build → Approval → Deploy. Adds a human gate before production deployment. The approval action can include a link to the build results and a custom message with variable substitution.

### Multi-Environment Promotion

Source → Build → Deploy Staging → Test Staging → Approve → Deploy Production. A full promotion pipeline where staging is tested before production deployment.

### Parallel Actions

Actions within the same stage that share `runOrder` execute in parallel:

```yaml
stages:
  - name: Build
    actions:
      - name: UnitTests
        runOrder: 1
        # ...
      - name: Linting
        runOrder: 1
        # ...
      - name: DockerBuild
        runOrder: 2
        # ...
```

In this example, UnitTests and Linting run in parallel (both `runOrder: 1`), and DockerBuild runs after both complete (`runOrder: 2`).

### Artifact Naming Conventions

Use descriptive artifact names that indicate content:

| Name | Content |
|------|---------|
| `SourceOutput` | Raw source code checkout |
| `BuildOutput` | Compiled application or packaged artifact |
| `TestResults` | Test reports and coverage data |
| `PackagedTemplates` | CloudFormation/IaC templates ready for deployment |
| `ImageDefinitions` | ECS `imagedefinitions.json` for container deployments |
