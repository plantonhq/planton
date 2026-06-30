# AWS ECS Service Pulumi Module

This module automates the provisioning and lifecycle management of an Amazon ECS service on AWS. It integrates with
Planton’s unified API resource model, letting you define your desired state using a single YAML manifest. Under
the hood, it supports both Pulumi and Terraform to accommodate different IaC preferences while maintaining consistent
resource definitions and validations via Protobuf.

By centralizing all ECS service configuration—including CPU, memory, VPC networking, and IAM roles—you can rapidly
deploy
and manage containerized applications without directly juggling multiple AWS console pages or CLI steps. Whether
launching a small service on Fargate or managing a web application with load balancing, **AwsEcsService** ensures the same
predictable, streamlined experience across dev, staging, and production environments.

---

## Key Features

### Unified API Resource Model

- **Kubernetes-Like Structure**  
  Uses `apiVersion`, `kind`, `metadata`, and `spec` fields for consistency across all Planton resources.

- **Protobuf Validations**  
  Ensures every field—like container image, CPU, memory, and subnets—is checked against the ECS service schema, reducing
  misconfigurations.

### Comprehensive ECS Provisioning

- **Fargate & EC2**  
  Deploy to AWS Fargate for serverless containers or to EC2 for more control, both using the same manifest
  specification.

- **Load Balancing & Networking**  
  Easily connect your ECS service to an ALB/NLB, configure private subnets, and attach security groups for a secure,
  production-ready setup.

- **Scalable & Observable**  
  Specify desired task counts, choose the container port to expose, and leverage ECS/CloudWatch for logging and metrics.

### Secure and Flexible

- **IAM Roles & Environment Variables**  
  Integrate custom task and execution IAM roles, inject environment variables, and reference secrets if needed.

- **Pulumi & Terraform**  
  Seamlessly switch between Pulumi or Terraform provisioning, applying the same YAML manifest without rewriting
  infrastructure logic.

---

## Installation

