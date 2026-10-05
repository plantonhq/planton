# AwsAppRunnerService

AWS App Runner is a fully managed service that makes it easy to deploy containerized web applications and APIs at scale, with no infrastructure to manage. You give it a container image or a source repository, and App Runner handles building, deploying, scaling, and load-balancing -- producing an HTTPS endpoint in minutes. It is the simplest path from container to production URL on AWS.

## When to use App Runner

| Use case | Best fit | Why |
| --- | --- | --- |
| Stateless HTTP APIs and web apps that need zero-ops deployment | **App Runner** | No cluster, no task definitions, no load balancer to configure |
| Event-driven, short-lived functions (< 15 min) | Lambda | Pay-per-invocation, sub-second billing, broader event-source integrations |
| Long-running services that need full control of networking, sidecars, or service mesh | ECS Fargate | Full task-definition control, service discovery, service connect |
| Kubernetes workloads or teams already invested in K8s tooling | EKS | Full Kubernetes API, Helm charts, GitOps |
| Batch or GPU workloads | ECS / EKS | App Runner does not support GPU instance types or batch scheduling |

**Rule of thumb:** If your workload is a stateless HTTP service and you want AWS to own the infrastructure decisions, start with App Runner. Move to ECS Fargate or EKS when you need capabilities App Runner does not expose (custom networking topologies, sidecars, GPU, gRPC passthrough, etc.).

## The App Runner family

The service composes with three shared, versioned companion resources -- each its own first-class kind, referenced by ARN and shared across any number of services:

| Companion | Kind | What it tunes |
| --- | --- | --- |
| Auto scaling configuration | `AwsAppRunnerAutoScalingConfiguration` | Concurrency-based instance scaling (warm floor, ceiling, requests per instance) |
| VPC connector | `AwsAppRunnerVpcConnector` | Outbound access into a VPC (databases, caches, internal APIs) |
| Observability configuration | `AwsAppRunnerObservabilityConfiguration` | X-Ray distributed tracing (the reference itself is the enable switch) |

## Spec fields

### Top-level (`AwsAppRunnerServiceSpec`)

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `image_source` | `AwsAppRunnerServiceImageSource` | -- | Deploy from a container image in ECR or ECR Public. **Exactly one** of `image_source` or `code_source` must be set. |
| `code_source` | `AwsAppRunnerServiceCodeSource` | -- | Deploy from a source repository. **Exactly one** of `image_source` or `code_source` must be set. |
| `port` | `string` | `"8080"` | Port the application listens on inside the container. App Runner routes inbound HTTPS (443) to this port. |
| `start_command` | `string` | -- | Override the container's ENTRYPOINT/CMD (image source) or define the start command (code source with `configuration_source=API`). |
| `environment_variables` | `map<string,string>` | -- | Plaintext environment variables. Keys prefixed with `AWSAPPRUNNER` are reserved. |
| `environment_secrets` | `map<string,string>` | -- | Secret environment variables. Values are ARNs of Secrets Manager secrets or SSM Parameter Store parameters. The `instance_role_arn` must grant read access. |
| `cpu` | `string` | `"1024"` | CPU per instance. Numeric (`"256"`, `"512"`, `"1024"`, `"2048"`, `"4096"`) or human-readable (`"0.25 vCPU"` .. `"4 vCPU"`). |
| `memory` | `string` | `"2048"` | Memory per instance in MB. Numeric (`"512"` .. `"12288"`) or human-readable (`"0.5 GB"` .. `"12 GB"`). Not all CPU/memory combos are valid -- the create call rejects invalid pairs. |
| `instance_role_arn` | `StringValueOrRef` | -- | IAM role assumed by running instances to call AWS APIs (S3, DynamoDB, etc.). **Not** the role used to pull images. |
| `health_check` | `AwsAppRunnerServiceHealthCheck` | TCP defaults | Health check configuration. See nested message below. |
| `auto_scaling_configuration_arn` | `StringValueOrRef` | account default | ARN of an `AwsAppRunnerAutoScalingConfiguration` revision. Omitted, AWS applies the account's default configuration. One-way on a live service: removing the reference does NOT return to the account default (provider contract) -- reference the default's ARN explicitly instead. |
| `vpc_connector_arn` | `StringValueOrRef` | -- | ARN of an `AwsAppRunnerVpcConnector` for outbound VPC access. Omitted, egress reaches public endpoints only. |
| `observability_configuration_arn` | `StringValueOrRef` | -- | ARN of an `AwsAppRunnerObservabilityConfiguration`. The reference itself enables tracing. One-way on a live service (upstream provider gap): removing the reference does NOT disable tracing -- disabling requires service replacement. |
| `is_publicly_accessible` | `bool` | `true` | Whether the service endpoint is publicly reachable. When `false`, pair with `vpc_ingress_connections` -- otherwise nothing can reach the service. |
| `vpc_ingress_connections` | `repeated AwsAppRunnerServiceVpcIngressConnection` | -- | Publish the service into VPCs through interface VPC endpoints (AWS PrivateLink), keyed by connection name. Per-connection private domain names surface in outputs. |
| `ip_address_type` | `string` | `"IPV4"` | `"IPV4"` or `"DUAL_STACK"` (IPv4 + IPv6). |
| `kms_key_arn` | `StringValueOrRef` | AWS-managed key | Customer-managed KMS key for encrypting stored source and logs. **ForceNew** -- changing this replaces the service. |
| `auto_deployments_enabled` | `bool` | `false` | Automatically deploy when the tracked source changes (new image on the tag, new commit on the branch). Supported for private ECR and code repositories only -- **AWS rejects it for ECR_PUBLIC images** (validated at spec level). |
| `custom_domains` | `repeated AwsAppRunnerServiceCustomDomain` | -- | Custom domains associated with the service, keyed by domain name. Per-domain certificate-validation records surface in outputs. |
| `web_acl_arn` | `StringValueOrRef` | -- | ARN of a REGIONAL `AwsWafWebAcl` to associate; all requests pass WAF inspection first. |

