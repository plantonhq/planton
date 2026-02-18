# AWS ECS Cluster: Production-Grade Container Orchestration Without the Kubernetes Complexity

## Introduction

A persistent debate in cloud infrastructure circles is whether teams need the full power and complexity of Kubernetes for every containerized workload. The answer, grounded in production evidence, is a resounding "no"—for teams all-in on AWS, Amazon Elastic Container Service (ECS) offers a compelling alternative that trades Kubernetes' ecosystem flexibility for powerful simplicity and deep AWS integration.

At the heart of ECS is the **ECS cluster**, a logical grouping or namespace that isolates containerized workloads. Unlike traditional infrastructure, the cluster itself is remarkably minimal—it's not a fleet of servers you manage, but rather a lightweight orchestration boundary that groups three critical elements:

1. **Compute capacity**: Fargate (serverless) or EC2 (self-managed) capacity providers
2. **Cluster-wide settings**: CloudWatch Container Insights, ECS Exec auditing
3. **Deployment target**: The logical destination for services and tasks

This document explores the landscape of ECS cluster deployment methods, from manual console operations through production-grade Infrastructure as Code (IaC) tooling, and explains why OpenMCF provides a streamlined, opinionated API that honors ECS's core value proposition: simplicity without sacrificing production-readiness.

## The Strategic Choice: ECS vs. EKS

Before diving into deployment methods, it's essential to understand the strategic decision point that precedes ECS adoption. AWS offers two fully managed container orchestration services: **Amazon ECS** and **Amazon Elastic Kubernetes Service (EKS)**. The choice between them is fundamentally about priorities.

**Amazon ECS** is AWS's opinionated, proprietary solution. It's designed for teams who want:
- Lower operational overhead with zero control plane costs
- Deep, seamless integration with AWS services (ALB, Route 53, IAM, CloudWatch)
- A simpler learning curve for teams with AWS skills but not Kubernetes expertise
- Focus on application development rather than infrastructure management

**Amazon EKS** provides a fully managed, conformant Kubernetes control plane. It's chosen for:
- Access to the vibrant Kubernetes ecosystem and tooling
- Multi-cloud or hybrid-cloud portability strategies
- Consistent Kubernetes API across environments
- Teams with existing Kubernetes expertise

The trade-off is clear: **simplicity versus flexibility**. ECS is for teams committed to AWS who value operational simplicity. EKS is for teams needing the portability and ecosystem of Kubernetes, accepting higher complexity and per-cluster hourly control plane fees.

**If your organization has chosen ECS**, you've already prioritized powerful simplicity and AWS-native integration. Your infrastructure tooling—including your IaC API—must honor and enhance that decision, not undermine it with unnecessary complexity.

## The Deployment Methods Landscape

Provisioning an ECS cluster spans a maturity spectrum from manual operations to fully declarative, abstracted control planes.

### Level 0: Manual Console Provisioning (The Anti-Pattern)

The AWS Management Console is valuable for learning and debugging, but it's an anti-pattern for production infrastructure. Manual provisioning is:
- **Non-repeatable**: Configuration exists only in someone's memory or incomplete screenshots
- **Error-prone**: Common mistakes include forgetting to enable CloudWatch Logs (causing task failures), neglecting to auto-assign public IPs in public subnets (preventing image pulls from ECR), and storing secrets as plain-text environment variables (severe security vulnerability)
- **Unauditable**: No version control, no peer review, no automated compliance checks

**Verdict**: Acceptable for initial exploration, unacceptable for any production workload.

### Level 1: CLI and SDKs (Imperative Building Blocks)

The AWS CLI and SDKs (like Boto3 for Python) provide programmatic access to AWS APIs. They're the foundational building blocks but lack state management. A single high-level action—deploying a containerized application—requires orchestrating dozens of low-level API calls across ECS, IAM, VPC, and CloudWatch. This is precisely the complexity that declarative IaC tools are designed to eliminate.

**Verdict**: Essential for automation scripts and CI/CD pipelines, but insufficient as the primary infrastructure provisioning method.

### Level 2: Configuration Management (Ansible)

