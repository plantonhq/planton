---
title: "SageMaker Domain"
description: "SageMaker Domain deployment documentation"
icon: "package"
order: 100
componentName: "awssagemakerdomain"
---

# AWS SageMaker Domain

Deploys an Amazon SageMaker Domain with VPC networking, default user settings including JupyterLab and KernelGateway configuration, optional Docker access, and idle timeout policies for cost management. The domain is the top-level workspace for SageMaker Studio, provisioning a dedicated EFS file system for user home directories and establishing IAM execution roles for ML workloads.

## What Gets Created

When you deploy an AwsSagemakerDomain resource, OpenMCF provisions:

- **SageMaker Domain** — a `sagemaker.Domain` resource placed in the specified VPC and subnets, with the configured authentication mode (IAM or SSO), default user settings, and optional domain-level settings (Docker, security groups)
- **Dedicated EFS File System** — automatically created by AWS for user home directories (the `home_efs_file_system_id` is exposed as a stack output)
- **Domain Boundary Security Group** — automatically created by AWS to control cross-app and cross-user traffic within the domain (the `security_group_id_for_domain_boundary` is exposed as a stack output)
- **IAM Identity Center Application** — created only when `authMode` is `SSO`, registers the domain as an SSO application for centralized identity management

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **A VPC** with DNS resolution and DNS hostnames enabled
- **At least one subnet** in the VPC for SageMaker network interfaces (private subnets recommended for production)
- **An IAM execution role** with a trust policy for `sagemaker.amazonaws.com`, granting access to S3, ECR, and other services the ML workloads need
- **AWS IAM Identity Center** configured in the account if using `SSO` authentication mode
- **A security group** allowing outbound traffic for notebook and training workloads

## Quick Start

Create a file `sagemaker-domain.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSagemakerDomain
metadata:
  name: my-domain
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsSagemakerDomain.my-domain
spec:
  authMode: IAM
  vpcId: vpc-0123456789abcdef0
  subnetIds:
    - subnet-0a1b2c3d4e5f00001
    - subnet-0a1b2c3d4e5f00002
  defaultUserSettings:
    executionRoleArn: arn:aws:iam::123456789012:role/SageMakerExecutionRole
```

Deploy:

```shell
openmcf apply -f sagemaker-domain.yaml
```

