# AWS ECS Service Terraform Module

This Terraform module deploys an AWS ECS service using Fargate with optional Application Load Balancer (ALB) integration.

## Features

- **Fargate-based ECS Service**: Serverless container execution
- **ALB Integration**: Optional path-based or hostname-based routing
- **CloudWatch Logging**: Automatic log group creation and configuration
- **Environment Variables**: Support for plain variables and secrets
- **S3 Environment Files**: Support for environment files from S3
- **IAM Role Integration**: Task execution and task roles
- **Health Checks**: Configurable ALB target group health checks

## Usage

Use the Planton CLI (tofu) with the default local backend:

```bash
planton tofu apply --manifest ecs-service.yaml --stack myorg/dev
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| aws | >= 5.0, < 6.0 |

## Inputs

The module accepts the standard Planton metadata and spec structure as defined in the protobuf schema.

## Outputs

| Name | Description |
|------|-------------|
| aws_ecs_service_name | The final name of the ECS service |
| ecs_cluster_name | Indicates which cluster the service is deployed in |
| load_balancer_dns_name | The DNS name of the ALB if configured |
| service_url | The final external endpoint if ALB is configured |
| cloudwatch_log_group_name | The name of the CloudWatch log group |
| cloudwatch_log_group_arn | The ARN of the CloudWatch log group |
| service_arn | The ARN of the ECS service |
| task_definition_arn | The ARN of the ECS task definition |
| target_group_arn | The ARN of the target group when ALB is enabled |
| target_group_name | The name of the target group when ALB is enabled |

