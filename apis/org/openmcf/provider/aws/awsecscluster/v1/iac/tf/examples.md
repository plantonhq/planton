# AWS ECS Cluster Examples

Below are several examples demonstrating how to define an AWS ECS Cluster component in
OpenMCF. After creating one of these YAML manifests, apply it with Terraform using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Basic ECS Cluster

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: basic-ecs-cluster
  version:
    message: "Initial cluster"
spec:
  region: us-east-1
  enableContainerInsights: true
  capacityProviders:
    - FARGATE
  enableExecuteCommand: false
```

This example creates a basic ECS cluster:
• Uses Fargate capacity provider for serverless containers.
• Enables CloudWatch Container Insights for monitoring.
• ECS Exec disabled for security.
• Simple configuration for development or testing.

---

## ECS Cluster with Multiple Capacity Providers

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: mixed-capacity-cluster
  version:
    message: "Multi capacity providers cluster"
spec:
  region: us-east-1
  enableContainerInsights: true
  capacityProviders:
    - FARGATE
    - FARGATE_SPOT
  enableExecuteCommand: false
```

This example optimizes for cost and availability:
• Combines Fargate and Fargate Spot for cost savings.
• Spot instances provide up to 70% cost reduction.
• Container Insights enabled for performance monitoring.
• Suitable for production workloads with cost optimization.

---

## ECS Cluster with ECS Exec Enabled

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: exec-enabled-cluster
  version:
    message: "Exec debugging example"
spec:
  region: us-east-1
  enableContainerInsights: true
  capacityProviders:
    - FARGATE
  enableExecuteCommand: true
```

This example enables debugging capabilities:
• ECS Exec allows shell access to running containers.
• Useful for troubleshooting and operational tasks.
• Container Insights for comprehensive monitoring.
• Requires proper IAM permissions for ECS Exec.

---

## Production ECS Cluster

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: production-ecs-cluster
  version:
    message: "Production cluster with full monitoring"
spec:
  region: us-east-1
  enableContainerInsights: true
  capacityProviders:
    - FARGATE
    - FARGATE_SPOT
  enableExecuteCommand: true
```

This example is production-ready:
• Full monitoring with Container Insights.
• Cost optimization with Spot instances.
• Debugging capabilities enabled.
• Comprehensive observability for production workloads.

---

## Development ECS Cluster

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: development-ecs-cluster
  version:
    message: "Development cluster"
spec:
  region: us-east-1
  enableContainerInsights: true
  capacityProviders:
    - FARGATE
  enableExecuteCommand: true
```

This example is optimized for development:
• ECS Exec enabled for debugging.
• Container Insights for development monitoring.
• Single Fargate provider for simplicity.
• Cost-effective for development environments.

---

## Minimal ECS Cluster

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: minimal-ecs-cluster
  version:
    message: "Minimal cluster example"
spec:
  region: us-east-1
  enableContainerInsights: true
```

A minimal configuration with:
• Only required fields specified.
• Container Insights enabled by default.
• No capacity providers (can be added later).
• ECS Exec disabled for security.

---

## After Deploying

Once you've applied your manifest with OpenMCF tofu, you can confirm that the ECS cluster is active via the AWS console or by
using the AWS CLI:

```shell
aws ecs list-clusters
```

You should see your new ECS cluster in the list. For more detailed information:

```shell
aws ecs describe-clusters --clusters <your-cluster-name>
```

This will show cluster details including capacity providers, container insights status, and ECS Exec configuration.