Ansible represents a middle ground with declarative modules for ECS resources (`community.aws.ecs_cluster`, `community.aws.ecs_service`). While an improvement over pure scripting, its state management is generally weaker than dedicated IaC tools. In complex scenarios, Ansible playbooks can degrade into executing raw shell commands, losing their declarative benefits.

**Verdict**: Viable for simpler deployments, but dedicated IaC tools provide stronger state management for production infrastructure.

### Level 3: Declarative IaC Tooling (The Production Standard)

This category represents the **de facto standard** for production infrastructure. These tools treat infrastructure as code, enabling repeatable, version-controlled, and auditable deployments.

#### **AWS CloudFormation**
The native AWS IaC solution. Users define resources in YAML or JSON templates, and AWS manages resource state in logical units called "stacks." The `AWS::ECS::Cluster` resource is authoritative but low-level, exposing properties like `CapacityProviders`, `ClusterSettings`, and `Configuration.ExecuteCommandConfiguration`.

**Strengths**:
- No third-party tooling required
- AWS-managed state (no state file to secure)
- Deep integration with AWS services

**Limitations**:
- Verbose, low-level syntax
- Slower deployment and drift detection compared to alternatives
- Less flexible for multi-cloud scenarios

#### **AWS Cloud Development Kit (CDK)**
A developer-centric framework that allows infrastructure definition in familiar programming languages (TypeScript, Python, Java, Go). CDK code synthesizes into CloudFormation templates. High-level "constructs" like `ecs.Cluster` encapsulate best practices with smart defaults.

**Strengths**:
- Developer-friendly (real programming languages)
- High-level abstractions reduce boilerplate
- AWS-managed state (synthesizes to CloudFormation)

**Limitations**:
- AWS-only (not multi-cloud)
- Learning curve for the construct library
- Requires understanding of both the language and CDK abstractions

#### **Terraform**
The multi-cloud market leader from HashiCorp, using the HashiCorp Configuration Language (HCL). Terraform manages state in a JSON state file, typically stored in remote backends (S3 with DynamoDB locking). The power of Terraform lies in its **community-vetted modules** like `terraform-aws-modules/ecs` and the "ECS Blueprints," which provide opinionated, production-ready patterns.

**Strengths**:
- Multi-cloud flexibility
- Mature ecosystem and strong community
- Excellent drift detection and state management
- Modular, reusable patterns

**Limitations**:
- Self-managed state file (operational overhead, security considerations)
- HCL learning curve for developers unfamiliar with declarative languages

#### **Pulumi**
Combines CDK's approach (using real programming languages) with Terraform's approach (self-managed state). Pulumi's AWS provider is largely a wrapper around the Terraform AWS provider, providing a familiar resource model.

**Strengths**:
- Multi-cloud with real programming languages
- Granular state management like Terraform
- Appeals to full-stack developers

**Limitations**:
- Smaller community than Terraform
- Self-managed state file
- Abstraction layer over Terraform provider can occasionally introduce friction

**Comparison Table: IaC Tool Selection Matrix**

| Tool | Language/Format | State Management | Abstraction Level | Ideal Use Case |
|------|----------------|------------------|-------------------|----------------|
| **CloudFormation** | YAML / JSON | AWS-Managed (Stack) | Low (Resource-level) | AWS-native "purists"; environments prohibiting third-party tooling |
| **Terraform** | HCL (Declarative) | Self-Managed (State File) | Low (Resource) / High (Module) | Platform teams; multi-cloud environments; mature ecosystem preference |
| **AWS CDK** | TypeScript, Python, etc. | AWS-Managed (Stack) | High (Construct-level) | Application developers "all-in" on AWS; preference for code over config |
| **Pulumi** | TypeScript, Python, Go | Self-Managed (State File) | High (Component-level) | Full-stack developers; multi-cloud with general-purpose languages |

### Level 4: Higher-Level Abstractions

Beyond provisioning individual resources, some tools manage entire application lifecycles.

#### **AWS Copilot**
An opinionated, application-centric CLI specifically for ECS and Fargate. Copilot uses a simple `manifest.yml` file to define services (e.g., "Load Balanced Web Service") and handles everything: building the image, pushing to ECR, provisioning the cluster, service, and load balancer. High-level fields like `exec: true` (for ECS Exec) and `count: { spot: true }` (for Fargate Spot) demonstrate the 80/20 abstraction principle in action.

