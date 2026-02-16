# AwsSagemakerDomain

Amazon SageMaker Domain resource for OpenMCF. Provisions the top-level organizational unit for Amazon SageMaker Studio — a shared, managed workspace where data scientists and ML engineers access JupyterLab notebooks, custom kernels, collaborative editing, and ML tooling. The domain handles VPC networking, EFS-backed home directories, IAM execution roles, Docker access, and idle-timeout cost controls so teams can focus on building models instead of managing infrastructure.

## When to use

- You need a managed ML workspace on AWS for data science or ML engineering teams.
- Teams require JupyterLab environments with shared storage, Git integration, and configurable compute instances.
- You want centralized governance over ML workspaces: approved container images, idle-timeout policies, network boundaries, and storage quotas.
- Your security posture requires VPC-only notebook access with no public internet exposure.
- You need Docker-in-notebook capability for building custom training containers or inference images.

## Prerequisites

| Prerequisite | Why | OpenMCF Resource |
|---|---|---|
| VPC with subnets in 1+ AZs | Domain ENIs for notebook/training traffic are placed in subnets; 2+ AZs recommended for HA | `AwsVpc` |
| IAM execution role | SageMaker assumes this role to access S3, ECR, Secrets Manager, etc. on behalf of users; must trust `sagemaker.amazonaws.com` | `AwsIamRole` |
| Security groups (optional) | Control inbound/outbound traffic for user notebooks and domain-scoped apps | `AwsSecurityGroup` |
| KMS key (optional) | Customer-managed encryption for the domain's EFS home directory volume | `AwsKmsKey` |
| IAM Identity Center (optional) | Required when `auth_mode` is `SSO` for enterprise single sign-on | (external) |

## API envelope

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSagemakerDomain
metadata:
  name: <resource-id>
