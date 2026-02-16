# AWS SageMaker Domain

Deploys an Amazon SageMaker Domain — the top-level organizational unit for SageMaker Studio that provides ML teams with JupyterLab notebooks, custom compute kernels, and shared workspace infrastructure backed by VPC networking, dedicated EFS storage, and configurable IAM execution roles.

## What Gets Created

When you deploy an AwsSagemakerDomain resource, OpenMCF provisions:

- **SageMaker Domain** — an `aws_sagemaker_domain` resource with the specified auth mode (IAM or SSO), VPC configuration, and default user settings
- **Dedicated EFS File System** — automatically created by AWS for user home directories, optionally encrypted with a customer-managed KMS key
- **Domain Boundary Security Group** — automatically created by AWS to control cross-app and cross-user traffic within the domain
- **SSO Application** — created only when `authMode` is `SSO`, integrating with AWS IAM Identity Center for user authentication

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **A VPC** with DNS resolution and DNS hostnames enabled
- **At least one subnet** (two or more in different AZs recommended for high availability)
- **An IAM execution role** with a trust policy allowing `sagemaker.amazonaws.com` to assume it, plus permissions for S3, ECR, and other AWS services the team needs
- **Security groups** (optional) for user-level and domain-level network isolation
- **A KMS key** (optional) for encrypting the EFS home directory file system

## Quick Start

Create a file `sagemaker-domain.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSagemakerDomain
metadata:
  name: ml-team-domain
spec:
  authMode: IAM
  vpcId:
    value: vpc-0abc123def456
  subnetIds:
    - value: subnet-0a1b2c3d4e5f00001
    - value: subnet-0a1b2c3d4e5f00002
  defaultUserSettings:
    executionRoleArn:
      value: arn:aws:iam::123456789012:role/SageMakerExecRole
```

Deploy:

```shell
openmcf apply -f sagemaker-domain.yaml
```

This creates a SageMaker Domain with IAM authentication, JupyterLab available by default, and the EFS home directory encrypted with the AWS-managed key.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `authMode` | `string` | Authentication mode: `IAM` or `SSO`. ForceNew. | Must be `IAM` or `SSO` |
| `vpcId` | `StringValueOrRef` | VPC where the domain is created. ForceNew. | Required. Default kind: `AwsVpc` |
| `subnetIds` | `StringValueOrRef[]` | VPC subnets for notebook ENIs. ForceNew. Max 16. | Minimum 1 item. Default kind: `AwsVpc` |
| `defaultUserSettings` | `object` | Default configuration for all user profiles. | Required |
| `defaultUserSettings.executionRoleArn` | `StringValueOrRef` | IAM role assumed by SageMaker for ML workloads. | Required. Default kind: `AwsIamRole` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `kmsKeyId` | `StringValueOrRef` | AWS-managed key | KMS key ARN for EFS encryption. ForceNew. Default kind: `AwsKmsKey`. |
| `appNetworkAccessType` | `string` | `PublicInternetOnly` | Network mode: `PublicInternetOnly` or `VpcOnly`. |
| `domainSecurityGroupIds` | `StringValueOrRef[]` | `[]` | Domain-level security groups. ForceNew. Max 3. Default kind: `AwsSecurityGroup`. |
| `dockerSettings.enableDockerAccess` | `string` | — | `ENABLED` or `DISABLED`. Allows Docker in notebooks. |
| `dockerSettings.vpcOnlyTrustedAccounts` | `string[]` | `[]` | AWS account IDs for trusted Docker pulls (VpcOnly mode). Max 20. |
| `defaultUserSettings.securityGroupIds` | `StringValueOrRef[]` | `[]` | User-level security groups. Max 5. Default kind: `AwsSecurityGroup`. |
| `defaultUserSettings.defaultLandingUri` | `string` | Platform default | Landing page URI (e.g., `studio::relative/JupyterLab`). |
| `defaultUserSettings.studioWebPortal` | `string` | `ENABLED` | `ENABLED` or `DISABLED`. Controls Studio web portal access. |
| `defaultUserSettings.jupyterLabAppSettings` | `object` | — | JupyterLab IDE configuration. See below. |
| `defaultUserSettings.kernelGatewayAppSettings` | `object` | — | Custom compute kernel configuration. See below. |
| `defaultUserSettings.sharingSettings` | `object` | — | Notebook output sharing to S3. See below. |
| `defaultUserSettings.spaceStorageSettings` | `object` | — | Default EBS volume sizes for spaces. See below. |

