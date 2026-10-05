# AwsSagemakerDomain

Amazon SageMaker Domain resource for Planton. Provisions the top-level organizational unit for Amazon SageMaker Studio — a shared, managed workspace where data scientists and ML engineers access JupyterLab notebooks, the Code Editor (VS Code) IDE, Canvas no-code ML, custom kernels, collaborative editing, and ML tooling. The domain handles VPC networking, EFS-backed home directories, IAM execution roles, Docker access, idle-timeout cost controls, RStudio (Posit) Workbench activation, and identity-propagation posture so teams can focus on building models instead of managing infrastructure.

## When to use

- You need a managed ML workspace on AWS for data science or ML engineering teams.
- Teams require JupyterLab or VS Code (Code Editor) environments with shared storage, Git integration, and configurable compute instances.
- You want centralized governance over ML workspaces: approved container images, idle-timeout policies, network boundaries, storage quotas, and UI-level hiding of expensive instance types.
- Your security posture requires VPC-only notebook access with no public internet exposure, per-user CloudTrail attribution (`executionRoleIdentityConfig`), or Identity Center trusted identity propagation into Athena/Redshift/Lake Formation.
- You need Docker-in-notebook capability for building custom training containers or inference images.
- Business teams need SageMaker Canvas (no-code ML) with governed capabilities: model registry instead of direct deployment, curated SaaS connectors, Bedrock/Forecast roles.

## Prerequisites

| Prerequisite | Why | Planton Resource |
|---|---|---|
| VPC with subnets in 1+ AZs | Domain ENIs for notebook/training traffic are placed in subnets; 2+ AZs recommended for HA | `AwsVpc` |
| IAM execution role | SageMaker assumes this role to access S3, ECR, Secrets Manager, etc. on behalf of users; must trust `sagemaker.amazonaws.com` | `AwsIamRole` |
| Security groups (optional) | Control inbound/outbound traffic for user notebooks and domain-scoped apps | `AwsSecurityGroup` |
| KMS key (optional) | Customer-managed encryption for the domain's EFS home directory volume | `AwsKmsKey` |
| EFS file system (optional) | Additional shared file systems mounted into apps via `customFileSystemConfigs` | `AwsElasticFileSystem` |
| IAM Identity Center (optional) | Required when `authMode` is `SSO`, and for trusted identity propagation | (external) |
| Posit license via AWS License Manager (optional) | Required to activate RStudio Server Pro on the domain | (external) |

## API envelope

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerDomain
metadata:
  name: <resource-name>   # becomes the AWS domain name: 1-63 chars of [0-9A-Za-z-]