**Verdict**: Excellent for application developers who want to deploy without deep infrastructure expertise. Less suitable for platform teams needing granular control.

#### **Crossplane**
An open-source, Kubernetes-native control plane that extends the Kubernetes API to manage external cloud resources using Custom Resource Definitions (CRDs). Interestingly, Crossplane is almost exclusively used to provision **EKS** clusters within the Kubernetes community—not ECS—reflecting the cultural divide between Kubernetes-native and AWS-native ecosystems.

**Verdict**: Powerful for Kubernetes-first organizations managing multi-cloud infrastructure via GitOps, but not a natural fit for ECS-centric workflows.

## Production Essentials: The Features That Matter

A provisioned cluster is not production-ready until it's observable, debuggable, cost-optimized, and secure. These operational capabilities are essential considerations for API design.

### CloudWatch Container Insights: Observability at the Cluster Level

**Function**: Container Insights provides comprehensive monitoring, collecting metrics at the cluster, service, and task levels—CPU, memory, disk, network utilization—and presenting them in automatic CloudWatch dashboards.

**Configuration**: Enabled at the *cluster level* by setting `ClusterSettings` to `name="containerInsights"` and `value="enabled"`.

**Cost vs. Value**: Container Insights is **not free**. It ingests high-cardinality performance data as log events into CloudWatch, billed at standard CloudWatch Logs ingestion and storage rates. While historically an "all-or-nothing" feature, customization is now possible via AWS Distro for OpenTelemetry (ADOT) to filter metrics and reduce costs.

**Abstraction Opportunity**: The underlying API mechanism (a specific `name/value` pair) is obscure, but the user's intent ("I want monitoring") is simple. A production-focused IaC API should provide a boolean: `enable_container_insights: true`.

### Fargate Capacity Providers: The Cost-Optimization Lever

This is the most important lever for Fargate cost optimization. **Fargate Spot** offers discounts of **up to 70%** compared to on-demand Fargate pricing. The trade-off: Fargate Spot tasks run on spare AWS capacity and can be interrupted with a two-minute warning when AWS needs that capacity back.

The production pattern for safely using Spot is the **base and weight strategy**, defined in the cluster's `defaultCapacityProviderStrategy`:

- **base** (integer): Defines a minimum number of tasks to run on a specified capacity provider (typically `FARGATE`). This guarantees stability. For example, `base: 1` on `FARGATE` ensures at least one task is always on stable, on-demand capacity.
- **weight** (integer): Defines the relative proportion for all tasks beyond the base. For example, a strategy of `FARGATE (base: 1, weight: 1)` and `FARGATE_SPOT (weight: 4)` means that for every 5 new tasks scaled, one is on-demand and four are Spot.

**A Fargate cluster that only supports on-demand FARGATE is not fully production-ready from a cost perspective.** Fargate Spot configuration must be treated as an essential feature, not an "advanced" option.

### ECS Exec: Secure Debugging Without SSH

ECS Exec provides secure shell access or command execution directly inside running containers, replacing the anti-pattern of SSH. This is critical for "break-glass" debugging in production.

Enabling ECS Exec requires configuration in **two distinct places**:

1. **At the Service Level**: The ECS service must be created with `--enable-execute-command` set to `true`. This *enables* the feature for that service's tasks.
2. **At the Cluster Level (for Auditing)**: The cluster can be configured with an `executeCommandConfiguration` block. This defines *auditing* for Exec sessions—whether commands and output are logged to CloudWatch Logs or S3, and if logging is encrypted with a KMS key.

**API Design Implication**: The `AwsEcsCluster` resource should expose the cluster-level `execute_command_configuration` (for auditing), while a corresponding `AwsEcsService` resource would expose the service-level `enable_execute_command` flag.

### Service Discovery with AWS Cloud Map

For microservices architectures, tasks need to discover each other. ECS integrates with AWS Cloud Map for service discovery. When enabled, ECS automatically registers a task's private IP with a Cloud Map namespace, making it discoverable via stable internal DNS (e.g., `payment-service.prod.local`).

This is configured at the *service level* using the `serviceRegistries` block. However, all services within a cluster typically share a common Cloud Map namespace. A production-ready IaC tool might offer a convenience feature on the *cluster* resource (e.g., `create_default_cloudmap_namespace: true`) to provision this shared namespace automatically.

