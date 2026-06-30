---
title: "CodeBuild Project"
description: "CodeBuild Project deployment documentation"
icon: "package"
order: 100
componentName: "awscodebuildproject"
---

# AWS CodeBuild Project

Deploys an AWS CodeBuild project with configurable source providers, build environments, artifact outputs, and an optional webhook for automatic build triggers. Supports GitHub, Bitbucket, GitLab, CodeCommit, S3, and CodePipeline source types.

## What Gets Created

When you deploy an AwsCodeBuildProject resource, Planton provisions:

- **CodeBuild Project** — an `aws_codebuild_project` resource with the specified source, environment, artifacts, and logging configuration
- **Webhook** — created only when `webhook` is configured, registers a webhook with the source provider for automatic build triggers on push and pull request events
- **VPC Configuration** — created only when `vpcConfig` is configured, places the build environment inside a VPC for access to private resources

## Prerequisites

- **AWS credentials** configured via environment variables or Planton provider config
- **An IAM service role** granting CodeBuild permission to access source code, write artifacts, and publish logs
- **A source repository** accessible to CodeBuild (GitHub via CodeStar Connections, CodeCommit, Bitbucket, GitLab, or S3)
- **An S3 bucket** if using S3 artifacts or S3 cache
- **VPC, subnets, and security groups** if running builds inside a VPC
- **A KMS key** if custom encryption for build artifacts is required

## Quick Start

Create a file `codebuild.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCodeBuildProject
metadata:
  name: my-build
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsCodeBuildProject.my-build
spec:
  region: us-west-2
  source:
    type: GITHUB
    location: https://github.com/my-org/my-app.git
  environment:
    type: LINUX_CONTAINER
    computeType: BUILD_GENERAL1_SMALL
    image: aws/codebuild/amazonlinux2-x86_64-standard:5.0
  artifacts:
    type: NO_ARTIFACTS
  serviceRole:
    value: arn:aws:iam::123456789012:role/codebuild-service-role
```

Deploy:

```shell
planton apply -f codebuild.yaml
```

