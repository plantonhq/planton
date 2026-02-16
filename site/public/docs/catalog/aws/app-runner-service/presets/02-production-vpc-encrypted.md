---
title: "Production VPC-Connected and Encrypted Service"
description: "This preset creates a production-grade App Runner service with private ECR image, VPC egress, customer-managed KMS encryption, tuned auto scaling, and HTTP health checks. It represents the..."
type: "preset"
rank: "02"
presetSlug: "02-production-vpc-encrypted"
componentSlug: "app-runner-service"
componentTitle: "App Runner Service"
provider: "aws"
icon: "package"
order: 2
---

# Production VPC-Connected and Encrypted Service

This preset creates a production-grade App Runner service with private ECR image, VPC egress, customer-managed KMS encryption, tuned auto scaling, and HTTP health checks. It represents the recommended baseline for production API workloads that need to reach VPC resources (databases, caches, internal services).

## When to Use

- Production APIs that connect to RDS, ElastiCache, or other VPC-internal resources
- Services handling sensitive data that require customer-managed encryption keys
- Workloads with SLAs that demand zero cold-start latency (warm instance pool)
- Environments with compliance requirements (SOC2, HIPAA, PCI-DSS)

## Key Configuration Choices

- **Private ECR image** (`imageRepositoryType: ECR`) -- Uses your own container registry. The `accessRoleArn` grants App Runner permission to pull images.
- **2 vCPU / 4 GB memory** (`cpu: 2048`, `memory: 4096`) -- Sized for production API workloads. Adjust based on your application's resource profile.
- **Instance role** (`instanceRoleArn`) -- IAM role assumed at runtime for calling AWS APIs. Follow least-privilege: only grant permissions your application actually needs.
- **VPC Connector** (`subnetIds`, `securityGroupIds`) -- Creates an inline VPC Connector so the service can reach private resources. Subnets must have a NAT Gateway for outbound internet access.
- **KMS encryption** (`kmsKeyArn`) -- Encrypts stored image copies and data logs with your key. **ForceNew**: changing this value replaces the entire service.
- **Tuned auto scaling** -- `minSize: 2` keeps 2 warm instances at all times (eliminates cold starts). `maxConcurrency: 50` scales out earlier, giving each instance more headroom.
- **HTTP health check** -- Validates application-level readiness, not just port availability.
- **Auto-deploy disabled** (`autoDeploymentsEnabled: false`) -- Production deployments should be triggered deliberately, not by image pushes.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<account-id>` | Your AWS account ID (12 digits) | AWS Console → Account Settings |
| `<region>` | AWS region (e.g., `us-east-1`) | Your deployment region |
| `<repo>` | ECR repository name | AWS ECR Console |
| `<tag>` | Image tag (e.g., `v1.0.0`, `latest`) | Your CI/CD pipeline |
| `<ecr-access-role-arn>` | IAM role ARN for ECR image pulling | IAM Console or `AwsIamRole` outputs |
| `<application-port>` | Port your app listens on (e.g., `8080`) | Your Dockerfile or app config |
| `<instance-role-arn>` | IAM role ARN for runtime AWS API access | IAM Console or `AwsIamRole` outputs |
| `<secrets-manager-arn-or-ssm-parameter-arn>` | ARN of the secret or parameter to inject | Secrets Manager or SSM Console |
| `<private-subnet-1>`, `<private-subnet-2>` | Private subnet IDs in 2+ AZs | VPC Console or `AwsVpc` outputs |
| `<security-group-id>` | Security group ID for the VPC Connector | VPC Console or `AwsSecurityGroup` outputs |
| `<kms-key-arn>` | KMS key ARN for encryption at rest | KMS Console or `AwsKmsKey` outputs |
| `<health-check-path>` | HTTP health check path (e.g., `/health`, `/healthz`) | Your application's health endpoint |

## Related Presets

- **01-basic-public-image** -- Use for quick prototyping without VPC or encryption.
- **03-github-code-source** -- Use when deploying from source code instead of a pre-built image.
