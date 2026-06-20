---
title: "ECS Service"
description: "ECS Service deployment documentation"
icon: "package"
order: 100
componentName: "awsecsservice"
---

# AWS ECS Service

Deploys a Fargate-based ECS service with a task definition, optional ALB integration (path-based or hostname-based routing), CloudWatch logging, and target-tracking autoscaling. The component creates the task definition, ECS service, target group, listener rules, and scaling policies from a single manifest.

## What Gets Created

When you deploy an AwsEcsService resource, OpenMCF provisions:

- **CloudWatch Log Group** — a `cloudwatch.LogGroup` named `/ecs/<serviceName>` with 30-day retention (created when logging is enabled, which is the default)
- **ECS Task Definition** — an `ecs.TaskDefinition` configured for Fargate with `awsvpc` networking, the specified CPU/memory, container image, environment variables, secrets, S3 environment files, and optional IAM roles
- **ECS Service** — an `ecs.Service` running the task definition on the specified cluster with the desired replica count and network configuration
- **ALB Target Group** — an `lb.TargetGroup` of type `ip` (created only when `alb.enabled` is `true`) with configurable health check settings
- **ALB Listener Rule** — an `lb.ListenerRule` for path-based or hostname-based routing (created only when `alb.enabled` is `true` and a routing type is specified)
- **Auto Scaling Target** — an `appautoscaling.Target` for the ECS service (created only when `autoscaling.enabled` is `true`)
- **CPU Scaling Policy** — an `appautoscaling.Policy` using `ECSServiceAverageCPUUtilization` target tracking (created when `autoscaling.targetCpuPercent` is set)
- **Memory Scaling Policy** — an `appautoscaling.Policy` using `ECSServiceAverageMemoryUtilization` target tracking (created when `autoscaling.targetMemoryPercent` is set)

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **An existing ECS cluster** (deploy with [AwsEcsCluster](/docs/catalog/aws/ecs-cluster) or provide an ARN)
- **At least one VPC subnet** for Fargate task placement
- **A security group** allowing traffic to the container port
- **An existing ALB** with a listener on the target port if enabling ALB integration (deploy with [AwsAlb](/docs/catalog/aws/alb))
- **IAM roles** for task execution (image pull, log writes) and optional task role (AWS API access)

## Quick Start

Create a file `ecs-service.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsService
metadata:
  name: my-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsEcsService.my-api
spec:
  clusterArn: arn:aws:ecs:us-east-1:123456789012:cluster/my-cluster
  container:
    image:
      repo: nginx
      tag: latest
    cpu: 256
    memory: 512
    port: 80
  network:
    subnets:
      - subnet-0a1b2c3d4e5f00001
      - subnet-0a1b2c3d4e5f00002
```

Deploy:

```shell
openmcf apply -f ecs-service.yaml
```

