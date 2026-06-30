# AWS ECS Service Deployment: Filling the Vacuum Left by Copilot

## Introduction

The AWS Elastic Container Service (ECS) deployment landscape reveals a critical truth: developers don't want to be infrastructure assemblers. They want to define a *service*—an image, resource requirements, routing rules—and have the platform handle the intricate orchestration of task definitions, IAM roles, target groups, security policies, and load balancer rules.

AWS understood this. They built `ecs-cli` for high-level service abstractions, deprecated it, then replaced it with AWS Copilot—a more mature, opinionated CLI that generated clean CloudFormation and abstracted the entire workflow from Dockerfile to deployed service. On February 3, 2025, AWS officially ended support for Copilot.

This deprecation created a vacuum. Developers who relied on Copilot's service-oriented simplicity are left with a choice: drop down to low-level CloudFormation/Terraform resource-by-resource definitions, or find an alternative that provides the same high-level abstraction without vendor abandonment risk.

Planton fills this vacuum. It provides a stable, production-grade, service-oriented API for ECS deployment—one that learns from both the successes and failures of its predecessors.

## The ECS Abstraction Lifecycle: Understanding the Layers

ECS deployment methods have evolved through four distinct abstraction layers. Understanding this progression clarifies what Planton is designed to be.

### Layer 1: The Raw API

At the foundation is the AWS SDK, exposing raw API calls like `create_service`. This is the "ground truth" of the ECS control plane—every tool must ultimately interact with this layer.

**The Assembler Problem**: A single `create_service` call doesn't create a functional service. It *assembles* pre-existing resources. You must first create the cluster, task definition, IAM roles, subnets, security groups, and ALB target group, then pass their ARNs and IDs to wire everything together.

**Verdict**: Essential for understanding the system, unsuitable for production workflows.

### Layer 2: 1:1 Declarative Wrappers

This layer includes AWS CloudFormation (`AWS::ECS::Service`) and Terraform/OpenTofu (`aws_ecs_service`). These tools provide state management and a direct 1:1 mapping to API concepts.

**The Pattern**: Developers explicitly define every resource. A "simple" Fargate service requires verbose, separate definitions for:
- TaskDefinition
- TaskRole (for application AWS API permissions)
- ExecutionRole (for ECS agent permissions to pull images, fetch secrets, write logs)
- TargetGroup (for load balancer health checks)
- ListenerRule (for routing traffic)
- SecurityGroup (for network policies)

The ARNs and IDs from these resources are then manually "wired" together within the main `aws_ecs_service` resource block.

**Trade-offs**:
- **Pros**: 100% API coverage, granular control, stable and widely adopted
- **Cons**: Extremely verbose (200-500 lines for a simple service), high cognitive load, steep learning curve requiring deep expertise in ECS, IAM, and VPC networking

**Verdict**: The current industry standard for production teams with dedicated platform engineering capacity. The foundation layer for higher-level abstractions.

### Layer 3: High-Level Constructs

This layer, exemplified by AWS CDK Patterns (`ApplicationLoadBalancedFargateService`) and Pulumi Crosswalk (`awsx.ecs.FargateService`), bundles multiple L2 resources into a single, service-oriented construct.

**The Pattern**: Developers define a *single component* with high-level inputs like `image: "nginx"`, `cpu: 512`, and `memory: 1024`. The L3 construct automatically and implicitly generates all 5-10 ancillary L2 resources (ALB, Target Groups, Roles, Security Groups, Task Definition) with sensible, secure defaults.

**Example** (Pulumi awsx):
```typescript
const service = new awsx.ecs.FargateService("my-service", {
    cluster: cluster.arn,
    taskDefinitionArgs: {
        container: {
            image: "nginx:latest",
            cpu: 512,
            memory: 1024,
            portMappings: [{ containerPort: 80 }],
        },
    },
});
```

This single declaration generates the entire stack: Task Definition, both IAM roles (with appropriate policies), Target Group, and wires them into the service configuration.

