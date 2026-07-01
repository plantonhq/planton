# AWS ML Workbench

Provisions a self-service ML development environment with SageMaker Studio, S3 for model artifacts, optional ECR for custom container images, and optional VPC network isolation. Enable only the components your data-science team needs.

## Architecture

```
                     Data Scientists
                           │
                           ▼
                  ┌─────────────────────┐
                  │ AwsSagemakerDomain   │
                  │ (Studio notebooks)  │
                  └───┬──────────┬──────┘
                      │          │
           ┌──────────┘          └──────────┐
           ▼                                ▼
  ┌──────────────────┐            ┌──────────────────┐
  │   AwsS3Bucket    │            │   AwsEcrRepo     │
  │ (artifacts/data) │            │ (custom images)  │
  └──────────────────┘            └──────────────────┘

  ┌──────────────────┐   ┌─────────────────────────────┐
  │   AwsIamRole     │   │  AwsVpc + AwsSecurityGroup  │
  │ (execution role) │   │  (optional network isolation)│
  └──────────────────┘   └─────────────────────────────┘
```

## Dependency Graph

```
Layer 0 (parallel):  AwsIamRole, AwsS3Bucket, AwsEcrRepo, AwsVpc
Layer 1 (dep VPC):   AwsSecurityGroup (when vpcEnabled)
Layer 2 (dep IAM):   AwsSagemakerDomain
```

## Included Cloud Resources

| Resource | Kind | Group | Condition | Purpose |
|----------|------|-------|-----------|---------|
| IAM Role | `AwsIamRole` | identity | Always | SageMaker execution role with S3 access |
| S3 Bucket | `AwsS3Bucket` | storage | Always | Model artifacts, datasets, and notebooks |
| ECR Repository | `AwsEcrRepo` | storage | `customImagesEnabled` | Custom SageMaker container images |
| VPC | `AwsVpc` | network | `vpcEnabled` | Network isolation for SageMaker |
| Security Group | `AwsSecurityGroup` | network | `vpcEnabled` | HTTPS-only ingress within VPC |
| SageMaker Domain | `AwsSagemakerDomain` | compute | Always | Studio IDE for notebooks and experiments |

## Parameters

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| **General** | | | |
| `aws_region` | AWS region | `us-east-1` | Yes |
| **SageMaker** | | | |
| `sagemaker_domain_name` | Studio domain name | `ml-workbench` | Yes |
| `auth_mode` | Authentication mode (`IAM` or `SSO`) | `IAM` | Yes |
| **Storage** | | | |
| `s3_bucket_name` | S3 bucket for model data | `ml-workbench-data` | Yes |
| **VPC** | | | |
| `vpcEnabled` | Deploy inside a VPC | `false` | No |
| `vpc_availability_zone_1` | First AZ (when VPC enabled) | `us-east-1a` | No |
| `vpc_availability_zone_2` | Second AZ (when VPC enabled) | `us-east-1b` | No |
| **Custom Images** | | | |
| `customImagesEnabled` | Create ECR repository | `false` | No |
| `ecr_repo_name` | ECR repository name | `ml-custom-images` | No |

## Common Configurations

### Minimal (SageMaker + S3, no VPC)

```yaml
vpcEnabled: false
customImagesEnabled: false
```

### Production (VPC-isolated with custom images)

```yaml
vpcEnabled: true
customImagesEnabled: true
auth_mode: SSO
```

## Important Notes

- When `vpcEnabled: false`, SageMaker uses **PublicInternetOnly** network access. Notebooks can reach the public internet directly. This is the simplest setup for experimentation.
- When `vpcEnabled: true`, SageMaker uses **VpcOnly** network access. All traffic is routed through the VPC. NAT Gateway is enabled so notebooks can still install packages from the internet.
- The IAM role grants `AmazonSageMakerFullAccess` and `AmazonS3FullAccess`. Scope these policies down for production workloads.
- S3 bucket versioning is enabled by default to protect against accidental deletions of model artifacts and datasets.
- ECR images are set to **mutable** tags (unlike compute charts) because ML images are frequently rebuilt during experimentation.

---

© 2025 Planton. All rights reserved.