1. **Planton CLI**  
   Ensure you have [Planton CLI](https://github.com/plantonhq/planton) installed. This command-line
   tool handles manifest validation, provisioning, and integration with Pulumi or Terraform.

2. **AWS Credentials**  
   Provide your AWS credentials in one of the following ways:
    - Set `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables, **or**
    - Create an AWS credential resource within Planton and reference it in your manifest.

3. **Pulumi or Terraform**
    - [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/) if you plan to use Pulumi.
    - [Terraform CLI](https://developer.hashicorp.com/terraform/downloads) if you prefer Terraform.

---

## Usage

1. **Define a Manifest**  
   Create a YAML file specifying your ECS service. For example:

   ```yaml
   apiVersion: aws.planton.dev/v1
   kind: AwsEcsService
   metadata:
     name: my-aws-ecs-service
     version:
       message: "Initial ECS service deployment"
   spec:
     cluster_name: my-ecs-cluster
     service_name: my-aws-ecs-service
     image: "123456789012.dkr.ecr.us-east-1.amazonaws.com/myapp:latest"
     container_port: 8080
     desired_count: 2
     cpu: 512
     memory: 1024
     subnets:
       - subnet-abc123
       - subnet-def456
     security_groups:
       - sg-123456
     assign_public_ip: false
   ```

2. **Validate (Optional)**  
   Optionally, validate your manifest to catch schema or field errors:
   ```bash
   planton validate --manifest awsecsservice.yaml
   ```

3. **Deploy**  
   Use the Planton CLI to deploy via Pulumi or Terraform:
   ```bash
   # Pulumi example
   planton pulumi up --manifest awsecsservice.yaml --stack myorg/dev

   # Terraform example
   planton terraform apply --manifest awsecsservice.yaml --stack myorg/dev
   ```

4. **Verify**  
   After provisioning completes, verify the service in your AWS console or using AWS CLI:
   ```bash
   aws ecs describe-services --cluster my-ecs-cluster --services my-aws-ecs-service
   ```

---

## API Resource Specification

Below are the key fields for the **AwsEcsService** resource. Consult the Protobuf definitions in the Planton
repository for the complete schema.

### `apiVersion`

- **Description**: Must be `"aws.planton.dev/v1"`.
- **Required**: Yes.

### `kind`

- **Description**: Must be `"AwsEcsService"`.
- **Required**: Yes.

### `metadata`

- **Description**: Standard metadata including resource name and versioning details.
- **Fields**:
    - **name**: Unique resource name (3–63 characters).
    - **version.message**: A mandatory string for version/change context.

### `spec.cluster_name`

- **Type**: `string`
- **Description**: Specifies the name or ARN of the ECS cluster where this service will run.
- **Required**: Yes.

### `spec.service_name`

- **Type**: `string`
- **Description**: Unique ECS service identifier within the specified cluster.
- **Required**: Yes (3–63 characters).

### `spec.image`

- **Type**: `string`
- **Description**: Container image URI (ECR, Docker Hub, etc.).
- **Required**: Yes.

### `spec.container_port`

- **Type**: `int32`
- **Default**: None (optional if no inbound traffic).
- **Description**: Port exposed on the container for incoming requests.

### `spec.desired_count`

- **Type**: `int32`
- **Default**: `1`
- **Description**: Number of ECS task replicas to run for the service.

### `spec.cpu`

- **Type**: `int32`
- **Required**: Yes
- **Description**: CPU units allocated to the task (e.g., 256, 512, 1024, etc.).

### `spec.memory`

- **Type**: `int32`
- **Required**: Yes
- **Description**: Memory (MiB) allocated to the task (e.g., 512, 1024, 2048, etc.).

### `spec.subnets`

- **Type**: `repeated string`
- **Required**: Yes
- **Description**: Subnet IDs where tasks will run (often private subnets for production).

### `spec.security_groups`

- **Type**: `repeated string`
- **Description**: Security group IDs attached to the Fargate/EC2 tasks. Omit to default to the VPC’s default SG (not
  recommended for production).

### `spec.assign_public_ip`

- **Type**: `bool`
- **Default**: `false`
- **Description**: Assign a public IP to the Fargate tasks. Typically false for private or load-balanced deployments.

### `spec.task_execution_role_arn`

- **Type**: `string`
- **Description**: IAM role ARN used by ECS to pull private images and write logs. Omit to use a default
  `ecsTaskExecutionRole` if available.

### `spec.task_role_arn`

- **Type**: `string`
- **Description**: IAM role ARN that your container can assume. Ideal for apps needing AWS API access.

### `spec.environment`

- **Type**: `repeated EnvironmentVar`
- **Description**: Environment variables to pass into the container. Useful for configuration or secret references.

---

## Customization and Extensibility

- **Load Balancing**  
  Integrate a load balancer by attaching a target group to the container port. This can be handled automatically if
  you’re using the **AwsEcsService** in a more advanced blueprint.

- **Autoscaling**  
  Enable AWS ECS Service Auto Scaling by referencing CloudWatch metrics and scaling policies (if desired) through
  Planton
  extended specs or manual AWS configuration.

- **Advanced IAM**  
  Provide custom task roles for use cases requiring AWS API interactions (e.g., S3 access, DynamoDB queries).

- **EC2 Launch Type**  
  If you prefer to run on EC2 instead of Fargate, the underlying Planton modules can adapt accordingly by setting
  the appropriate capacity provider or launch type.

---

## Further Reading

- **[examples.md](./examples.md)**: Step-by-step sample manifests for various ECS service scenarios.
- **[Planton Guide](https://github.com/plantonhq/planton/blob/main/docs/Guide.md)**: Learn how to
  leverage multi-cloud workflows, the CLI, and advanced usage patterns.
- **AWS Documentation**: [Amazon ECS](https://docs.aws.amazon.com/ecs) for an in-depth look at ECS features and
  best practices.