**Trade-offs**:
- **Pros**: Massive reduction in boilerplate, codifies best practices, low cognitive load
- **Cons**: Opinionated (may not match 100% of edge cases), requires learning construct-specific patterns, CloudFormation synthesis can be slow (CDK)

**Verdict**: The right abstraction for developer productivity. This is where AWS Copilot operated, and where Planton aims.

### Layer 4: Platform-Level Abstractions

Tools like Crossplane Compositions and Planton present a simplified, abstract "platform API" to developers, hiding even the L3 constructs.

**The Vision**: A developer declares `AwsEcsService` with minimal, service-oriented fields. The platform handles everything: provisioning, networking, secrets management, observability, security policies—often enforcing organizational standards automatically.

**Verdict**: The future for large organizations with dedicated platform teams building internal developer platforms (IDPs).

## The Strategic Vacuum: Why Copilot's Deprecation Matters

AWS's abandonment of Copilot is significant. It's not just the loss of a tool—it's the loss of AWS's official stance that *developers deserve service-oriented abstractions*.

Copilot users are now faced with an unpleasant choice:
1. **Regress to Layer 2**: Drop down to verbose Terraform or CloudFormation, becoming infrastructure assemblers again
2. **Adopt Layer 3 CDK/Pulumi**: Learn a new imperative programming model, accept synthesis overhead, and manage state complexity
3. **Build Custom Abstractions**: Invest significant engineering time to recreate what Copilot provided

Planton offers a fourth path: stable, declarative, service-oriented abstractions built on proven open-source foundations (Terraform/Pulumi), designed for production use, and free from vendor deprecation risk.

## Production-Grade ECS: The Non-Negotiables

A robust L3 abstraction must correctly handle the patterns and pitfalls that define production ECS deployments.

### Compute: Fargate as the Sane Default

**Fargate** is the serverless, default option. AWS manages the underlying compute infrastructure. Developers define `cpu` and `memory` for their task and pay per-second for those allocated resources.

**When to Use Fargate**:
- Spiky or unpredictable workloads
- New or unproven workloads where utilization is unknown
- Teams prioritizing operational simplicity over cost optimization
- Workloads where compute utilization is below 60-80%

**EC2 Launch Type**: The alternative model where developers manage a cluster of EC2 instances (via Auto Scaling Groups) that tasks are placed onto. This is only cost-effective for **sustained high-utilization workloads** where tasks can be densely "bin-packed" onto instances to achieve >80% utilization.

**The Break-Even Point**: Analysis shows the cost break-even is between 60-85% EC2 utilization. Below this threshold, Fargate is cheaper. Achieving sustained >80% utilization in dynamic microservice environments is described as "more or less unachievable" in practice.

**Savings Plans Revolution**: The old argument for EC2 was Reserved Instances (RIs). This is now obsolete. AWS Compute Savings Plans offer up to 66% discounts and apply *equally* to EC2, **Fargate**, and Lambda. This allows teams to choose Fargate for its operational benefits without cost penalty.

**Planton Approach**: Default to `FARGATE` launch type. Advanced users can override for specialized cases (GPU workloads, extreme cost optimization with proven high utilization).

### Networking: The awsvpc Standard

The `awsvpc` network mode is *required* for all Fargate tasks and is the production standard for EC2 tasks. Every task receives its own Elastic Network Interface (ENI).

**Why This Matters**:
1. **Per-Task Security**: Each task ENI can have its own Security Group, enabling granular, zero-trust network policies
2. **No Port Conflicts**: Multiple copies of the same container can run on the same host without port collision
3. **Service Discovery**: Prerequisite for seamless integration with AWS Cloud Map

**Subnet Configuration**:
- **Private Subnets** (Recommended): Tasks are secure and not directly exposed to the internet. They access the internet (to pull images, hit APIs) via a NAT Gateway. This is the 80% production use case.
- **Public Subnets**: Used for tasks that need a public IP. **Critical**: `assignPublicIp: true` must be set, otherwise tasks cannot pull container images and deployments will fail.

