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
- **VPC Connector** (`vpcConnectorArn`) -- References a shared first-class `AwsAppRunnerVpcConnector` so the service can reach private resources; one connector serves many services. The connector's subnets need a NAT Gateway for outbound internet access.
- **Auto scaling configuration** (`autoScalingConfigurationArn`) -- References a shared `AwsAppRunnerAutoScalingConfiguration`; a warm floor of 2+ instances eliminates cold starts, and a lowered concurrency ceiling gives each instance headroom.
- **Observability** (`observabilityConfigurationArn`) -- References a shared `AwsAppRunnerObservabilityConfiguration`; the reference itself enables X-Ray tracing.
- **KMS encryption** (`kmsKeyArn`) -- Encrypts stored image copies and data logs with your key. **ForceNew**: changing this value replaces the entire service.
- **WAF inspection** (`webAclArn`) -- Associates a REGIONAL `AwsWafWebAcl`; every request passes WAF before reaching the application.
- **HTTP health check** -- Validates application-level readiness, not just port availability.
- **Auto-deploy enabled** (`autoDeploymentsEnabled: true`) -- New pushes to the tracked private-ECR tag deploy automatically. Set to false where deployments must be deliberate.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `imageIdentifier` | Your private ECR image path -- the preset ships a format-valid example (`123456789012.dkr.ecr.us-east-1.amazonaws.com/my-api:v1.0.0`); replace account, region, repo, and tag with your own | AWS ECR Console |
| `<aws-region>` | AWS region (e.g., `us-east-1`) | Your deployment region |
| `<ecr-access-role>` | Name of the `AwsIamRole` granting ECR pull access | Your resource graph |
| `<application-port>` | Port your app listens on (e.g., `8080`) | Your Dockerfile or app config |
| `<instance-role>` | Name of the `AwsIamRole` for runtime AWS API access | Your resource graph |
| `<secrets-manager-arn-or-ssm-parameter-arn>` | ARN of the secret or parameter to inject | Secrets Manager or SSM Console |
| `<vpc-connector>` | Name of the `AwsAppRunnerVpcConnector` to route egress through | Your resource graph |
| `<auto-scaling-configuration>` | Name of the `AwsAppRunnerAutoScalingConfiguration` to adopt | Your resource graph |
| `<observability-configuration>` | Name of the `AwsAppRunnerObservabilityConfiguration` to adopt | Your resource graph |
| `<kms-key>` | Name of the `AwsKmsKey` for encryption at rest | Your resource graph |
| `<web-acl>` | Name of the REGIONAL `AwsWafWebAcl` to associate | Your resource graph |
| `<health-check-path>` | HTTP health check path (e.g., `/health`, `/healthz`) | Your application's health endpoint |

## Related Presets

- **01-basic-public-image** -- Use for quick prototyping without VPC or encryption.
- **03-github-code-source** -- Use when deploying from source code instead of a pre-built image.