This creates a CodeBuild project that pulls from a GitHub repository, runs builds in a standard Linux container, and produces no stored artifacts (CI-only).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the project will be created (e.g., `us-west-2`). | Required |
| `source.type` | `string` | Source provider: `GITHUB`, `BITBUCKET`, `CODECOMMIT`, `CODEPIPELINE`, `GITHUB_ENTERPRISE`, `GITLAB`, `GITLAB_SELF_MANAGED`, `NO_SOURCE`, `S3` | Must be one of the listed values |
| `source.location` | `string` | Repository URL or S3 path | Required unless type is `CODEPIPELINE` or `NO_SOURCE` |
| `environment.type` | `string` | Container type: `LINUX_CONTAINER`, `LINUX_GPU_CONTAINER`, `ARM_CONTAINER`, `WINDOWS_SERVER_2019_CONTAINER`, `WINDOWS_SERVER_2022_CONTAINER`, `LINUX_LAMBDA_CONTAINER`, `ARM_LAMBDA_CONTAINER` | Must be one of the listed values |
| `environment.computeType` | `string` | Compute capacity: `BUILD_GENERAL1_SMALL` through `BUILD_GENERAL1_2XLARGE`, or `BUILD_LAMBDA_1GB` through `BUILD_LAMBDA_10GB` | Must be one of the listed values |
| `environment.image` | `string` | Docker image for the build environment (e.g., `aws/codebuild/amazonlinux2-x86_64-standard:5.0`) | Required, non-empty |
| `artifacts.type` | `string` | Artifact output type: `NO_ARTIFACTS`, `S3`, `CODEPIPELINE` | Must be one of the listed values. Must be `CODEPIPELINE` when source is `CODEPIPELINE` |
| `serviceRole` | `StringValueOrRef` | IAM role ARN for CodeBuild. Can reference an AwsIamRole resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Project description (max 255 characters) |
| `encryptionKey` | `StringValueOrRef` | AWS-managed key | KMS key ARN for artifact encryption. Can reference AwsKmsKey via `valueFrom`. |
| `buildTimeout` | `int` | `60` | Build timeout in minutes (5-2160) |
| `queuedTimeout` | `int` | `480` | Queue timeout in minutes (5-480) |
| `concurrentBuildLimit` | `int` | Unlimited | Maximum concurrent builds (minimum 1) |
| `sourceVersion` | `string` | — | Default branch, tag, or commit to build |
| `source.buildspec` | `string` | `buildspec.yml` | Build specification file path or inline YAML. Required when type is `NO_SOURCE`. |
| `source.gitCloneDepth` | `int` | Full clone | Git clone depth (0 = full clone) |
| `source.reportBuildStatus` | `bool` | `false` | Report build status back to the source provider |
| `source.fetchSubmodules` | `bool` | `false` | Fetch Git submodules during source download |
| `environment.privilegedMode` | `bool` | `false` | Enable Docker daemon access inside the build container |
| `environment.imagePullCredentialsType` | `string` | `CODEBUILD` | Image pull credentials: `CODEBUILD` or `SERVICE_ROLE` |
| `environment.environmentVariables` | `list` | `[]` | Build environment variables with `name`, `value`, and optional `type` (`PLAINTEXT`, `PARAMETER_STORE`, `SECRETS_MANAGER`) |
| `environment.registryCredential` | `object` | — | Private registry credentials (`credential` ARN, `credentialProvider`: `SECRETS_MANAGER`) |
| `artifacts.location` | `StringValueOrRef` | — | S3 bucket name. Required when type is `S3`. Can reference AwsS3Bucket via `valueFrom`. |
| `artifacts.name` | `string` | — | Artifact output name |
| `artifacts.path` | `string` | — | S3 prefix path for artifacts |
| `artifacts.packaging` | `string` | `NONE` | Packaging type: `NONE` or `ZIP` |
| `artifacts.namespaceType` | `string` | `NONE` | Namespace: `NONE` or `BUILD_ID` |
| `artifacts.encryptionDisabled` | `bool` | `false` | Disable artifact encryption |
| `cache.type` | `string` | `NO_CACHE` | Cache type: `NO_CACHE`, `S3`, or `LOCAL` |
| `cache.location` | `StringValueOrRef` | — | S3 cache location. Required when type is `S3`. Can reference AwsS3Bucket via `valueFrom`. |
| `cache.modes` | `string[]` | `[]` | Local cache modes: `LOCAL_SOURCE_CACHE`, `LOCAL_DOCKER_LAYER_CACHE`, `LOCAL_CUSTOM_CACHE` |
| `logsConfig.cloudwatchLogs.status` | `string` | `ENABLED` | CloudWatch logging: `ENABLED` or `DISABLED` |
| `logsConfig.cloudwatchLogs.groupName` | `StringValueOrRef` | Auto-generated | Log group name. Can reference AwsCloudwatchLogGroup via `valueFrom`. |
| `logsConfig.cloudwatchLogs.streamName` | `string` | Auto-generated | Log stream name prefix |
| `logsConfig.s3Logs.status` | `string` | `DISABLED` | S3 logging: `ENABLED` or `DISABLED` |
| `logsConfig.s3Logs.location` | `StringValueOrRef` | — | S3 bucket and prefix for logs. Can reference AwsS3Bucket via `valueFrom`. |
| `logsConfig.s3Logs.encryptionDisabled` | `bool` | `false` | Disable log file encryption |
| `vpcConfig.vpcId` | `StringValueOrRef` | — | VPC ID. Can reference AwsVpc via `valueFrom`. Required if vpcConfig is set. |
| `vpcConfig.subnetIds` | `StringValueOrRef[]` | — | VPC subnets (max 16). Can reference AwsVpc via `valueFrom`. Required if vpcConfig is set. |
| `vpcConfig.securityGroupIds` | `StringValueOrRef[]` | — | Security groups (max 5). Can reference AwsSecurityGroup via `valueFrom`. Required if vpcConfig is set. |
| `webhook.buildType` | `string` | `BUILD` | Webhook build type: `BUILD` or `BUILD_BATCH` |
| `webhook.filterGroups` | `list` | `[]` | Filter groups (OR'd). Each group contains filters (AND'd) with `type`, `pattern`, and optional `excludeMatchedPattern`. |

## Examples

### GitHub CI with Webhook

A standard CI project triggered by pushes and pull requests on the main branch:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCodeBuildProject
metadata:
  name: app-ci
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsCodeBuildProject.app-ci
spec:
  region: us-west-2
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
    value: arn:aws:iam::123456789012:role/codebuild-role
  webhook:
    filterGroups:
      - filters:
          - type: EVENT
            pattern: PUSH, PULL_REQUEST_CREATED, PULL_REQUEST_UPDATED
          - type: HEAD_REF
            pattern: ^refs/heads/main$
