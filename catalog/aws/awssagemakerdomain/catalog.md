# AWS SageMaker Domain

Deploys an Amazon SageMaker Domain providing a shared workspace for JupyterLab notebooks, the Code Editor IDE, Canvas no-code ML, custom kernels, and ML tooling with configurable authentication (IAM or SSO), VPC networking, EFS home directories, idle shutdown policies, RStudio Workbench, and Docker build support. The domain also declares its people and collaboration surfaces as code: user profiles and shared spaces are name-keyed lists on the spec, so team membership changes are one-line diffs.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SageMaker Domain** -- a Studio workspace configured with the specified authentication mode, VPC placement, two inheritance planes (default user settings and default shared-space settings), per-IDE app defaults (JupyterLab, Code Editor, classic Jupyter Server, KernelGateway, TensorBoard), Canvas capabilities, optional RStudio Workbench, and domain-wide governance dials (Docker access, Studio UI hiding, tag propagation)
- **User Profiles** -- one per `userProfiles` entry: the per-person workspaces inside the domain, each inheriting the domain's defaults with optional per-user overrides (adding a teammate is adding one list entry)
- **Spaces** -- one per `spaces` entry: named shared (or private) workspaces whose runtime and EBS volume are shared by everyone with access -- the collaboration plane, declared beside the per-user plane
- **EFS File System** -- automatically created by AWS to provide persistent home directories for each user profile in the domain (retention on domain deletion is your choice via `homeEfsRetentionPolicy`)
- **Domain Boundary Security Group** -- automatically created by AWS to control cross-app and cross-user traffic within the domain
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance; set `tagPropagation: ENABLED` to flow them onto the apps, spaces, and user profiles created inside the domain

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A VPC** with DNS resolution and DNS hostnames enabled. All domain network interfaces are placed in this VPC. Provide the VPC ID directly or reference an AwsVpc Infra Component via ValueFromRef. Changing the VPC forces domain replacement.
- **At least one subnet** (two recommended for high availability) in the target VPC. Private subnets are required for `VpcOnly` network mode. Provide subnet IDs directly or reference an AwsVpc Infra Component via ValueFromRef. Changing subnets forces domain replacement.
- **An IAM execution role** with a trust policy allowing `sagemaker.amazonaws.com` to assume it. This role governs what AWS resources users can access from Studio sessions. Provide the ARN directly or reference an AwsIamRole Infra Component via ValueFromRef.
- **Security groups** (optional) for domain-level and user-level network isolation. Provide IDs directly or reference AwsSecurityGroup Infra Components via ValueFromRef.
- **A KMS key** (optional) for encrypting the EFS home directory volume. If omitted, AWS uses the default `aws/elasticfilesystem` service key. Changing the KMS key forces domain replacement.
- **AWS IAM Identity Center** (required for SSO auth mode) configured in the account with user assignments.

## Deploy

### Console

Open the deployment store, find **AWS SageMaker Domain**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic JupyterLab Domain** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerDomain
metadata:
  name: ml-workspace
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  authMode: IAM
  vpcId:
    value: "vpc-0a1b2c3d4e5f00001"
  subnetIds:
    - value: "subnet-0a1b2c3d4e5f00001"
    - value: "subnet-0a1b2c3d4e5f00002"
  defaultUserSettings:
    executionRoleArn:
      value: "arn:aws:iam::123456789012:role/sagemaker-execution-role"
```

```shell
planton apply -f sagemaker-domain.yaml
```

This creates a SageMaker Domain with IAM authentication, public internet network access, default JupyterLab settings, and AWS-managed EFS encryption. No idle shutdown, custom images, or Docker access is configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the SageMaker Domain to a VPC, IAM role, security groups, and KMS key deployed in the same InfraPipeline:

```yaml
spec:
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: ml-vpc
      fieldPath: status.outputs.vpc_id
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: ml-vpc
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: ml-vpc
        fieldPath: status.outputs.private_subnets.[1].id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: ml-efs-key
      fieldPath: status.outputs.key_arn
  defaultUserSettings:
    executionRoleArn:
      valueFrom:
        kind: AwsIamRole
        name: sagemaker-role
        fieldPath: status.outputs.role_arn
    securityGroupIds:
      - valueFrom:
          kind: AwsSecurityGroup
          name: ml-user-sg
          fieldPath: status.outputs.security_group_id