**Planton Approach**: Treat `awsvpc` as the default. Require `subnets` and `security_groups` as essential configuration fields.

### IAM: The Two-Role Model

This is the single most confused concept for new ECS users. A production-ready abstraction must handle this correctly and intuitively.

**Task Execution Role** (for the **ECS Agent**):
- Grants permissions *to ECS itself* to perform actions *on your behalf* before the container starts
- Required permissions:
  - Pull images from ECR (`ecr:GetAuthorizationToken`, `ecr:BatchGetImage`)
  - Fetch secrets (`secretsmanager:GetSecretValue`, `ssm:GetParameters`)
  - Write logs (`logs:CreateLogStream`, `logs:PutLogEvents`)

**Task Role** (for your **Application Container**):
- Grants permissions *to your application code* to interact with other AWS services *after* it has started
- Example permissions: `s3:GetObject`, `sqs:SendMessage`, `dynamodb:PutItem`

**Common Anti-Pattern**: Granting application permissions (S3, DynamoDB) to the Task Execution Role. This is a security violation—the ECS agent should not have access to business data.

**Planton Approach**: Provide sensible defaults for Task Execution Role (ECR, logs, secrets). Expose Task Role ARN as an explicit field for users to attach application-specific policies.

### Secret Management: The Right Way

**Anti-Pattern**: Defining secrets in plain-text environment variables in infrastructure code. This hardcodes sensitive data into task definitions, IaC state files, and version control.

**Best Practice**: Sensitive data *must* be stored in AWS Secrets Manager or Systems Manager (SSM) Parameter Store as encrypted SecureStrings.

**The IaC Pattern**:
1. Store the secret value in Secrets Manager/SSM (outside of IaC)
2. In the task definition, reference the secret's ARN in a `secrets` block (not `environment`)
3. Grant the **Task Execution Role** permission to `secretsmanager:GetSecretValue` or `ssm:GetParameters`
4. ECS fetches the secret at runtime and injects it into the container as an environment variable

**SSM vs. Secrets Manager**:
- **Secrets Manager**: Use when automatic rotation, random generation, or cross-account sharing is required
- **SSM Parameter Store (SecureString)**: Use for simpler key-value storage without rotation requirements

**Planton Approach**: Provide a dedicated `secrets` map accepting `{name: valueFrom ARN}` pairs. Automatically add required permissions to the auto-generated Task Execution Role.

### Load Balancer Integration: The Single ALB Pattern

The standard architecture for microservices on ECS is the **"Single ALB, Multiple Services"** pattern. This is cost-effective and simplifies routing.

**Architecture**:
1. **One Shared ALB**: A single Application Load Balancer for the entire environment (e.g., "production-alb")
2. **One Shared Listener**: Typically HTTPS:443, configured with a wildcard SSL certificate
3. **N Services**: Each microservice (users-service, orders-service, auth-service) is a separate ECS service
4. **N Target Groups**: Each service gets its own Target Group tracking its task IPs
5. **N Listener Rules**: The shared listener routes traffic based on path or hostname:
   - Priority 10: IF `path == /users/*` → FORWARD to users-target-group
   - Priority 20: IF `path == /orders/*` → FORWARD to orders-target-group
   - Priority 30: IF `host == auth.example.com` → FORWARD to auth-target-group

**Critical Configuration**: `healthCheckGracePeriodSeconds` prevents a common deployment failure race condition:
1. ECS starts a new task. Container state becomes `RUNNING`.
2. Application inside is still booting (may take 30-60 seconds for Spring Boot, JVM apps).
3. ALB health check pings the task before it's ready.
4. Health check fails → ALB marks task Unhealthy → ECS kills the task → deployment fails.

**Solution**: Set `healthCheckGracePeriodSeconds` (e.g., 60-120 seconds) to instruct ECS to ignore ALB health status during initial boot.

