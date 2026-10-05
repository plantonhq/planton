# AwsCodeBuildProject

Deploy and manage AWS CodeBuild projects — build environments, multi-source/multi-artifact wiring, batch builds, reserved-capacity fleet membership, and webhook triggers for automated CI/CD.

## Overview

AWS CodeBuild is a fully managed build service that compiles source code, runs tests, and produces deployable artifacts. This component creates a CodeBuild project — the build configuration unit — defining where source code comes from, how to build it, and where to put the output. A folded webhook enables automatic build triggers from source providers, and a folded resource policy grants cross-account access to the project.

**Supported source types:**

| Source | Description | Use Case |
|--------|-------------|----------|
| `GITHUB` | GitHub.com repository | Most common for open-source and SaaS teams |
| `GITHUB_ENTERPRISE` | GitHub Enterprise Server | Self-hosted GitHub |
| `BITBUCKET` | Bitbucket Cloud repository | Atlassian ecosystem |
| `CODECOMMIT` | AWS CodeCommit repository | Fully AWS-native CI/CD |
| `CODEPIPELINE` | Source provided by CodePipeline | Stage in a multi-step pipeline |
| `S3` | S3 bucket containing source archive | Artifact-based builds |
| `GITLAB` | GitLab.com repository | GitLab ecosystem |
| `GITLAB_SELF_MANAGED` | Self-hosted GitLab instance | Self-hosted GitLab |
| `NO_SOURCE` | No source — inline buildspec | Utility/maintenance builds |

Up to 12 **secondary sources** ride alongside the primary source (each named by a `sourceIdentifier` the buildspec addresses as `$CODEBUILD_SRC_DIR_<identifier>`), with optional per-source version pins. Up to 12 **secondary artifacts** fan the output to multiple destinations.

**Bundled resources:**
- **CodeBuild project** — build configuration (sources, environment, artifacts, cache, logs, VPC, EFS mounts, batch builds, visibility)
- **Webhook** (optional) — automatic build triggers, including organization/group-scoped webhooks, manual-registration mode, and fork-PR approval gating
- **Resource policy** (optional) — a resource-based IAM policy on the project for cross-account access

**Not included:** source credentials (an account/region-wide Git credential store — the folded `source.auth` CODECONNECTIONS arm is the modern per-project path), report groups (shared across projects), and reserved-capacity fleets (shared account-level resources — the project joins one through `environment.fleetArn`).

## Prerequisites

- An IAM service role for CodeBuild with appropriate permissions (use `AwsIamRole`)
- For VPC builds: a VPC with private subnets and security groups (use `AwsVpc`, `AwsSubnet`, `AwsSecurityGroup`)
- For S3 artifacts/cache/logs: an S3 bucket (use `AwsS3Bucket`)
- For encrypted artifacts: a KMS key (use `AwsKmsKey`)
- For EFS mounts: an EFS file system reachable from the build subnets (use `AwsElasticFileSystem`)

## Quick Start

### Minimal (GitHub CI)

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCodeBuildProject
metadata:
  name: my-app-ci
spec:
  region: us-east-1
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

### Production (Docker Build with Persistent Docker Server)

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCodeBuildProject
metadata:
  name: docker-build