```

The InfraPipeline resolves the dependency graph, deploys the VPC, KMS key, IAM role, and security group first, then provisions the SageMaker Domain with the resolved values.

## Key Configuration

These are the most important decisions when configuring a SageMaker Domain. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Authentication mode** -- Set `authMode` to `IAM` for AWS credential-based login or `SSO` for centralized identity via IAM Identity Center. SSO is recommended for enterprise teams, and is required for `trustedIdentityPropagationStatus` (forwarding the Identity Center user into Athena, Redshift, and Lake Formation). Auth mode is immutable after creation.

**Network access** -- Set `appNetworkAccessType` to `PublicInternetOnly` (default) for internet-routable notebooks or `VpcOnly` to keep all traffic within the VPC. `VpcOnly` is recommended for production to prevent data exfiltration and is required for Docker trusted account restrictions.

**Idle shutdown** -- Configure `defaultUserSettings.jupyterLabAppSettings.idleSettings` (and its Code Editor sibling). The block's presence turns idle auto-shutdown on, and all three timeouts are required together: `idleTimeoutInMinutes` (e.g., 120), `minIdleTimeoutInMinutes`, and `maxIdleTimeoutInMinutes` (the bounds individual users may pick for their own apps). Setting `lifecycleManagement: DISABLED` keeps the timeouts as published guardrails without enforcing auto-shutdown -- and flipping it back genuinely re-enables enforcement. Without idle shutdown, notebook instances run 24/7 at full compute cost.

**Instance types** -- Set default instance types per app in each `defaultResourceSpec.instanceType` (JupyterLab, Code Editor, KernelGateway, TensorBoard, and the classic Jupyter Server's lightweight `system` instance). Start with `ml.t3.medium` for exploration and scale to GPU instances (`ml.g4dn.xlarge`) for training workloads. Pin custom images with `sagemakerImageVersionAlias` OR `sagemakerImageVersionArn` -- never both.

**Home EFS retention** -- `homeEfsRetentionPolicy` decides what happens to the auto-created home file system when the domain is deleted: `Retain` (the AWS default -- it survives and keeps billing until deleted by hand) or `Delete` (right for dev/test domains). The decision is fixed at create time.

**Canvas governance** -- Each block under `defaultUserSettings.canvasAppSettings` is an independent capability. The governed pattern pairs `directDeployStatus: DISABLED` with `modelRegisterSettings.status: ENABLED`, so analyst-built models flow through the Model Registry instead of deploying straight to billable endpoints. Role-backed capabilities (Forecast, Bedrock, EMR Serverless) are enabled by wiring their IAM role.

**RStudio Workbench** -- Configuring `rStudioServerProDomainSettings` (its required `domainExecutionRoleArn` plus optional Connect/Package Manager URLs) is what activates the RStudio app plane; it requires a Posit license via AWS License Manager. The per-user switch lives in `defaultUserSettings.rStudioServerProAppSettings`, and `appSecurityGroupManagement` is only honored while RStudio is configured.

**Shared spaces** -- `defaultSpaceSettings` is a second, independent inheritance plane for collaborative spaces, with its own execution role, app baselines, storage sizing, and POSIX identity. Omit it entirely to leave shared spaces on AWS defaults.

**Team provisioning as code** -- Declare the people and the shared workspaces beside the domain: `userProfiles` (one per team member, each optionally overriding the defaults) and `spaces` (named workspaces with `Private` or `Shared` posture; ownership and sharing are declared together, and the owner must be a profile in the domain). Both are name-keyed, so membership changes are one-line diffs -- and removals are destructive (a removed profile takes its home-directory surfaces with it).

**Docker access** -- Enable `dockerSettings.enableDockerAccess` to allow users to build and run custom containers. In `VpcOnly` mode, restrict image pulls to trusted accounts via `vpcOnlyTrustedAccounts`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsVpc** | `vpcId` | `status.outputs.vpc_id` |
| **AwsVpc** | `subnetIds` | `status.outputs.private_subnets.[*].id` |
| **AwsIamRole** | `defaultUserSettings.executionRoleArn` | `status.outputs.role_arn` |
| **AwsIamRole** (optional) | `defaultSpaceSettings.executionRoleArn` | `status.outputs.role_arn` |
| **AwsIamRole** (optional) | `rStudioServerProDomainSettings.domainExecutionRoleArn` | `status.outputs.role_arn` |
| **AwsIamRole** (optional) | `defaultUserSettings.jupyterLabAppSettings.emrSettings.assumableRoleArns` / `executionRoleArns` | `status.outputs.role_arn` |
| **AwsIamRole** (optional) | `defaultUserSettings.canvasAppSettings.generativeAiBedrockRoleArn` | `status.outputs.role_arn` |
| **AwsIamRole** (optional) | `defaultUserSettings.canvasAppSettings.emrServerlessSettings.executionRoleArn` | `status.outputs.role_arn` |
| **AwsIamRole** (optional) | `defaultUserSettings.canvasAppSettings.timeSeriesForecastingSettings.amazonForecastRoleArn` | `status.outputs.role_arn` |
| **AwsSecurityGroup** (optional) | `domainSecurityGroupIds` | `status.outputs.security_group_id` |
| **AwsSecurityGroup** (optional) | `defaultUserSettings.securityGroupIds` / `defaultSpaceSettings.securityGroupIds` | `status.outputs.security_group_id` |
| **AwsKmsKey** (optional) | `kmsKeyId` | `status.outputs.key_arn` |
| **AwsKmsKey** (optional) | `defaultUserSettings.sharingSettings.s3KmsKeyId` | `status.outputs.key_arn` |
| **AwsKmsKey** (optional) | `defaultUserSettings.canvasAppSettings.workspaceSettings.s3KmsKeyId` | `status.outputs.key_arn` |
| **AwsElasticFileSystem** (optional) | `defaultUserSettings.customFileSystemConfigs.[*].efsFileSystemConfig.fileSystemId` (and the `defaultSpaceSettings` twin) | `status.outputs.file_system_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `domain_id` | SageMaker Domain identifier | User profile creation, API calls |
| `domain_arn` | Amazon Resource Name | IAM policies, CloudWatch metrics, resource tagging |
| `domain_url` | Studio web interface URL | Team access, bookmarks |
| `home_efs_file_system_id` | EFS file system ID for home directories | Backup policies, monitoring, lifecycle management |
| `security_group_id_for_domain_boundary` | Auto-created domain boundary security group | Additional security group rules |
| `single_sign_on_application_arn` | IAM Identity Center application ARN (SSO only) | SSO configuration, access management |
| `single_sign_on_managed_application_instance_id` | IAM Identity Center managed application instance ID (SSO only) | Programmatic user/group assignment to the domain |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic JupyterLab domain** -- IAM authentication with public internet access and default settings. Suited for development, exploration, and small teams getting started with SageMaker Studio. Start from the **Basic JupyterLab Domain** preset.