This creates a SageMaker Domain with IAM authentication, public internet access for notebooks, and the SageMaker-provided default JupyterLab environment.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `authMode` | `string` | Authentication mode for the domain. ForceNew. | Must be `IAM` or `SSO` |
| `vpcId` | `StringValueOrRef` | VPC where the domain is created. ForceNew. Can reference AwsVpc via `valueFrom`. | Required |
| `subnetIds` | `StringValueOrRef[]` | Subnets for SageMaker network interfaces. ForceNew. Can reference AwsVpc via `valueFrom`. | Minimum 1 item |
| `defaultUserSettings.executionRoleArn` | `StringValueOrRef` | IAM role assumed by SageMaker for notebooks, training, and inference. Must trust `sagemaker.amazonaws.com`. Can reference AwsIamRole via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `kmsKeyId` | `StringValueOrRef` | AWS-managed key | KMS key ARN for encrypting the domain's EFS volume. ForceNew. Can reference AwsKmsKey via `valueFrom`. |
| `appNetworkAccessType` | `string` | `PublicInternetOnly` | Network access for notebooks. `VpcOnly` keeps all traffic in the VPC (recommended for production). |
| `domainSecurityGroupIds` | `StringValueOrRef[]` | `[]` | Domain-level security groups for shared resources. Maximum 3. ForceNew. Can reference AwsSecurityGroup via `valueFrom`. |
| `defaultUserSettings.securityGroupIds` | `StringValueOrRef[]` | `[]` | User-level security groups for notebook ENIs. Maximum 5. Can reference AwsSecurityGroup via `valueFrom`. |
| `defaultUserSettings.defaultLandingUri` | `string` | — | Default app opened on domain access. Common: `studio::relative/JupyterLab`. |
| `defaultUserSettings.studioWebPortal` | `string` | `ENABLED` | Whether the Studio web portal is accessible. Valid values: `ENABLED`, `DISABLED`. |
| `defaultUserSettings.jupyterLabAppSettings.defaultResourceSpec.instanceType` | `string` | — | Default compute instance type (e.g., `ml.t3.medium`, `ml.m5.large`, `ml.g4dn.xlarge`). |
| `defaultUserSettings.jupyterLabAppSettings.defaultResourceSpec.lifecycleConfigArn` | `string` | — | Lifecycle config script ARN for JupyterLab startup. |
| `defaultUserSettings.jupyterLabAppSettings.lifecycleConfigArns` | `string[]` | `[]` | Lifecycle config script ARNs that run on JupyterLab app start. |
| `defaultUserSettings.jupyterLabAppSettings.customImages` | `object[]` | `[]` | Custom Docker images available as JupyterLab kernels. Maximum 200. |
| `defaultUserSettings.jupyterLabAppSettings.customImages[].appImageConfigName` | `string` | — | SageMaker AppImageConfig name. Required per image. |
| `defaultUserSettings.jupyterLabAppSettings.customImages[].imageName` | `string` | — | SageMaker Image resource name. Required per image. |
| `defaultUserSettings.jupyterLabAppSettings.customImages[].imageVersionNumber` | `int32` | latest | Pin to a specific image version. |
| `defaultUserSettings.jupyterLabAppSettings.codeRepositories` | `object[]` | `[]` | Git repos cloned into JupyterLab on startup. Maximum 10. |
| `defaultUserSettings.jupyterLabAppSettings.codeRepositories[].repositoryUrl` | `string` | — | HTTPS Git URL (max 1024 chars). Required per repo. |
| `defaultUserSettings.jupyterLabAppSettings.idleSettings.lifecycleManagement` | `string` | — | Enable idle shutdown. Valid values: `ENABLED`, `DISABLED`. |
| `defaultUserSettings.jupyterLabAppSettings.idleSettings.idleTimeoutInMinutes` | `int32` | — | Minutes of inactivity before shutdown. Range: 60-525600. Requires `lifecycleManagement: ENABLED`. |
| `defaultUserSettings.jupyterLabAppSettings.idleSettings.minIdleTimeoutInMinutes` | `int32` | — | Minimum idle timeout users can set. Range: 60-525600. |
| `defaultUserSettings.jupyterLabAppSettings.idleSettings.maxIdleTimeoutInMinutes` | `int32` | — | Maximum idle timeout users can set. Range: 60-525600. |
| `defaultUserSettings.kernelGatewayAppSettings.defaultResourceSpec.instanceType` | `string` | — | Default compute instance type for KernelGateway apps. |
| `defaultUserSettings.kernelGatewayAppSettings.lifecycleConfigArns` | `string[]` | `[]` | Lifecycle config ARNs for KernelGateway apps. |
| `defaultUserSettings.kernelGatewayAppSettings.customImages` | `object[]` | `[]` | Custom Docker images for KernelGateway kernels. Maximum 200. |
| `defaultUserSettings.sharingSettings.notebookOutputOption` | `string` | `Disabled` | Persist notebook outputs to S3. Valid values: `Allowed`, `Disabled`. |
| `defaultUserSettings.sharingSettings.s3OutputPath` | `string` | — | S3 URI for shared outputs. Required when `notebookOutputOption` is `Allowed`. |
| `defaultUserSettings.sharingSettings.s3KmsKeyId` | `StringValueOrRef` | — | KMS key for encrypting shared outputs. Can reference AwsKmsKey via `valueFrom`. |
| `defaultUserSettings.spaceStorageSettings.defaultEbsVolumeSizeInGb` | `int32` | — | Default EBS volume size (GB) for user spaces. Required when `spaceStorageSettings` is set. |
| `defaultUserSettings.spaceStorageSettings.maximumEbsVolumeSizeInGb` | `int32` | — | Maximum EBS volume size (GB) users can request. Must be >= default. Required when `spaceStorageSettings` is set. |
| `dockerSettings.enableDockerAccess` | `string` | — | Enable Docker in notebooks/terminals. Valid values: `ENABLED`, `DISABLED`. |
| `dockerSettings.vpcOnlyTrustedAccounts` | `string[]` | `[]` | AWS account IDs allowed for Docker image pulls in VpcOnly mode. Maximum 20. |

## Examples

### IAM Authentication with JupyterLab Idle Shutdown

Cost-optimized domain with automatic shutdown of idle JupyterLab instances after 2 hours:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSagemakerDomain
metadata:
  name: ml-team
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsSagemakerDomain.ml-team
spec:
  authMode: IAM
  vpcId: vpc-0123456789abcdef0
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
  defaultUserSettings:
    executionRoleArn: arn:aws:iam::123456789012:role/SageMakerExecutionRole
    defaultLandingUri: "studio::relative/JupyterLab"
    jupyterLabAppSettings:
      defaultResourceSpec:
        instanceType: ml.t3.medium
      idleSettings:
        lifecycleManagement: ENABLED
        idleTimeoutInMinutes: 120
