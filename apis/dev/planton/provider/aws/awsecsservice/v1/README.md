# Overview

The **AwsEcsService** API resource provides a standardized and straightforward way to deploy containerized applications
onto an existing Amazon ECS cluster on AWS. By focusing on essential configurations like image definition, compute
capacity (Fargate or EC2), networking, and environment variables, it makes running services on ECS far more accessible
within the Planton multi-cloud deployment framework.

## Purpose

Deploying ECS services typically involves handling multiple moving parts—task definitions, networking, autoscaling,
IAM roles, and more. The **AwsEcsService** resource aims to streamline that process by:

- **Simplifying ECS Deployments**: Offer an easy-to-use interface for spinning up microservices on ECS (Fargate or EC2).
- **Aligning with Best Practices**: Provide recommended defaults (e.g., CPU, memory) to ensure users have a
  production-ready
  baseline without repetitive configuration.
- **Promoting Consistency**: Enforce standardized naming and validations, reducing misconfigurations across
  multiple services and environments.

## Key Features

### Single-Container Focus

- **Minimal, Opinionated Spec**: Focuses on the 80-20 use case—a single-container service—while still exposing fields
  for
  resource requirements, environment variables, and networking.

### Flexible Compute Options

- **Fargate or EC2**: Operate serverless via AWS Fargate, or integrate with your existing EC2-backed ECS environment.
- **Resource Control**: Define CPU and memory precisely, aligned with ECS constraints (e.g., 256, 512, 1024 CPU units).

### Automatic Networking Setup

- **Subnets & Security Groups**: Attach your service to specific VPC subnets, choosing whether to assign a public IP.
- **Public or Private**: Easily configure production deployments in private subnets, or set up a publicly accessible
  service when needed.

### Environment Management

- **Environment Variables**: Pass configuration to your container, including references to secrets from AWS Secrets
  Manager or SSM.
- **Role Separation**: Separate `task_execution_role_arn` (for pulling container images and writing logs) from
  `task_role_arn` (for runtime AWS API access).

### Auto Scaling

- **Target Tracking Scaling**: Automatically scale your ECS service based on CPU or memory utilization.
- **Flexible Configuration**: Set minimum and maximum task counts, and define target CPU/memory percentages.
- **Production-Ready**: Uses AWS Application Auto Scaling with sensible cooldown periods (300s scale-in, 60s scale-out).

### Health Check Grace Period

- **Prevent Premature Failures**: Configure a grace period during which ECS ignores ALB health check failures.
- **Essential for Slow-Start Apps**: Critical for applications like Spring Boot or JVM-based services that take 30-120
  seconds to fully boot.
- **Recommended Default**: Defaults to 60 seconds, configurable up to several minutes for complex applications.

### Seamless Integration

- **Planton CLI**: Deploy the same resource across multiple stacks using either Pulumi or Terraform under the
  hood.
- **Multi-Cloud Ready**: Combine AwsEcsService on AWS with other providers in the same manifest, adopting Planton’s
  uniform resource model.

## Benefits

- **Reduced Complexity**: A single definition for your ECS service—container image, CPU/memory, subnets, and more—means
  fewer files and less overhead.
- **Automatic Scaling**: Built-in support for target tracking auto scaling eliminates manual capacity management.
- **Production-Ready Defaults**: Health check grace periods and auto scaling prevent common deployment failures.
- **Infrastructure Consistency**: Enforce naming conventions, validations, and recommended defaults for CPU/memory
  allocations so your deployments remain predictable and repeatable.
- **Enhanced Observability**: Integrate seamlessly with ECS cluster features like CloudWatch metrics and logs—no extra
  manual setup needed.
- **Service-Oriented Abstraction**: Focus on defining *what* you want to run (a service), not *how* to wire together
  infrastructure components.

## Example Usage

Below is a production-ready example demonstrating how to configure and deploy an ECS service with autoscaling and health
check grace period using Planton:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsEcsService
metadata:
  name: production-api-service
spec:
  clusterArn: "arn:aws:ecs:us-east-1:123456789012:cluster/production-cluster"
  container:
    image:
      repo: "123456789012.dkr.ecr.us-east-1.amazonaws.com/myapp"
      tag: "v1.0.0"
    port: 8080
    replicas: 2
    cpu: 512
    memory: 1024
    env:
      variables:
        ENVIRONMENT: "production"
      secrets:
        DATABASE_URL: "arn:aws:secretsmanager:us-east-1:123456789012:secret:prod/db-url"
  network:
    subnets:
      - "subnet-1abc234d"
      - "subnet-2abc345e"
    securityGroups:
      - "sg-111aaabbb"
  alb:
    enabled: true
    arn: "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/prod-alb/xyz"
    routingType: "hostname"
    hostname: "api.example.com"
    listenerPort: 443
  healthCheckGracePeriodSeconds: 60
  autoscaling:
    enabled: true
    minTasks: 2
    maxTasks: 10
    targetCpuPercent: 75
```

### Deploying with Planton

Once your YAML manifest is ready, you can deploy using Planton’s CLI. Planton will validate the manifest
against the Protobuf schema and orchestrate everything in Pulumi or Terraform.

- **Using Pulumi**:
  ```bash
  planton pulumi up --manifest awsecsservice.yaml --stack org/project/my-stack
  ```
- **Using Terraform**:
  ```bash
  planton terraform apply --manifest awsecsservice.yaml --stack org/project/my-stack
  ```

Planton will provision the ECS service, create or update the necessary IAM roles (if specified), assign the
service to the given subnets and security groups, and ensure you have the correct number of running tasks.

---

Happy deploying! If you have questions or run into issues, feel free to open an issue on our GitHub repository or
reach out through our community channels for support.
