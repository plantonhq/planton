# Overview

The AWS ECR Repo API resource provides a consistent and streamlined interface for deploying and managing AWS Elastic Container Registry (ECR) repositories within our cloud infrastructure. By abstracting the complexities of ECR configurations and lifecycle policies, this resource allows you to define your container image storage effortlessly while ensuring security, compliance, and cost control across different environments.

## Why We Created This API Resource

Configuring AWS ECR repositories can be intricate due to numerous security considerations, lifecycle management requirements, and compliance needs. The AWS CLI and base IaC tools often default to insecure configurations (mutable tags, no scanning). To simplify this process and promote production-ready defaults, we developed this API resource. It enables you to:

- **Secure by Default**: Automatically enables image scanning and supports immutable tags to prevent production incidents.
- **Simplify Cost Control**: Provides high-level lifecycle policy configuration without dealing with complex JSON syntax.
- **Ensure Consistency**: Maintain uniform repository configurations across different environments and teams.
- **Enhance Productivity**: Reduce the time and effort required to configure ECR repositories, allowing you to focus on building and deploying applications.

## Key Features

### Environment Integration

- **Environment Info**: Seamlessly integrates with our environment management system to deploy ECR repositories within specific environments.
- **Infra Job Settings**: Supports custom stack-update settings for infrastructure-as-code deployments.

### AWS Credential Management

- **AWS Credential ID**: Utilizes specified AWS credentials to ensure secure and authorized deployments.

### Customizable Repository Specifications

#### Repository Identity

- **Repository Name**: Define unique repository names that can be namespaced (e.g., `my-org/service-name`).
- **Metadata Tags**: Automatically applies organization, environment, and resource tags for cost tracking and governance.

#### Security Configuration

- **Tag Mutability**: Four modes — `MUTABLE`, `IMMUTABLE`, and the exclusion-filtered pair (`IMMUTABLE_WITH_EXCLUSION` / `MUTABLE_WITH_EXCLUSION`) that freeze release tags while floating tags like `latest` stay movable, so `my-app:v1.0` always refers to exactly one image build.
- **Image Scanning**: Enable automatic vulnerability scanning when images are pushed (defaults to enabled for security).
- **Encryption**: Choose AWS-managed encryption (AES256), a customer-managed KMS key, or dual-layer KMS_DSSE for the strictest data-at-rest regimes. The whole encryption configuration is create-time.

#### Cost Control

- **Structured Lifecycle Rules**: The full ECR lifecycle rule model — priority-ordered rules selecting images by tag state (`tagged` with prefix or wildcard-pattern lists, `untagged`, `any`) and expiring them by age (`sinceImagePushed`) or count (`imageCountMoreThan`) — validated at authoring time so a policy AWS would reject never reaches an apply.

#### Access Control

- **Repository Policy**: A folded resource-based IAM policy for cross-account pulls and service principals (e.g. AWS Lambda pulling container images).

#### Safety Features

- **Force Delete**: Control whether repositories can be deleted when they contain images (defaults to false to prevent accidental data loss).

## Production Best Practices

This API resource encodes production best practices identified through industry research:

### Stability Guarantee

- **Immutable Tags**: Prevent overwrites of image tags, ensuring reproducible deployments and reliable rollbacks.

### Security Baseline

- **Scan on Push**: Automatically scan images for CVE vulnerabilities when pushed.
- **Encryption at Rest**: All images encrypted by default (AES256 or KMS).

### Cost Control

- **Lifecycle Policies**: Automated cleanup of old and untagged images prevents runaway storage costs from active CI/CD pipelines.

## Benefits

- **Simplified Deployment**: Abstracts the complexities of AWS ECR configurations, lifecycle policies, and scanning into an easy-to-use API.
- **Secure Defaults**: Production-ready configurations out of the box - image scanning enabled, encryption enforced, immutability supported.
- **Consistency**: Ensures all ECR repositories adhere to organizational standards for security, compliance, and cost control.
- **No JSON Complexity**: Define lifecycle policies as structured, validated rule objects instead of hand-written policy JSON — constraint violations fail at authoring time, not at apply.
- **Compliance Ready**: Support for KMS encryption with customer-managed keys for HIPAA, PCI-DSS, and other compliance regimes.
- **Cost Efficiency**: Automated lifecycle policies prevent storage costs from growing indefinitely due to CI/CD pipelines creating thousands of images.
- **Production Stability**: Immutable tags prevent the "worked yesterday" problem where production images get accidentally overwritten.

## Integration Patterns

### With Amazon ECS/Fargate

ECR repositories integrate seamlessly with ECS and Fargate for container deployments. The Task Execution Role requires ECR pull permissions.

### With Amazon EKS

For Kubernetes workloads, use IRSA (IAM Roles for Service Accounts) to grant pods transparent access to ECR without managing static credentials.

### With CI/CD Pipelines

CI/CD systems authenticate using `aws ecr get-login-password` and push images with unique tags (Git SHA recommended). Lifecycle policies automatically clean up old images.

## Security Considerations

- **IAM Permissions**: Separate pull and push permissions - runtime services need only pull access, CI/CD systems need push access.
- **Image Scanning**: Enable scan-on-push and configure EventBridge to alert on critical vulnerabilities.
- **Encryption**: Use KMS encryption when compliance requires demonstrable key lifecycle control and rotation policies.
- **Network Access**: For private subnet workloads, use VPC endpoints for ECR to avoid NAT Gateway costs and improve security.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
