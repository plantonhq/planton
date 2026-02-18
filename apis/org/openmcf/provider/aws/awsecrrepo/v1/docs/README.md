# AWS ECR Repository Deployment: The Evolution from Manual to Production-Ready

## Introduction

Amazon Elastic Container Registry (ECR) is the container image storage foundation for AWS-based applications, yet its deployment patterns reveal a striking evolution. While the AWS Management Console guides users toward secure defaults with mandatory choices for tag immutability, scanning, and encryption, the programmatic tooling that most teams rely on for automation has historically defaulted to **insecure configurations**. A naive `aws ecr create-repository --repository-name my-repo` creates a repository with mutable tags and no vulnerability scanning—a setup that is fundamentally not production-ready.

This document explores the landscape of ECR deployment methods, from manual console operations to fully declarative Infrastructure as Code (IaC), and examines why the industry has decisively moved toward high-level abstractions that bundle secure defaults with simplified policy management. Understanding this evolution is essential for making informed decisions about how to manage ECR repositories at scale.

## The Deployment Landscape: From Manual to Automated

### Level 0: The Manual Baseline (AWS Console)

The AWS Management Console workflow for creating an ECR repository is intentionally prescriptive:

1. Navigate to ECR in your desired region
2. Choose "Private repository"
3. Provide a repository name (can be namespaced: `project-a/nginx-web-app`)
4. **Make a mandatory selection** for Image Tag Immutability (Mutable vs. Immutable)
5. Configure Image Scanning settings
6. Select Encryption configuration (AES-256 or AWS KMS)

The console forces conscious security decisions upfront—a "golden path" by design. However, this approach has critical limitations:

- **No Auditability**: Changes are logged in CloudTrail but disconnected from code review processes
- **Configuration Drift**: Post-creation modifications via the console are the primary source of environment inconsistencies
- **No Reproducibility**: Not programmatically reproducible across accounts or regions
- **SCP Gotchas**: Service Control Policies can silently block even console operations, a non-obvious pitfall for teams in AWS Organizations

**Verdict**: Suitable only for initial exploration or one-off test repositories. Not viable for production or multi-environment workflows.

### Level 1: Scripting and CLI Automation (AWS CLI, Boto3)

The AWS CLI (`aws ecr create-repository`) and SDKs like Boto3 enable scripting but inherit a critical flaw: **insecure defaults**.

```bash
# This creates a repository with MUTABLE tags and no scanning
aws ecr create-repository --repository-name my-repo
```

To achieve production readiness, teams must explicitly override every setting:

```bash
aws ecr create-repository \
  --repository-name my-repo \
  --image-tag-mutability IMMUTABLE \
  --image-scanning-configuration scanOnPush=true \
  --encryption-configuration encryptionType=KMS,kmsKey=arn:aws:kms:...
```

**Additional Pain Points:**

- **No Idempotency**: Scripts must manually check if a repository exists before creation
- **API Throttling**: The `GetAuthorizationToken` action (required for `docker login`) is limited to 20 transactions per second sustained, with bursts to 200. Systems exceeding this receive `ThrottleException` errors, requiring exponential backoff logic

**Verdict**: A step toward automation, but brittle and error-prone. Lacks state management and forces users to build their own safety guardrails.

### Level 2: Infrastructure as Code - The Baseline Resources

This is where most teams operate. Tools like Terraform, CloudFormation, and Pulumi provide declarative, state-managed infrastructure. However, the base resources expose a significant **abstraction gap**, particularly for policy management.

#### CloudFormation: The Foundation

The `AWS::ECR::Repository` resource provides core repository properties but requires **embedded JSON strings** for policies:

- `LifecyclePolicyText`: Raw JSON for lifecycle rules
- `RepositoryPolicyText`: Raw JSON for cross-account or service access

This is error-prone, difficult to validate, and lacks readability.

#### Terraform: The Same Gap

The `aws_ecr_repository` resource is nearly identical to CloudFormation, with lifecycle and repository policies split into **separate resources** (`aws_ecr_lifecycle_policy`, `aws_ecr_repository_policy`) that also require raw JSON strings.

The Terraform community's response to this pain point was the creation of the **terraform-aws-ecr module**, which provides the abstraction users actually wanted: simple inputs that generate complex policy JSON behind the scenes.

#### Pulumi (Classic Provider): The 1:1 Mapping

Pulumi's `aws.ecr.Repository` resource is a direct port from Terraform, inheriting the same limitations. Separate resources, embedded JSON policies, no high-level abstractions.

**Verdict**: These "Level 1.5" resources work but force users to manage the most complex aspects (lifecycle policies for cost control, repository policies for integrations) as raw, verbose JSON. The market signal is clear: **users want better abstractions**.

### Level 3: Production-Ready Abstractions (AWS CDK, Pulumi AWSX)