spec: { ... }
```

## Spec fields reference

### Core (ForceNew)

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `authMode` | string | **yes** | — | Authentication mode: `IAM` or `SSO`. IAM uses AWS credentials; SSO uses IAM Identity Center. **ForceNew**. |
| `vpcId` | StringValueOrRef | **yes** | — | VPC for domain network interfaces. Must have DNS resolution and DNS hostnames enabled. Supports `value` or `valueFrom` (AwsVpc). **ForceNew**. |
| `subnetIds` | list(StringValueOrRef) | **yes** (≥1) | — | Subnets for notebook/training ENIs. Max 16 subnets; 2+ AZs recommended. **ForceNew**. |
| `kmsKeyId` | StringValueOrRef | no | aws/elasticfilesystem | KMS key ARN for EFS home directory encryption. **ForceNew**. |

### Network

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `appNetworkAccessType` | string | no | `PublicInternetOnly` | `PublicInternetOnly` — ENIs have internet access via VPC routing. `VpcOnly` — all traffic stays within VPC; requires NAT for internet. Recommended for production. |

### Domain Settings

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `domainSecurityGroupIds` | list(StringValueOrRef) | no | [] | Security groups for domain-scoped apps and shared resources. Max 3. **ForceNew**. |
| `dockerSettings` | DockerSettings | no | — | Docker access configuration (see nested message below). |

### Default User Settings

Nested message `defaultUserSettings` — inherited by all user profiles.

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `executionRoleArn` | StringValueOrRef | **yes** | — | IAM role assumed by SageMaker for notebooks/training. Must trust `sagemaker.amazonaws.com`. |
| `securityGroupIds` | list(StringValueOrRef) | no | [] | User-level security groups for notebook ENIs. Max 5. |
| `defaultLandingUri` | string | no | platform default | URI of the default app opened on login. Common: `studio::relative/JupyterLab`, `studio::relative/JupyterServer:`, `studio::`. |
| `studioWebPortal` | string | no | `ENABLED` | `ENABLED` for full Studio web UI; `DISABLED` for programmatic-only access. |
| `jupyterLabAppSettings` | JupyterLabAppSettings | no | — | JupyterLab IDE configuration (see nested message below). |
| `kernelGatewayAppSettings` | KernelGatewayAppSettings | no | — | Custom kernel configuration (see nested message below). |
| `sharingSettings` | SharingSettings | no | — | Notebook output sharing to S3 (see nested message below). |
| `spaceStorageSettings` | SpaceStorageSettings | no | — | Default EBS volume sizes for user spaces (see nested message below). |

### JupyterLabAppSettings

Nested message within `defaultUserSettings.jupyterLabAppSettings`:

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `defaultResourceSpec` | ResourceSpec | no | — | Default compute instance type and lifecycle config for JupyterLab apps. |
| `lifecycleConfigArns` | list(string) | no | [] | ARNs of lifecycle scripts run at JupyterLab startup (install packages, configure extensions). |
| `customImages` | list(CustomImage) | no | [] | Custom Docker images available as JupyterLab kernels. Max 200. |
| `codeRepositories` | list(CodeRepository) | no | [] | Git repos auto-cloned into JupyterLab on startup. Max 10. |
| `idleSettings` | IdleSettings | no | — | Automatic shutdown of idle JupyterLab instances (see nested message below). |

### KernelGatewayAppSettings

Nested message within `defaultUserSettings.kernelGatewayAppSettings`:

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `defaultResourceSpec` | ResourceSpec | no | — | Default compute instance type for KernelGateway apps. |
| `lifecycleConfigArns` | list(string) | no | [] | ARNs of lifecycle scripts for KernelGateway startup. |
| `customImages` | list(CustomImage) | no | [] | Custom Docker images as KernelGateway kernels. Max 200. |

### ResourceSpec

Shared message used by JupyterLab and KernelGateway settings:

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `instanceType` | string | no | — | EC2 instance type: `ml.t3.medium` (dev), `ml.m5.large` (general), `ml.g4dn.xlarge` (GPU), `ml.p3.2xlarge` (heavy training), `system` (lightweight). |
| `lifecycleConfigArn` | string | no | — | ARN of a lifecycle script for this app type. |
| `sagemakerImageArn` | string | no | — | ARN of a custom SageMaker Image (replaces default container). |
| `sagemakerImageVersionAlias` | string | no | — | Human-readable alias for image version (e.g., `latest`, `v2.0`). Mutually exclusive with `sagemakerImageVersionArn`. |
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

Nested message within `jupyterLabAppSettings.idleSettings`:

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `lifecycleManagement` | string | no | — | `ENABLED` to auto-shutdown idle instances; `DISABLED` to keep running. |
| `idleTimeoutInMinutes` | int32 | conditional | — | Minutes of inactivity before shutdown (60–525600). Required when `lifecycleManagement` is `ENABLED`. |
| `minIdleTimeoutInMinutes` | int32 | no | — | Minimum idle timeout users can set for their own apps (60–525600). |
| `maxIdleTimeoutInMinutes` | int32 | no | — | Maximum idle timeout users can set (60–525600). |

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

## Examples

### Minimal domain

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSagemakerDomain
metadata:
  name: dev-ml-workspace
spec:
  authMode: IAM
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
apiVersion: aws.openmcf.org/v1
kind: AwsSagemakerDomain
metadata:
  name: prod-ml-platform
spec:
  authMode: SSO
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: ml-platform-vpc
      fieldPath: status.outputs.vpc_id
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: ml-platform-vpc
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: ml-platform-vpc
        fieldPath: status.outputs.private_subnets.[1].id
    - valueFrom:
        kind: AwsVpc
        name: ml-platform-vpc
        fieldPath: status.outputs.private_subnets.[2].id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: ml-encryption-key
      fieldPath: status.outputs.key_arn
  appNetworkAccessType: VpcOnly
  domainSecurityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: ml-domain-sg
        fieldPath: status.outputs.security_group_id
  dockerSettings:
    enableDockerAccess: ENABLED
    vpcOnlyTrustedAccounts:
      - "111122223333"
      - "444455556666"
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
    defaultLandingUri: "studio::relative/JupyterLab"
    studioWebPortal: ENABLED
    jupyterLabAppSettings:
      defaultResourceSpec:
        instanceType: ml.m5.large
      codeRepositories:
        - repositoryUrl: "https://github.com/org/ml-notebooks.git"
        - repositoryUrl: "https://github.com/org/ml-pipelines.git"
      idleSettings:
        lifecycleManagement: ENABLED
        idleTimeoutInMinutes: 120
        minIdleTimeoutInMinutes: 60
        maxIdleTimeoutInMinutes: 480
    sharingSettings:
      notebookOutputOption: Allowed
      s3OutputPath: "s3://ml-team-artifacts/notebook-outputs/"
      s3KmsKeyId:
        valueFrom:
          kind: AwsKmsKey
          name: ml-encryption-key
          fieldPath: status.outputs.key_arn
    spaceStorageSettings:
      defaultEbsVolumeSizeInGb: 50
      maximumEbsVolumeSizeInGb: 500
```

