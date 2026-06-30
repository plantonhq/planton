# AwsAppRunnerService

AWS App Runner is a fully managed service that makes it easy to deploy containerized web applications and APIs at scale, with no infrastructure to manage. You give it a container image or a GitHub repository, and App Runner handles building, deploying, scaling, and load-balancing -- producing an HTTPS endpoint in minutes. It is the simplest path from container to production URL on AWS.

## When to use App Runner

| Use case | Best fit | Why |
| --- | --- | --- |
| Stateless HTTP APIs and web apps that need zero-ops deployment | **App Runner** | No cluster, no task definitions, no load balancer to configure |
| Event-driven, short-lived functions (< 15 min) | Lambda | Pay-per-invocation, sub-second billing, broader event-source integrations |
| Long-running services that need full control of networking, sidecars, or service mesh | ECS Fargate | Full task-definition control, service discovery, service connect |
| Kubernetes workloads or teams already invested in K8s tooling | EKS | Full Kubernetes API, Helm charts, GitOps |
| Batch or GPU workloads | ECS / EKS | App Runner does not support GPU instance types or batch scheduling |

**Rule of thumb:** If your workload is a stateless HTTP service and you want AWS to own the infrastructure decisions, start with App Runner. Move to ECS Fargate or EKS when you need capabilities App Runner does not expose (custom networking topologies, sidecars, GPU, gRPC passthrough, etc.).

## Spec fields

### Top-level (`AwsAppRunnerServiceSpec`)

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `image_source` | `AwsAppRunnerServiceImageSource` | -- | Deploy from a container image in ECR or ECR Public. **Exactly one** of `image_source` or `code_source` must be set. |
| `code_source` | `AwsAppRunnerServiceCodeSource` | -- | Deploy from a GitHub repository. **Exactly one** of `image_source` or `code_source` must be set. |
| `port` | `string` | `"8080"` | Port the application listens on inside the container. App Runner routes inbound HTTPS (443) to this port. |
| `start_command` | `string` | -- | Override the container's ENTRYPOINT/CMD (image source) or define the start command (code source with `configuration_source=API`). |
| `environment_variables` | `map<string,string>` | -- | Plaintext environment variables. Keys prefixed with `AWSAPPRUNNER` are reserved. |
| `environment_secrets` | `map<string,string>` | -- | Secret environment variables. Values are ARNs of Secrets Manager secrets or SSM Parameter Store parameters. The `instance_role_arn` must grant read access. |
| `cpu` | `string` | `"1024"` | CPU per instance. Numeric (`"256"`, `"512"`, `"1024"`, `"2048"`, `"4096"`) or human-readable (`"0.25 vCPU"` .. `"4 vCPU"`). |
| `memory` | `string` | `"2048"` | Memory per instance in MB. Numeric (`"512"` .. `"12288"`) or human-readable (`"0.5 GB"` .. `"12 GB"`). Not all CPU/memory combos are valid -- see AWS docs. |
| `instance_role_arn` | `StringValueOrRef` | -- | IAM role assumed by running instances to call AWS APIs (S3, DynamoDB, etc.). **Not** the role used to pull images. |
| `health_check` | `AwsAppRunnerServiceHealthCheck` | TCP defaults | Health check configuration. See nested message below. |
| `auto_scaling` | `AwsAppRunnerServiceAutoScaling` | 1 min / 25 max / 100 concurrency | Auto scaling configuration. See nested message below. |
| `vpc_connector_arn` | `StringValueOrRef` | -- | ARN of an existing VPC Connector for outbound VPC access. Mutually exclusive with `subnet_ids`. |
| `subnet_ids` | `repeated StringValueOrRef` | -- | Subnet IDs for an inline VPC Connector. Provide subnets in 2+ AZs. Mutually exclusive with `vpc_connector_arn`. |
| `security_group_ids` | `repeated StringValueOrRef` | -- | Security group IDs for the inline VPC Connector. Only used when `subnet_ids` is provided. |
| `is_publicly_accessible` | `bool` | `true` | Whether the service endpoint is publicly reachable. When `false`, the service requires a VPC Ingress Connection. |
| `ip_address_type` | `string` | `"IPV4"` | `"IPV4"` or `"DUAL_STACK"` (IPv4 + IPv6). |
| `kms_key_arn` | `StringValueOrRef` | AWS-managed key | Customer-managed KMS key for encrypting stored images and logs. **ForceNew** -- changing this replaces the service. |
| `observability_enabled` | `bool` | `false` | Enable AWS X-Ray tracing. Requires `observability_configuration_arn`. |
| `observability_configuration_arn` | `StringValueOrRef` | -- | ARN of an App Runner Observability Configuration. Required when `observability_enabled` is `true`. |
| `auto_deployments_enabled` | `bool` | `true` | Automatically redeploy when the source (image tag or code branch) changes. |

### `AwsAppRunnerServiceImageSource`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `image_identifier` | `string` | yes | Full image URI with tag or digest. ECR: `ACCOUNT.dkr.ecr.REGION.amazonaws.com/REPO:TAG`. ECR Public: `public.ecr.aws/ALIAS/REPO:TAG`. |
| `image_repository_type` | `string` | yes | `"ECR"` (private) or `"ECR_PUBLIC"`. |
| `access_role_arn` | `StringValueOrRef` | for ECR | IAM role that grants App Runner permission to pull from private ECR. Must have `ecr:GetDownloadUrlForLayer`, `ecr:BatchGetImage`, `ecr:GetAuthorizationToken`. |