spec: { ... }
```

## Spec fields reference

### Core (ForceNew)

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `authMode` | string | **yes** | — | Authentication mode: `IAM` or `SSO`. IAM uses AWS credentials; SSO uses IAM Identity Center. **ForceNew**. |
| `vpcId` | StringValueOrRef | **yes** | — | VPC for domain network interfaces. Must have DNS resolution and DNS hostnames enabled. Supports `value` or `valueFrom` (AwsVpc). **ForceNew**. |
| `subnetIds` | list(StringValueOrRef) | **yes** (1–16) | — | Subnets for notebook/training ENIs. 2+ AZs recommended. **ForceNew**. |
| `kmsKeyId` | StringValueOrRef | no | aws/elasticfilesystem | KMS key for EFS home directory encryption. **ForceNew**. |
| `homeEfsRetentionPolicy` | string | no | `Retain` | What happens to the domain's auto-created EFS file system on destroy: `Retain` preserves it (and its storage charges); `Delete` removes it with the domain — the right choice for ephemeral domains. **ForceNew**. |

### Network & lifecycle posture

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `appNetworkAccessType` | string | no | `PublicInternetOnly` | `PublicInternetOnly` — ENIs have internet access via AWS-managed networking. `VpcOnly` — all traffic stays within the VPC; requires NAT for internet. Recommended for production. |
| `appSecurityGroupManagement` | string | no | — | `Service` (SageMaker manages app-ENI security groups) or `Customer`. Only honored by AWS when RStudio is configured — the spec requires the pairing. |
| `tagPropagation` | string | no | `DISABLED` | `ENABLED` copies the domain's tags onto the apps, spaces, and user profiles created inside it — recommended for cost allocation. |

### Domain-scoped administration

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `domainSecurityGroupIds` | list(StringValueOrRef) | no | [] | Security groups for domain-scoped apps and shared resources. Max 3. **ForceNew**. |
| `dockerSettings` | DockerSettings | no | — | Docker access configuration (see nested message below). |
| `executionRoleIdentityConfig` | string | no | — | `USER_PROFILE_NAME` stamps each session's `sts:SourceIdentity` with the acting user profile (per-user CloudTrail attribution through the shared role); `DISABLED` leaves sessions role-only. |
| `rStudioServerProDomainSettings` | RStudioServerProDomainSettings | no | — | Activates RStudio (Posit) Workbench on the domain (see nested message below). Requires a Posit license via AWS License Manager. |
| `trustedIdentityPropagationStatus` | string | no | — | `ENABLED` forwards the Identity Center user identity through SageMaker to downstream analytics services (Athena, Redshift, Lake Formation). Requires `authMode: SSO`. |

### Default User Settings

Nested message `defaultUserSettings` — inherited by all user profiles.

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `executionRoleArn` | StringValueOrRef | **yes** | — | IAM role assumed by SageMaker for notebooks/training. Must trust `sagemaker.amazonaws.com`. |
| `securityGroupIds` | list(StringValueOrRef) | no | [] | User-level security groups for notebook ENIs. Max 5. |
| `defaultLandingUri` | string | no | platform default | URI of the default app opened on login. Common: `studio::relative/JupyterLab`, `studio::relative/JupyterServer:`, `studio::`. |
| `studioWebPortal` | string | no | `ENABLED` | `ENABLED` for full Studio web UI; `DISABLED` for programmatic-only access. |
| `autoMountHomeEfs` | string | no | — | `Enabled` auto-mounts each user's EFS home directory into apps; `Disabled` starts apps without it (pair with space storage or custom file systems). |
| `jupyterLabAppSettings` | JupyterLabAppSettings | no | — | JupyterLab IDE configuration (see nested message below). |
| `jupyterServerAppSettings` | JupyterServerAppSettings | no | — | Classic Jupyter Server (Studio Classic) configuration. |
| `kernelGatewayAppSettings` | KernelGatewayAppSettings | no | — | Custom kernel configuration (see nested message below). |
| `codeEditorAppSettings` | CodeEditorAppSettings | no | — | Code Editor (VS Code / Code-OSS) configuration — same controls as JupyterLab. |
| `tensorBoardAppSettings` | TensorBoardAppSettings | no | — | TensorBoard default resource spec. |
| `rSessionAppSettings` | RSessionAppSettings | no | — | RSession (R kernel) resource spec and custom images. Requires RStudio on the domain. |
| `rStudioServerProAppSettings` | RStudioServerProAppSettings | no | — | Per-user RStudio access (`accessStatus`) and authorization level (`userGroup`). |
| `canvasAppSettings` | CanvasAppSettings | no | — | SageMaker Canvas (no-code ML) capabilities (see nested message below). |
| `sharingSettings` | SharingSettings | no | — | Notebook output sharing to S3 (see nested message below). |
| `spaceStorageSettings` | SpaceStorageSettings | no | — | Default EBS volume sizes for user spaces (see nested message below). |
| `customFileSystemConfigs` | list(CustomFileSystemConfig) | no | [] | Additional file systems mounted into every user's apps — shared datasets, feature stores, model trees. |
| `customPosixUserConfig` | CustomPosixUserConfig | no | — | POSIX identity (UID ≥ 10000 / GID ≥ 1001) apps run as when accessing EFS and custom file systems. |
| `studioWebPortalSettings` | StudioWebPortalSettings | no | — | Hide app types, instance types, or ML tools from the Studio UI. |

### Default Space Settings

Nested message `defaultSpaceSettings` — inherited by all shared (collaborative) spaces. Spaces run under their own execution role, separate from the per-user plane.

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `executionRoleArn` | StringValueOrRef | **yes** | — | IAM role for workloads in shared spaces. Must trust `sagemaker.amazonaws.com`. |
| `securityGroupIds` | list(StringValueOrRef) | no | [] | Security groups for shared-space app ENIs. Max 5. |
| `jupyterLabAppSettings` | JupyterLabAppSettings | no | — | JupyterLab baseline for spaces (same shape as the user-level block). |
| `jupyterServerAppSettings` | JupyterServerAppSettings | no | — | Classic Jupyter Server baseline for spaces. |
| `kernelGatewayAppSettings` | KernelGatewayAppSettings | no | — | KernelGateway baseline for spaces. |
| `spaceStorageSettings` | SpaceStorageSettings | no | — | Default/maximum EBS sizes for spaces. |
| `customFileSystemConfigs` | list(CustomFileSystemConfig) | no | [] | Additional file systems mounted into space apps. |
| `customPosixUserConfig` | CustomPosixUserConfig | no | — | POSIX identity for space file-system operations. |

### User Profiles (folded satellites)

Repeated `userProfiles` — the per-person workspaces inside the domain, one
`aws_sagemaker_user_profile` per entry, keyed by name (adding or removing a
person never disturbs the others; removing an entry deletes that profile
and its home-directory surfaces). Profiles created outside the manifest
(SSO auto-provisioning, console) are not managed or removed by it.

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `userProfileName` | string | **yes** | — | Unique in the domain; 1–63 chars, alphanumeric + hyphens. **ForceNew**. |
| `singleSignOnUserIdentifier` | string | no | — | SSO domains: must be `UserName`; set together with the value below. **ForceNew**. |
| `singleSignOnUserValue` | string | no | — | SSO domains: the Identity Center username this profile belongs to. **ForceNew**. |
| `userSettings` | UserSettings | no | — | Per-user overrides of `defaultUserSettings` — the SAME settings tree (`executionRoleArn` required inside when set). |

### Spaces (folded satellites)

Repeated `spaces` — named shared (or private) workspaces, one
`aws_sagemaker_space` per entry, keyed by name. A space's runtime and EBS
volume are shared by everyone with access. The space settings tree is
deliberately different from `defaultSpaceSettings` (AWS uses distinct
types): it adds `appType`, requires a resource spec per configured app,
carries a timeout-only idle dial, and mounts existing EFS file systems by
id. Profiles create before spaces (ownership references the profile by
name).

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `spaceName` | string | **yes** | — | Unique in the domain; 1–63 chars. **ForceNew**. |
| `displayName` | string | no | — | Studio display name (mutable; max 64 chars). |
| `ownershipSettings.ownerUserProfileName` | string | with sharing | — | The owning profile. Declared together with `spaceSharingSettings`; not updatable after create. |
| `spaceSharingSettings.sharingType` | string | with ownership | — | `Private` or `Shared`. Not updatable after create. |
| `spaceSettings` | SpaceSettings | no | — | `appType` (`JupyterLab`/`CodeEditor`/`JupyterServer`/`KernelGateway`), per-app blocks (resource spec required), `customFileSystems` (EFS by id), `spaceStorageSettings.ebsVolumeSizeInGb` (5–16384). |

### JupyterLabAppSettings

Used by `defaultUserSettings.jupyterLabAppSettings` and `defaultSpaceSettings.jupyterLabAppSettings`:

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `defaultResourceSpec` | ResourceSpec | no | — | Default compute instance type and image configuration for JupyterLab apps. |
| `lifecycleConfigArns` | list(string) | no | [] | ARNs of lifecycle scripts run at JupyterLab startup (install packages, configure extensions). |
| `builtInLifecycleConfigArn` | string | no | — | ARN of an AWS-curated (built-in) lifecycle configuration. |
| `customImages` | list(CustomImage) | no | [] | Custom Docker images available as JupyterLab kernels. Max 200. |
| `codeRepositories` | list(CodeRepository) | no | [] | Git repos auto-cloned into JupyterLab on startup. Max 10. |
| `idleSettings` | IdleSettings | no | — | Automatic shutdown of idle JupyterLab instances (see nested message below). |
| `emrSettings` | EmrSettings | no | — | IAM roles for discovering and connecting to EMR clusters from notebooks. |

### JupyterServerAppSettings

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `defaultResourceSpec` | ResourceSpec | no | — | Default configuration for the classic Jupyter Server app (`system` instance type). |
| `lifecycleConfigArns` | list(string) | no | [] | ARNs of lifecycle scripts for Jupyter Server startup. |
| `codeRepositories` | list(CodeRepository) | no | [] | Git repos auto-cloned on startup. Max 10. |

### KernelGatewayAppSettings

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `defaultResourceSpec` | ResourceSpec | no | — | Default compute instance type for KernelGateway apps. |
| `lifecycleConfigArns` | list(string) | no | [] | ARNs of lifecycle scripts for KernelGateway startup. |
| `customImages` | list(CustomImage) | no | [] | Custom Docker images as KernelGateway kernels. Max 200. |

### CodeEditorAppSettings

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `defaultResourceSpec` | ResourceSpec | no | — | Default compute instance type and image configuration for Code Editor apps. |
| `lifecycleConfigArns` | list(string) | no | [] | ARNs of lifecycle scripts for Code Editor startup. |
| `builtInLifecycleConfigArn` | string | no | — | ARN of an AWS-curated (built-in) lifecycle configuration. |
| `customImages` | list(CustomImage) | no | [] | Custom Docker images available in Code Editor. Max 200. |
| `idleSettings` | IdleSettings | no | — | Automatic shutdown of idle Code Editor instances. |

### TensorBoardAppSettings / RSessionAppSettings

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `defaultResourceSpec` | ResourceSpec | no | — | Default compute instance type and image configuration. |
| `customImages` (RSession only) | list(CustomImage) | no | [] | Custom Docker images as RSession kernels. Max 200. |

### RStudioServerProAppSettings

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `accessStatus` | string | no | — | `ENABLED` grants the user access to RStudio; `DISABLED` hides it. |
| `userGroup` | string | no | `R_STUDIO_USER` | `R_STUDIO_ADMIN` for the Workbench admin dashboard; `R_STUDIO_USER` for regular use. Only honored when `accessStatus` is `ENABLED`. |

### RStudioServerProDomainSettings

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `domainExecutionRoleArn` | StringValueOrRef | **yes** | — | IAM role SageMaker assumes to run the RStudio server (license validation, server lifecycle). |
| `rStudioConnectUrl` | string | no | — | RStudio Connect server URL for publishing Shiny apps and R Markdown documents. |
| `rStudioPackageManagerUrl` | string | no | — | RStudio Package Manager URL for resolving R packages (instead of public CRAN). |
| `defaultResourceSpec` | ResourceSpec | no | — | Compute for the RStudio server app itself. |

### CanvasAppSettings

Each block is an independent Canvas capability:

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `directDeployStatus` | string | no | — | `ENABLED`/`DISABLED` — whether Canvas users can deploy models straight to real-time endpoints. Governance-minded teams disable this and use the model registry instead. |
| `emrServerlessSettings` | object | no | — | `executionRoleArn` (StringValueOrRef) + `status` — run large data-prep jobs on EMR Serverless. |
| `generativeAiBedrockRoleArn` | StringValueOrRef | no | — | IAM role Canvas assumes to call Amazon Bedrock; setting it enables generative-AI features. |
| `identityProviderOauthSettings` | list(object) | no | [] | Up to 20 SaaS OAuth connectors: `dataSourceName` (`SalesforceGenie` or `Snowflake`), `secretArn` (Secrets Manager secret holding OAuth client credentials), `status`. |
| `kendraSettingsStatus` | string | no | — | `ENABLED`/`DISABLED` — Canvas document querying against Amazon Kendra indexes. |
| `modelRegisterSettings` | object | no | — | `status` + optional `crossAccountModelRegisterRoleArn` — register Canvas models into a (possibly cross-account) SageMaker Model Registry. |
| `timeSeriesForecastingSettings` | object | no | — | `amazonForecastRoleArn` (StringValueOrRef) + `status` — Canvas forecasting via Amazon Forecast. |
| `workspaceSettings` | object | no | — | `s3ArtifactPath` (s3:// or https:// URI) + `s3KmsKeyId` — where Canvas stores its working artifacts. |

### ResourceSpec

Shared message used by every app-settings block:

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `instanceType` | string | no | — | EC2 instance type: `ml.t3.medium` (dev), `ml.m5.large` (general), `ml.g4dn.xlarge` (GPU), `ml.p3.2xlarge` (heavy training), `system` (lightweight). |
| `lifecycleConfigArn` | string | no | — | ARN of a lifecycle script for this app type. |
| `sagemakerImageArn` | string | no | — | ARN of a custom SageMaker Image (replaces default container). |
| `sagemakerImageVersionAlias` | string | no | — | Human-readable alias for image version (e.g., `latest`). Mutually exclusive with `sagemakerImageVersionArn`. |
| `sagemakerImageVersionArn` | string | no | — | ARN of a specific image version for reproducibility. Mutually exclusive with `sagemakerImageVersionAlias`. |

### CustomImage

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `appImageConfigName` | string | **yes** | — | Name of the SageMaker AppImageConfig defining kernel specs and file system layout. |
| `imageName` | string | **yes** | — | Name of the SageMaker Image resource containing the container image. |
| `imageVersionNumber` | int32 | no | latest | Pin to a specific image version number. |

### CodeRepository

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `repositoryUrl` | string | **yes** | — | HTTPS URL of the Git repository to clone (max 1024 chars). SSH URLs not supported. |

### IdleSettings

Used by JupyterLab and Code Editor app settings. Configuring the block IS the enable
switch — omit it to leave instances running until manually stopped. All three timeouts
are required whenever the block is set (AWS rejects a partial block: absent members
transmit as 0, below the 60-minute floor):

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `idleTimeoutInMinutes` | int32 | **yes** | — | Minutes of inactivity before shutdown (60–525600). |
| `minIdleTimeoutInMinutes` | int32 | **yes** | — | Minimum idle timeout users can set for their own apps (60–525600). |
| `maxIdleTimeoutInMinutes` | int32 | **yes** | — | Maximum idle timeout users can set (60–525600). Must be ≥ the minimum. |

### EmrSettings

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `assumableRoleArns` | list(StringValueOrRef) | no | [] | IAM roles JupyterLab can assume to CONNECT to EMR clusters (including cross-account). |
| `executionRoleArns` | list(StringValueOrRef) | no | [] | IAM runtime roles available for EMR workloads submitted from notebooks. |

### SharingSettings

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `notebookOutputOption` | string | no | `Disabled` | `Allowed` to persist notebook outputs to S3; `Disabled` to skip. |
| `s3KmsKeyId` | StringValueOrRef | no | bucket default | KMS key for encrypting shared outputs. |
| `s3OutputPath` | string | conditional | — | S3 URI for shared outputs. Required when `notebookOutputOption` is `Allowed`. |

### SpaceStorageSettings

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `defaultEbsVolumeSizeInGb` | int32 | **yes** | — | Default EBS volume size (GB) for new spaces. |
| `maximumEbsVolumeSizeInGb` | int32 | **yes** | — | Maximum EBS volume size (GB) users can request. Must be ≥ default. |

### CustomFileSystemConfig

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `efsFileSystemConfig.fileSystemId` | StringValueOrRef | **yes** | — | EFS file system to mount (must be reachable from the domain's subnets). Supports `valueFrom` (AwsElasticFileSystem). |
| `efsFileSystemConfig.fileSystemPath` | string | **yes** | — | Path within the EFS file system to mount into apps. |

### CustomPosixUserConfig

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `uid` | int64 | **yes** | — | POSIX user ID. Must be ≥ 10000. |
| `gid` | int64 | **yes** | — | POSIX group ID. Must be ≥ 1001. |

### StudioWebPortalSettings

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `hiddenAppTypes` | list(string) | no | [] | Studio app types to hide (e.g. `JupyterServer`, `Canvas`, `CodeEditor`). |
| `hiddenInstanceTypes` | list(string) | no | [] | Instance types to hide from app-creation pickers (e.g. `ml.p3.2xlarge`). |
| `hiddenMlTools` | list(string) | no | [] | Studio ML tools to hide (e.g. `DataWrangler`, `FeatureStore`, `EmrClusters`). |

### DockerSettings

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `enableDockerAccess` | string | no | — | `ENABLED` to allow Docker commands; `DISABLED` to block. |
| `vpcOnlyTrustedAccounts` | list(string) | no | [] | AWS account IDs allowed as Docker image sources in VpcOnly mode. Max 20. |

## Output fields reference

| Output | Type | Description |
|---|---|---|
| `domain_id` | string | Unique identifier of the SageMaker Domain. Used when creating user profiles and spaces. |
| `domain_arn` | string | ARN of the domain. Used in IAM policies and cross-service references. |
| `domain_url` | string | HTTPS URL for accessing SageMaker Studio web interface. |
| `home_efs_file_system_id` | string | ID of the auto-created EFS file system for user home directories. |
| `security_group_id_for_domain_boundary` | string | ID of the AWS-managed security group controlling cross-app/cross-user traffic. |
| `single_sign_on_application_arn` | string | ARN of the IAM Identity Center application (only when `authMode` is `SSO`). |
| `single_sign_on_managed_application_instance_id` | string | Identity Center managed application instance ID (only when `authMode` is `SSO`); used when assigning users and groups programmatically. |

## Examples

### Minimal domain

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerDomain
metadata:
  name: dev-ml-workspace
spec:
  region: us-west-2
  authMode: IAM
  homeEfsRetentionPolicy: Delete
  vpcId:
    value: vpc-0abc123def456789
  subnetIds:
    - value: subnet-0aaa1111
    - value: subnet-0bbb2222
  defaultUserSettings:
    executionRoleArn:
      value: arn:aws:iam::111122223333:role/SageMakerExecutionRole
```