### Logging: awslogs vs. awsfirelens

Logging is configured in the **task definition**, not the cluster, but the strategy impacts the entire deployment.

- **awslogs**: The default log driver. Ships container stdout/stderr directly to a CloudWatch Log group. This is the 80% use case for straightforward applications.
- **awsfirelens**: An advanced log router using a sidecar container (Fluentd or Fluent Bit) to intercept logs and route them to various destinations (Amazon OpenSearch, S3, Splunk, Datadog). Enables centralized logging and advanced filtering.

### Secrets Management: Never Plain-Text

The primary security anti-pattern is storing credentials in plain-text environment variables. The production standard: store all sensitive data (API keys, database passwords) in **AWS Secrets Manager** or **AWS Systems Manager (SSM) Parameter Store**.

Secrets are injected securely into containers at runtime, configured in the task definition using the `secrets` or `valueFrom` parameter. The cluster's Task Execution Role must be granted IAM permissions (`secretsmanager:GetSecretValue` or `ssm:GetParameters`) for this to work.

### Networking and Security Best Practices

**Network Mode**: Fargate tasks **require** the `awsvpc` network mode. This is a major security benefit—each task gets a dedicated Elastic Network Interface (ENI) with its own private IP and, critically, its own **security group**, enabling granular, task-to-task network policies.

**IAM Roles**: Two distinct roles are required:
1. **Task Execution Role**: Used by the ECS agent to *start* the task (pull ECR images, write logs)
2. **Task Role**: Used by the *application code* running inside the container (call AWS services like S3, DynamoDB)

**Common Anti-Patterns**:
- Overly-permissive roles (single "god role" for all tasks)
- Overly-permissive security groups (allowing all traffic `0.0.0.0/0`)
- On EC2 launch type only: ECScape risk (cross-task credential exposure when tasks with different privilege levels share the same host; Fargate avoids this by design)

## OpenMCF's Design: The 80/20 Principle in Action

The AWS CloudFormation `AWS::ECS::Cluster` resource is already minimal. The 80/20 rule for an ECS API is not about *removing* fields, but about *abstracting* and *elevating* the few key settings that transform a basic cluster into a production-ready one.

### Current OpenMCF API

The existing `AwsEcsClusterSpec` protobuf includes:

1. **`enable_container_insights`** (bool): Simple boolean abstraction for enabling CloudWatch monitoring
2. **`capacity_providers`** (repeated string): List of capacity providers (`FARGATE`, `FARGATE_SPOT`)
3. **`enable_execute_command`** (bool): Controls whether ECS Exec is allowed on tasks

This is a strong foundation that captures the most essential production features. However, based on the research findings, there's a **critical gap**: the **default capacity provider strategy** (the base/weight configuration) is missing.

### The Missing Piece: Cost-Optimization Strategy

The `capacity_providers` field specifies *which* capacity providers are available (e.g., `["FARGATE", "FARGATE_SPOT"]`), but it doesn't define *how* they should be used. Without the base/weight strategy, users cannot implement the production pattern of "guarantee one on-demand task for stability, scale the rest on Spot for cost savings."

**Recommendation**: Expand the API to include `default_capacity_provider_strategy` as a structured field. This directly maps to the CloudFormation `DefaultCapacityProviderStrategy` property and is the core lever for Fargate cost optimization.

### Enhanced API Design Recommendation

