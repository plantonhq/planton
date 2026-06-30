# Overview

The **AwsEcsCluster** API resource provides a standardized, streamlined way to deploy and manage an Amazon ECS cluster on
AWS. It encapsulates key configurations—such as Fargate support, capacity providers, and ECS Exec options—into a single,
uniform resource that seamlessly integrates with the Planton multi-cloud deployment framework.

## Purpose

Deploying ECS clusters on AWS can involve numerous configurations, from defining capacity providers to enabling
container insights for production monitoring. The **AwsEcsCluster** resource aims to:

- **Simplify Cluster Provisioning**: Provide an easy-to-use interface for creating ECS clusters tailored to Fargate or
  EC2-based workloads.
- **Enable Production-Ready Monitoring**: Optionally enable CloudWatch Container Insights to gain observability into
  containerized workloads.
- **Streamline Debugging**: Allow for ECS Exec configuration so teams can `exec` into running containers for
  troubleshooting and operational tasks.
- **Encourage Standards**: Enforce consistent naming and recommended defaults to maintain a uniform, predictable setup
  across multiple deployments and environments.

## Key Features

### Container Insights

- **CloudWatch Container Insights**: Easily enable container-level metrics, logs, and traces. Helps teams quickly
  diagnose issues and optimize resource usage.

### Capacity Providers with Cost Optimization

- **Fargate and Fargate Spot**: Effortlessly configure a Fargate-only cluster or mix in Spot capacity. This allows cost
  optimization while retaining a serverless container environment.
- **Default Capacity Provider Strategy**: Define the base/weight distribution for tasks across capacity providers, 
  enabling production cost optimization patterns like guaranteed on-demand base + Spot scaling. This is the **primary 
  cost-optimization lever** for Fargate workloads, enabling up to 70% cost savings with Fargate Spot.
- **Production Pattern**: Configure strategies like "1 guaranteed on-demand task + 80/20 Spot/on-demand scaling" to 
  balance reliability and cost.

### ECS Exec Support with Production-Grade Auditing

- **Enable Execute Command**: Optionally allow shell access into running ECS tasks for live debugging. This feature is
  extremely useful during development or in production for resolving critical issues.
- **Cluster-Level Auditing**: Configure comprehensive audit logging for exec sessions with CloudWatch Logs and/or S3 
  storage for compliance and security monitoring.
- **Encryption Support**: Enable KMS encryption for exec session audit logs to meet security and compliance requirements.
- **Flexible Logging Options**: Choose from AWS-managed defaults, custom CloudWatch log groups, S3 buckets, or combined 
  logging strategies.

### Seamless Integration

- **Planton CLI**: Manifests validated and deployed using the same familiar CLI, ensuring a consistent developer
  experience.
- **Multi-Cloud Ready**: Combine AWS ECS with other providers or resources in the same workflow, using the uniform
  Planton resource model.

## Benefits

- **Consistent Deployments**: A single resource definition for ECS clusters reduces the likelihood of misconfiguration
  or fragmentation across environments.
- **Improved Observability**: Built-in support for enabling CloudWatch Container Insights ensures teams can monitor and
  troubleshoot services without complex manual setup.
- **Cost Optimization**: Leverage capacity providers and advanced scheduling options (e.g., Spot, Fargate) to balance
  performance and budget constraints.
- **Future-Proof**: As AWS evolves ECS capabilities, updates to the AwsEcsCluster resource keep your deployments aligned
  with best practices and new features.

## Example Usage

Below is a production-ready YAML example demonstrating how to configure and deploy an ECS cluster with cost optimization 
and exec auditing using Planton:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsEcsCluster
metadata:
  name: production-cluster
  version:
    message: "Production ECS cluster with cost optimization"
spec:
  # Enable CloudWatch Container Insights for comprehensive monitoring
  enable_container_insights: true
  
  # Cost-optimized capacity providers (up to 70% savings with Spot)
  capacity_providers:
    - "FARGATE"
    - "FARGATE_SPOT"
  
  # Production cost-optimization: 1 guaranteed on-demand + 80/20 Spot scaling
  default_capacity_provider_strategy:
    - capacity_provider: "FARGATE"
      base: 1              # Guarantee 1 on-demand task for stability
      weight: 1            # 1 part on-demand for scaling (20%)
    - capacity_provider: "FARGATE_SPOT"
      base: 0              # No minimum Spot tasks required
      weight: 4            # 4 parts Spot for scaling (80%)
  
  # Production-grade exec auditing with CloudWatch logging
  execute_command_configuration:
    logging: OVERRIDE
    log_configuration:
      cloud_watch_log_group_name: "/aws/ecs/prod/exec"
      cloud_watch_encryption_enabled: true
    kms_key_id: "arn:aws:kms:us-east-1:123456789012:key/prod-key"
```

You would then deploy this using the Planton CLI (either Pulumi or Terraform under the hood):

```bash
planton pulumi up --manifest awsecscluster.yaml --stack org/project/stack
```

Or:

```bash
planton terraform apply --manifest awsecscluster.yaml --stack org/project/stack
```

This deploys a production-ready ECS cluster with:
- CloudWatch Container Insights for observability
- Fargate Spot cost optimization (up to 70% savings) with the 80/20 base/weight strategy
- Production-grade exec auditing with CloudWatch logging and KMS encryption

---

Happy deploying! If you have any questions or issues, feel free to open an issue in our GitHub repository or reach out
via the Planton community channels.