```

### Docker Build with Layer Caching

A privileged build project for Docker image builds with local layer caching:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCodeBuildProject
metadata:
  name: docker-builder
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsCodeBuildProject.docker-builder
spec:
  region: us-west-2
  source:
    type: GITHUB
    location: https://github.com/my-org/my-service.git
    gitCloneDepth: 1
    reportBuildStatus: true
  environment:
    type: LINUX_CONTAINER
    computeType: BUILD_GENERAL1_LARGE
    image: aws/codebuild/amazonlinux2-x86_64-standard:5.0
    privilegedMode: true
    environmentVariables:
      - name: ECR_REPO
        value: 123456789012.dkr.ecr.us-east-1.amazonaws.com/my-service
      - name: DOCKER_BUILDKIT
        value: "1"
  artifacts:
    type: NO_ARTIFACTS
  serviceRole:
    value: arn:aws:iam::123456789012:role/codebuild-docker-role
  buildTimeout: 30
  cache:
    type: LOCAL
    modes:
      - LOCAL_DOCKER_LAYER_CACHE
      - LOCAL_SOURCE_CACHE
  webhook:
    filterGroups:
      - filters:
          - type: EVENT
            pattern: PUSH
          - type: HEAD_REF
            pattern: ^refs/heads/main$
```

### CodePipeline Build Stage

A build project designed as a stage in AWS CodePipeline with Secrets Manager variables:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCodeBuildProject
metadata:
  name: pipeline-build
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsCodeBuildProject.pipeline-build
spec:
  region: us-west-2
  source:
    type: CODEPIPELINE
    buildspec: buildspec.yml
  environment:
    type: LINUX_CONTAINER
    computeType: BUILD_GENERAL1_MEDIUM
    image: aws/codebuild/amazonlinux2-x86_64-standard:5.0
    environmentVariables:
      - name: STAGE
        value: production
      - name: DB_CONNECTION_STRING
        value: prod/db-connection-string
        type: SECRETS_MANAGER
  artifacts:
    type: CODEPIPELINE
  serviceRole:
    value: arn:aws:iam::123456789012:role/codebuild-pipeline-role
  buildTimeout: 20
  concurrentBuildLimit: 3
```

### Using Foreign Key References

Reference other Planton-managed resources instead of hardcoding IDs:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCodeBuildProject
metadata:
  name: connected-build
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsCodeBuildProject.connected-build
spec:
  region: us-west-2
  source:
    type: GITHUB
    location: https://github.com/my-org/my-app.git
  environment:
    type: LINUX_CONTAINER
    computeType: BUILD_GENERAL1_SMALL
    image: aws/codebuild/amazonlinux2-x86_64-standard:5.0
  artifacts:
    type: NO_ARTIFACTS
  serviceRole:
    valueFrom:
      kind: AwsIamRole
      name: codebuild-role
      field: status.outputs.role_arn
  encryptionKey:
    valueFrom:
      kind: AwsKmsKey
      name: build-key
      field: status.outputs.key_arn
  vpcConfig:
    vpcId:
      valueFrom:
        kind: AwsVpc
        name: main-vpc
        field: status.outputs.vpc_id
    subnetIds:
      - valueFrom:
          kind: AwsSubnet
          name: main-private-subnet-a
          fieldPath: status.outputs.subnet_id
      - valueFrom:
          kind: AwsSubnet
          name: main-private-subnet-b
          fieldPath: status.outputs.subnet_id
    securityGroupIds:
      - valueFrom:
          kind: AwsSecurityGroup
          name: codebuild-sg
          field: status.outputs.security_group_id
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `project_arn` | `string` | ARN of the CodeBuild project, used for IAM policies and cross-resource references |
| `project_name` | `string` | Name of the CodeBuild project |
| `service_role_arn` | `string` | IAM service role ARN used by the project |
| `webhook_url` | `string` | Webhook URL from the source provider (empty when no webhook is configured) |
| `webhook_payload_url` | `string` | URL that receives webhook payloads (empty when no webhook is configured) |

## Related Components

- [AwsIamRole](/docs/catalog/aws/iam-role) — provides the service role for CodeBuild
- [AwsVpc](/docs/catalog/aws/vpc) — provides VPC and subnets for builds that access private resources
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls network access for VPC-enabled builds
- [AwsS3Bucket](/docs/catalog/aws/s3-bucket) — stores build artifacts and cache
- [AwsCloudwatchLogGroup](/docs/catalog/aws/cloudwatch-log-group) — hosts build logs
