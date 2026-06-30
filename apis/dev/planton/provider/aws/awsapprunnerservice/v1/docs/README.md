# AWS App Runner -- Comprehensive Research Documentation

This document provides in-depth technical context for the `AwsAppRunnerService` Planton component. It is intended for contributors, platform engineers, and anyone who needs to understand the design decisions behind the component spec.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Source Types Deep Dive](#source-types-deep-dive)
3. [Networking](#networking)
4. [Auto Scaling Model](#auto-scaling-model)
5. [Cost Model](#cost-model)
6. [Security](#security)
7. [Health Checks](#health-checks)
8. [Observability](#observability)
9. [Deployment Lifecycle](#deployment-lifecycle)
10. [Comparison with Alternatives](#comparison-with-alternatives)
11. [Service Limits and Quotas](#service-limits-and-quotas)
12. [v2 Roadmap Ideas](#v2-roadmap-ideas)

---

## Architecture Overview

AWS App Runner is a fully managed compute service built on top of AWS Fargate. When you create an App Runner service, the following happens under the hood:

1. **Image acquisition** -- App Runner pulls your container image from ECR/ECR Public, or clones and builds your GitHub repository.
2. **Image caching** -- A copy of the image is stored in App Runner's internal registry, encrypted at rest (with your KMS key or an AWS-managed key).
3. **Instance provisioning** -- App Runner launches one or more Fargate-based instances. Each instance runs your container in an isolated microVM (Firecracker).
4. **Load balancing** -- An internal Application Load Balancer distributes incoming HTTPS requests across healthy instances. TLS termination happens at the ALB; traffic to your container is HTTP on the configured port.
5. **Auto scaling** -- A concurrency-based auto scaler monitors the number of in-flight requests per instance and scales horizontally.
6. **Health monitoring** -- Each instance is probed via TCP or HTTP health checks. Unhealthy instances are replaced automatically.

### Request flow

```
Client (HTTPS:443)
  └─> App Runner managed ALB (TLS termination)
        └─> Instance 1 (HTTP:port)
        └─> Instance 2 (HTTP:port)
        └─> Instance N (HTTP:port)
              └─> (optional) VPC Connector → VPC resources
```

App Runner manages the ALB, TLS certificate (*.awsapprunner.com), DNS, health checks, and scaling. You only define the container, port, scaling rules, and optional networking.

### Resource model

A single `AwsAppRunnerService` manifest can create up to three AWS resources:

| Resource | When created | Lifecycle |
| --- | --- | --- |
| `aws:apprunner:Service` | Always | Core resource |
| `aws:apprunner:VpcConnector` | When `subnet_ids` are provided | Managed inline |
| `aws:apprunner:AutoScalingConfigurationVersion` | When `auto_scaling` is provided | Managed inline |

This bundling simplifies UX -- users define one manifest instead of three separate resources.

---

## Source Types Deep Dive

App Runner supports two fundamentally different source types. Understanding the differences is critical for choosing the right deployment model.

### Image source (ECR / ECR Public)

**How it works:** You provide a fully built container image URI. App Runner pulls it and deploys it directly. No build step happens inside App Runner.

**ECR Private (`image_repository_type: ECR`):**
- Requires an IAM access role (`access_role_arn`) with permissions:
  - `ecr:GetDownloadUrlForLayer`
  - `ecr:BatchGetImage`
  - `ecr:GetAuthorizationToken`
- The access role must trust the `build.apprunner.amazonaws.com` service principal.
- Supports image tags and SHA256 digests.
- When `auto_deployments_enabled` is true, App Runner uses ECR event rules to detect pushes to the configured tag and automatically redeploys.

**ECR Public (`image_repository_type: ECR_PUBLIC`):**
- No authentication required. Anyone can pull ECR Public images.
- `access_role_arn` is not needed (and is ignored if provided).
- Useful for open-source images, demos, and quick prototyping.
- Auto-deployment is also supported for ECR Public.

**Best for:** Teams with existing CI/CD pipelines that already build and push container images. Gives full control over the build process, base images, and image scanning.

### Code source (GitHub)

**How it works:** You provide a GitHub repository URL and branch. App Runner clones the repo, builds the application using a managed runtime, and deploys the resulting container.

**Prerequisites:**
- An **App Runner Connection** (`connection_arn`) that authorizes GitHub access. Connections are created via the AWS Console or CLI (they require an interactive OAuth handshake with GitHub). A single connection can be shared across multiple services.
- The connection must be in the `AVAILABLE` state (not `PENDING_HANDSHAKE`).

**Configuration source:**

| Mode | `configuration_source` | Build config source | When to use |
| --- | --- | --- | --- |
| API | `"API"` | Inline in the Planton spec (`runtime`, `build_command`, `start_command`, `port`) | Simple apps where you want all config in one place |
| Repository | `"REPOSITORY"` | `apprunner.yaml` in the repo root (or `source_directory`) | Teams that want build config co-located with code |

**Supported managed runtimes:**

| Runtime ID | Language | Notes |
| --- | --- | --- |
| `PYTHON_3` | Python 3 | pip-based builds |
| `NODEJS_12` | Node.js 12 | Legacy, avoid for new projects |
| `NODEJS_14` | Node.js 14 | Legacy |
| `NODEJS_16` | Node.js 16 | Maintenance LTS |
| `NODEJS_18` | Node.js 18 | Recommended for Node.js |
| `CORRETTO_8` | Java (Corretto 8) | Maven/Gradle builds |
| `CORRETTO_11` | Java (Corretto 11) | Maven/Gradle builds |
| `GO_1` | Go 1.x | `go build` based |
| `DOTNET_6` | .NET 6 | `dotnet publish` based |
| `PHP_81` | PHP 8.1 | Composer-based builds |
| `RUBY_31` | Ruby 3.1 | Bundler-based builds |

**Source directory:** For monorepos, set `source_directory` to the subdirectory containing the application. App Runner treats this as the build context root.

**Auto-deploy:** When `auto_deployments_enabled` is true, App Runner listens for push events on the configured branch and automatically triggers a new deployment.

**Best for:** Teams that want a fully managed build-and-deploy pipeline without maintaining a CI/CD system. Great for prototyping, hackathons, and small teams.

---

## Networking

App Runner networking has two distinct aspects: **ingress** (how traffic reaches your service) and **egress** (how your service reaches external resources).

### Ingress

By default, every App Runner service gets a public HTTPS endpoint at `SERVICE_ID.REGION.awsapprunner.com`. This endpoint:
- Terminates TLS with an AWS-managed certificate.
- Supports HTTP/1.1 and HTTP/2 (but not gRPC passthrough as of 2024).
- Is backed by an internal Application Load Balancer.
- Supports `IPV4` or `DUAL_STACK` (IPv4 + IPv6) via `ip_address_type`.

**Private ingress:** Setting `is_publicly_accessible: false` makes the service unreachable from the public internet. To access it, you need a **VPC Ingress Connection** (a separate AWS resource not managed by this component). This is useful for internal microservices that should only be reachable from within a VPC.

**Custom domains:** App Runner supports associating custom domains with a service, but this is managed separately (via `aws apprunner associate-custom-domain` or a dedicated resource). This component does not manage custom domains.

### Egress

By default, App Runner instances have outbound internet access through AWS-managed NAT. They can reach public APIs, SaaS endpoints, and AWS service endpoints without any VPC configuration.

**VPC egress (VPC Connector):** When your service needs to reach resources inside a VPC (RDS databases, ElastiCache clusters, internal APIs, PrivateLink endpoints), you need a VPC Connector. This component supports two patterns:

1. **Inline VPC Connector** -- Provide `subnet_ids` (and optionally `security_group_ids`). The module creates a VPC Connector automatically. This is the recommended approach for most use cases.

2. **Existing VPC Connector** -- Provide `vpc_connector_arn` to reference a VPC Connector you created separately. Useful when sharing a single connector across multiple services.

These two patterns are mutually exclusive (enforced by validation).

**Important networking behavior with VPC Connector:**
- When a VPC Connector is attached, **all** outbound traffic from the service is routed through the VPC.
- To reach the public internet (e.g., third-party APIs), the VPC subnets must have a NAT Gateway or NAT Instance.
- To reach AWS services, use VPC endpoints (PrivateLink) for better performance and cost.
- The VPC Connector supports subnets in multiple AZs. Use at least 2 AZs for high availability.

### DNS resolution

App Runner instances use the VPC's DNS resolver when a VPC Connector is attached. This means they can resolve private DNS names (e.g., RDS endpoints, Route 53 private hosted zones).

---

## Auto Scaling Model

App Runner uses a **concurrency-based** auto scaling model, which is fundamentally different from CPU/memory-based scaling in ECS or Kubernetes.

### How it works

1. Each instance has a `max_concurrency` setting (default: 100). This is the maximum number of concurrent HTTP requests the instance will handle simultaneously.
2. When the aggregate concurrency across all instances approaches capacity, App Runner launches new instances (up to `max_size`).
3. When traffic decreases, App Runner removes instances (down to `min_size`).
4. The `min_size` instances are always warm and ready to serve traffic.

### Scaling timeline

| Event | Approximate time |
| --- | --- |
| New instance cold start | 2--30 seconds (depends on image size and startup time) |
| Scale-out trigger to instance ready | 10--60 seconds |
| Scale-in (instance removal) | Gradual, after sustained low traffic |

### Warm instances vs. provisioned instances

- **Active instances**: Serving or ready to serve traffic. You pay the active instance rate.
- **Provisioned (idle) instances**: Warm instances at `min_size` that are not receiving traffic. You pay a lower provisioned instance rate (approximately 1/10th of the active rate for compute).

### Tuning guidance

| Scenario | `min_size` | `max_size` | `max_concurrency` |
| --- | --- | --- | --- |
| Development/staging | 1 | 2 | 100 |
| Low-traffic API | 1 | 5 | 100 |
| Production API with SLA | 2--3 | 10--15 | 50--80 |
| High-traffic public website | 3--5 | 25 | 50--100 |
| CPU-intensive workload | 2 | 25 | 10--25 |

**Key insight:** Lower `max_concurrency` means App Runner scales out sooner, giving each instance more headroom. This is better for CPU-intensive workloads. Higher `max_concurrency` is more cost-efficient for I/O-bound workloads (e.g., proxy, gateway).

---

## Cost Model

App Runner pricing has two dimensions:

### 1. Compute charges

| Charge type | Rate (us-east-1, approximate) | When it applies |
| --- | --- | --- |
| Active instance (vCPU-second) | ~$0.064 / vCPU-hour | Instance is serving or ready to serve traffic |
| Provisioned instance (vCPU-second) | ~$0.007 / vCPU-hour | Warm instances at `min_size` with no traffic |
| Memory (GB-second) | ~$0.007 / GB-hour | Always (active + provisioned) |

### 2. Additional charges

- **Build minutes** (code source only): ~$0.005 per build minute
- **Data transfer**: Standard AWS data transfer rates (ingress free, egress to internet charged)
- **NAT Gateway** (if using VPC Connector with NAT): Standard NAT Gateway rates

### Cost optimization tips

1. **Right-size CPU and memory.** Start with `1024` CPU / `2048` memory and benchmark. Don't over-provision.
2. **Use `min_size: 1` for non-critical services.** Saves provisioned instance costs but introduces cold-start latency.
3. **Tune `max_concurrency`.** Higher values mean fewer instances for the same traffic, but watch latency.
4. **Disable auto-deploy in production.** Unintended deployments can cause unnecessary build charges and downtime risk.
5. **Use ARM-based images** when available -- App Runner does not currently expose instance architecture selection, but this may change.

### Example monthly estimate

A service with 1 vCPU / 2 GB memory, `min_size: 1`, `max_size: 5`, averaging 2 active instances for 8 hours/day and 1 provisioned instance for 16 hours/day:

```
Active:       2 instances × 1 vCPU × 8h × 30 days × $0.064/vCPU-hr  = $30.72
              2 instances × 2 GB  × 8h × 30 days × $0.007/GB-hr     = $6.72
Provisioned:  1 instance  × 1 vCPU × 16h × 30 days × $0.007/vCPU-hr = $3.36
              1 instance  × 2 GB  × 16h × 30 days × $0.007/GB-hr    = $6.72
Total:                                                                ≈ $47.52/month
```

---

## Security

### Encryption at rest

App Runner encrypts the stored copy of your container image and any data logs. By default, it uses an AWS-managed key. You can provide a customer-managed KMS key via `kms_key_arn` for:
- Compliance requirements (HIPAA, SOC2, PCI-DSS)
- Key rotation control
- Cross-account key sharing
- Audit trail via CloudTrail

**Important:** `kms_key_arn` is a **ForceNew** field. Changing it requires replacing the entire service (all instances are terminated and recreated).

### Encryption in transit

- Client-to-service: TLS 1.2+ enforced at the App Runner managed ALB.
- Service-to-VPC: Traffic through the VPC Connector is within the AWS network but is **not** encrypted at the application layer by default. Use TLS for connections to databases and downstream services.

### IAM roles

App Runner uses two distinct IAM roles:

| Role | Field | Purpose | Trust principal |
| --- | --- | --- | --- |
| **Access role** | `image_source.access_role_arn` | Pull images from private ECR | `build.apprunner.amazonaws.com` |
| **Instance role** | `instance_role_arn` | Runtime AWS API access (S3, DynamoDB, Secrets Manager, etc.) | `tasks.apprunner.amazonaws.com` |

**Access role minimum permissions:**
```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
      "ecr:DescribeImages",
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability"
    ],
    "Resource": "*"
  }]
}
```

**Instance role:** Permissions depend on what your application needs. Follow least-privilege principles. Common patterns:
- S3 read/write for file storage
- DynamoDB CRUD for data
- Secrets Manager / SSM read for secrets
- SQS send/receive for async messaging
- SNS publish for notifications

### Secrets management

The `environment_secrets` field supports two secret sources:

| Source | ARN format | Example |
| --- | --- | --- |
| Secrets Manager | `arn:aws:secretsmanager:REGION:ACCOUNT:secret:NAME` | Database passwords, API keys |
| SSM Parameter Store | `arn:aws:ssm:REGION:ACCOUNT:parameter/PATH` | Configuration values, feature flags |

App Runner resolves secrets at deploy time and injects them as plaintext environment variables. The `instance_role_arn` must have `secretsmanager:GetSecretValue` or `ssm:GetParameter` permissions for the referenced secrets.

**Security note:** Secrets are injected as environment variables, which means they appear in the process environment. They are not exposed in App Runner logs, but they are visible to anyone who can exec into the instance (not possible in App Runner) or who has access to the App Runner API.

---

## Health Checks

App Runner health checks determine instance readiness. Unhealthy instances are automatically replaced.

### TCP health check (default)

- Checks that the configured port is open and accepting TCP connections.
- No application-level validation.
- Best for: Simple services where "port is open" means "ready to serve."

### HTTP health check

- Sends an HTTP GET request to the configured `path` on the instance port.
- Expects an HTTP 200 response.
- Best for: Services that need application-level readiness validation (e.g., database connection established, cache warmed).

### Timing parameters

| Parameter | Default | Range | Guidance |
| --- | --- | --- | --- |
| `interval_seconds` | 5 | 1--20 | Lower = faster detection, higher overhead |
| `timeout_seconds` | 2 | 1--20 | Must be less than `interval_seconds` |
| `healthy_threshold` | 1 | 1--20 | 1 is usually fine (fast recovery) |
| `unhealthy_threshold` | 5 | 1--20 | Lower = faster replacement, higher false-positive risk |

### Health check design tips

1. **Keep health check endpoints fast** (< 100ms). Avoid database queries in health checks unless you specifically want to validate DB connectivity.
2. **Use HTTP health checks in production.** TCP health checks only verify the port is open, not that your application is functioning correctly.
3. **Separate readiness from liveness.** App Runner uses a single health check for both. If your app needs time to warm up, use a higher `healthy_threshold`.
4. **Return 200 from your health endpoint even during graceful degradation.** Only return non-200 when the instance truly cannot serve traffic.

---

## Observability

### Built-in observability

App Runner provides:
- **CloudWatch metrics:** `RequestCount`, `2xxStatusResponses`, `4xxStatusResponses`, `5xxStatusResponses`, `RequestLatency`, `ActiveInstances`, `ConcurrencyUtilization`
- **CloudWatch Logs:** Application stdout/stderr is automatically collected and sent to CloudWatch Logs at `/aws/apprunner/SERVICE_NAME/SERVICE_ID/application`
- **System logs:** App Runner lifecycle events are logged at `/aws/apprunner/SERVICE_NAME/SERVICE_ID/service`

### X-Ray tracing

When `observability_enabled: true` and an `observability_configuration_arn` is provided, App Runner instruments requests with AWS X-Ray. This gives you:
- End-to-end request traces
- Service maps showing dependencies
- Latency analysis by segment

The observability configuration is created separately (via AWS Console or CLI) and defines the X-Ray sampling rate.

---

## Deployment Lifecycle

### Create

1. App Runner validates the configuration.
2. The image is pulled (or code is built).
3. A cached copy of the image is stored (encrypted).
4. Instances are launched (starting at `min_size`).
5. Health checks pass, and the service enters `RUNNING` state.
6. The service URL becomes active.

Total create time: 2--10 minutes (depends on image size and startup time).

### Update

App Runner supports two update strategies:

- **Rolling deployment** (default): New instances are launched with the new configuration. Once healthy, old instances are drained and terminated. There is no downtime.
- **Pause and resume**: You can pause a service (terminates all instances, stops billing) and resume it later.

### Auto-deploy

When `auto_deployments_enabled: true`:
- **Image source:** App Runner monitors the ECR image tag for new pushes. A push triggers automatic redeployment.
- **Code source:** App Runner monitors the GitHub branch for new commits. A push triggers build + deploy.

### ForceNew fields

Changing these fields requires replacing the entire service (creates a new service and deletes the old one):
- `kms_key_arn` (encryption configuration)
- `service_name` (derived from metadata)

All other fields can be updated in-place.

---

## Comparison with Alternatives

### App Runner vs. ECS Fargate

| Dimension | App Runner | ECS Fargate |
| --- | --- | --- |
| **Complexity** | Minimal (no cluster, task def, ALB to manage) | Moderate (cluster, task definition, service, ALB, target group) |
| **Scaling model** | Concurrency-based | CPU/memory/request-based, step/target tracking |
| **Networking** | Managed ALB, optional VPC Connector | Full VPC control, service connect, service mesh |
| **Custom domains** | Supported (manual association) | Via ALB + Route 53 (full control) |
| **Sidecars** | Not supported | Supported (Firelens, Datadog agent, etc.) |
| **gRPC** | Not supported natively | Supported via ALB gRPC target groups |
| **GPU** | Not supported | Supported (EC2 launch type) |
| **Cost** | Slightly higher per-instance (convenience premium) | Lower per-instance, higher operational cost |
| **Deploy speed** | 2--5 min | 3--10 min (depends on ALB health check) |

**Choose App Runner when:** You want the fastest path to production for a stateless HTTP service and don't need sidecars, gRPC, or fine-grained networking.

**Choose ECS Fargate when:** You need sidecars, service mesh, gRPC, custom health check logic, or full networking control.

### App Runner vs. Lambda

| Dimension | App Runner | Lambda |
| --- | --- | --- |
| **Execution model** | Long-running process | Event-driven, invocation-based |
| **Max execution time** | Unlimited (always running) | 15 minutes |
| **Scaling** | Concurrency-based, 1--25 instances | Invocation-based, 0--thousands |
| **Cold starts** | 2--30s (mitigated by `min_size`) | 100ms--10s (mitigated by provisioned concurrency) |
| **Billing** | Per-second (instance time) | Per-invocation + duration |
| **WebSocket** | Not supported | Via API Gateway WebSocket |
| **State** | In-memory state persists across requests | Stateless per invocation |

**Choose App Runner when:** Your workload receives continuous HTTP traffic and benefits from warm instances.

**Choose Lambda when:** Your workload is event-driven, bursty, or needs to scale to thousands of concurrent executions.

### App Runner vs. EKS

| Dimension | App Runner | EKS |
| --- | --- | --- |
| **Complexity** | Minimal | High (Kubernetes expertise required) |
| **Ecosystem** | Limited to App Runner features | Full Kubernetes ecosystem (Helm, operators, Istio, etc.) |
| **Multi-cloud** | AWS only | Portable across clouds |
| **Cost** | Simple pricing | Complex (control plane + nodes + ALB + add-ons) |
| **Team skills** | No Kubernetes knowledge needed | Kubernetes expertise required |

**Choose App Runner when:** Your team is small, doesn't have Kubernetes expertise, and the workload is a simple HTTP service.

**Choose EKS when:** You're running many services, need the Kubernetes ecosystem, or need multi-cloud portability.

### App Runner vs. AWS Copilot

AWS Copilot is a CLI that deploys to ECS Fargate (or App Runner). It is an **orchestration tool**, not a compute service. You can use Copilot to deploy App Runner services, but Planton serves a similar orchestration role with a declarative manifest approach.

---

## Service Limits and Quotas

| Resource | Default limit | Adjustable |
| --- | --- | --- |
| Services per region per account | 25 | Yes (via quota request) |
| Active instances per service | 25 | Yes |
| Max concurrency per instance | 200 | No |
| CPU per instance | 4 vCPU | No |
| Memory per instance | 12 GB | No |
| Max image size | 12 GB | No |
| Custom domains per service | 5 | Yes |
| VPC Connectors per region | 10 | Yes |
| Security groups per VPC Connector | 5 | No |
| Subnets per VPC Connector | 16 | No |
| Environment variables + secrets | 200 total | No |
| Request timeout | 120 seconds | No |
| Request body size | 6 MB | No |
| Response body size | 6 MB | No |
| Concurrent deployments per service | 1 | No |

### Important operational limits

- **Request timeout:** App Runner has a fixed 120-second request timeout. Long-polling, SSE, and WebSocket are not supported.
- **Request/response body size:** 6 MB limit. For large payloads, use pre-signed S3 URLs.
- **No persistent storage:** App Runner instances have ephemeral storage only. Use S3, EFS (not natively supported), or external databases.
- **No SSH/exec access:** You cannot shell into App Runner instances. Use structured logging and X-Ray for debugging.
- **Single container per instance:** No sidecar support. Logging agents, APM agents, etc., must be embedded in your container.

---

## v2 Roadmap Ideas

The following items are not currently in the spec but are candidates for future versions:

1. **Custom domain management** -- Associate custom domains with the service (currently done out-of-band).
2. **WAF integration** -- Attach AWS WAF Web ACL to the App Runner service for request filtering and rate limiting.
3. **Tags on sub-resources** -- Propagate tags to the VPC Connector and Auto Scaling Configuration.
4. **IP-based access control** -- Allow/deny lists for ingress (currently requires WAF).
5. **VPC Ingress Connection** -- Manage private ingress as part of the same manifest for fully private services.
6. **Deployment strategy options** -- Expose canary/blue-green deployment options when App Runner adds support.
7. **gRPC support** -- Map to App Runner's gRPC capabilities when they become available.
8. **Managed runtime version pinning** -- Allow specifying exact runtime versions for code source builds.
9. **EFS volume mounts** -- If App Runner adds EFS support, expose it in the spec.
10. **Service-to-service authorization** -- Integrate with App Runner's future service auth features.