```protobuf
message AwsEcsClusterSpec {
  // region is the AWS region where the ECS cluster will be created.
  string region = 1;

  // enable_container_insights determines whether to enable CloudWatch
  // Container Insights for this cluster. Highly recommended for
  // production monitoring (incurs CloudWatch costs).
  bool enable_container_insights = 2;

  // capacity_providers is a list of capacity providers attached
  // to this cluster. For Fargate-only: ["FARGATE"] or
  // ["FARGATE", "FARGATE_SPOT"] for Spot cost optimization.
  repeated string capacity_providers = 3;

  // default_capacity_provider_strategy defines the base/weight
  // distribution for tasks across capacity providers. This is the
  // primary cost-optimization lever for Fargate workloads.
  repeated CapacityProviderStrategy default_capacity_provider_strategy = 3;

  // execute_command_configuration defines cluster-level auditing
  // settings for ECS Exec. This is the auditing layer; the service
  // must separately enable exec with enable_execute_command.
  ExecConfiguration execute_command_configuration = 4;
}

// CapacityProviderStrategy defines the base/weight model for
// distributing tasks across capacity providers.
message CapacityProviderStrategy {
  // The capacity provider name (e.g., "FARGATE" or "FARGATE_SPOT").
  string provider = 1;

  // The minimum number of tasks to run on this provider.
  // Typically used with FARGATE to guarantee stability.
  int32 base = 2;

  // The relative weight for scaling tasks beyond the base.
  // Example: FARGATE (weight: 1) + FARGATE_SPOT (weight: 4)
  // results in 20% on-demand, 80% Spot for scaled tasks.
  int32 weight = 3;
}

// ExecConfiguration defines cluster-level auditing for ECS Exec.
message ExecConfiguration {
  enum Logging {
    // AWS-managed defaults (CloudWatch or S3 if enabled on account).
    DEFAULT = 0;
    // Explicitly disable exec auditing.
    NONE = 1;
    // Override defaults with custom log_configuration below.
    OVERRIDE = 2;
  }

  // The logging behavior for Exec sessions.
  Logging logging = 1;

  // Custom log configuration (only used if logging is OVERRIDE).
  ExecLogConfiguration log_configuration = 2;
}

// ExecLogConfiguration specifies custom destinations for Exec audit logs.
message ExecLogConfiguration {
  // CloudWatch log group to send logs to.
  string cloud_watch_log_group = 1;

  // S3 bucket to send logs to.
  string s3_bucket = 2;

  // S3 key prefix for log files.
  string s3_key_prefix = 3;

  // (Optional) KMS key for S3 encryption.
  string s3_encryption_kms_key_id = 4;

  // (Optional) KMS key for CloudWatch encryption.
  string cloud_watch_encryption_kms_key_id = 5;
}
```

### Configuration Examples

#### Example 1: Development Cluster (Minimal)

**User Intent**: "I want a Fargate cluster to test my app. No Spot, minimal monitoring."

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: dev-cluster
spec:
  region: us-east-1
  capacity_providers:
    - FARGATE
  enable_container_insights: false
  execute_command_configuration:
    logging: DEFAULT  # Enable exec with AWS-managed defaults
```

#### Example 2: Production Cluster (Cost-Optimized and Audited)

**User Intent**: "Production cluster, cost-optimized with Spot, fully monitored, all Exec commands audited to S3."

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcsCluster
metadata:
  name: prod-cluster
spec:
  region: us-east-1
  capacity_providers:
    - FARGATE
    - FARGATE_SPOT
  default_capacity_provider_strategy:
    - provider: FARGATE
      base: 1       # Guarantee 1 on-demand task for stability
      weight: 1     # 1 part on-demand for scaling
    - provider: FARGATE_SPOT
      weight: 4     # 4 parts Spot for cost (80% of scaled tasks)
  enable_container_insights: true
  execute_command_configuration:
    logging: OVERRIDE
    log_configuration:
      cloud_watch_log_group: "/aws/ecs/prod-exec-logs"
      s3_bucket: "prod-exec-audit-logs"
      s3_key_prefix: "ecs-exec"
```

## Service Deployment and Lifecycle Management

While the cluster is foundational infrastructure, the **ECS Service** manages application lifecycles. The cluster's design directly enables these deployment patterns.

### Deploying Services to the Cluster

An ECS Service launches and maintains a `desiredCount` of tasks from a specific `taskDefinition` on a cluster. When deployed, the service uses the cluster's `defaultCapacityProviderStrategy` (unless overridden at the service level) to determine task placement.

### Rolling Updates and the Deployment Circuit Breaker

The default deployment type is the **rolling update**, which gradually replaces tasks running the old task definition with tasks running the new one.

The critical safety feature is the **Deployment Circuit Breaker**, a service-level configuration that monitors new task health during deployment. If new tasks consistently fail to start (bad image, failed health checks, resource constraints), the circuit breaker "trips," immediately stopping the deployment. If configured with `rollback: true`, it automatically rolls back to the last known-good deployment, preventing a bad deployment from taking down the entire service.