### `AwsAppRunnerServiceCodeSource`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `repository_url` | `string` | yes | GitHub repository URL (e.g., `https://github.com/owner/repo`). |
| `branch` | `string` | yes | Branch to deploy from (e.g., `main`). |
| `source_directory` | `string` | no | Subdirectory containing the app source. Defaults to repo root. Useful for monorepos. |
| `connection_arn` | `StringValueOrRef` | yes | ARN of an App Runner Connection authorizing GitHub access. Created via AWS Console/CLI (requires OAuth). |
| `configuration_source` | `string` | yes | `"API"` (build config in this spec) or `"REPOSITORY"` (reads `apprunner.yaml` from the repo). |
| `runtime` | `string` | for API | Runtime identifier. Values: `PYTHON_3`, `NODEJS_12`/`14`/`16`/`18`, `CORRETTO_8`/`11`, `GO_1`, `DOTNET_6`, `PHP_81`, `RUBY_31`. |
| `build_command` | `string` | for API | Shell command to build the app (e.g., `npm ci && npm run build`). |

### `AwsAppRunnerServiceHealthCheck`

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `protocol` | `string` | `"TCP"` | `"TCP"` (port-open check) or `"HTTP"` (GET request expecting 200). |
| `path` | `string` | `"/"` | URL path for HTTP health checks. Ignored for TCP. |
| `interval_seconds` | `int32` | `5` | Seconds between checks (1--20). |
| `timeout_seconds` | `int32` | `2` | Max seconds to wait for a response (1--20). |
| `healthy_threshold` | `int32` | `1` | Consecutive successes to mark healthy (1--20). |
| `unhealthy_threshold` | `int32` | `5` | Consecutive failures to mark unhealthy and replace (1--20). |

### `AwsAppRunnerServiceAutoScaling`

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `min_size` | `int32` | `1` | Minimum warm instances (1--25). Higher values reduce cold starts. |
| `max_size` | `int32` | `25` | Maximum instances during traffic spikes (1--25). |
| `max_concurrency` | `int32` | `100` | Concurrent requests per instance before scaling out (1--200). |

## Stack outputs

| Output | Description |
| --- | --- |
| `service_arn` | Full ARN of the App Runner Service. |
| `service_id` | Unique identifier assigned by App Runner. |
| `service_url` | Public HTTPS URL (e.g., `abc123.us-east-1.awsapprunner.com`). |
| `service_name` | Computed service name (derived from metadata). |
| `service_status` | Current operational status (`RUNNING`, `CREATE_FAILED`, etc.). |
| `vpc_connector_arn` | ARN of the VPC Connector (empty if VPC egress was not configured). |
| `auto_scaling_configuration_arn` | ARN of the Auto Scaling Configuration Version. |

## Prerequisites

1. **AWS credentials** -- Provided via stack input (`provider_credential`), not in the spec.
2. **IAM access role** (private ECR only) -- An IAM role with `ecr:GetDownloadUrlForLayer`, `ecr:BatchGetImage`, and `ecr:GetAuthorizationToken` permissions. Pass its ARN as `image_source.access_role_arn`.
3. **IAM instance role** (optional) -- If your application calls AWS APIs at runtime (S3, DynamoDB, Secrets Manager, etc.), create a role with the necessary policies and pass it as `instance_role_arn`.
4. **App Runner Connection** (code source only) -- Created via the AWS Console or CLI. The connection performs an OAuth handshake with GitHub and can be shared across multiple services.
5. **VPC and subnets** (optional) -- Only needed when the service must reach resources in a VPC (RDS, ElastiCache, internal APIs). Provide subnet IDs in at least two AZs.
6. **KMS key** (optional) -- Only needed for customer-managed encryption. The key must allow the App Runner service principal to use it.

## Quick start

Deploy a public Nginx container with a single manifest:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsAppRunnerService
metadata:
  name: my-web-app
spec:
  imageSource:
    imageIdentifier: "public.ecr.aws/nginx/nginx:latest"
    imageRepositoryType: "ECR_PUBLIC"
  port: "80"
```

## How it works

This resource is orchestrated by the Planton CLI as part of a stack-update. The CLI validates your manifest, generates stack inputs, and invokes the IaC backend:

- **Pulumi** (Go modules under `iac/pulumi/`)

The module automatically creates bundled sub-resources (VPC Connector, Auto Scaling Configuration) based on your spec, so you get a complete deployment from a single manifest.

Credentials and region live in stack input (`provider_credential`), not in the spec.

## References

- AWS App Runner: https://docs.aws.amazon.com/apprunner/latest/dg/what-is-apprunner.html
- App Runner pricing: https://aws.amazon.com/apprunner/pricing/
- Supported runtimes: https://docs.aws.amazon.com/apprunner/latest/dg/service-source-code.html
- VPC Connector: https://docs.aws.amazon.com/apprunner/latest/dg/network-vpc.html
- Auto scaling: https://docs.aws.amazon.com/apprunner/latest/dg/manage-autoscaling.html
- Health checks: https://docs.aws.amazon.com/apprunner/latest/dg/manage-configure-healthcheck.html
- Encryption: https://docs.aws.amazon.com/apprunner/latest/dg/security-data-protection-encryption.html