**Planton Approach**: The `AwsEcsService` resource should *not* create the shared ALB or Listener. It takes the listener ARN as input and creates *only* the Target Group and Listener Rule. Default `healthCheckGracePeriodSeconds` to 60.

### CI/CD: The GitOps Pattern for Image Updates

A critical integration point: how does an ECS service, defined in IaC, get updated when a new Docker image is built?

**Anti-Pattern 1: The :latest Tag**: Using `image: "my-app:latest"` breaks deployments and rollbacks. ECS/IaC sees no change (the string is identical) even if the ECR image has updated. Rollbacks fail because the previous task definition also points to `:latest`.

**Anti-Pattern 2: The CLI Push**: App CI runs `aws ecs update-service --force-new-deployment`. This creates state drift—the live environment no longer matches the Git-defined state. The next `terraform apply` may revert the deployment.

**Recommended Pattern: The IaC Pull (GitOps)**:
1. App CI builds and pushes `my-app:git-sha-12345`
2. App CI opens a PR to the *IaC repo*, updating a variable: `image_tag = "git-sha-12345"`
3. Change is reviewed, merged, and triggers the IaC pipeline
4. Terraform/Pulumi sees a diff in the task definition and performs a safe, state-managed rolling deployment

**Alternative Pattern: The Data Source Pull**:
1. IaC uses a data source to query ECR for the digest of a stable tag (e.g., `stable`)
2. Task definition references the immutable digest
3. App CI's job is to push the new image and re-tag `stable` to its digest
4. Next IaC pipeline run fetches the new digest, forcing a diff and deployment

**Planton Approach**: Support both patterns. Provide `container.image.repo` and `container.image.tag` fields for explicit tagging (GitOps pattern). Document the data source pattern for advanced users.

### Observability: Logs, Metrics, and Traces

A production service is blind without observability.

**Logging**: The `awslogs` log driver is the standard. L3 abstractions should automatically create a CloudWatch Log Group (e.g., `/ecs/{service-name}`) and configure the task definition to send container stdout/stderr to it.

**Metrics (Container Insights)**: Provides cluster, service, and task-level metrics for CPU, memory, network, and disk. It is *not* enabled by default and must be explicitly enabled at the **cluster level**.

**Tracing (X-Ray)**: The standard pattern uses the AWS Distro for OpenTelemetry (ADOT) collector deployed as a *sidecar container* in the same task. The application container sends traces to the sidecar, which forwards them to X-Ray. This requires:
1. Adding the ADOT container definition to the task
2. Adding `AWSXrayWriteOnlyAccess` managed policy to the **Task Role**

**Planton Approach**: Auto-create CloudWatch Log Groups by default. Document Container Insights and ADOT as opt-in production enhancements.

### Auto Scaling: Target Tracking

The modern, recommended pattern for service scaling is **Target Tracking Scaling**.

**How It Works**: Developers define a minimum and maximum task count and a target metric (typically `cpu_utilization: 75%` or `memory_utilization: 75%`). AWS Application Auto Scaling automatically creates CloudWatch Alarms to scale the service's `desired_count` up or down to maintain the target.

**Behavior**: Designed for availability—scales out aggressively and quickly, scales in conservatively and gradually.

**Planton Approach**: Expose a simple `autoscaling` block accepting `min_tasks`, `max_tasks`, `target_cpu_percent`, and optionally `target_memory_percent`.

### Cost Optimization: The Hidden Drivers

The compute cost (Fargate/EC2) is often not the dominant expense. Production architectures must account for:

**NAT Gateway**: A major, often unexpected cost. Billed per-hour and per-GB processed. A private Fargate task must route through a NAT Gateway to pull images from ECR or hit external APIs.
- **Mitigation**: Use **VPC Endpoints** (Interface) for ECR, S3, Secrets Manager, and CloudWatch Logs. Tasks access these services over the AWS internal network, bypassing the NAT Gateway entirely.

**Cross-AZ Data Transfer**: Traffic between Availability Zones is billed (~$0.01-0.02/GB). Chatty microservices across AZs can generate significant costs.