These tools represent a paradigm shift: treating ECR as a **bundled construct** rather than atomic resources.

#### AWS CDK: The Model of Abstraction

The `aws_ecr.Repository` construct is a "Level 2" abstraction that solves the JSON pain point elegantly:

```python
from aws_cdk import aws_ecr as ecr

repo = ecr.Repository(self, "MyRepo",
    repository_name="my-app",
    lifecycle_rules=[
        ecr.LifecycleRule(
            max_image_age=cdk.Duration.days(14),
            tag_status=ecr.TagStatus.UNTAGGED
        ),
        ecr.LifecycleRule(
            max_image_count=30,
            tag_status=ecr.TagStatus.ANY
        )
    ]
)

# High-level helper methods
repo.grant_pull(other_account)
repo.add_to_resource_policy(statement)
```

The CDK accepts **native, typed objects** instead of JSON strings and provides helper methods for the 80% use case (granting pull access, adding policy statements).

#### Pulumi AWSX: The Component Resource

The `awsx.ecr.Repository` component resource bundles the base repository with its lifecycle policy and provides **sensible defaults**:

- Automatically expires untagged images after one day
- Accepts simplified `lifecyclePolicy` inputs
- Abstracts away the JSON complexity

**Verdict**: This is the developer experience the industry has converged on. These abstractions don't just wrap the API—they encode best practices and make the secure path the easy path.

### Level 4: Kubernetes-Native Management (Crossplane, ACK)

For teams operating in a GitOps model, ECR repositories can be managed as native Kubernetes resources:

- **Crossplane** (via Upbound AWS provider): Defines a `Repository.ecr.aws.upbound.io` Kind
- **AWS Controllers for Kubernetes (ACK)**: Official AWS project with a dedicated `ecr-controller`

This approach enables declarative repository management via ArgoCD or Flux. For runtime operations, a separate class of operators (like `ecr-secret-operator`) solves the 12-hour token expiration problem by automatically refreshing Kubernetes `imagePullSecrets`.

**Verdict**: Essential for Kubernetes-centric organizations, but adds operational complexity (managing the operator itself). Most valuable when combined with IRSA (IAM Roles for Service Accounts) on EKS.

## Production Essentials: The Non-Negotiables

A production-ready ECR repository is defined by a set of configurations that ensure **stability, security, and cost control**.

### 1. Tag Immutability: The Stability Guarantee

**Setting:** `image_tag_mutability = IMMUTABLE`

This is the single most important production configuration. Immutable tags prevent overwrites, ensuring that a tag like `my-app:prod` or `my-app:a1b2c3d` always refers to exactly one, unique image build.

- **Mutable Tags (ANTI-PATTERN)**: Allows `my-app:prod` to be overwritten, breaking traceability and making rollbacks unreliable
- **Immutable Tags (BEST PRACTICE)**: Attempting to push to an existing tag returns `ImageTagAlreadyExistsException`, enforcing the cultural practice that new code = new tag

Immutability is non-negotiable for production. It guarantees that your deployments are reproducible and traceable.

### 2. Image Scanning: Shift-Left Security

**Setting:** `scan_on_push = true`

Enabling scan-on-push triggers immediate vulnerability scanning when an image is pushed.

- **Basic Scanning**: Uses the CVEs database for known vulnerabilities
- **Enhanced Scanning**: Integrates with Amazon Inspector for continuous, automated scanning of OS and programming language packages

Scanning alone is not enough—mature workflows subscribe to ECR scan completion events in EventBridge and automate responses, such as blocking CI/CD pipelines from deploying images with critical vulnerabilities.

### 3. Encryption: The Default vs. The Compliance Switch

**Default:** `encryption_type = AES256` (AWS-managed keys)  
**Compliance:** `encryption_type = KMS` (customer-managed keys)

AES256 is transparent, free, and sufficient for most use cases. It encrypts images at rest using Amazon S3-managed encryption.

KMS provides an **auditable control plane**. This is not "more secure" encryption—it's about demonstrating **key lifecycle control** for compliance regimes like HIPAA or PCI-DSS. Regulators require evidence that you can manage key rotation policies and revoke access.

### 4. Lifecycle Policies: Automated Cost Control

Without lifecycle policies, ECR storage costs grow indefinitely. An active CI/CD pipeline can generate thousands of untagged or temporary images.

**The 80% Use Case** is covered by two rules:

1. **Expire Untagged Images**: Remove images with no tags after 14 days (cleans up intermediate build layers)
2. **Expire Old Tagged Images**: Keep only the last 30 images, expiring all older ones

Lifecycle policies are the single biggest pain point in the base IaC resources (CloudFormation, Terraform, Pulumi classic) because they require complex, verbose JSON. High-level abstractions (CDK, Pulumi AWSX) solve this by providing simple fields like `expire_untagged_after_days` or `max_image_count`.