```

### SSO Authentication with VPC-Only Networking

Enterprise domain using IAM Identity Center and VPC-only access for compliance:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSagemakerDomain
metadata:
  name: enterprise-ml
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsSagemakerDomain.enterprise-ml
spec:
  authMode: SSO
  vpcId: vpc-0123456789abcdef0
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
    - subnet-private-az3
  kmsKeyId: arn:aws:kms:us-east-1:123456789012:key/mrk-abc123
  appNetworkAccessType: VpcOnly
  defaultUserSettings:
    executionRoleArn: arn:aws:iam::123456789012:role/SageMakerExecutionRole
    securityGroupIds:
      - sg-user-notebooks
    jupyterLabAppSettings:
      defaultResourceSpec:
        instanceType: ml.m5.large
      idleSettings:
        lifecycleManagement: ENABLED
        idleTimeoutInMinutes: 120
        minIdleTimeoutInMinutes: 60
        maxIdleTimeoutInMinutes: 480
  domainSecurityGroupIds:
    - sg-domain-boundary
```

### Full-Featured with Docker, Custom Images, and Code Repos

Production domain with Docker access, custom KernelGateway images, auto-cloned repositories, notebook sharing, and space storage limits:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSagemakerDomain
metadata:
  name: data-science
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsSagemakerDomain.data-science
spec:
  authMode: SSO
  vpcId: vpc-0123456789abcdef0
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
  kmsKeyId: arn:aws:kms:us-east-1:123456789012:key/mrk-abc123
  appNetworkAccessType: VpcOnly
  defaultUserSettings:
    executionRoleArn: arn:aws:iam::123456789012:role/SageMakerExecutionRole
    securityGroupIds:
      - sg-user-notebooks
    defaultLandingUri: "studio::relative/JupyterLab"
    jupyterLabAppSettings:
      defaultResourceSpec:
        instanceType: ml.m5.large
      codeRepositories:
        - repositoryUrl: "https://github.com/org/ml-notebooks.git"
        - repositoryUrl: "https://github.com/org/shared-utils.git"
      idleSettings:
        lifecycleManagement: ENABLED
        idleTimeoutInMinutes: 120
    kernelGatewayAppSettings:
      defaultResourceSpec:
        instanceType: ml.g4dn.xlarge
      customImages:
        - appImageConfigName: pytorch-gpu-config
          imageName: pytorch-gpu
          imageVersionNumber: 3
    sharingSettings:
      notebookOutputOption: Allowed
      s3OutputPath: "s3://ml-team-outputs/notebooks/"
    spaceStorageSettings:
      defaultEbsVolumeSizeInGb: 50
      maximumEbsVolumeSizeInGb: 200
  domainSecurityGroupIds:
    - sg-domain-boundary
  dockerSettings:
    enableDockerAccess: ENABLED
    vpcOnlyTrustedAccounts:
      - "123456789012"
      - "987654321098"
```

### Using Foreign Key References

Reference OpenMCF-managed VPC, IAM role, security groups, and KMS key instead of hardcoding ARNs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSagemakerDomain
metadata:
  name: ref-domain
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsSagemakerDomain.ref-domain
spec:
  authMode: IAM
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: ml-vpc
      field: status.outputs.vpc_id
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: ml-vpc
        field: status.outputs.private_subnets[0].id
    - valueFrom:
        kind: AwsVpc
        name: ml-vpc
        field: status.outputs.private_subnets[1].id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: ml-key
      field: status.outputs.key_arn
  defaultUserSettings:
    executionRoleArn:
      valueFrom:
        kind: AwsIamRole
        name: sagemaker-exec
        field: status.outputs.role_arn
    securityGroupIds:
      - valueFrom:
          kind: AwsSecurityGroup
          name: notebook-sg
          field: status.outputs.security_group_id
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `domain_id` | `string` | Unique identifier of the SageMaker Domain, used when creating user profiles and spaces |
| `domain_arn` | `string` | Amazon Resource Name for IAM policies and CloudWatch metrics |
| `domain_url` | `string` | HTTPS URL for accessing the SageMaker Studio web interface |
| `home_efs_file_system_id` | `string` | ID of the EFS file system created for user home directories |
| `security_group_id_for_domain_boundary` | `string` | ID of the AWS-managed security group controlling cross-app traffic |
| `single_sign_on_application_arn` | `string` | IAM Identity Center application ARN. Only populated when `authMode` is `SSO`. |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides the VPC and subnets for domain networking
- [AwsIamRole](/docs/catalog/aws/iam-role) — provides the execution role assumed by SageMaker for ML workloads
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls network access for domain and user-level resources
- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides customer-managed encryption for the domain's EFS volume
- [AwsS3Bucket](/docs/catalog/aws/s3-bucket) — stores notebook outputs when sharing is enabled