### `AwsAppRunnerServiceImageSource`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `image_identifier` | `string` | yes | Full image URI with tag or digest. ECR: `ACCOUNT.dkr.ecr.REGION.amazonaws.com/REPO:TAG`. ECR Public: `public.ecr.aws/ALIAS/REPO:TAG`. Stays a literal string: it carries a repository-plus-tag coordinate no single upstream output represents. |
| `image_repository_type` | `string` | yes | `"ECR"` (private) or `"ECR_PUBLIC"`. |
| `access_role_arn` | `StringValueOrRef` | for ECR | IAM role that grants App Runner permission to pull from private ECR. Must be assumable by `build.apprunner.amazonaws.com`; the AWS-managed `AWSAppRunnerServicePolicyForECRAccess` policy covers the permissions. |

### `AwsAppRunnerServiceCodeSource`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `repository_url` | `string` | yes | Repository URL (e.g., `https://github.com/owner/repo`). |
| `branch` | `string` | yes | Branch to deploy from (e.g., `main`). |
| `source_directory` | `string` | no | Subdirectory containing the app source. Defaults to repo root. Useful for monorepos. |
| `connection_arn` | `StringValueOrRef` | yes | ARN of an App Runner connection authorizing repository access. Created out-of-band (requires a one-time OAuth handshake in the AWS console); shared across services. |
| `configuration_source` | `string` | yes | `"API"` (build config in this spec) or `"REPOSITORY"` (reads `apprunner.yaml` from the repo). |
| `runtime` | `string` | for API | Managed runtime. Values: `PYTHON_3`, `PYTHON_311`, `NODEJS_12`/`14`/`16`/`18`/`22`, `CORRETTO_8`/`11`, `GO_1`, `DOTNET_6`, `PHP_81`, `RUBY_31`. |
| `build_command` | `string` | for API | Shell command to build the app (e.g., `npm ci && npm run build`). |

### `AwsAppRunnerServiceHealthCheck`

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `protocol` | `string` | `"TCP"` | `"TCP"` (port-open check) or `"HTTP"` (GET request expecting success). |
| `path` | `string` | `"/"` | URL path for HTTP health checks. Ignored for TCP. |
| `interval` | `int32` | `5` | Seconds between checks (1--20). |
| `timeout` | `int32` | `2` | Max seconds to wait for a response (1--20). |
| `healthy_threshold` | `int32` | `1` | Consecutive successes to mark healthy (1--20). |
| `unhealthy_threshold` | `int32` | `5` | Consecutive failures to mark unhealthy and replace (1--20). |