### Production-ready domain

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerDomain
metadata:
  name: prod-ml-platform
spec:
  region: us-west-2
  authMode: SSO
  appNetworkAccessType: VpcOnly
  tagPropagation: ENABLED
  executionRoleIdentityConfig: USER_PROFILE_NAME
  trustedIdentityPropagationStatus: ENABLED
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: ml-platform-vpc
      fieldPath: status.outputs.vpc_id
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: ml-platform-private-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: ml-platform-private-subnet-b
        fieldPath: status.outputs.subnet_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: ml-encryption-key
      fieldPath: status.outputs.key_arn
  domainSecurityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: ml-domain-sg
        fieldPath: status.outputs.security_group_id
  dockerSettings:
    enableDockerAccess: ENABLED
    vpcOnlyTrustedAccounts:
      - "111122223333"
  defaultUserSettings:
    executionRoleArn:
      valueFrom:
        kind: AwsIamRole
        name: sagemaker-execution-role
        fieldPath: status.outputs.role_arn
    securityGroupIds:
      - valueFrom:
          kind: AwsSecurityGroup
          name: ml-notebook-sg
          fieldPath: status.outputs.security_group_id
    studioWebPortal: ENABLED
    jupyterLabAppSettings:
      defaultResourceSpec:
        instanceType: ml.m5.large
      codeRepositories:
        - repositoryUrl: "https://github.com/org/ml-notebooks.git"
      idleSettings:
        idleTimeoutInMinutes: 120
        minIdleTimeoutInMinutes: 60
        maxIdleTimeoutInMinutes: 480
    studioWebPortalSettings:
      hiddenInstanceTypes:
        - ml.p3.16xlarge
    sharingSettings:
      notebookOutputOption: Allowed
      s3OutputPath: "s3://ml-team-artifacts/notebook-outputs/"
    spaceStorageSettings:
      defaultEbsVolumeSizeInGb: 50
      maximumEbsVolumeSizeInGb: 500