### Blue/Green Deployments with CodeDeploy

A **blue/green deployment** is a more advanced, zero-downtime strategy managed by AWS CodeDeploy:

1. Provision a full new set of "green" tasks alongside existing "blue" production tasks
2. Validate the green deployment, often shifting a small percentage of test traffic
3. Once validated, shift 100% of production traffic from blue to green *instantly*
4. Keep the old "blue" environment running for a "bake time" (e.g., 1 hour) to allow immediate rollback

This provides zero downtime and instant rollbacks, at the cost of running double infrastructure during the deployment window.

### Integration with CI/CD Pipelines

The ECS cluster is static infrastructure, typically provisioned by IaC tools. The service and task definition are dynamic application components updated by CI/CD pipelines (GitHub Actions, AWS CodePipeline):

1. **Build**: Code change triggers the pipeline, builds new Docker image
2. **Push**: Image is tagged (e.g., git commit SHA) and pushed to Amazon ECR
3. **Deploy**: Pipeline updates the ECS service's task definition to the new image tag, triggering a rolling or blue/green deployment

## Cost Optimization Strategies

Effective Fargate cost management requires a multi-faceted strategy.

### Understanding Fargate Cost Drivers

Fargate pricing is "serverless"—you pay only for resources you *request* (not *use*), billed per-second with a one-minute minimum. Cost drivers:

- vCPU-per-second
- Memory (GB)-per-second
- Ephemeral Storage (GB)-per-second (beyond 20 GB free tier)

**Over-provisioning is the primary source of waste.** A task requesting 4 vCPU but using 0.5 vCPU wastes 87.5% of compute cost.

### Fargate Spot for Active Cost Savings

**Benefit**: Up to 70% discount on Fargate compute

**Trade-off**: Spot tasks can be interrupted with a two-minute warning

**Use Cases**: Interrupt-tolerant, stateless workloads—dev/test environments, batch processing, horizontally-scaled web services where the loss of a single task is not critical

**Strategy**: Never use Fargate Spot in isolation for critical services. Use the base (on-demand) and weight (Spot) strategy to blend reliability with cost savings.

### Compute Savings Plans for Passive Cost Savings

**Benefit**: Up to 50% discount on Fargate in exchange for a 1- or 3-year commitment to a specific dollar-per-hour compute spend

**Flexibility**: Compute Savings Plans are extremely flexible—not tied to a specific instance or region. The discount *automatically* applies to compute usage across EC2, Lambda, and Fargate (for both ECS and EKS).

**Hybrid Strategy**: Use a **Compute Savings Plan** to cover the predictable, 24/7 baseline of on-demand Fargate spend (the base tasks). Use **Fargate Spot** to cover dynamic, scalable, interruptible portions (the weight tasks).

### Right-Sizing Containers

**Problem**: Developers often "guess" at vCPU and memory, leading to significant over-provisioning.

**Solution**: Use **AWS Compute Optimizer**. This service analyzes at least 24 hours of CloudWatch metrics from running services and provides specific, actionable recommendations (e.g., "Change task size from 4 vCPU / 8GB to 2 vCPU / 4GB") with quantified cost savings.

This data-driven approach, combined with load testing, is the professional-grade method for eliminating waste and ensuring efficient Fargate deployments.

## Conclusion

Amazon ECS represents a deliberate choice: powerful simplicity and deep AWS integration over the ecosystem flexibility of Kubernetes. For teams committed to AWS who want to focus on application development rather than infrastructure complexity, ECS is the strategically sound choice.

OpenMCF's `AwsEcsCluster` API honors this decision by providing a streamlined, opinionated interface that abstracts the essential production features—monitoring, cost-optimization, debugging—without overwhelming users with low-level complexity. The cluster is intentionally minimal, but the few settings it exposes are the ones that matter most:

- **CloudWatch Container Insights** for comprehensive observability
- **Fargate and Fargate Spot capacity providers** with the base/weight strategy for cost optimization
- **ECS Exec auditing configuration** for secure, production-grade debugging

By elevating these features as first-class, high-level abstractions in the protobuf API, OpenMCF delivers the same experience ECS itself provides: production-grade capabilities without the operational burden. The cluster becomes the solid, observable, cost-efficient foundation on which containerized applications thrive.