**CloudWatch Logs**: Billed per-GB ingested (~$0.50/GB). Verbose logging at scale can cost thousands.
- **Mitigation**: Set log levels to INFO/WARN in production, use log sampling

**ELB**: ALBs/NLBs are billed per-hour and per-LCU (a unit of traffic/connections).
- **Mitigation**: Share ALBs across services (Single ALB, Multiple Services pattern)

## What Planton Supports

Planton provides a service-oriented API for deploying Fargate-based ECS services into existing clusters. The design follows the 80/20 principle: make the common case simple while making the advanced case possible.

### Design Philosophy: Service-First, Not Infrastructure-First

The core insight from the research is clear: developers want to define *services*, not assemble infrastructure components. Planton's `AwsEcsServiceSpec` reflects this philosophy.

**Essential Fields (The 80% Case)**:
- `cluster_arn`: The ECS cluster to deploy into
- `container.image.repo` and `container.image.tag`: The Docker image
- `container.port`: The port to expose for load balancing
- `container.cpu` and `container.memory`: Task-level resource requirements
- `container.replicas`: Number of tasks to run
- `network.subnets` and `network.security_groups`: The awsvpc configuration
- `alb.arn`, `alb.routing_type` (path or hostname), `alb.path` or `alb.hostname`: How to route traffic from the shared ALB

**Common Fields (The 19% Case)**:
- `container.env.variables`: Plain-text environment variables
- `container.env.secrets`: Secure secret injection from Secrets Manager/SSM
- `container.env.s3_files`: Environment files from S3
- `container.logging.enabled`: Auto-create CloudWatch Log Groups (default: true)
- `iam.task_execution_role_arn` and `iam.task_role_arn`: Override auto-generated IAM roles
- `alb.health_check`: Customize ALB target group health check settings
- `alb.listener_priority`: Control listener rule priority

**Advanced Fields (The 1% Case)**: Not yet exposed, reserved for future needs:
- Launch type override (EC2)
- Capacity provider strategies (Fargate + Fargate Spot blends)
- Placement constraints (EC2-only)
- Service discovery / service mesh integration
- Sidecar containers (ADOT, Envoy)
- Blue/Green deployments (CODE_DEPLOY controller)

### Current Implementation Patterns

**Fargate-First**: The implementation defaults to Fargate. No EC2 cluster management overhead.

**Shared ALB Integration**: The API assumes a shared ALB pattern. Users provide the ALB ARN and routing configuration (path-based or hostname-based). Planton creates the Target Group and Listener Rule.

**Secure Secrets Management**: Secrets are configured separately from plain-text environment variables, ensuring proper IAM permissions and runtime injection.

**Automatic Logging**: CloudWatch Log Groups are auto-created unless explicitly disabled, codifying the production standard.

### Multi-Environment Best Practice

Following AWS best practices, Planton encourages separate ECS clusters for each environment:
- `dev-cluster` → Development services
- `staging-cluster` → Staging services
- `prod-cluster` → Production services

Each environment references a different `cluster_arn`, providing complete isolation for resources, security policies, and cost allocation.

## Conclusion: The Service-Oriented Future

AWS ECS deployment has evolved through four abstraction layers: from raw API calls to 1:1 declarative resources to high-level constructs to platform-level abstractions. AWS's own opinionated tooling (ecs-cli, Copilot) attempted to provide Layer 3 abstractions but were ultimately deprecated, leaving a vacuum for teams seeking service-oriented simplicity.

Planton fills this vacuum with a production-grade, declarative API that makes deploying ECS services feel like deploying to a platform—not like assembling infrastructure components. By codifying best practices (Fargate defaults, awsvpc networking, secure secret management, shared ALB patterns, automatic logging), Planton helps teams avoid common pitfalls and deploy production-ready services with confidence.

The paradigm shift is clear: developers should define *what* they want to run (a service), not *how* to wire together the dozen resources required to run it. Planton provides that abstraction—stable, open-source, and built on the proven foundations of Terraform and Pulumi.