### JupyterLab App Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `defaultResourceSpec.instanceType` | `string` | — | Default compute instance (e.g., `ml.t3.medium`, `ml.m5.large`). |
| `defaultResourceSpec.lifecycleConfigArn` | `string` | — | Lifecycle config ARN for default startup scripts. |
| `defaultResourceSpec.sagemakerImageArn` | `string` | — | Custom SageMaker image ARN. |
| `lifecycleConfigArns` | `string[]` | `[]` | Additional lifecycle config ARNs. |
| `customImages` | `object[]` | `[]` | Custom Docker images as kernels. Max 200. |
| `customImages[].appImageConfigName` | `string` | — | AppImageConfig name. Required. |
| `customImages[].imageName` | `string` | — | SageMaker Image name. Required. |
| `customImages[].imageVersionNumber` | `int` | Latest | Pin to specific version. |
| `codeRepositories` | `object[]` | `[]` | Git repos cloned on startup. Max 10. |
| `codeRepositories[].repositoryUrl` | `string` | — | HTTPS Git URL. Required. Max 1024 chars. |
| `idleSettings.lifecycleManagement` | `string` | — | `ENABLED` or `DISABLED`. Auto-shutdown for idle instances. |
| `idleSettings.idleTimeoutInMinutes` | `int` | — | Minutes before auto-shutdown. Range: 60-525600. |
| `idleSettings.minIdleTimeoutInMinutes` | `int` | — | Minimum user-configurable timeout. Range: 60-525600. |
| `idleSettings.maxIdleTimeoutInMinutes` | `int` | — | Maximum user-configurable timeout. Range: 60-525600. |

### KernelGateway App Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `defaultResourceSpec` | `object` | — | Same structure as JupyterLab `defaultResourceSpec`. |
| `lifecycleConfigArns` | `string[]` | `[]` | Lifecycle config ARNs for KernelGateway. |
| `customImages` | `object[]` | `[]` | Custom Docker images. Same structure as JupyterLab. Max 200. |

### Sharing Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `notebookOutputOption` | `string` | `Disabled` | `Allowed` or `Disabled`. Controls S3 output sharing. |
| `s3OutputPath` | `string` | — | S3 URI for outputs. Required when `Allowed`. |
| `s3KmsKeyId` | `StringValueOrRef` | S3 default | KMS key for output encryption. Default kind: `AwsKmsKey`. |

### Space Storage Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `defaultEbsVolumeSizeInGb` | `int` | — | Default EBS size per space. Required when block is set. |
| `maximumEbsVolumeSizeInGb` | `int` | — | Maximum EBS size. Must be >= default. Required when block is set. |

## Examples

### VPC-Only Domain with KMS Encryption

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSagemakerDomain
metadata:
  name: secure-ml-domain
spec:
  authMode: SSO
  appNetworkAccessType: VpcOnly
  vpcId:
    value: vpc-0abc123def456
  subnetIds:
    - value: subnet-0a1b2c3d4e5f00001
    - value: subnet-0a1b2c3d4e5f00002
  kmsKeyId:
    value: arn:aws:kms:us-east-1:123456789012:key/mrk-abc123
  domainSecurityGroupIds:
    - value: sg-domain001
  defaultUserSettings:
    executionRoleArn:
      value: arn:aws:iam::123456789012:role/SageMakerExecRole
    securityGroupIds:
      - value: sg-user001
```

### JupyterLab with Idle Timeout and Code Repos

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSagemakerDomain
metadata:
  name: ml-workspace
spec:
  authMode: IAM
  vpcId:
    value: vpc-0abc123def456
  subnetIds:
    - value: subnet-0a1b2c3d4e5f00001
    - value: subnet-0a1b2c3d4e5f00002
  defaultUserSettings:
    executionRoleArn:
      value: arn:aws:iam::123456789012:role/SageMakerExecRole
    defaultLandingUri: "studio::relative/JupyterLab"
    jupyterLabAppSettings:
      defaultResourceSpec:
        instanceType: ml.t3.medium
      idleSettings:
        lifecycleManagement: ENABLED
        idleTimeoutInMinutes: 120
      codeRepositories:
        - repositoryUrl: "https://github.com/team/ml-notebooks.git"
```

### Cross-Resource References with valueFrom

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSagemakerDomain
metadata:
  name: composable-domain
spec:
  authMode: IAM
  vpcId:
    valueFrom:
      name: ml-vpc
  subnetIds:
    - valueFrom:
        name: ml-vpc
  kmsKeyId:
    valueFrom:
      name: ml-kms-key
  defaultUserSettings:
    executionRoleArn:
      valueFrom:
        name: sagemaker-exec-role
    securityGroupIds:
      - valueFrom:
          name: notebook-sg
```

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `domain_id` | `string` | SageMaker Domain ID used in API calls and user profile creation. |
| `domain_arn` | `string` | Domain ARN for IAM policies and cross-service references. |
| `domain_url` | `string` | HTTPS URL for accessing the SageMaker Studio web interface. |
| `home_efs_file_system_id` | `string` | ID of the auto-created EFS file system for user home directories. |
| `security_group_id_for_domain_boundary` | `string` | ID of the AWS-managed security group for domain network boundary. |
| `single_sign_on_application_arn` | `string` | IAM Identity Center application ARN. Populated when `authMode` is `SSO`. |

## Related Components

- **[AwsVpc](/docs/catalog/aws/vpc)** — VPC where the domain's network interfaces are placed
- **[AwsIamRole](/docs/catalog/aws/iam-role)** — execution role granting SageMaker access to AWS resources
- **[AwsSecurityGroup](/docs/catalog/aws/security-group)** — network isolation for domain and user traffic
- **[AwsKmsKey](/docs/catalog/aws/kms-key)** — customer-managed encryption for EFS and S3 outputs
- **[AwsS3Bucket](/docs/catalog/aws/s3-bucket)** — storage for shared notebook outputs
- **[AwsCloudwatchLogGroup](/docs/catalog/aws/cloudwatch-log-group)** — logging for SageMaker training jobs