**Production VPC-only domain** -- SSO authentication, VPC-only networking, customer-managed KMS encryption, idle shutdown, and layered security groups. Suited for production ML platforms with compliance requirements. Start from the **Production VPC-Only Domain** preset.

**ML team with custom images** -- SSO authentication, VPC-only networking, Docker access with trusted accounts, custom JupyterLab and KernelGateway images, GPU compute, notebook sharing to S3, and auto-cloned Git repositories. Suited for advanced ML teams building custom training containers. Start from the **ML Team with Custom Images** preset.

**Governed Canvas workspace** -- IAM authentication with Canvas configured for business analysts: direct endpoint deployment disabled, Model Registry registration enabled, time-series forecasting under a dedicated role, a pinned S3 workspace, and the code-first Studio surfaces hidden from the launcher. Start from the **Governed Canvas Workspace** preset.

**Team workspaces** -- The domain plus its people and collaboration surfaces in one manifest: `userProfiles` for each team member and named `spaces` with `Private` or `Shared` posture. Membership changes become one-line diffs reviewed like any other code. Start from the **Team Workspaces SageMaker Domain** preset.

## Works With

- [**AWS VPC**](/infra-catalog/aws-vpc) -- provides the VPC and subnets for domain network interfaces
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- provides the execution roles for Studio sessions, shared spaces, RStudio licensing, EMR access, and the Canvas service capabilities
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- provides domain-level, user-level, and space-level network isolation
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for EFS encryption, notebook output encryption, and the Canvas workspace
- [**AWS Elastic File System**](/infra-catalog/aws-elastic-file-system) -- provides additional shared file systems mounted into user and space apps beyond the domain's own home EFS