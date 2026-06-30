# AwsEcsCluster Pulumi/Terraform Module

This module automates the provisioning of an Amazon ECS (Elastic Container Service) cluster, integrating seamlessly with
Planton’s unified API resource model. It supports both Pulumi and Terraform under the hood, allowing teams to
adopt the IaC workflow that best suits their environment. With the **AwsEcsCluster** resource, you can quickly spin up
Fargate- or EC2-backed clusters, manage capacity providers (including Spot), and enable key features like ECS Exec and
CloudWatch Container Insights.

By relying on a declarative YAML specification, you can define all cluster settings—such as capacity providers,
monitoring, and debugging—without the need to juggle multiple AWS interfaces or create custom scripts. This consistent,
validated approach reduces operational complexity and ensures every environment follows the same configuration
standards.

---

## Key Features

### Unified API Resource Model

- **Kubernetes-Like Structure**  
  Uses `apiVersion`, `kind`, `metadata`, and `spec` fields for consistency across all Planton resources.
- **Protobuf Validations**  
  Each input field is validated against Protobuf schemas, catching misconfigurations before deployment.

### Comprehensive ECS Provisioning

- **Fargate & EC2 Support**  
  Choose from Fargate, Fargate Spot, or custom EC2 capacity providers to match workload requirements.
- **Built-In Observability**  
  Easily enable CloudWatch Container Insights for richer metrics and logs, streamlining troubleshooting.

### Secure and Flexible

- **ECS Exec**  
  Grant on-demand shell access to containers, ideal for debugging or limited maintenance tasks.
- **Pulumi & Terraform**  
  Switch between Pulumi or Terraform provisioning with no changes to your YAML manifest.

---

## Installation

1. **Planton CLI**  
   Make sure you have [Planton CLI](https://github.com/plantonhq/planton) installed and configured.
   This CLI handles resource validation, manifest processing, and integration with Pulumi or Terraform.

2. **AWS Credentials**  
   Ensure your AWS credentials are properly set up:
    - AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY environment variables, **or**
    - An AWS credential resource in Planton, referencing your AWS credential via the manifest.

3. **Pulumi or Terraform**  
   Depending on your chosen backend:
    - [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/) is required for Pulumi-based deployments.
    - [Terraform CLI](https://developer.hashicorp.com/terraform/downloads) is required for Terraform-based deployments.

---

## Usage

1. **Define a Manifest**  
   Create a YAML file that includes the **AwsEcsCluster** resource. For example:

   ```yaml
   apiVersion: aws.planton.dev/v1
   kind: AwsEcsCluster
   metadata:
     name: my-aws-ecs-cluster
     version:
       message: "Initial ECS cluster"
   spec:
     enable_container_insights: true
     capacity_providers:
       - "FARGATE"
       - "FARGATE_SPOT"
     enable_execute_command: false
   ```

2. **Validate (Optional)**  
   Use the Planton CLI to validate your manifest:
   ```bash
   planton validate --manifest awsecscluster.yaml
   ```

3. **Deploy**  
   Deploy using either Pulumi or Terraform through the same CLI:
   ```bash
   # Pulumi example
   planton pulumi up --manifest awsecscluster.yaml --stack myorg/dev
   
   # Terraform example
   planton terraform apply --manifest awsecscluster.yaml --stack myorg/dev
   ```

4. **Verify**  
   After successful provisioning, use the AWS Console or CLI to confirm the cluster’s status:
   ```bash
   aws ecs list-clusters
   ```

---

## API Resource Specification

Below are the key fields for the **AwsEcsCluster** resource. For a full listing, see the Protobuf definitions in the
Planton repository.

### `apiVersion`

- **Description**: Must be `"aws.planton.dev/v1"`.
- **Required**: Yes.

### `kind`

- **Description**: Must be `"AwsEcsCluster"`.
- **Required**: Yes.

### `metadata`

- **Description**: Standard metadata for naming and versioning.
- **Fields**:
    - **name**: A unique name (3–63 characters).
    - **version.message**: A mandatory string describing the version or change context.

### `spec.enable_container_insights`

- **Type**: `bool`
- **Default**: `true`
- **Description**: Toggles CloudWatch Container Insights for enhanced metrics.

### `spec.capacity_providers`

- **Type**: `repeated string`
- **Allowed Values**: `"FARGATE"`, `"FARGATE_SPOT"`
- **Description**: Defines which capacity providers the cluster uses.
    - Use a single `"FARGATE"` for basic serverless capacity.
    - Include `"FARGATE_SPOT"` for cost-saving spot capacity.

### `spec.enable_execute_command`

- **Type**: `bool`
- **Default**: `false`
- **Description**: Enables ECS Exec to allow secure shell access into running tasks. Useful for debugging.

---

## Customization and Extensibility

- **Hybrid Fargate & EC2**  
  Extend the `capacity_providers` list to include custom EC2 providers or scaling policies.
- **Additional Observability**  
  Integrate AWS X-Ray or third-party tools for deeper tracing and logging.
- **Advanced Security**  
  Configure IAM roles or custom VPC networking directly within your ECS tasks through Planton’s flexible YAML
  interface.
- **Multi-Cluster Strategies**  
  Deploy multiple ECS clusters across different environments (dev, staging, production) using the same standardized
  approach.

---

## Further Reading

- **[examples.md](./examples.md)**: Explore real-world usage scenarios for single-cluster and multi-capacity setups.
- **[Planton Guide](https://github.com/plantonhq/planton/blob/main/docs/Guide.md)**: Learn about
  the multi-cloud workflow, referencing the CLI commands and ecosystem best practices.
- **AWS Documentation**: [Amazon ECS](https://docs.aws.amazon.com/ecs) for in-depth ECS concepts.