```

## Related resources

| Resource | Relationship |
|---|---|
| `AwsVpc` | Provides VPC ID and subnets for domain ENI placement. |
| `AwsIamRole` | Execution roles: user plane, space plane, RStudio domain role, Canvas capability roles, EMR roles. |
| `AwsSecurityGroup` | User-level and domain-level network access control. |
| `AwsKmsKey` | Customer-managed encryption for EFS home directories, shared notebook outputs, and Canvas artifacts. |
| `AwsElasticFileSystem` | Additional shared file systems mounted into apps via `customFileSystemConfigs`. |

## Cross-field validations

The spec enforces the following cross-field validations at the protobuf level:

1. **auth_mode valid** — must be `IAM` or `SSO`.
2. **app_network_access_type valid** — must be `PublicInternetOnly` or `VpcOnly`.
3. **app_security_group_management valid + requires RStudio** — must be `Service` or `Customer`, and is only accepted when `rStudioServerProDomainSettings` is configured (AWS silently ignores it otherwise).
4. **tag_propagation valid** — must be `ENABLED` or `DISABLED`.
5. **home_efs_retention_policy valid** — must be `Retain` or `Delete`.
6. **execution_role_identity_config valid** — must be `USER_PROFILE_NAME` or `DISABLED`.
7. **trusted identity propagation requires SSO** — `ENABLED` is rejected unless `authMode` is `SSO`.
8. **studio_web_portal valid** — must be `ENABLED` or `DISABLED`.
9. **auto_mount_home_efs valid** — must be `Enabled` or `Disabled` (`DefaultAsDomain` is profile-only and rejected at the domain level).
10. **idle settings complete** — all three idle timeouts are required whenever `idleSettings` is set (block presence is the enable switch; AWS rejects partial blocks), bounded to 60–525600 minutes, with max ≥ min.
11. **image version alias XOR ARN** — a resource spec cannot pin both `sagemakerImageVersionAlias` and `sagemakerImageVersionArn`.
12. **RStudio user_group requires ENABLED access** — `userGroup` is only accepted when `accessStatus` is `ENABLED`.
13. **Canvas statuses valid** — every Canvas capability status must be `ENABLED` or `DISABLED`; OAuth `dataSourceName` must be `SalesforceGenie` or `Snowflake`; `secretArn` is required per connector; the workspace path must be an `s3://` or `https://` URI.
14. **notebook_output_option valid + s3_output_path required when allowed**.
15. **max EBS ≥ default** — `maximumEbsVolumeSizeInGb` must be ≥ `defaultEbsVolumeSizeInGb`.
16. **POSIX identity bounds** — `uid` ≥ 10000, `gid` ≥ 1001.
17. **enable_docker_access valid** — must be `ENABLED` or `DISABLED`.

