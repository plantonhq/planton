# AWS CodeBuild Project

Deploys a CodeBuild project with configurable source providers, build environments, artifact destinations, and an optional webhook for source-triggered builds. The service role, encryption key, artifact and cache buckets, log destinations, and VPC networking all accept ValueFromRef wiring, so a build project composes with its IAM, storage, and network dependencies in one InfraChart — including VPC connectivity for builds that reach private resources.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **CodeBuild Project** -- a build project configured with the specified source provider, build container image, compute type, environment variables, artifact output, and optional VPC networking, caching, and log destinations
- **CodeBuild Webhook** -- created only when `webhook` is configured; registers a webhook with the source provider (GitHub, Bitbucket, GitLab, CodeCommit) to trigger builds on push or pull request events
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An IAM service role** with permissions for source access, log writing, artifact storage, and any additional AWS services invoked during the build. Provide the ARN directly or reference an AwsIamRole Infra Component via ValueFromRef.
- **An S3 bucket** (optional) -- required when artifact type is S3 or when S3 caching or S3 log delivery is enabled. Provide the bucket name directly or reference an AwsS3Bucket Infra Component.
- **A KMS key** (optional) -- for encrypting build artifacts with a customer-managed key instead of the AWS-managed S3 key. Provide the ARN directly or reference an AwsKmsKey Infra Component.
- **VPC subnets and security groups** (optional) -- required when the build needs access to private resources (RDS, ElastiCache, internal APIs). Provide IDs directly or reference AwsVpc and AwsSecurityGroup Infra Components.

## Deploy

### Console

Open the deployment store, find **AWS CodeBuild Project**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **GitHub CI (Linux)** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCodeBuildProject
metadata:
  name: api-ci
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  source:
    type: GITHUB
    location: "https://github.com/acme-corp/api.git"
  environment:
    type: LINUX_CONTAINER
    computeType: BUILD_GENERAL1_SMALL
    image: "aws/codebuild/amazonlinux2-x86_64-standard:5.0"
  artifacts:
    type: NO_ARTIFACTS
  serviceRole:
    value: "arn:aws:iam::123456789012:role/CodeBuildServiceRole"
```

```shell
planton apply -f codebuild-project.yaml
```

This creates a CI-only CodeBuild project pulling from GitHub with a small Linux build container and no artifacts. No webhook, VPC connectivity, or caching is configured. An Infra Job tracks the provisioning and streams progress in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the CodeBuild project to an IAM role, KMS key, and VPC deployed in the same InfraPipeline:

```yaml
spec:
  serviceRole:
    valueFrom:
      kind: AwsIamRole
      name: codebuild-service-role
      fieldPath: status.outputs.role_arn
  encryptionKey:
    valueFrom:
      kind: AwsKmsKey
      name: artifact-encryption-key
      fieldPath: status.outputs.key_arn
  vpcConfig:
    vpcId:
      valueFrom:
        kind: AwsVpc
        name: production-vpc
        fieldPath: status.outputs.vpc_id
    subnetIds:
      - valueFrom:
          kind: AwsSubnet
          name: private-subnet-a
          fieldPath: status.outputs.subnet_id
    securityGroupIds:
      - valueFrom:
          kind: AwsSecurityGroup
          name: codebuild-sg
          fieldPath: status.outputs.security_group_id
