# AwsCodeBuildProject

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsCodeBuildProjectSpec defines the specification for an AWS CodeBuild
project with an optional webhook for source-triggered builds.

AWS CodeBuild is a fully managed build service that compiles source code,
runs tests, and produces deployable artifacts. This component creates a
CodeBuild project — the build configuration unit — and optionally a webhook
for automatic build triggers from source providers.

Bundling rationale: a webhook is 1:1 with a project and useless in
isolation — source-triggered projects are incomplete without one, while
CodePipeline-triggered and manual projects omit it. A resource policy is a
single project-keyed document, so it folds. Source credentials (an
account/region-wide Git credential store), report groups, and reserved
capacity fleets have independent lifecycles shared across projects and are
deliberately excluded; the project joins a fleet through
environment.fleet_arn.

## Example

```yaml
# Full-surface offline-plan manifest. This exercises every optional block the
# module renders -- including the arms excluded from live E2E lanes (webhook
# with manual creation, organization scope, and PR build policy; source auth;
# public visibility with its access role; EFS mounts; batch builds; the
# host-kernel selection) -- so the `tofu plan` / `pulumi preview` proofs
# cover the whole contract.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCodeBuildProject
metadata:
  name: test-codebuild
  id: test-codebuild
  org: test-org
  env: dev
spec:
  region: us-west-2
  source:
    type: GITHUB
    location: https://github.com/example/repo.git
    buildspec: ci/buildspec.yml
    gitCloneDepth: 1
    gitSubmodulesConfig:
      fetchSubmodules: true
    reportBuildStatus: true
    buildStatusConfig:
      context: ci/codebuild
      targetUrl: https://ci.example.com/builds
    auth:
      type: CODECONNECTIONS
      resource: arn:aws:codeconnections:us-west-2:123456789012:connection/aaaa1111-22bb-33cc-44dd-5555eeee6666
  secondarySources:
    - sourceIdentifier: tooling
      type: GITHUB
      location: https://github.com/example/build-tooling.git
      gitCloneDepth: 1
  secondarySourceVersions:
    - sourceIdentifier: tooling
      sourceVersion: main
  environment:
    type: LINUX_CONTAINER
    computeType: BUILD_GENERAL1_MEDIUM
    image: aws/codebuild/amazonlinux2-x86_64-standard:5.0
    certificate: certs-bucket/private-ca.pem
    privilegedMode: true
    imagePullCredentialsType: CODEBUILD
    hostKernel: LINUX_KERNEL_6
    environmentVariables:
      - name: STAGE
        value: dev
      - name: DB_PASSWORD
        value: arn:aws:secretsmanager:us-west-2:123456789012:secret:db-pass
        type: SECRETS_MANAGER
    dockerServer:
      computeType: BUILD_GENERAL1_MEDIUM
      securityGroupIds:
        - value: sg-0123456789abcdef0
  artifacts:
    type: S3
    location:
      value: my-artifact-bucket
    name: app.zip
    path: releases
    packaging: ZIP
    namespaceType: BUILD_ID
    overrideArtifactName: true
    bucketOwnerAccess: READ_ONLY
  secondaryArtifacts:
    - artifactIdentifier: reports
      type: S3
      location:
        value: my-reports-bucket
      path: test-reports
  serviceRole:
    value: arn:aws:iam::123456789012:role/codebuild-service-role
  description: Full-surface CodeBuild plan proof
  encryptionKey:
    value: arn:aws:kms:us-west-2:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab
  buildTimeout: 90
  queuedTimeout: 120
  concurrentBuildLimit: 4
  autoRetryLimit: 2
  badgeEnabled: true
  sourceVersion: main
  # Public visibility with its required read-access role (PUBLIC_READ is
  # CEL-gated on resourceAccessRole being set).
  projectVisibility: PUBLIC_READ
  resourceAccessRole:
    value: arn:aws:iam::123456789012:role/codebuild-public-read-role
  cache:
    type: S3
    location:
      value: my-cache-bucket
    cacheNamespace: main-branch
  logsConfig:
    cloudwatchLogs:
      status: ENABLED
      groupName:
        value: /codebuild/test-codebuild
      streamName: builds
    s3Logs:
      status: ENABLED
      bucket:
        value: my-log-bucket
      prefix: build-logs
      bucketOwnerAccess: FULL
  vpcConfig:
    vpcId:
      value: vpc-0123456789abcdef0
    subnetIds:
      - value: subnet-0123456789abcdef0
      - value: subnet-0123456789abcdef1
    securityGroupIds:
      - value: sg-0123456789abcdef0
  fileSystemLocations:
    - identifier: build_cache
      location: fs-0123456789abcdef0.efs.us-west-2.amazonaws.com:/build-cache
      mountPoint: /mnt/build-cache
  buildBatchConfig:
    serviceRole:
      value: arn:aws:iam::123456789012:role/codebuild-batch-role
    combineArtifacts: true
    timeoutInMins: 120
    restrictions:
      computeTypesAllowed:
        - BUILD_GENERAL1_SMALL
        - BUILD_GENERAL1_MEDIUM
      maximumBuildsAllowed: 10
  resourcePolicy:
    Version: "2012-10-17"
    Statement:
      - Sid: SharedBuildAccess
        Effect: Allow
        Principal:
          AWS: arn:aws:iam::210987654321:root
        Action:
          - codebuild:BatchGetProjects
          - codebuild:StartBuild
        Resource: "*"
        Condition:
          StringEquals:
            aws:ResourceAccount: "${aws:PrincipalAccount}"
  webhook:
    buildType: BUILD
    # manual_creation: CodeBuild creates the webhook definition without
    # registering it with the source provider -- the payload URL and secret
    # are then wired into GitHub by hand (GitHub-source projects only).
    manualCreation: true
    # An organization-scoped webhook: one webhook covering every repository
    # in the GitHub organization (runner/monorepo fan-in patterns).
    scopeConfiguration:
      name: example-org
      scope: GITHUB_ORGANIZATION
    filterGroups:
      - filters:
          - type: EVENT
            pattern: PUSH, PULL_REQUEST_CREATED, PULL_REQUEST_UPDATED
          - type: HEAD_REF
            pattern: ^refs/heads/main$
          - type: FILE_PATH
            pattern: ^src/.*
            excludeMatchedPattern: false
    pullRequestBuildPolicy:
      requiresCommentApproval: FORK_PULL_REQUESTS
      approverRoles:
        - GITHUB_WRITE
        - GITHUB_MAINTAIN
        - GITHUB_ADMIN
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.source` | `AwsCodeBuildSource` | yes |  |  |
| `spec.source.type` | `string` | yes |  |  |
| `spec.source.location` | `string` |  |  |  |
| `spec.source.buildspec` | `string` |  |  |  |
| `spec.source.gitCloneDepth` | `int32` |  |  |  |
| `spec.source.gitSubmodulesConfig` | `AwsCodeBuildGitSubmodulesConfig` |  |  |  |
| `spec.source.gitSubmodulesConfig.fetchSubmodules` | `bool` |  |  |  |
| `spec.source.insecureSsl` | `bool` |  |  |  |
| `spec.source.reportBuildStatus` | `bool` |  |  |  |
| `spec.source.buildStatusConfig` | `AwsCodeBuildBuildStatusConfig` |  |  |  |
| `spec.source.buildStatusConfig.context` | `string` |  |  |  |
| `spec.source.buildStatusConfig.targetUrl` | `string` |  |  |  |
| `spec.source.auth` | `AwsCodeBuildSourceAuth` |  |  |  |
| `spec.source.auth.type` | `string` | yes |  |  |
| `spec.source.auth.resource` | `string` | yes |  |  |
| `spec.source.sourceIdentifier` | `string` |  |  |  |
| `spec.secondarySources` | `[]AwsCodeBuildSource` |  |  |  |
| `spec.secondarySources[].type` | `string` | yes |  |  |
| `spec.secondarySources[].location` | `string` |  |  |  |
| `spec.secondarySources[].buildspec` | `string` |  |  |  |
| `spec.secondarySources[].gitCloneDepth` | `int32` |  |  |  |
| `spec.secondarySources[].gitSubmodulesConfig` | `AwsCodeBuildGitSubmodulesConfig` |  |  |  |
| `spec.secondarySources[].gitSubmodulesConfig.fetchSubmodules` | `bool` |  |  |  |
| `spec.secondarySources[].insecureSsl` | `bool` |  |  |  |
| `spec.secondarySources[].reportBuildStatus` | `bool` |  |  |  |
| `spec.secondarySources[].buildStatusConfig` | `AwsCodeBuildBuildStatusConfig` |  |  |  |
| `spec.secondarySources[].buildStatusConfig.context` | `string` |  |  |  |
| `spec.secondarySources[].buildStatusConfig.targetUrl` | `string` |  |  |  |
| `spec.secondarySources[].auth` | `AwsCodeBuildSourceAuth` |  |  |  |
| `spec.secondarySources[].auth.type` | `string` | yes |  |  |
| `spec.secondarySources[].auth.resource` | `string` | yes |  |  |
| `spec.secondarySources[].sourceIdentifier` | `string` |  |  |  |
| `spec.secondarySourceVersions` | `[]AwsCodeBuildSecondarySourceVersion` |  |  |  |
| `spec.secondarySourceVersions[].sourceIdentifier` | `string` | yes |  |  |
| `spec.secondarySourceVersions[].sourceVersion` | `string` | yes |  |  |
| `spec.environment` | `AwsCodeBuildEnvironment` | yes |  |  |
| `spec.environment.type` | `string` | yes |  |  |
| `spec.environment.computeType` | `string` | yes |  |  |
| `spec.environment.image` | `string` | yes |  |  |
| `spec.environment.certificate` | `string` |  |  |  |
| `spec.environment.privilegedMode` | `bool` |  |  |  |
| `spec.environment.imagePullCredentialsType` | `string` |  | `CODEBUILD` |  |
| `spec.environment.environmentVariables` | `[]AwsCodeBuildEnvironmentVariable` |  |  |  |
| `spec.environment.environmentVariables[].name` | `string` | yes |  |  |
| `spec.environment.environmentVariables[].value` | `string` | yes |  |  |
| `spec.environment.environmentVariables[].type` | `string` |  | `PLAINTEXT` |  |
| `spec.environment.registryCredential` | `AwsCodeBuildRegistryCredential` |  |  |  |
| `spec.environment.registryCredential.credential` | `string` | yes |  |  |
| `spec.environment.registryCredential.credentialProvider` | `string` | yes |  |  |
| `spec.environment.dockerServer` | `AwsCodeBuildDockerServer` |  |  |  |
| `spec.environment.dockerServer.computeType` | `string` | yes |  |  |
| `spec.environment.dockerServer.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.environment.fleetArn` | `string` |  |  |  |
| `spec.environment.hostKernel` | `string` |  |  |  |
| `spec.artifacts` | `AwsCodeBuildArtifacts` | yes |  |  |
| `spec.artifacts.type` | `string` | yes |  |  |
| `spec.artifacts.location` | `string \| valueFrom` |  |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.artifacts.name` | `string` |  |  |  |
| `spec.artifacts.path` | `string` |  |  |  |
| `spec.artifacts.packaging` | `string` |  |  |  |
| `spec.artifacts.namespaceType` | `string` |  |  |  |
| `spec.artifacts.encryptionDisabled` | `bool` |  |  |  |
| `spec.artifacts.overrideArtifactName` | `bool` |  |  |  |
| `spec.artifacts.bucketOwnerAccess` | `string` |  |  |  |
| `spec.artifacts.artifactIdentifier` | `string` |  |  |  |
| `spec.secondaryArtifacts` | `[]AwsCodeBuildArtifacts` |  |  |  |
| `spec.secondaryArtifacts[].type` | `string` | yes |  |  |
| `spec.secondaryArtifacts[].location` | `string \| valueFrom` |  |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.secondaryArtifacts[].name` | `string` |  |  |  |
| `spec.secondaryArtifacts[].path` | `string` |  |  |  |
| `spec.secondaryArtifacts[].packaging` | `string` |  |  |  |
| `spec.secondaryArtifacts[].namespaceType` | `string` |  |  |  |
| `spec.secondaryArtifacts[].encryptionDisabled` | `bool` |  |  |  |
| `spec.secondaryArtifacts[].overrideArtifactName` | `bool` |  |  |  |
| `spec.secondaryArtifacts[].bucketOwnerAccess` | `string` |  |  |  |
| `spec.secondaryArtifacts[].artifactIdentifier` | `string` |  |  |  |
| `spec.serviceRole` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.description` | `string` |  |  |  |
| `spec.encryptionKey` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.buildTimeout` | `int32` |  | `60` |  |
| `spec.queuedTimeout` | `int32` |  | `480` |  |
| `spec.concurrentBuildLimit` | `int32` |  |  |  |
| `spec.autoRetryLimit` | `int32` |  |  |  |
| `spec.badgeEnabled` | `bool` |  |  |  |
| `spec.sourceVersion` | `string` |  |  |  |
| `spec.cache` | `AwsCodeBuildCache` |  |  |  |
| `spec.cache.type` | `string` |  | `NO_CACHE` |  |
| `spec.cache.location` | `string \| valueFrom` |  |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.cache.modes` | `[]string` |  |  |  |
| `spec.cache.cacheNamespace` | `string` |  |  |  |
| `spec.logsConfig` | `AwsCodeBuildLogsConfig` |  |  |  |
| `spec.logsConfig.cloudwatchLogs` | `AwsCodeBuildCloudWatchLogs` |  |  |  |
| `spec.logsConfig.cloudwatchLogs.status` | `string` |  | `ENABLED` |  |
| `spec.logsConfig.cloudwatchLogs.groupName` | `string \| valueFrom` |  |  | AwsCloudwatchLogGroup (`status.outputs.log_group_name`) |
| `spec.logsConfig.cloudwatchLogs.streamName` | `string` |  |  |  |
| `spec.logsConfig.s3Logs` | `AwsCodeBuildS3Logs` |  |  |  |
| `spec.logsConfig.s3Logs.status` | `string` |  | `DISABLED` |  |
| `spec.logsConfig.s3Logs.bucket` | `string \| valueFrom` |  |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.logsConfig.s3Logs.encryptionDisabled` | `bool` |  |  |  |
| `spec.logsConfig.s3Logs.bucketOwnerAccess` | `string` |  |  |  |
| `spec.logsConfig.s3Logs.prefix` | `string` |  |  |  |
| `spec.vpcConfig` | `AwsCodeBuildVpcConfig` |  |  |  |
| `spec.vpcConfig.vpcId` | `string \| valueFrom` | yes |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.vpcConfig.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.vpcConfig.securityGroupIds` | `[]string \| valueFrom` | yes |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.fileSystemLocations` | `[]AwsCodeBuildFileSystemLocation` |  |  |  |
| `spec.fileSystemLocations[].identifier` | `string` | yes |  |  |
| `spec.fileSystemLocations[].location` | `string` | yes |  |  |
| `spec.fileSystemLocations[].mountPoint` | `string` | yes |  |  |
| `spec.fileSystemLocations[].mountOptions` | `string` |  |  |  |
| `spec.fileSystemLocations[].type` | `string` |  | `EFS` |  |
| `spec.buildBatchConfig` | `AwsCodeBuildBatchConfig` |  |  |  |
| `spec.buildBatchConfig.serviceRole` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.buildBatchConfig.combineArtifacts` | `bool` |  |  |  |
| `spec.buildBatchConfig.timeoutInMins` | `int32` |  |  |  |
| `spec.buildBatchConfig.restrictions` | `AwsCodeBuildBatchRestrictions` |  |  |  |
| `spec.buildBatchConfig.restrictions.computeTypesAllowed` | `[]string` |  |  |  |
| `spec.buildBatchConfig.restrictions.maximumBuildsAllowed` | `int32` |  |  |  |
| `spec.projectVisibility` | `string` |  | `PRIVATE` |  |
| `spec.resourceAccessRole` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.resourcePolicy` | `object` |  |  |  |
| `spec.webhook` | `AwsCodeBuildWebhook` |  |  |  |
| `spec.webhook.buildType` | `string` |  |  |  |
| `spec.webhook.manualCreation` | `bool` |  |  |  |
| `spec.webhook.filterGroups` | `[]AwsCodeBuildWebhookFilterGroup` |  |  |  |
| `spec.webhook.filterGroups[].filters` | `[]AwsCodeBuildWebhookFilter` | yes |  |  |
| `spec.webhook.filterGroups[].filters[].type` | `string` | yes |  |  |
| `spec.webhook.filterGroups[].filters[].pattern` | `string` | yes |  |  |
| `spec.webhook.filterGroups[].filters[].excludeMatchedPattern` | `bool` |  |  |  |
| `spec.webhook.scopeConfiguration` | `AwsCodeBuildWebhookScopeConfiguration` |  |  |  |
| `spec.webhook.scopeConfiguration.name` | `string` | yes |  |  |
| `spec.webhook.scopeConfiguration.scope` | `string` | yes |  |  |
| `spec.webhook.scopeConfiguration.domain` | `string` |  |  |  |
| `spec.webhook.pullRequestBuildPolicy` | `AwsCodeBuildWebhookPullRequestBuildPolicy` |  |  |  |
| `spec.webhook.pullRequestBuildPolicy.requiresCommentApproval` | `string` | yes |  |  |
| `spec.webhook.pullRequestBuildPolicy.approverRoles` | `[]string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.source

`AwsCodeBuildSource` · required

source defines where the primary build input comes from and how to
fetch it. Its source_identifier must stay empty — identifiers exist only
to name secondary sources.

- rule: {"required":true}

### spec.source.type

`string` · required

type is the source provider.
  GITHUB:                GitHub.com repository
  BITBUCKET:             Bitbucket Cloud repository
  CODECOMMIT:            AWS CodeCommit repository
  CODEPIPELINE:          Source provided by CodePipeline (no location needed)
  GITHUB_ENTERPRISE:     GitHub Enterprise Server
  GITLAB:                GitLab.com repository
  GITLAB_SELF_MANAGED:   Self-hosted GitLab instance
  NO_SOURCE:             No source — buildspec must be provided inline
  S3:                    S3 bucket containing source archive

- rule: {"required":true,"string":{"in":["GITHUB","BITBUCKET","CODECOMMIT","CODEPIPELINE","GITHUB_ENTERPRISE","GITLAB","GITLAB_SELF_MANAGED","NO_SOURCE","S3"]}}

### spec.source.location

`string`

location is the source code repository URL or S3 path.
Required for all types except CODEPIPELINE and NO_SOURCE.
Format depends on type:
  GITHUB/BITBUCKET/GITLAB: https URL (e.g., https://github.com/owner/repo.git)
  CODECOMMIT: HTTPS clone URL
  S3: bucket/path (e.g., my-bucket/source.zip)
  GITHUB_ENTERPRISE: HTTPS URL of the repo

### spec.source.buildspec

`string`

buildspec is the build specification, either as an inline YAML string or
a path relative to the source root (e.g., "buildspec.yml").
If omitted, CodeBuild looks for buildspec.yml at the source root.
Required when source type is NO_SOURCE.

### spec.source.gitCloneDepth

`int32`

git_clone_depth limits the Git clone depth. 0 means full clone.
Only applicable for Git-based source types.

- rule: {"int32":{"gte":0}}

### spec.source.gitSubmodulesConfig

`AwsCodeBuildGitSubmodulesConfig`

git_submodules_config controls Git submodule fetching during the source
download phase. Presence of the block opts into explicit submodule
handling. Only supported for BITBUCKET, CODECOMMIT, GITHUB, and
GITHUB_ENTERPRISE sources.

### spec.source.gitSubmodulesConfig.fetchSubmodules

`bool`

fetch_submodules fetches all Git submodules during the source download
phase when true.

### spec.source.insecureSsl

`bool`

insecure_ssl skips TLS certificate verification when fetching source —
only for GitHub Enterprise / self-managed GitLab instances with
self-signed certificates. Never enable against public providers.

### spec.source.reportBuildStatus

`bool`

report_build_status reports build start and finish status back to the
source provider (e.g., GitHub commit status checks).
Only applicable for GITHUB, BITBUCKET, GITHUB_ENTERPRISE, GITLAB,
GITLAB_SELF_MANAGED.

### spec.source.buildStatusConfig

`AwsCodeBuildBuildStatusConfig`

build_status_config customizes how the commit status is reported —
the status-check context label and the URL it links to. Only applicable
to the same source types as report_build_status.

### spec.source.buildStatusConfig.context

`string`

context is the status-check label shown by the provider (e.g., the
GitHub status context). Defaults to the CodeBuild project identity when
omitted.

### spec.source.buildStatusConfig.targetUrl

`string`

target_url is the URL the reported status links to. Defaults to the
CodeBuild console page for the build when omitted.

### spec.source.auth

`AwsCodeBuildSourceAuth`

auth pins how CodeBuild authenticates to this source, overriding the
account-level source credential. The modern path is CODECONNECTIONS: a
CodeConnections (formerly CodeStar Connections) connection ARN that
grants repository access without long-lived tokens. SECRETS_MANAGER
points at a secret holding a provider token; OAUTH uses the legacy
account-level OAuth grant.

### spec.source.auth.type

`string` · required

type is the authorization mechanism.
  CODECONNECTIONS:  A CodeConnections connection (recommended; no stored
                    tokens, resource is the connection ARN)
  SECRETS_MANAGER:  A Secrets Manager secret holding the provider token
                    (resource is the secret ARN)
  OAUTH:            The account-level OAuth authorization (legacy)

- rule: {"required":true,"string":{"in":["CODECONNECTIONS","SECRETS_MANAGER","OAUTH"]}}

### spec.source.auth.resource

`string` · required

resource is the ARN of the authorization object — the CodeConnections
connection ARN or the Secrets Manager secret ARN, depending on type.
Always a reference, never the credential value itself.

- rule: {"required":true}

### spec.source.sourceIdentifier

`string`

source_identifier names this source inside the project. REQUIRED for
secondary sources (the buildspec addresses the checkout as
$CODEBUILD_SRC_DIR_<source_identifier>); must stay EMPTY on the primary
source. Alphanumeric and underscore.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"128","pattern":"^[A-Za-z0-9_]+$"}}

### spec.secondarySources

`[]AwsCodeBuildSource`

secondary_sources add up to 12 extra inputs alongside the primary
source (e.g., a shared build-tooling repository next to the application
repository). Each entry MUST set source_identifier; the buildspec
addresses the checkout via $CODEBUILD_SRC_DIR_<source_identifier>.

- rule: {"repeated":{"maxItems":"12"}}

### spec.secondarySources[].type

`string` · required

type is the source provider.
  GITHUB:                GitHub.com repository
  BITBUCKET:             Bitbucket Cloud repository
  CODECOMMIT:            AWS CodeCommit repository
  CODEPIPELINE:          Source provided by CodePipeline (no location needed)
  GITHUB_ENTERPRISE:     GitHub Enterprise Server
  GITLAB:                GitLab.com repository
  GITLAB_SELF_MANAGED:   Self-hosted GitLab instance
  NO_SOURCE:             No source — buildspec must be provided inline
  S3:                    S3 bucket containing source archive

- rule: {"required":true,"string":{"in":["GITHUB","BITBUCKET","CODECOMMIT","CODEPIPELINE","GITHUB_ENTERPRISE","GITLAB","GITLAB_SELF_MANAGED","NO_SOURCE","S3"]}}

### spec.secondarySources[].location

`string`

location is the source code repository URL or S3 path.
Required for all types except CODEPIPELINE and NO_SOURCE.
Format depends on type:
  GITHUB/BITBUCKET/GITLAB: https URL (e.g., https://github.com/owner/repo.git)
  CODECOMMIT: HTTPS clone URL
  S3: bucket/path (e.g., my-bucket/source.zip)
  GITHUB_ENTERPRISE: HTTPS URL of the repo

### spec.secondarySources[].buildspec

`string`

buildspec is the build specification, either as an inline YAML string or
a path relative to the source root (e.g., "buildspec.yml").
If omitted, CodeBuild looks for buildspec.yml at the source root.
Required when source type is NO_SOURCE.

### spec.secondarySources[].gitCloneDepth

`int32`

git_clone_depth limits the Git clone depth. 0 means full clone.
Only applicable for Git-based source types.

- rule: {"int32":{"gte":0}}

### spec.secondarySources[].gitSubmodulesConfig

`AwsCodeBuildGitSubmodulesConfig`

git_submodules_config controls Git submodule fetching during the source
download phase. Presence of the block opts into explicit submodule
handling. Only supported for BITBUCKET, CODECOMMIT, GITHUB, and
GITHUB_ENTERPRISE sources.

### spec.secondarySources[].gitSubmodulesConfig.fetchSubmodules

`bool`

fetch_submodules fetches all Git submodules during the source download
phase when true.

### spec.secondarySources[].insecureSsl

`bool`

insecure_ssl skips TLS certificate verification when fetching source —
only for GitHub Enterprise / self-managed GitLab instances with
self-signed certificates. Never enable against public providers.

### spec.secondarySources[].reportBuildStatus

`bool`

report_build_status reports build start and finish status back to the
source provider (e.g., GitHub commit status checks).
Only applicable for GITHUB, BITBUCKET, GITHUB_ENTERPRISE, GITLAB,
GITLAB_SELF_MANAGED.

### spec.secondarySources[].buildStatusConfig

`AwsCodeBuildBuildStatusConfig`

build_status_config customizes how the commit status is reported —
the status-check context label and the URL it links to. Only applicable
to the same source types as report_build_status.

### spec.secondarySources[].buildStatusConfig.context

`string`

context is the status-check label shown by the provider (e.g., the
GitHub status context). Defaults to the CodeBuild project identity when
omitted.

### spec.secondarySources[].buildStatusConfig.targetUrl

`string`

target_url is the URL the reported status links to. Defaults to the
CodeBuild console page for the build when omitted.

### spec.secondarySources[].auth

`AwsCodeBuildSourceAuth`

auth pins how CodeBuild authenticates to this source, overriding the
account-level source credential. The modern path is CODECONNECTIONS: a
CodeConnections (formerly CodeStar Connections) connection ARN that
grants repository access without long-lived tokens. SECRETS_MANAGER
points at a secret holding a provider token; OAUTH uses the legacy
account-level OAuth grant.

### spec.secondarySources[].auth.type

`string` · required

type is the authorization mechanism.
  CODECONNECTIONS:  A CodeConnections connection (recommended; no stored
                    tokens, resource is the connection ARN)
  SECRETS_MANAGER:  A Secrets Manager secret holding the provider token
                    (resource is the secret ARN)
  OAUTH:            The account-level OAuth authorization (legacy)

- rule: {"required":true,"string":{"in":["CODECONNECTIONS","SECRETS_MANAGER","OAUTH"]}}

### spec.secondarySources[].auth.resource

`string` · required

resource is the ARN of the authorization object — the CodeConnections
connection ARN or the Secrets Manager secret ARN, depending on type.
Always a reference, never the credential value itself.

- rule: {"required":true}

### spec.secondarySources[].sourceIdentifier

`string`

source_identifier names this source inside the project. REQUIRED for
secondary sources (the buildspec addresses the checkout as
$CODEBUILD_SRC_DIR_<source_identifier>); must stay EMPTY on the primary
source. Alphanumeric and underscore.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"128","pattern":"^[A-Za-z0-9_]+$"}}

### spec.secondarySourceVersions

`[]AwsCodeBuildSecondarySourceVersion`

secondary_source_versions pin a branch, tag, or commit per secondary
source (by its source_identifier). Secondary sources without an entry
build from their default branch.

- rule: {"repeated":{"maxItems":"12"}}

### spec.secondarySourceVersions[].sourceIdentifier

`string` · required

source_identifier matches the source_identifier of an entry in
secondary_sources.

- rule: {"required":true,"string":{"maxLen":"128","pattern":"^[A-Za-z0-9_]+$"}}

### spec.secondarySourceVersions[].sourceVersion

`string` · required

source_version is the branch name, tag, commit SHA, or S3 object
version for that source.

- rule: {"required":true}

### spec.environment

`AwsCodeBuildEnvironment` · required

environment defines the build container: image, compute, and variables.

- rule: {"required":true}
- rule: host_kernel applies only to LINUX_CONTAINER, ARM_CONTAINER, LINUX_EC2, and ARM_EC2 environment types

### spec.environment.type

`string` · required

type is the build environment type.
  LINUX_CONTAINER:               Standard Linux (x86_64)
  LINUX_GPU_CONTAINER:           Linux with GPU support
  ARM_CONTAINER:                 Linux (ARM64)
  WINDOWS_CONTAINER:             Windows (legacy alias; prefer the
                                 versioned Windows Server types)
  WINDOWS_SERVER_2019_CONTAINER: Windows Server 2019
  WINDOWS_SERVER_2022_CONTAINER: Windows Server 2022
  LINUX_LAMBDA_CONTAINER:        Lambda-based Linux (x86_64) — fastest
                                 start, no privileged mode or timeouts
  ARM_LAMBDA_CONTAINER:          Lambda-based Linux (ARM64)
  LINUX_EC2:                     EC2-based Linux (reserved fleets)
  ARM_EC2:                       EC2-based Linux ARM (reserved fleets)
  WINDOWS_EC2:                   EC2-based Windows (reserved fleets)
  MAC_ARM:                       macOS on Apple Silicon (reserved
                                 fleets only — iOS/macOS builds)

- rule: {"required":true,"string":{"in":["LINUX_CONTAINER","LINUX_GPU_CONTAINER","ARM_CONTAINER","WINDOWS_CONTAINER","WINDOWS_SERVER_2019_CONTAINER","WINDOWS_SERVER_2022_CONTAINER","LINUX_LAMBDA_CONTAINER","ARM_LAMBDA_CONTAINER","LINUX_EC2","ARM_EC2","WINDOWS_EC2","MAC_ARM"]}}

### spec.environment.computeType

`string` · required

compute_type is the compute capacity for the build container.
  Standard:  BUILD_GENERAL1_SMALL, BUILD_GENERAL1_MEDIUM,
             BUILD_GENERAL1_LARGE, BUILD_GENERAL1_XLARGE,
             BUILD_GENERAL1_2XLARGE
  Lambda:    BUILD_LAMBDA_1GB through BUILD_LAMBDA_10GB
  Fleets:    ATTRIBUTE_BASED_COMPUTE (the fleet picks machines by
             attributes), CUSTOM_INSTANCE_TYPE (the fleet pins an EC2
             instance type) — both only meaningful when the project or
             its fleet uses reserved capacity

- rule: {"required":true,"string":{"in":["BUILD_GENERAL1_SMALL","BUILD_GENERAL1_MEDIUM","BUILD_GENERAL1_LARGE","BUILD_GENERAL1_XLARGE","BUILD_GENERAL1_2XLARGE","BUILD_LAMBDA_1GB","BUILD_LAMBDA_2GB","BUILD_LAMBDA_4GB","BUILD_LAMBDA_8GB","BUILD_LAMBDA_10GB","ATTRIBUTE_BASED_COMPUTE","CUSTOM_INSTANCE_TYPE"]}}

### spec.environment.image

`string` · required

image is the Docker image identifier for the build environment.
Use AWS managed images (e.g., "aws/codebuild/amazonlinux2-x86_64-standard:5.0")
or a custom image URI from ECR or Docker Hub.

- rule: {"required":true}

### spec.environment.certificate

`string`

certificate is an S3 path (bucket/key ending in .pem or .zip) to a
certificate bundle the build trusts — for reaching source providers or
registries behind private CAs.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"\\.(pem|zip)$"}}

### spec.environment.privilegedMode

`bool`

privileged_mode enables Docker daemon access inside the build container.
Required for building Docker images. Not supported for Lambda types.

### spec.environment.imagePullCredentialsType

`string` · optional (explicit presence)

image_pull_credentials_type controls how the build image is pulled.
  CODEBUILD:    CodeBuild uses its own credentials (default, for AWS images)
  SERVICE_ROLE: CodeBuild uses the project's service role (for ECR private images)

- default: `CODEBUILD`
- rule: {"string":{"in":["CODEBUILD","SERVICE_ROLE"]}}

### spec.environment.environmentVariables

`[]AwsCodeBuildEnvironmentVariable`

environment_variables define key-value pairs available during the build.
Supports plaintext values, SSM Parameter Store references, and
Secrets Manager references.

### spec.environment.environmentVariables[].name

`string` · required

name is the environment variable name.

- rule: {"required":true}

### spec.environment.environmentVariables[].value

`string` · required

value is the environment variable value. For PARAMETER_STORE type,
this is the SSM parameter name. For SECRETS_MANAGER, this is the
secret ARN or name. Never put secret material in a PLAINTEXT value —
it is visible to anyone who can describe the project.

- rule: {"required":true}

### spec.environment.environmentVariables[].type

`string` · optional (explicit presence)

type controls how the value is interpreted.
  PLAINTEXT:       Value is used as-is (default)
  PARAMETER_STORE: Value is an SSM Parameter Store parameter name
  SECRETS_MANAGER: Value is a Secrets Manager secret ARN or name

- default: `PLAINTEXT`
- rule: {"string":{"in":["PLAINTEXT","PARAMETER_STORE","SECRETS_MANAGER"]}}

### spec.environment.registryCredential

`AwsCodeBuildRegistryCredential`

registry_credential provides credentials for pulling the build image
from a private Docker registry. Only needed when image_pull_credentials_type
is SERVICE_ROLE and the image is in a non-ECR private registry.

### spec.environment.registryCredential.credential

`string` · required

credential is the ARN or name of the Secrets Manager secret containing
the Docker registry credentials (username + password).

- rule: {"required":true}

### spec.environment.registryCredential.credentialProvider

`string` · required

credential_provider is the credential provider type.
Currently only SECRETS_MANAGER is supported by AWS.

- rule: {"required":true,"string":{"const":"SECRETS_MANAGER"}}

### spec.environment.dockerServer

`AwsCodeBuildDockerServer`

docker_server provisions a persistent, dedicated Docker server for the
project's builds — Docker layer state survives across builds, giving
dramatically faster image builds than a per-build daemon. Choose the
server's compute size independently of the build compute.

### spec.environment.dockerServer.computeType

`string` · required

compute_type sizes the Docker server (same scale as the build compute
types, e.g. BUILD_GENERAL1_MEDIUM).

- rule: {"required":true,"string":{"in":["BUILD_GENERAL1_SMALL","BUILD_GENERAL1_MEDIUM","BUILD_GENERAL1_LARGE","BUILD_GENERAL1_XLARGE","BUILD_GENERAL1_2XLARGE"]}}

### spec.environment.dockerServer.securityGroupIds

`[]string | valueFrom`

security_group_ids attach VPC security groups to the Docker server when
the project runs inside a VPC. Maximum 5.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"maxItems":"5"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.environment.fleetArn

`string`

fleet_arn joins this project to a reserved-capacity fleet — a pool of
pre-provisioned, always-warm build machines (zero queue/provisioning
time; required for MAC_ARM and the EC2 environment types). The fleet is
a shared, account-level resource managed outside this project; reference
it by ARN.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^arn:aws[a-zA-Z-]*:codebuild:[a-z0-9-]+:[0-9]{12}:fleet/.+$"}}

### spec.environment.hostKernel

`string`

host_kernel selects the Linux kernel version for the build hosts.
  LINUX_KERNEL_4:      Kernel 4.x (the long-standing default)
  LINUX_KERNEL_6:      Kernel 6.x (newer eBPF/io_uring features)
  LINUX_KERNEL_LATEST: Always the newest kernel CodeBuild offers
Applies only to the LINUX_CONTAINER, ARM_CONTAINER, LINUX_EC2, and
ARM_EC2 environment types — AWS's contract; Windows, Lambda, and Mac
environments have no kernel selection. Omit to let AWS choose.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["LINUX_KERNEL_4","LINUX_KERNEL_6","LINUX_KERNEL_LATEST"]}}

### spec.artifacts

`AwsCodeBuildArtifacts` · required

artifacts defines where the primary build output goes. Its
artifact_identifier is optional on the primary output.

- rule: {"required":true}

### spec.artifacts.type

`string` · required

type is the artifact output type.
  NO_ARTIFACTS:  No output artifacts (CI-only builds, or push to ECR in buildspec)
  S3:            Write artifacts to an S3 bucket
  CODEPIPELINE:  Artifacts managed by CodePipeline

- rule: {"required":true,"string":{"in":["NO_ARTIFACTS","S3","CODEPIPELINE"]}}

### spec.artifacts.location

`string | valueFrom`

location is the S3 bucket name for artifact output.
Required when type is S3.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.artifacts.name

`string`

name is the artifact output name (object key in S3).

### spec.artifacts.path

`string`

path is the S3 prefix for the artifact output.

### spec.artifacts.packaging

`string`

packaging controls how artifacts are packaged.
  NONE: No packaging (files uploaded as-is)
  ZIP:  Files packaged into a ZIP archive

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NONE","ZIP"]}}

### spec.artifacts.namespaceType

`string`

namespace_type controls whether artifact paths include the build ID.
  NONE:     No namespace (artifacts at path/name)
  BUILD_ID: Artifacts at path/<build-id>/name

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NONE","BUILD_ID"]}}

### spec.artifacts.encryptionDisabled

`bool`

encryption_disabled disables server-side encryption for artifacts.
By default, CodeBuild encrypts artifacts using the project's encryption key.

### spec.artifacts.overrideArtifactName

`bool`

override_artifact_name lets the buildspec's artifacts.name override the
name configured here — used for per-build artifact names (e.g., embedding
the version or commit).

### spec.artifacts.bucketOwnerAccess

`string`

bucket_owner_access grants the owning account of the destination bucket
access to the uploaded artifacts (for cross-account artifact buckets).
  NONE:      Bucket owner gets no access (default)
  READ_ONLY: Bucket owner can read the artifacts
  FULL:      Bucket owner has full control

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NONE","READ_ONLY","FULL"]}}

### spec.artifacts.artifactIdentifier

`string`

artifact_identifier names this output. REQUIRED for secondary artifacts
(matching the buildspec's secondary-artifacts key); optional on the
primary output. Alphanumeric and underscore.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"128","pattern":"^[A-Za-z0-9_]+$"}}

### spec.secondaryArtifacts

`[]AwsCodeBuildArtifacts`

secondary_artifacts add up to 12 extra output locations (e.g., publish
a container context to one bucket and test reports to another). Each
entry MUST set artifact_identifier, matching the identifier used in the
buildspec's secondary-artifacts section.

- rule: {"repeated":{"maxItems":"12"}}

### spec.secondaryArtifacts[].type

`string` · required

type is the artifact output type.
  NO_ARTIFACTS:  No output artifacts (CI-only builds, or push to ECR in buildspec)
  S3:            Write artifacts to an S3 bucket
  CODEPIPELINE:  Artifacts managed by CodePipeline

- rule: {"required":true,"string":{"in":["NO_ARTIFACTS","S3","CODEPIPELINE"]}}

### spec.secondaryArtifacts[].location

`string | valueFrom`

location is the S3 bucket name for artifact output.
Required when type is S3.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.secondaryArtifacts[].name

`string`

name is the artifact output name (object key in S3).

### spec.secondaryArtifacts[].path

`string`

path is the S3 prefix for the artifact output.

### spec.secondaryArtifacts[].packaging

`string`

packaging controls how artifacts are packaged.
  NONE: No packaging (files uploaded as-is)
  ZIP:  Files packaged into a ZIP archive

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NONE","ZIP"]}}

### spec.secondaryArtifacts[].namespaceType

`string`

namespace_type controls whether artifact paths include the build ID.
  NONE:     No namespace (artifacts at path/name)
  BUILD_ID: Artifacts at path/<build-id>/name

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NONE","BUILD_ID"]}}

### spec.secondaryArtifacts[].encryptionDisabled

`bool`

encryption_disabled disables server-side encryption for artifacts.
By default, CodeBuild encrypts artifacts using the project's encryption key.

### spec.secondaryArtifacts[].overrideArtifactName

`bool`

override_artifact_name lets the buildspec's artifacts.name override the
name configured here — used for per-build artifact names (e.g., embedding
the version or commit).

### spec.secondaryArtifacts[].bucketOwnerAccess

`string`

bucket_owner_access grants the owning account of the destination bucket
access to the uploaded artifacts (for cross-account artifact buckets).
  NONE:      Bucket owner gets no access (default)
  READ_ONLY: Bucket owner can read the artifacts
  FULL:      Bucket owner has full control

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NONE","READ_ONLY","FULL"]}}

### spec.secondaryArtifacts[].artifactIdentifier

`string`

artifact_identifier names this output. REQUIRED for secondary artifacts
(matching the buildspec's secondary-artifacts key); optional on the
primary output. Alphanumeric and underscore.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"128","pattern":"^[A-Za-z0-9_]+$"}}

### spec.serviceRole

`string | valueFrom` · required

service_role is the IAM role ARN that grants CodeBuild permission to
access source code, write artifacts, publish logs, and interact with
other AWS services during the build.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.description

`string`

description is a human-readable description of the project (max 255 chars).

- rule: {"string":{"maxLen":"255"}}

### spec.encryptionKey

`string | valueFrom`

encryption_key is the ARN of a KMS key used to encrypt build artifacts.
If omitted, CodeBuild uses the AWS-managed key for S3.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.buildTimeout

`int32` · optional (explicit presence)

build_timeout is the maximum duration of a single build, in minutes.
Range: 5-2160 (36 hours). Default: 60. Not supported for Lambda compute
environment types (LINUX_LAMBDA_CONTAINER, ARM_LAMBDA_CONTAINER), where
AWS caps every build at the Lambda maximum instead.

- default: `60`
- rule: {"int32":{"lte":2160,"gte":5}}

### spec.queuedTimeout

`int32` · optional (explicit presence)

queued_timeout is the maximum time a build can wait in the queue before
timing out, in minutes. Range: 5-480 (8 hours). Default: 480. Not
supported for Lambda compute environment types.

- default: `480`
- rule: {"int32":{"lte":480,"gte":5}}

### spec.concurrentBuildLimit

`int32`

concurrent_build_limit caps the number of concurrent builds. Useful for
cost control. Minimum 1. Omit to allow unlimited concurrency.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.autoRetryLimit

`int32`

auto_retry_limit is the number of ADDITIONAL automatic retries after a
failed build (e.g., 2 means up to three attempts total). AWS allows up
to 10. Omit (or 0) to disable automatic retry.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":10,"gte":1}}

### spec.badgeEnabled

`bool`

badge_enabled publishes a dynamic build badge for the project. The badge
URL is exported as the badge_url output and can be embedded in a
repository README. Not supported for CODEPIPELINE or S3 sources.

### spec.sourceVersion

`string`

source_version is the default branch, tag, or commit ID to build.
For GitHub: branch name, tag, or full commit SHA.
For S3: object version ID.

### spec.cache

`AwsCodeBuildCache`

cache configures build caching to speed up subsequent builds.

### spec.cache.type

`string` · optional (explicit presence)

type is the cache type.
  NO_CACHE: Caching disabled (default)
  S3:       Cache stored in an S3 bucket
  LOCAL:    Cache stored on the build host (ephemeral, useful for Docker layers)

- default: `NO_CACHE`
- rule: {"string":{"in":["NO_CACHE","S3","LOCAL"]}}

### spec.cache.location

`string | valueFrom`

location is the S3 bucket and optional prefix for cache storage.
Required when type is S3. Format: "bucket-name" or "bucket-name/prefix".

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.cache.modes

`[]string`

modes specifies what to cache when type is LOCAL.
  LOCAL_SOURCE_CACHE:        Cache Git metadata
  LOCAL_DOCKER_LAYER_CACHE:  Cache Docker layers
  LOCAL_CUSTOM_CACHE:        Cache paths specified in buildspec

- rule: {"repeated":{"items":{"string":{"in":["LOCAL_SOURCE_CACHE","LOCAL_DOCKER_LAYER_CACHE","LOCAL_CUSTOM_CACHE"]}}}}

### spec.cache.cacheNamespace

`string`

cache_namespace scopes S3 cache keys so multiple projects (or branches)
can share one cache bucket without collisions.

### spec.logsConfig

`AwsCodeBuildLogsConfig`

logs_config controls where build logs are sent.

### spec.logsConfig.cloudwatchLogs

`AwsCodeBuildCloudWatchLogs`

cloudwatch_logs configures CloudWatch Logs for build output.

### spec.logsConfig.cloudwatchLogs.status

`string` · optional (explicit presence)

status controls whether CloudWatch logging is enabled.
Default: ENABLED.

- default: `ENABLED`
- rule: {"string":{"in":["ENABLED","DISABLED"]}}

### spec.logsConfig.cloudwatchLogs.groupName

`string | valueFrom`

group_name is the CloudWatch Logs log group name.
If omitted, CodeBuild creates a default log group.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_name}} -- a bare string does not parse

### spec.logsConfig.cloudwatchLogs.streamName

`string`

stream_name is the CloudWatch Logs log stream name prefix.
If omitted, CodeBuild generates a default stream name.

### spec.logsConfig.s3Logs

`AwsCodeBuildS3Logs`

s3_logs configures S3 logging for build output.

- rule: an ENABLED s3_logs requires bucket and prefix — AWS stores S3 build logs only under "bucket/prefix"

### spec.logsConfig.s3Logs.status

`string` · optional (explicit presence)

status controls whether S3 logging is enabled.
Default: DISABLED.

- default: `DISABLED`
- rule: {"string":{"in":["ENABLED","DISABLED"]}}

### spec.logsConfig.s3Logs.bucket

`string | valueFrom`

bucket is the S3 bucket receiving build logs — a reference to an
AwsS3Bucket's bucket_id output, a literal bucket name, or the bucket
ARN. AWS stores S3 build logs only under a prefix, so the bucket and
`prefix` are separate fields here and both IaC modules compose the
provider's single location argument as "{bucket}/{prefix}".

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.logsConfig.s3Logs.encryptionDisabled

`bool`

encryption_disabled disables server-side encryption for log files.

### spec.logsConfig.s3Logs.bucketOwnerAccess

`string`

bucket_owner_access grants the owning account of the log bucket access
to the uploaded log files (for centralized cross-account log buckets).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NONE","READ_ONLY","FULL"]}}

### spec.logsConfig.s3Logs.prefix

`string`

prefix is the path under the bucket where log objects land. AWS
REQUIRES it for S3 build logs — a bare bucket with no prefix is
rejected at CreateProject, which is why it is a separate required
companion to `bucket` rather than a "bucket/prefix" literal format.

### spec.vpcConfig

`AwsCodeBuildVpcConfig`

vpc_config places the build in a VPC, giving it access to private
resources such as RDS databases, ElastiCache clusters, or internal APIs.

### spec.vpcConfig.vpcId

`string | valueFrom` · required

vpc_id is the VPC where build containers are launched.

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.vpcConfig.subnetIds

`[]string | valueFrom` · required

subnet_ids are the VPC subnets where build containers are placed.
Use private subnets for security. Maximum 16 subnets.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true,"repeated":{"minItems":"1","maxItems":"16"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.vpcConfig.securityGroupIds

`[]string | valueFrom` · required

security_group_ids are the VPC security groups applied to build containers.
Maximum 5 security groups.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"required":true,"repeated":{"minItems":"1","maxItems":"5"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.fileSystemLocations

`[]AwsCodeBuildFileSystemLocation`

file_system_locations mount Amazon EFS file systems into the build
container — useful for large shared caches (e.g., a Gradle or Bazel
cache) that outlive individual builds. Requires vpc_config placing the
build in subnets that can reach the file system's mount targets.

### spec.fileSystemLocations[].identifier

`string` · required

identifier names this mount; CodeBuild exposes it to the build as the
environment variable CODEBUILD_<identifier> (uppercased).

- rule: {"required":true}

### spec.fileSystemLocations[].location

`string` · required

location is the EFS mount source in the form
<file-system-dns>:/<path>, e.g.
"fs-0abc123.efs.us-east-1.amazonaws.com:/build-cache". The DNS name
composes from an AwsElasticFileSystem's file_system_id as
<file_system_id>.efs.<region>.amazonaws.com.

- rule: {"required":true}

### spec.fileSystemLocations[].mountPoint

`string` · required

mount_point is the absolute path inside the build container where the
file system is mounted (e.g., "/mnt/build-cache").

- rule: {"required":true}

### spec.fileSystemLocations[].mountOptions

`string`

mount_options are NFS mount options. Omit for the EFS recommended
defaults (nfsvers=4.1,rsize=1048576,wsize=1048576,hard,timeo=600,retrans=2).

### spec.fileSystemLocations[].type

`string` · optional (explicit presence)

type is the file system protocol. AWS currently supports only EFS.

- default: `EFS`
- rule: {"string":{"const":"EFS"}}

### spec.buildBatchConfig

`AwsCodeBuildBatchConfig`

build_batch_config enables batch builds — a single StartBuildBatch call
fans out into multiple coordinated builds defined by the buildspec's
batch section (build graphs, build lists, matrix builds).

### spec.buildBatchConfig.serviceRole

`string | valueFrom` · required

service_role is the IAM role CodeBuild assumes to launch the child
builds of a batch. May be the project's service role or a dedicated one.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.buildBatchConfig.combineArtifacts

`bool`

combine_artifacts merges the child builds' artifacts into a single
artifact location for the whole batch.

### spec.buildBatchConfig.timeoutInMins

`int32`

timeout_in_mins is the maximum duration of the entire batch, in minutes
(5-2160).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":2160,"gte":5}}

### spec.buildBatchConfig.restrictions

`AwsCodeBuildBatchRestrictions`

restrictions bound what the batch's child builds may consume.

### spec.buildBatchConfig.restrictions.computeTypesAllowed

`[]string`

compute_types_allowed restricts which compute types child builds may
request (values from the environment compute_type set). Empty allows all.

- rule: {"repeated":{"items":{"string":{"in":["BUILD_GENERAL1_SMALL","BUILD_GENERAL1_MEDIUM","BUILD_GENERAL1_LARGE","BUILD_GENERAL1_XLARGE","BUILD_GENERAL1_2XLARGE","BUILD_LAMBDA_1GB","BUILD_LAMBDA_2GB","BUILD_LAMBDA_4GB","BUILD_LAMBDA_8GB","BUILD_LAMBDA_10GB","ATTRIBUTE_BASED_COMPUTE","CUSTOM_INSTANCE_TYPE"]}}}}

### spec.buildBatchConfig.restrictions.maximumBuildsAllowed

`int32`

maximum_builds_allowed caps how many child builds one batch may spawn
(1-100).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":100,"gte":1}}

### spec.projectVisibility

`string` · optional (explicit presence)

project_visibility controls whether the project's builds are publicly
readable.
  PRIVATE:     Only principals in the account can view builds (default)
  PUBLIC_READ: Build results, logs, and artifacts are world-readable —
               used by open-source projects publishing CI results

- default: `PRIVATE`
- rule: {"string":{"in":["PRIVATE","PUBLIC_READ"]}}

### spec.resourceAccessRole

`string | valueFrom`

resource_access_role is the IAM role CodeBuild uses to read the CloudWatch
logs and S3 artifacts it exposes for public builds. Required when
project_visibility is PUBLIC_READ.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.resourcePolicy

`object`

resource_policy is a resource-based IAM policy document attached to the
project — the mechanism for cross-account access (e.g., letting a
central CI account start builds in this account). Provide the policy as
a structured JSON document.

### spec.webhook

`AwsCodeBuildWebhook`

webhook configures automatic build triggers from the source provider.
Only valid when source type supports webhooks: GITHUB, BITBUCKET,
GITHUB_ENTERPRISE, GITLAB, GITLAB_SELF_MANAGED, CODECOMMIT.
Omit for CodePipeline-triggered or manual-only projects.

### spec.webhook.buildType

`string`

build_type controls the build type triggered by the webhook.
  BUILD:                  Standard single build (default)
  BUILD_BATCH:            Batch build (requires build_batch_config)
  RUNNER_BUILDKITE_BUILD: The project acts as a Buildkite runner —
                          webhook events dispatch Buildkite jobs

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["BUILD","BUILD_BATCH","RUNNER_BUILDKITE_BUILD"]}}

### spec.webhook.manualCreation

`bool`

manual_creation makes CodeBuild return the payload URL and HMAC secret
WITHOUT registering the webhook with the provider — you configure the
repository webhook by hand from the webhook_payload_url and
webhook_secret outputs. Required for GitHub Enterprise; useful
when the connection lacks admin rights on the repository.

### spec.webhook.filterGroups

`[]AwsCodeBuildWebhookFilterGroup`

filter_groups define which repository events trigger a build.
Multiple groups are OR'd together — a build triggers if ANY group matches.
Within a group, filters are AND'd — ALL filters in the group must match.

### spec.webhook.filterGroups[].filters

`[]AwsCodeBuildWebhookFilter` · required

filters are the individual conditions in this group.

- rule: {"repeated":{"minItems":"1"}}

### spec.webhook.filterGroups[].filters[].type

`string` · required

type is the event field to match against.
  EVENT:             Event type (PUSH, PULL_REQUEST_CREATED, PULL_REQUEST_UPDATED, ...)
  BASE_REF:          Base branch for PRs (regex)
  HEAD_REF:          Branch or tag name (regex)
  ACTOR_ACCOUNT_ID:  Source provider account ID
  FILE_PATH:         Changed file paths (regex)
  COMMIT_MESSAGE:    Commit message (regex)
  WORKFLOW_NAME:     GitHub Actions workflow name (runner projects)
  TAG_NAME:          Release tag name (regex)
  RELEASE_NAME:      Release name (regex)
  REPOSITORY_NAME:   Repository name (org-scoped webhooks; regex)
  ORGANIZATION_NAME: Organization name (global-scoped webhooks; regex)

- rule: {"required":true,"string":{"in":["EVENT","BASE_REF","HEAD_REF","ACTOR_ACCOUNT_ID","FILE_PATH","COMMIT_MESSAGE","WORKFLOW_NAME","TAG_NAME","RELEASE_NAME","REPOSITORY_NAME","ORGANIZATION_NAME"]}}

### spec.webhook.filterGroups[].filters[].pattern

`string` · required

pattern is the regex pattern or comma-separated event list to match.
For EVENT type: comma-separated values like "PUSH, PULL_REQUEST_CREATED".
For other types: a regex pattern (e.g., "^refs/heads/main$").

- rule: {"required":true}

### spec.webhook.filterGroups[].filters[].excludeMatchedPattern

`bool`

exclude_matched_pattern inverts the match — the filter passes when the
pattern does NOT match. Useful for excluding branches or paths.

### spec.webhook.scopeConfiguration

`AwsCodeBuildWebhookScopeConfiguration`

scope_configuration widens the webhook beyond one repository — an
organization- or group-level webhook (used with runner projects and
org-wide CI). The webhook then fires for every repository in scope.

### spec.webhook.scopeConfiguration.name

`string` · required

name is the organization (GitHub) or group (GitLab) the webhook covers.

- rule: {"required":true}

### spec.webhook.scopeConfiguration.scope

`string` · required

scope is the webhook's coverage.
  GITHUB_ORGANIZATION: All repositories in a GitHub organization
  GITHUB_GLOBAL:       All repositories the connection can reach
  GITLAB_GROUP:        All projects in a GitLab group

- rule: {"required":true,"string":{"in":["GITHUB_ORGANIZATION","GITHUB_GLOBAL","GITLAB_GROUP"]}}

### spec.webhook.scopeConfiguration.domain

`string`

domain is the self-hosted provider domain (GitHub Enterprise Server /
self-managed GitLab). Omit for the cloud-hosted providers.

### spec.webhook.pullRequestBuildPolicy

`AwsCodeBuildWebhookPullRequestBuildPolicy`

pull_request_build_policy gates pull-request-triggered builds behind a
comment approval — protection against untrusted code running in CI from
fork PRs.

### spec.webhook.pullRequestBuildPolicy.requiresCommentApproval

`string` · required

requires_comment_approval controls which pull requests need an approval
comment before building.
  DISABLED:           All PRs build immediately (default provider behavior)
  FORK_PULL_REQUESTS: Only PRs from forks wait for approval — the common
                      open-source posture (protects secrets from
                      untrusted fork code)
  ALL_PULL_REQUESTS:  Every PR waits for an approval comment

- rule: {"required":true,"string":{"in":["DISABLED","FORK_PULL_REQUESTS","ALL_PULL_REQUESTS"]}}

### spec.webhook.pullRequestBuildPolicy.approverRoles

`[]string`

approver_roles are the repository roles whose comments count as
approval. Defaults are provider-managed when empty.

- rule: {"repeated":{"items":{"string":{"in":["GITHUB_READ","GITHUB_TRIAGE","GITHUB_WRITE","GITHUB_MAINTAIN","GITHUB_ADMIN","GITLAB_GUEST","GITLAB_PLANNER","GITLAB_REPORTER","GITLAB_DEVELOPER","GITLAB_MAINTAINER","GITLAB_OWNER","BITBUCKET_READ","BITBUCKET_WRITE","BITBUCKET_ADMIN"]}}}}

## Validation Rules

- `codepipeline_source_artifacts_match`: when source.type is CODEPIPELINE, artifacts.type must also be CODEPIPELINE (and vice versa)
- `source_location_required`: source.location is required when source.type is not CODEPIPELINE or NO_SOURCE
- `buildspec_required_for_no_source`: source.buildspec is required when source.type is NO_SOURCE
- `primary_source_no_identifier`: source.source_identifier must be empty on the primary source; it is only used to name secondary sources
- `secondary_sources_require_identifier`: every secondary_sources entry must set source_identifier
- `secondary_artifacts_require_identifier`: every secondary_artifacts entry must set artifact_identifier
- `s3_artifacts_require_location`: artifacts.location is required when artifacts.type is S3
- `s3_cache_requires_location`: cache.location is required when cache.type is S3
- `lambda_env_no_timeouts`: build_timeout and queued_timeout are not supported for Lambda environment types (LINUX_LAMBDA_CONTAINER, ARM_LAMBDA_CONTAINER)
- `lambda_env_no_privileged_mode`: environment.privileged_mode is not supported for Lambda environment types
- `registry_credential_requires_service_role_pull`: environment.registry_credential requires environment.image_pull_credentials_type to be SERVICE_ROLE
- `public_visibility_requires_resource_access_role`: resource_access_role is required when project_visibility is PUBLIC_READ
- `webhook_requires_compatible_source`: webhook is only valid when source.type supports webhooks (GITHUB, BITBUCKET, GITHUB_ENTERPRISE, GITLAB, GITLAB_SELF_MANAGED, CODECOMMIT)
- `badge_requires_repository_source`: badge_enabled is not supported when source.type is CODEPIPELINE, S3, or NO_SOURCE
- `git_submodules_require_git_source`: git_submodules_config is only supported for BITBUCKET, CODECOMMIT, GITHUB, and GITHUB_ENTERPRISE sources
- `build_status_requires_supported_source`: report_build_status and build_status_config are only supported for GITHUB, BITBUCKET, GITHUB_ENTERPRISE, GITLAB, and GITLAB_SELF_MANAGED sources

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsCodeBuildProject, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.project_arn` | `string` | The Amazon Resource Name (ARN) of the CodeBuild project. Use this for IAM policies, EventBridge targets, and cross-resource references. |
| `status.outputs.project_name` | `string` | The name of the CodeBuild project. Use this in CodePipeline Build action configurations (the ProjectName key). |
| `status.outputs.service_role_arn` | `string` | The IAM service role ARN used by the project. Echoed back for downstream reference in job definitions or policies. |
| `status.outputs.badge_url` | `string` | The build badge URL, when badge_enabled is true. Embed it in a repository README to show live build status. Empty when the badge is disabled. |
| `status.outputs.public_project_alias` | `string` | The public alias of the project, when project_visibility is PUBLIC_READ. This is the identifier in the public build results URL. Empty for private projects. |
| `status.outputs.webhook_url` | `string` | The webhook URL for the source provider, if a webhook was created. Empty when no webhook is configured. |
| `status.outputs.webhook_payload_url` | `string` | The webhook payload URL, if a webhook was created. This is the URL the source provider posts events to; with webhook.manual_creation, register it on the repository by hand. Empty when no webhook is configured. |
| `status.outputs.webhook_secret` | `string` (sensitive) | The webhook's HMAC signing secret (sensitive — a provider-minted credential, only returned at webhook creation). With webhook.manual_creation, configure it on the repository webhook so CodeBuild can authenticate payloads. Empty when no webhook is configured. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.environment.dockerServer.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.artifacts.location` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.secondaryArtifacts[].location` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.serviceRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.encryptionKey` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.cache.location` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.logsConfig.cloudwatchLogs.groupName` | AwsCloudwatchLogGroup | `status.outputs.log_group_name` |
| `spec.logsConfig.s3Logs.bucket` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.vpcConfig.vpcId` | AwsVpc | `status.outputs.vpc_id` |
| `spec.vpcConfig.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.vpcConfig.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.buildBatchConfig.serviceRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.resourceAccessRole` | AwsIamRole | `status.outputs.role_arn` |

## See Also

- [Overview](../README.md)