spec:
  region: us-east-1
  source:
    type: GITHUB
    location: https://github.com/my-org/my-app.git
    gitCloneDepth: 1
    reportBuildStatus: true
    auth:
      type: CODECONNECTIONS
      resource: arn:aws:codeconnections:us-east-1:123456789012:connection/aaaa1111
  environment:
    type: LINUX_CONTAINER
    computeType: BUILD_GENERAL1_LARGE
    image: aws/codebuild/amazonlinux2-x86_64-standard:5.0
    privilegedMode: true
    dockerServer:
      computeType: BUILD_GENERAL1_MEDIUM
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
  autoRetryLimit: 1
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
| `region` | string | **Yes** | — | AWS region |
| `source` | object | **Yes** | — | Primary source (its `sourceIdentifier` must stay empty) |
| `secondarySources` | list | No | — | Up to 12 named extra sources |
| `secondarySourceVersions` | list | No | — | Per-secondary-source branch/tag/commit pins |
| `environment` | object | **Yes** | — | Build container configuration |
| `artifacts` | object | **Yes** | — | Primary build output |
| `secondaryArtifacts` | list | No | — | Up to 12 named extra outputs |
| `serviceRole` | StringValueOrRef | **Yes** | — | IAM service role ARN |
| `description` | string | No | — | Project description (max 255 chars) |
| `encryptionKey` | StringValueOrRef | No | AWS-managed key | KMS key for artifact encryption |
| `buildTimeout` | int | No | 60 | Build timeout in minutes (5-2160; not supported for Lambda types) |
| `queuedTimeout` | int | No | 480 | Queue timeout in minutes (5-480; not supported for Lambda types) |
| `concurrentBuildLimit` | int | No | Unlimited | Max concurrent builds (min 1) |
| `autoRetryLimit` | int | No | 0 | Additional automatic retries after a failed build (max 10) |
| `badgeEnabled` | bool | No | false | Publish a build badge (repository sources only) |
| `sourceVersion` | string | No | — | Default branch/tag/commit to build |
| `cache` | object | No | NO_CACHE | Build caching configuration |
| `logsConfig` | object | No | CloudWatch ENABLED | Logging configuration |
| `vpcConfig` | object | No | — | VPC placement for private resource access |
| `fileSystemLocations` | list | No | — | EFS mounts (require `vpcConfig`) |
| `buildBatchConfig` | object | No | — | Batch builds (build graphs/lists/matrices) |
| `projectVisibility` | string | No | PRIVATE | PRIVATE or PUBLIC_READ |
| `resourceAccessRole` | StringValueOrRef | Conditional | — | Role for public log/artifact access (required with PUBLIC_READ) |
| `resourcePolicy` | object | No | — | Resource-based IAM policy (cross-account access) |
| `webhook` | object | No | — | Automatic build trigger configuration |

### Source (primary and secondary)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | **Yes** | Source provider (table above) |
| `location` | string | Conditional | Repository URL or S3 path (required unless CODEPIPELINE/NO_SOURCE) |
| `buildspec` | string | Conditional | Buildspec path or inline YAML (required for NO_SOURCE) |
| `gitCloneDepth` | int | No | Git clone depth (0 = full clone) |
| `gitSubmodulesConfig` | object | No | Submodule fetching (BITBUCKET/CODECOMMIT/GITHUB/GITHUB_ENTERPRISE only) |
| `insecureSsl` | bool | No | Skip TLS verification (self-hosted providers with private CAs only) |
| `reportBuildStatus` | bool | No | Report status to the provider (commit checks) |
| `buildStatusConfig` | object | No | Custom status context + target URL |
| `auth` | object | No | Per-source authorization: CODECONNECTIONS / SECRETS_MANAGER / OAUTH + resource ARN |
| `sourceIdentifier` | string | Secondary only | Names a secondary source (`$CODEBUILD_SRC_DIR_<identifier>`) |