```

The InfraPipeline resolves the dependency graph, deploys the IAM role, KMS key, VPC, and security group first, then provisions the CodeBuild project with the resolved values.

## Key Configuration

These are the most important decisions when configuring a CodeBuild project. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Source type** -- Determines where build input comes from. Use `GITHUB`, `BITBUCKET`, or `GITLAB` for webhook-triggered CI. Use `CODEPIPELINE` when the project is a stage in an AWS CodePipeline (both `source.type` and `artifacts.type` must be CODEPIPELINE). Use `S3` for archive-based builds.

**Compute type and environment** -- `BUILD_GENERAL1_SMALL` (3 GB, 2 vCPUs) handles most test suites. Use `BUILD_GENERAL1_LARGE` (15 GB, 8 vCPUs) for Docker builds. Enable `environment.privilegedMode` when the build needs Docker daemon access. Lambda compute types (`BUILD_LAMBDA_*`) provide faster cold starts for lightweight builds. On Linux container and EC2 environment types, `environment.hostKernel` selects the build host's kernel line (`LINUX_KERNEL_4`, `LINUX_KERNEL_6`, or `LINUX_KERNEL_LATEST`) for builds that depend on newer kernel features (eBPF, io_uring); omit it to let AWS choose.

**Build caching** -- Set `cache.type` to `LOCAL` with `LOCAL_DOCKER_LAYER_CACHE` for Docker builds or `S3` for shared caches across concurrent builds. Local caching is faster but ephemeral.

**Webhook configuration** -- Omit `webhook` for CodePipeline-triggered or manual-only projects. Configure filter groups to control which branches and event types trigger builds.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `serviceRole` | `status.outputs.role_arn` |
| **AwsIamRole** (optional) | `resourceAccessRole` | `status.outputs.role_arn` |
| **AwsIamRole** (optional) | `buildBatchConfig.serviceRole` | `status.outputs.role_arn` |
| **AwsKmsKey** (optional) | `encryptionKey` | `status.outputs.key_arn` |
| **AwsS3Bucket** (optional) | `artifacts.location` | `status.outputs.bucket_id` |
| **AwsS3Bucket** (optional) | `cache.location` | `status.outputs.bucket_id` |
| **AwsCloudwatchLogGroup** (optional) | `logsConfig.cloudwatchLogs.groupName` | `status.outputs.log_group_name` |
| **AwsS3Bucket** (optional) | `logsConfig.s3Logs.bucket` | `status.outputs.bucket_id` |
| **AwsVpc** (optional) | `vpcConfig.vpcId` | `status.outputs.vpc_id` |
| **AwsSubnet** (optional) | `vpcConfig.subnetIds` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `vpcConfig.securityGroupIds` | `status.outputs.security_group_id` |
| **AwsSecurityGroup** (optional) | `environment.dockerServer.securityGroupIds` | `status.outputs.security_group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `project_arn` | CodeBuild project ARN | IAM policies, EventBridge targets, CodePipeline action configuration |
| `project_name` | CodeBuild project name | CodePipeline build stage action, CLI invocations |
| `service_role_arn` | IAM service role ARN used by the project | Audit, cross-resource IAM policy references |
| `badge_url` | Dynamic build badge URL (when `badgeEnabled`) | Repository README status badge |
| `public_project_alias` | Public alias (when `projectVisibility` is PUBLIC_READ) | Public build results URL |
| `webhook_url` | Webhook URL for the source provider | Source provider webhook verification |
| `webhook_payload_url` | Webhook payload delivery URL | Manual webhook registration in the repository |
| `webhook_secret` | Webhook HMAC secret (manual registration) | Authenticating provider deliveries when registering the webhook by hand |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**GitHub CI pipeline** -- GitHub source with webhook triggers on push and pull requests, small compute, no artifacts. Standard configuration for automated testing with build status reported back to GitHub. Start from the **GitHub CI (Linux)** preset.

**Docker image builder** -- GitHub source with privileged mode, large compute, local Docker layer caching, and environment variables for ECR push. Optimized for container image builds with fast incremental rebuilds. Start from the **Docker Build with ECR Push** preset.

**CodePipeline build stage** -- CODEPIPELINE source and artifact types with no webhook. Designed to run as a stage within an AWS CodePipeline where the pipeline handles source fetching and artifact passing. Start from the **CodePipeline Build Stage** preset.

## Works With

- [**AWS CodePipeline**](/infra-catalog/aws-code-pipeline) -- runs this project as a build stage, consuming the `project_name` output in its CodeBuild action configuration
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- provides the service role for source access, log writing, and artifact storage
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for build artifact encryption
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) -- provides storage for artifacts, build cache, and S3 log delivery
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) -- provides a custom log group for build output
- [**AWS VPC**](/infra-catalog/aws-vpc) -- provides VPC and subnets for builds needing private resource access
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- controls network access for VPC-connected builds