### `AwsAppRunnerServiceCustomDomain`

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `domain_name` | `string` | -- | The domain to associate (e.g., `app.example.com`). |
| `enable_www_subdomain` | `bool` | `true` | Also associate the `www.` subdomain -- meaningful mainly for apex domains. |

### `AwsAppRunnerServiceVpcIngressConnection`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `name` | `string` | yes | Connection name (4--40 chars), unique across active VPC Ingress Connections in the account/region. Entries are keyed by name; changing an entry's VPC or endpoint replaces that one connection. |
| `vpc_id` | `StringValueOrRef` | yes | The VPC to publish the service into. |
| `vpc_endpoint_id` | `StringValueOrRef` | yes | The interface VPC endpoint (for `com.amazonaws.REGION.apprunner.requests`) in that VPC that carries the traffic. |

## Outputs

| Output | Description |
| --- | --- |
| `service_arn` | Full ARN of the App Runner service -- the join key for IAM policies and deployment triggers. |
| `service_id` | Unique identifier assigned by App Runner. |
| `service_url` | Default HTTPS endpoint (scheme-less, e.g., `abc123.us-east-1.awsapprunner.com`). |
| `service_name` | Service name (derived from metadata). |
| `service_status` | Operational status at the end of deployment (`RUNNING` when serving). |
| `custom_domains` | Per-domain DNS material: the `dns_target` to point each domain at, plus certificate-validation CNAMEs -- each composes into an `AwsRoute53DnsRecord`. |
| `vpc_ingress_connections` | Per-connection identifiers: the private `domain_name` clients inside each connected VPC resolve, plus the connection ARN and status. |

## Prerequisites

1. **AWS credentials** -- Provided via IaC input, not in the spec.
2. **IAM access role** (private ECR only) -- assumable by `build.apprunner.amazonaws.com` with ECR read permissions. Pass its ARN as `image_source.access_role_arn`.
3. **IAM instance role** (optional) -- If your application calls AWS APIs at runtime (S3, DynamoDB, Secrets Manager, X-Ray), create a role with the necessary policies and pass it as `instance_role_arn`.
4. **App Runner connection** (code source only) -- Created via the AWS console (one-time OAuth handshake); shared across services.
5. **VPC connector** (optional) -- an `AwsAppRunnerVpcConnector` when the service must reach VPC resources.
6. **KMS key** (optional) -- Only for customer-managed encryption. The key must allow the App Runner service principal.

## Quick start

Deploy a public sample container with a single manifest:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAppRunnerService
metadata:
  name: my-web-app
spec:
  region: us-east-1
  imageSource:
    imageIdentifier: "public.ecr.aws/aws-containers/hello-app-runner:latest"
    imageRepositoryType: "ECR_PUBLIC"
  port: "8000"
```

## Custom domains

Associate domains directly on the spec; App Runner issues and renews the TLS certificates. Prove ownership by creating the exported validation CNAMEs (and a CNAME/alias from each domain to the exported `dns_target`):

```yaml
customDomains:
  - domainName: app.example.com
    enableWwwSubdomain: false
```

The per-domain records surface in `status.outputs.custom_domains` and compose directly into `AwsRoute53DnsRecord` resources.

## References

- AWS App Runner: https://docs.aws.amazon.com/apprunner/latest/dg/what-is-apprunner.html
- App Runner pricing: https://aws.amazon.com/apprunner/pricing/
- Supported runtimes: https://docs.aws.amazon.com/apprunner/latest/dg/service-source-code.html
- VPC connectors: https://docs.aws.amazon.com/apprunner/latest/dg/network-vpc.html
- Auto scaling: https://docs.aws.amazon.com/apprunner/latest/dg/manage-autoscaling.html
- Health checks: https://docs.aws.amazon.com/apprunner/latest/dg/manage-configure-healthcheck.html
- Custom domains: https://docs.aws.amazon.com/apprunner/latest/dg/manage-custom-domains.html

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