## Related resources

| Resource | Relationship |
|---|---|
| `AwsVpc` | Provides VPC ID and subnets for domain ENI placement. |
| `AwsIamRole` | Execution role assumed by SageMaker for user workloads. |
| `AwsSecurityGroup` | User-level and domain-level network access control. |
| `AwsKmsKey` | Customer-managed encryption for EFS home directories and shared notebook outputs. |

## Cross-field validations

The spec enforces the following cross-field validations at the protobuf level:

1. **auth_mode valid** — must be `IAM` or `SSO`.
2. **app_network_access_type valid** — must be `PublicInternetOnly` or `VpcOnly`.
3. **studio_web_portal valid** — must be `ENABLED` or `DISABLED`.
4. **lifecycle_management valid** — must be `ENABLED` or `DISABLED`.
5. **idle_timeout requires enabled** — `idleTimeoutInMinutes` requires `lifecycleManagement` to be `ENABLED`.
6. **notebook_output_option valid** — must be `Allowed` or `Disabled`.
7. **s3_output_path required when allowed** — `s3OutputPath` is required when `notebookOutputOption` is `Allowed`.
8. **max EBS ≥ default** — `maximumEbsVolumeSizeInGb` must be ≥ `defaultEbsVolumeSizeInGb`.
9. **enable_docker_access valid** — must be `ENABLED` or `DISABLED`.

## Deliberately omitted features

The following SageMaker features are **not** covered by this v1 API. They may be added in future versions:

| Feature | Reason |
|---|---|
| User Profiles | Per-user overrides of domain defaults; separate resource kind (`AwsSagemakerUserProfile`). |
| Spaces | Shared or private collaboration environments within a domain; separate resource kind. |
| Apps (explicit) | JupyterLab/KernelGateway app instances are created on-demand by users, not at domain provisioning time. |
| MLflow Tracking Server | Optional Studio component; would be a companion resource or domain extension. |
| RStudio Server Pro | Alternative IDE within Studio; requires separate license and configuration model. |
| Canvas (no-code ML) | No-code AutoML product within Studio; separate user-facing component. |
| SageMaker Pipelines | ML workflow orchestration; separate API surface. |
| Model Registry | Model versioning and approval; separate API surface. |
| Feature Store | Feature engineering and storage; separate API surface. |

## How it works

OpenMCF provisions the SageMaker Domain via Pulumi or Terraform modules defined in this repository. The API contract is protobuf-based (`api.proto`, `spec.proto`, `stack_outputs.proto`) and stack execution is orchestrated by the platform using the `AwsSagemakerDomainStackInput` (includes provider credentials and target resource).

## References

- [Amazon SageMaker Studio Documentation](https://docs.aws.amazon.com/sagemaker/latest/dg/studio.html)
- [SageMaker Domain](https://docs.aws.amazon.com/sagemaker/latest/dg/sm-domain.html)
- [SageMaker Studio Pricing](https://aws.amazon.com/sagemaker/pricing/)
- [SageMaker Network Configuration](https://docs.aws.amazon.com/sagemaker/latest/dg/studio-notebooks-and-internet-access.html)
- [SageMaker Docker Access](https://docs.aws.amazon.com/sagemaker/latest/dg/studio-updated-local.html)