This creates a single-replica Fargate service running nginx with CloudWatch logging enabled by default.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `clusterArn` | `StringValueOrRef` | ARN of the ECS cluster where this service runs. Can reference an AwsEcsCluster resource via `valueFrom`. | Required |
| `clusterArn.value` | `string` | Direct cluster ARN value | — |
| `clusterArn.valueFrom` | `object` | Foreign key reference | Default kind: `AwsEcsCluster`, field: `status.outputs.cluster_arn` |
| `container.cpu` | `int32` | vCPU units for the task. Valid Fargate values: 256, 512, 1024, 2048, 4096. | Required |
| `container.memory` | `int32` | Memory in MiB for the task. Valid values depend on CPU (e.g., 256 CPU supports 512-2048 MiB). | Required |
| `network.subnets` | `StringValueOrRef[]` | VPC subnet IDs for Fargate task placement. Can reference AwsVpc resources via `valueFrom`. | Required, at least 1 item |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `container.image.repo` | `string` | `""` | Container image repository (e.g., `nginx`, `123456789012.dkr.ecr.us-east-1.amazonaws.com/my-app`). |
| `container.image.tag` | `string` | `""` | Container image tag (e.g., `latest`, `v1.2.3`). |
| `container.port` | `int32` | `0` | Container port to expose. Omit for background workers that do not receive inbound traffic. |
| `container.replicas` | `int32` | `1` | Number of task replicas. Higher values improve availability at increased cost. |
| `container.env.variables` | `map<string,string>` | `{}` | Environment variables injected into the container as key-value pairs. |
| `container.env.secrets` | `map<string,string>` | `{}` | Secret values injected as environment variables. Values can be plaintext or ARN references to Secrets Manager / SSM Parameter Store. |
| `container.env.s3Files` | `string[]` | `[]` | S3 URIs loaded as environment files via the ECS `environmentFiles` feature. Must be unique. |
| `container.logging.enabled` | `bool` | `true` | Auto-creates a CloudWatch Log Group at `/ecs/<serviceName>` with 30-day retention and configures the `awslogs` driver. |
| `network.securityGroups` | `StringValueOrRef[]` | `[]` | Security group IDs for task ENIs. Can reference AwsSecurityGroup resources via `valueFrom`. |
| `iam.taskExecutionRoleArn` | `StringValueOrRef` | — | IAM role for ECS to pull images and write logs. Can reference an AwsIamRole resource via `valueFrom`. |
| `iam.taskRoleArn` | `StringValueOrRef` | — | IAM role the container assumes for AWS API calls. Can reference an AwsIamRole resource via `valueFrom`. |
| `alb.enabled` | `bool` | `false` | Attaches the service to an ALB via a target group and listener rule. |
| `alb.arn` | `StringValueOrRef` | — | ARN of the ALB. Required when `alb.enabled` is `true`. Can reference an AwsAlb resource via `valueFrom`. |
| `alb.routingType` | `string` | — | Routing mode. Valid values: `path`, `hostname`. |
| `alb.path` | `string` | `""` | URL path pattern for routing (e.g., `/api/*`). Required when `routingType` is `path`. |
| `alb.hostname` | `string` | `""` | Hostname for routing (e.g., `api.example.com`). Required when `routingType` is `hostname`. |
| `alb.listenerPort` | `int32` | `80` | Port on the ALB listener to attach the rule to. |
| `alb.listenerPriority` | `int32` | `100` | Priority of the listener rule. Lower numbers have higher priority. Must be unique per ALB. |
| `alb.healthCheck.protocol` | `string` | `HTTP` | Health check protocol. Valid values: `HTTP`, `HTTPS`, `TCP`. |
| `alb.healthCheck.path` | `string` | `/` | Health check path (HTTP/HTTPS only). |
| `alb.healthCheck.port` | `string` | `traffic-port` | Health check port. Use `traffic-port` or an explicit port number as a string. |
| `alb.healthCheck.interval` | `int32` | `30` | Seconds between health checks. |
| `alb.healthCheck.timeout` | `int32` | `5` | Seconds before a health check times out. |
| `alb.healthCheck.healthyThreshold` | `int32` | `5` | Consecutive successes before a target is considered healthy. |
| `alb.healthCheck.unhealthyThreshold` | `int32` | `2` | Consecutive failures before a target is considered unhealthy. |
| `healthCheckGracePeriodSeconds` | `int32` | `60` | Seconds ECS ignores ALB health check failures during container startup. Only applies when `alb.enabled` is `true`. |
| `autoscaling.enabled` | `bool` | `false` | Enables target-tracking autoscaling for the service. |
| `autoscaling.minTasks` | `int32` | — | Minimum number of tasks. Must be >= 1. Required when autoscaling is enabled. |
| `autoscaling.maxTasks` | `int32` | — | Maximum number of tasks. Must be >= 1 and >= `minTasks`. Required when autoscaling is enabled. |
| `autoscaling.targetCpuPercent` | `int32` | `75` | Target average CPU utilization percentage (1-100). Scaling out occurs when CPU exceeds this threshold. |
| `autoscaling.targetMemoryPercent` | `int32` | — | Target average memory utilization percentage (1-100). Optional; most services scale on CPU alone. |

## Examples

### Background Worker (No Ingress)

A background processing service with no exposed port and no ALB:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsService
metadata:
  name: queue-worker
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsEcsService.queue-worker
spec:
  clusterArn: arn:aws:ecs:us-east-1:123456789012:cluster/my-cluster
  container:
    image:
      repo: 123456789012.dkr.ecr.us-east-1.amazonaws.com/worker
      tag: v1.0.0
    cpu: 512
    memory: 1024
    replicas: 2
    env:
      variables:
        QUEUE_URL: https://sqs.us-east-1.amazonaws.com/123456789012/my-queue
        WORKER_CONCURRENCY: "10"
  network:
    subnets:
      - subnet-private-az1
      - subnet-private-az2
    securityGroups:
      - sg-worker
  iam:
    taskExecutionRoleArn: arn:aws:iam::123456789012:role/ecsTaskExecutionRole
    taskRoleArn: arn:aws:iam::123456789012:role/workerTaskRole
```

### Service with Path-Based ALB Routing

An API service fronted by an ALB using path-based routing:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsService
metadata:
  name: api-service
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEcsService.api-service
spec:
  clusterArn: arn:aws:ecs:us-east-1:123456789012:cluster/prod-cluster
  container:
    image:
      repo: 123456789012.dkr.ecr.us-east-1.amazonaws.com/api
      tag: v2.1.0
    cpu: 1024
    memory: 2048
    port: 8080
    replicas: 3
  network:
    subnets:
      - subnet-private-az1
      - subnet-private-az2
    securityGroups:
      - sg-api
  iam:
    taskExecutionRoleArn: arn:aws:iam::123456789012:role/ecsTaskExecutionRole
  alb:
    enabled: true
    arn: arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/prod-alb/1234567890
    routingType: path
    path: /api/*
    listenerPort: 80
    listenerPriority: 10
    healthCheck:
      protocol: HTTP
      path: /health
      interval: 15
      timeout: 5
      healthyThreshold: 3
      unhealthyThreshold: 2
  healthCheckGracePeriodSeconds: 90
```