### 5. Repository Policies: Cross-Account and Service Access

By default, ECR repositories are private to the AWS account. Repository policies are **IAM resource-based policies** that grant access to:

- **Other AWS Accounts**: Allow a production account to pull images from a shared tooling account
- **AWS Service Principals**: Grant Lambda (`lambda.amazonaws.com`), ECS (`ecs-tasks.amazonaws.com`), or other services pull permissions
- **CI/CD Roles**: Grant specific IAM roles push permissions (`ecr:PutImage`, `ecr:InitiateLayerUpload`)

**Anti-Pattern Alert**: Using `Principal: "*"` in a repository policy is a major security risk unless combined with very strict IAM-level controls.

## Common Anti-Patterns to Avoid

| Anti-Pattern | Impact | Fix |
|-------------|--------|-----|
| **No Lifecycle Policy** | Uncontrolled storage costs, thousands of orphaned images | Implement at minimum: expire untagged after 14 days, keep last 30 tagged |
| **Mutable Tags in Production** | Unreliable deployments, impossible rollbacks, lost traceability | Set `image_tag_mutability = IMMUTABLE` |
| **No Scan-on-Push** | Security vulnerabilities go undetected until runtime | Enable `scan_on_push = true` |
| **Overly-Broad Repository Policy** | Unintended access, potential data exposure | Use least-privilege IAM policies with specific principals |
| **Ignoring Token Expiration (EKS)** | `ImagePullBackOff` errors every 12 hours | Use IRSA on EKS or deploy `ecr-secret-operator` |

## The 80/20 Configuration Principle

Just as APIs should focus on the 20% of configuration that 80% of users need, ECR deployment should prioritize essential settings with secure defaults.

### The Essential 80%

1. **region**: The AWS region where the repository will be created
2. **repository_name**: The repository name (unique within account and region)
3. **image_tag_mutability**: Default to `IMMUTABLE` for production stability
4. **scan_on_push**: Default to `true` for security
5. **encryption_type**: Default to `AES256` (sufficient for most use cases)

### The Advanced 20%

1. **Lifecycle Policies**: Nearly universal need, but complex configuration (prime target for abstraction)
2. **Repository Policies**: Required for any cross-account or service integration
3. **KMS Key ARN**: The "compliance switch" for regulated industries
4. **Replication**: Rare, advanced use case for multi-region disaster recovery

### Example: Production Repository Configuration

Here's what a production ECR repository configuration looks like with best practices:

```hcl
# Terraform with secure defaults
resource "aws_ecr_repository" "prod" {
  name                 = "project-a/prod-service"
  image_tag_mutability = "IMMUTABLE"  # Stability guarantee
  
  image_scanning_configuration {
    scan_on_push = true  # Security baseline
  }
  
  encryption_configuration {
    encryption_type = "KMS"  # Compliance requirement
    kms_key        = "arn:aws:kms:us-east-1:123456789012:key/prod-key"
  }
}

# Lifecycle policy for cost control
resource "aws_ecr_lifecycle_policy" "prod_policy" {
  repository = aws_ecr_repository.prod.name
  
  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 50 production images"
        selection = {
          tagStatus     = "tagged"
          countType     = "imageCountMoreThan"
          countNumber   = 50
        }
        action = { type = "expire" }
      },
      {
        rulePriority = 2
        description  = "Expire untagged images after 1 day"
        selection = {
          tagStatus   = "untagged"
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 1
        }
        action = { type = "expire" }
      }
    ]
  })
}
```

## Integration Patterns

### With Amazon ECS and Fargate

Integration is seamless. Set the ECS Task Definition's `image` property to the ECR repository URL:

```
<account_id>.dkr.ecr.<region>.amazonaws.com/<repo_name>:<tag>
```

The **Task Execution Role** (not the Task Role) must have ECR pull permissions:
- `ecr:GetAuthorizationToken`
- `ecr:BatchGetImage`
- `ecr:GetDownloadUrlForLayer`

### With Amazon EKS (Kubernetes)

Integration presents an authentication challenge: ECR tokens expire every 12 hours.

**Solution 1 (Recommended): IRSA (IAM Roles for Service Accounts)**
- Associate the pod's Service Account with an IAM role that has ECR pull permissions
- The kubelet uses an ECR credential provider to automatically fetch credentials
- This is the best practice for EKS—fully transparent and automatic

**Solution 2 (For Non-EKS or Legacy Clusters): ECR Secret Operator**
- Deploy an operator like `ecr-secret-operator` that runs in the cluster
- Continuously refreshes the ECR auth token and patches the Kubernetes `imagePullSecret` before expiration
- Prevents `ImagePullBackOff` errors

### With CI/CD Pipelines

