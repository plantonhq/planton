# Create using CLI

Create a yaml manifest using one of the examples below. After the YAML is created, use the command below to apply with
OpenMCF:

```shell
openmcf pulumi up --manifest <yaml-path> --stack <stack-name>
```

Or, if using Terraform:

```shell
openmcf terraform apply --manifest <yaml-path> --stack <stack-name>
```

(You can also use a shorter form like `planton apply -f <yaml-path>` if your environment is configured accordingly.)

---

# Basic Example

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: dev-cluster
  version:
    message: "Development cluster"
spec:
  region: us-east-1

  # Recommended to be true for production, enabling CloudWatch Container Insights
  enable_container_insights: true

  # Using default Fargate capacity
  capacity_providers:
    - "FARGATE"
```

This example creates a simple AWS ECS cluster that runs on Fargate, with container insights enabled.

---

# Production Example with Cost Optimization

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: production-cluster
  version:
    message: "Production cluster with cost optimization"
spec:
  region: us-east-1

  # Enable CloudWatch insights for comprehensive monitoring
  enable_container_insights: true

  # Include both FARGATE and FARGATE_SPOT for cost optimization
  capacity_providers:
    - "FARGATE"
    - "FARGATE_SPOT"

  # Define default capacity provider strategy for cost optimization
  # This guarantees 1 on-demand task for stability, and scales with 80% Spot / 20% on-demand
  default_capacity_provider_strategy:
    - capacity_provider: "FARGATE"
      base: 1              # Guarantee 1 on-demand task for stability
      weight: 1            # 1 part on-demand for scaling (20%)
    - capacity_provider: "FARGATE_SPOT"
      base: 0              # No minimum Spot tasks required
      weight: 4            # 4 parts Spot for scaling (80%)
```

This production example demonstrates the **80/20 cost optimization pattern**: guarantee stability with a base on-demand task, 
while scaling 80% on Spot capacity (up to 70% cost savings) and 20% on-demand. This is the primary cost-optimization lever 
for Fargate workloads.

---

# Example with ECS Exec (Default Logging)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: debug-cluster
  version:
    message: "Cluster with ECS Exec for debugging"
spec:
  region: us-east-1

  # Container Insights are highly recommended for monitoring
  enable_container_insights: true

  # Single capacity provider for a basic Fargate cluster
  capacity_providers:
    - "FARGATE"

  # Enable ECS Exec with AWS-managed default logging
  execute_command_configuration:
    logging: DEFAULT
```

When ECS Exec is enabled with `DEFAULT` logging, you can connect to running tasks for debugging or operational 
troubleshooting using `aws ecs execute-command`, with AWS-managed audit logging.

---

# Example with ECS Exec (Custom CloudWatch Logging)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: audit-cluster
  version:
    message: "Cluster with custom exec auditing"
spec:
  region: us-east-1

  # Enable CloudWatch insights for monitoring
  enable_container_insights: true

  # Production capacity providers
  capacity_providers:
    - "FARGATE"
    - "FARGATE_SPOT"

  # Enable ECS Exec with custom CloudWatch logging for compliance
  execute_command_configuration:
    logging: OVERRIDE
    log_configuration:
      cloud_watch_log_group_name: "/aws/ecs/prod/exec-audit"
      cloud_watch_encryption_enabled: true
    kms_key_id: "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
```

This example enables ECS Exec with custom CloudWatch logging for compliance and security monitoring. All exec session 
commands and output are logged to a specific CloudWatch log group with KMS encryption.

---

# Example with ECS Exec (S3 Audit Logging)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: compliance-cluster
  version:
    message: "Cluster with S3 exec audit logging"
spec:
  region: us-east-1

  # Enable CloudWatch insights for monitoring
  enable_container_insights: true

  # Production capacity providers
  capacity_providers:
    - "FARGATE"

  # Enable ECS Exec with S3 logging for long-term audit retention
  execute_command_configuration:
    logging: OVERRIDE
    log_configuration:
      s3_bucket_name: "my-compliance-audit-bucket"
      s3_key_prefix: "ecs-exec-logs/"
      s3_encryption_enabled: true
    kms_key_id: "arn:aws:kms:us-east-1:123456789012:key/compliance-key"
```

This example stores ECS Exec audit logs in S3 for long-term retention and compliance requirements, with KMS encryption enabled.

---

# Full Production Configuration

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: production-full
  version:
    message: "Production cluster with all features"
spec:
  region: us-east-1

  # Enable CloudWatch Container Insights for comprehensive monitoring
  enable_container_insights: true

  # Cost-optimized capacity providers
  capacity_providers:
    - "FARGATE"
    - "FARGATE_SPOT"

  # Production cost-optimization: 1 guaranteed on-demand + 80/20 Spot scaling
  default_capacity_provider_strategy:
    - capacity_provider: "FARGATE"
      base: 1
      weight: 1
    - capacity_provider: "FARGATE_SPOT"
      base: 0
      weight: 4

  # Production-grade exec auditing with both CloudWatch and S3
  execute_command_configuration:
    logging: OVERRIDE
    log_configuration:
      cloud_watch_log_group_name: "/aws/ecs/prod/exec"
      cloud_watch_encryption_enabled: true
      s3_bucket_name: "prod-ecs-audit-logs"
      s3_key_prefix: "exec-sessions/"
      s3_encryption_enabled: true
    kms_key_id: "arn:aws:kms:us-east-1:123456789012:key/prod-key"
```

This comprehensive production example includes:
- **CloudWatch Container Insights** for observability
- **Fargate Spot cost optimization** (up to 70% savings) with the base/weight strategy
- **Production-grade exec auditing** with both CloudWatch and S3 logging for compliance
- **KMS encryption** for all audit logs

---

# Minimal Development Configuration

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: dev-minimal
  version:
    message: "Minimal development cluster"
spec:
  region: us-east-1
  # Accept defaults; Container Insights recommended default is true
  # No capacity providers (can be added later)
  # No exec configuration (disabled by default)
```

If you need a bare-bones cluster definition for development, you can omit most fields. However, for production-ready 
setups, explicitly configure capacity providers with the default strategy and consider enabling exec auditing for 
operational debugging.

---

After creating a YAML manifest for your ECS cluster, apply the configuration using either Pulumi or Terraform with the
OpenMCF CLI. The CLI will validate your manifest against the Protobuf schema, generate the required
infrastructure code, and provision the cluster on AWS.

For example:

```shell
openmcf pulumi up --manifest minimal-ecs.yaml --stack myorg/dev
```

Or:

```shell
openmcf terraform apply --manifest minimal-ecs.yaml --stack myorg/dev
```

Upon completion, you can check the newly created ECS cluster in the AWS Console or with the AWS CLI:

```shell
aws ecs list-clusters
```

This confirms that your ECS cluster has been created and is ready for workloads.