## Companion resources (separate kinds)

The domain is the root of the SageMaker resource family. The following are deliberately **not** part of the domain spec because they have independent lifecycles, are many-per-domain, or belong to separate product surfaces:

| Feature | Reason |
|---|---|
| User Profiles | Per-user membership and overrides; many per domain with independent lifecycles. Joins via the `domain_id` output. |
| Spaces | Shared or private collaboration environments; many per domain. |
| Apps (explicit) | JupyterLab/KernelGateway app instances are created on demand by users, not at provisioning time. |
| Studio Lifecycle Configs | Referenced by ARN from resource specs; own create/delete lifecycle. |
| App Image Configs / Images | Referenced by name from `customImages`; own versioned lifecycles. |
| MLflow Tracking Server | Separate SageMaker resource with its own sizing and lifecycle. |
| SageMaker Pipelines / Model Registry / Feature Store / Endpoints | Separate API surfaces for ML workflow, model management, and inference. |

## How it works

Planton provisions the SageMaker Domain via Pulumi or Terraform modules defined in this repository. The API contract is protobuf-based (`api.proto`, `spec.proto`, `outputs.proto`) and stack execution is orchestrated by the platform using the `AwsSagemakerDomainIacInput` (includes provider credentials and target resource).

## References

- [Amazon SageMaker Studio Documentation](https://docs.aws.amazon.com/sagemaker/latest/dg/studio.html)
- [SageMaker Domain](https://docs.aws.amazon.com/sagemaker/latest/dg/sm-domain.html)
- [SageMaker Studio Pricing](https://aws.amazon.com/sagemaker/pricing/)
- [SageMaker Network Configuration](https://docs.aws.amazon.com/sagemaker/latest/dg/studio-notebooks-and-internet-access.html)
- [SageMaker Docker Access](https://docs.aws.amazon.com/sagemaker/latest/dg/studio-updated-local.html)
- [RStudio on SageMaker](https://docs.aws.amazon.com/sagemaker/latest/dg/rstudio.html)
- [SageMaker Canvas](https://docs.aws.amazon.com/sagemaker/latest/dg/canvas.html)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