The typical workflow:

1. **Authenticate**: Use `aws ecr get-login-password | docker login` to authenticate the Docker client
2. **Build**: Run `docker build` to create the image
3. **Tag**: Tag with a unique identifier (Git SHA preferred for immutability)
4. **Push**: Run `docker push` to upload to ECR

**Best Practice**: Use OIDC to grant CI/CD systems temporary, role-based credentials instead of static IAM access keys.

### Cost Optimization

- **Storage**: Managed exclusively through lifecycle policies (expire untagged, limit image count)
- **Data Transfer**: Same-region pulls (ECS/EKS to ECR) are free; cross-region or internet pulls incur standard AWS data transfer costs
- **VPC Endpoints**: For workloads in private subnets, VPC endpoints eliminate NAT Gateway data processing costs, which can be significant at scale

## Security and Compliance

### IAM: Least Privilege by Role

**Pull Role (Runtime Services):**
```json
{
  "Effect": "Allow",
  "Action": [
    "ecr:GetAuthorizationToken",
    "ecr:BatchGetImage",
    "ecr:GetDownloadUrlForLayer"
  ],
  "Resource": "*"
}
```

**Push Role (CI/CD):**
```json
{
  "Effect": "Allow",
  "Action": [
    "ecr:GetAuthorizationToken",
    "ecr:InitiateLayerUpload",
    "ecr:UploadLayerPart",
    "ecr:CompleteLayerUpload",
    "ecr:PutImage"
  ],
  "Resource": "arn:aws:ecr:us-east-1:123456789012:repository/my-app"
}
```

### Network Security: VPC Endpoints

For workloads in private subnets, create three VPC endpoints to keep all ECR traffic on the AWS private network:

1. `com.amazonaws.<region>.ecr.api` (Interface): ECR API calls
2. `com.amazonaws.<region>.ecr.dkr` (Interface): Docker data plane (push/pull layers)
3. `com.amazonaws.<region>.s3` (Gateway): ECR stores layers in S3

This improves security, reduces latency, and eliminates NAT Gateway costs.

### Audit and Logging

All ECR API calls are logged in AWS CloudTrail, providing:
- **Security**: Alerts on high-risk actions (DeleteRepository, policy changes)
- **Compliance**: Immutable audit trail of image pushes
- **Debugging**: Identification of which principal pushed a specific image

## OpenMCF's Approach

The `AwsEcrRepo` API in OpenMCF follows the principle of **secure defaults with escape hatches for advanced use cases**.

### What's Included

- **Region**: The AWS region where the ECR repository will be created (required)
- **Repository Name**: The name of the ECR repository (required, with validation for length and format)
- **Image Immutability**: Boolean flag to enforce immutable tags (recommended for production)
- **Encryption**: Defaults to AES256, with KMS option via `kms_key_id` for compliance scenarios
- **Force Delete**: Safety flag to prevent accidental deletion of repositories with images

### Philosophy

OpenMCF's API design prioritizes:

1. **Minimal Required Configuration**: Only `region` and `repository_name` are mandatory
2. **Secure Defaults**: Encryption enabled by default (AES256)
3. **Flexibility for Compliance**: KMS encryption available when needed
4. **Safety First**: Force delete disabled by default to prevent data loss

The API intentionally stays focused on repository provisioning fundamentals. For advanced features like lifecycle policies, repository policies, and scanning configuration, OpenMCF follows the philosophy of composability—these can be managed through companion configurations or post-deployment automation.

### When to Use OpenMCF's AwsEcrRepo

- **Multi-cloud teams**: Consistent API patterns across AWS, GCP, Azure
- **GitOps workflows**: Declarative, protobuf-defined infrastructure
- **Teams prioritizing simplicity**: Minimal required configuration with secure defaults
- **Compliance-first organizations**: Built-in support for KMS encryption

## Conclusion

The evolution of ECR deployment methods reveals a consistent industry trend: the shift from low-level, atomic resources to high-level, opinionated abstractions. While the AWS Console enforces secure choices, the programmatic tools that teams rely on for automation have historically defaulted to insecure configurations.

The market has responded with solutions like AWS CDK, Pulumi AWSX, and community modules like terraform-aws-ecr—all converging on the same pattern: **bundle repository configuration with policy management, provide secure defaults, and abstract away JSON complexity**.

For production ECR deployments, the non-negotiables are clear:
- **Immutable tags** for stability and traceability
- **Scan-on-push** for shift-left security
- **Lifecycle policies** for cost control
- **Appropriate encryption** (AES256 for most, KMS for compliance)

OpenMCF's `AwsEcrRepo` API embraces these principles, providing a minimal, secure-by-default interface that works seamlessly in multi-cloud, GitOps-driven environments. The true power lies not in the number of configuration options, but in making the right choices the easy choices.