### Service with Hostname-Based ALB Routing

A web application routed by hostname through an ALB:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsService
metadata:
  name: web-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEcsService.web-app
spec:
  clusterArn: arn:aws:ecs:us-east-1:123456789012:cluster/prod-cluster
  container:
    image:
      repo: 123456789012.dkr.ecr.us-east-1.amazonaws.com/web
      tag: v3.0.0
    cpu: 512
    memory: 1024
    port: 3000
    replicas: 2
  network:
    subnets:
      - subnet-private-az1
      - subnet-private-az2
    securityGroups:
      - sg-web
  alb:
    enabled: true
    arn: arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/prod-alb/1234567890
    routingType: hostname
    hostname: app.example.com
    listenerPort: 80
```

### Service with Autoscaling

A service that scales between 2 and 10 replicas based on CPU utilization:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsService
metadata:
  name: scalable-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEcsService.scalable-api
spec:
  clusterArn: arn:aws:ecs:us-east-1:123456789012:cluster/prod-cluster
  container:
    image:
      repo: 123456789012.dkr.ecr.us-east-1.amazonaws.com/api
      tag: v2.5.0
    cpu: 1024
    memory: 2048
    port: 8080
    replicas: 2
  network:
    subnets:
      - subnet-private-az1
      - subnet-private-az2
    securityGroups:
      - sg-api
  iam:
    taskExecutionRoleArn: arn:aws:iam::123456789012:role/ecsTaskExecutionRole
  alb:
    enabled: true
    arn: arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/prod-alb/1234567890
    routingType: hostname
    hostname: api.example.com
    listenerPort: 80
    healthCheck:
      protocol: HTTP
      path: /health
  autoscaling:
    enabled: true
    minTasks: 2
    maxTasks: 10
    targetCpuPercent: 70
    targetMemoryPercent: 80
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding ARNs and IDs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsService
metadata:
  name: ref-service
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEcsService.ref-service
spec:
  clusterArn:
    valueFrom:
      kind: AwsEcsCluster
      name: prod-cluster
      field: status.outputs.cluster_arn
  container:
    image:
      repo: 123456789012.dkr.ecr.us-east-1.amazonaws.com/app
      tag: v1.0.0
    cpu: 512
    memory: 1024
    port: 8080
  network:
    subnets:
      - valueFrom:
          kind: AwsSubnet
          name: prod-private-subnet-a
          fieldPath: status.outputs.subnet_id
      - valueFrom:
          kind: AwsSubnet
          name: prod-private-subnet-b
          fieldPath: status.outputs.subnet_id
    securityGroups:
      - valueFrom:
          kind: AwsSecurityGroup
          name: app-sg
          field: status.outputs.security_group_id
  iam:
    taskExecutionRoleArn:
      valueFrom:
        kind: AwsIamRole
        name: ecs-exec-role
        field: status.outputs.role_arn
    taskRoleArn:
      valueFrom:
        kind: AwsIamRole
        name: app-task-role
        field: status.outputs.role_arn
  alb:
    enabled: true
    arn:
      valueFrom:
        kind: AwsAlb
        name: prod-alb
        field: status.outputs.load_balancer_arn
    routingType: hostname
    hostname: app.example.com
    listenerPort: 80
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `aws_ecs_service_name` | `string` | Name of the created ECS service |
| `ecs_cluster_name` | `string` | Cluster ARN/name the service is deployed in |
| `load_balancer_dns_name` | `string` | DNS name of the ALB (empty if ALB is not enabled) |
| `service_url` | `string` | External URL constructed from `alb.hostname` when hostname routing is used (empty otherwise) |
| `service_discovery_name` | `string` | Internal DNS name if service discovery is configured |
| `cloudwatch_log_group_name` | `string` | Name of the CloudWatch log group (e.g., `/ecs/my-api`) |
| `cloudwatch_log_group_arn` | `string` | ARN of the CloudWatch log group |
| `service_arn` | `string` | ARN of the ECS service |
| `target_group_arn` | `string` | ARN of the ALB target group (empty if ALB is not enabled) |

## Related Components

- [AwsEcsCluster](/docs/catalog/aws/ecs-cluster) — provides the cluster where this service runs
- [AwsAlb](/docs/catalog/aws/alb) — provides the Application Load Balancer for ingress traffic
- [AwsVpc](/docs/catalog/aws/vpc) — provides subnets for Fargate task placement
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls network access to task ENIs
- [AwsIamRole](/docs/catalog/aws/iam-role) — provides task execution and task roles
- [AwsEcrRepo](/docs/catalog/aws/ecr-repo) — hosts container images deployed by this service