### Environment

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | **Yes** | Container type (LINUX/ARM/Windows/Lambda/EC2 variants, MAC_ARM) |
| `computeType` | string | **Yes** | Compute size (BUILD_GENERAL1_*, BUILD_LAMBDA_*, fleet-driven ATTRIBUTE_BASED_COMPUTE / CUSTOM_INSTANCE_TYPE) |
| `image` | string | **Yes** | Docker image identifier |
| `certificate` | string | No | S3 path to a trusted certificate bundle (.pem/.zip) |
| `privilegedMode` | bool | No | Docker daemon access (not for Lambda types) |
| `imagePullCredentialsType` | string | No | CODEBUILD (default) or SERVICE_ROLE |
| `hostKernel` | string | No | Linux kernel selection: `LINUX_KERNEL_4`, `LINUX_KERNEL_6`, or `LINUX_KERNEL_LATEST` (LINUX/ARM container and EC2 types only; omit to let AWS choose) |
| `environmentVariables` | list | No | PLAINTEXT / PARAMETER_STORE / SECRETS_MANAGER variables |
| `registryCredential` | object | No | Private registry secret (requires SERVICE_ROLE pulls) |
| `dockerServer` | object | No | Persistent dedicated Docker server (compute + security groups) |
| `fleetArn` | string | No | Reserved-capacity fleet membership (required for MAC_ARM / EC2 types) |

### Artifacts (primary and secondary)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | **Yes** | NO_ARTIFACTS, S3, or CODEPIPELINE |
| `location` | StringValueOrRef | Conditional | S3 bucket (required for S3 type) |
| `name` / `path` | string | No | Object key and S3 prefix |
| `packaging` | string | No | NONE or ZIP |
| `namespaceType` | string | No | NONE or BUILD_ID |
| `encryptionDisabled` | bool | No | Disable artifact encryption |
| `overrideArtifactName` | bool | No | Let the buildspec override the artifact name per build |
| `bucketOwnerAccess` | string | No | NONE / READ_ONLY / FULL (cross-account buckets) |
| `artifactIdentifier` | string | Secondary only | Names a secondary artifact |

### Webhook

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `buildType` | string | No | BUILD (default), BUILD_BATCH, or RUNNER_BUILDKITE_BUILD |
| `manualCreation` | bool | No | Mint the payload URL + secret without registering (wire the repository by hand) |
| `filterGroups` | list | No | OR'd groups of AND'd filters (11 filter types incl. WORKFLOW_NAME, TAG_NAME, REPOSITORY_NAME, ORGANIZATION_NAME) |
| `scopeConfiguration` | object | No | Organization/group-wide webhooks (GITHUB_ORGANIZATION / GITHUB_GLOBAL / GITLAB_GROUP) |
| `pullRequestBuildPolicy` | object | No | Comment-approval gate for PR builds (DISABLED / FORK_PULL_REQUESTS / ALL_PULL_REQUESTS + approver roles) |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `project_arn` | string | ARN of the CodeBuild project |
| `project_name` | string | Name of the project (what CodePipeline Build actions reference) |
| `service_role_arn` | string | IAM service role ARN |
| `badge_url` | string | Build badge URL (empty when the badge is disabled) |
| `public_project_alias` | string | Public alias (empty for private projects) |
| `webhook_url` | string | Webhook URL (empty if no webhook) |
| `webhook_payload_url` | string | Webhook payload URL (empty if no webhook) |
| `webhook_secret` | string | Webhook HMAC secret — sensitive, only minted at creation (empty if no webhook) |

## Presets

| Preset | Description |
|--------|-------------|
| `01-github-ci-linux` | GitHub source, Linux container, CI-only (no artifacts), webhook |
| `02-docker-build-ecr` | GitHub source, privileged mode, Docker layer caching |
| `03-codepipeline-stage` | CODEPIPELINE source + artifacts, designed as a pipeline stage |
| `04-fork-pr-gated-oss-ci` | CodeConnections auth, fork-PR approval gating, badge, auto-retry |

## Deliberate Exclusions

- **`environment.host_kernel`** — not yet released in either IaC engine's provider; add when both carry it.
- **Legacy `branch_filter` webhook argument** — `filterGroups` with a `HEAD_REF` filter is the same capability in the modern, composable shape.
- **Source credentials (`aws_codebuild_source_credential`)** — an account/region-wide credential import; the per-source `auth` block (CodeConnections/Secrets Manager) is the modern path.
- **Report groups / fleets** — shared account-level resources with independent lifecycles; the project composes with fleets through `environment.fleetArn`.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
