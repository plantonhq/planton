# AwsAppRunnerService

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsAppRunnerServiceSpec defines the desired state of an AWS App Runner
service.

App Runner is a fully managed container application service: give it a
container image or a source repository and it handles build, deploy, TLS,
load balancing, and concurrency-based auto scaling. It is the shortest
path from container image to HTTPS endpoint on AWS.

The service supports two deployment source types (exactly one must be
provided):
  - image_source: deploy a container image from Amazon ECR (private) or
    ECR Public Gallery.
  - code_source: deploy from a source repository; App Runner builds and
    runs the code through a managed runtime.

Shared, versioned companion resources compose by reference rather than
being embedded here -- each is its own first-class resource shared across
any number of services:
  - AwsAppRunnerAutoScalingConfiguration (auto_scaling_configuration_arn)
    tunes concurrency-based scaling; omitted, AWS applies its account
    default configuration.
  - AwsAppRunnerVpcConnector (vpc_connector_arn) routes OUTBOUND traffic
    into a VPC; omitted, egress goes to the public internet only.
  - AwsAppRunnerObservabilityConfiguration
    (observability_configuration_arn) enables X-Ray request tracing;
    omitted, tracing is off.
INBOUND private access is service-owned, not shared, so it is modeled
inline: vpc_ingress_connections publishes the service into named VPCs
through interface VPC endpoints (AWS PrivateLink).

ForceNew honesty (changing these replaces the service):
  - kms_key_arn (the stored-source encryption key)
  - the service name (derived from metadata.name, not a spec field)
Everything else updates in place; source changes roll a new deployment.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAppRunnerService
metadata:
  name: apprunner-demo
  org: test-org
  env: dev
  id: apprunner-demo-dev
spec:
  region: us-west-2
  imageSource:
    imageIdentifier: "public.ecr.aws/aws-containers/hello-app-runner:latest"
    imageRepositoryType: "ECR_PUBLIC"
  port: "8000"
  cpu: "1024"
  memory: "2048"
  environmentVariables:
    EXAMPLE: "true"
  healthCheck:
    protocol: "HTTP"
    path: "/"
    interval: 10
    timeout: 5
  customDomains:
    - domainName: "app.example.com"
      enableWwwSubdomain: false
  isPubliclyAccessible: false
  vpcIngressConnections:
    - name: "apprunner-demo-private"
      vpcId:
        value: "vpc-0abc123def4567890"
      vpcEndpointId:
        value: "vpce-0abc123def4567890"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.imageSource` | `AwsAppRunnerServiceImageSource` |  |  |  |
| `spec.imageSource.imageIdentifier` | `string` | yes |  |  |
| `spec.imageSource.imageRepositoryType` | `string` | yes |  |  |
| `spec.imageSource.accessRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.codeSource` | `AwsAppRunnerServiceCodeSource` |  |  |  |
| `spec.codeSource.repositoryUrl` | `string` | yes |  |  |
| `spec.codeSource.branch` | `string` | yes |  |  |
| `spec.codeSource.sourceDirectory` | `string` |  |  |  |
| `spec.codeSource.connectionArn` | `string \| valueFrom` | yes |  |  |
| `spec.codeSource.configurationSource` | `string` | yes |  |  |
| `spec.codeSource.runtime` | `string` |  |  |  |
| `spec.codeSource.buildCommand` | `string` |  |  |  |
| `spec.port` | `string` |  | `8080` |  |
| `spec.startCommand` | `string` |  |  |  |
| `spec.environmentVariables` | `map<string, string>` |  |  |  |
| `spec.environmentSecrets` | `map<string, string>` |  |  |  |
| `spec.cpu` | `string` |  | `1024` |  |
| `spec.memory` | `string` |  | `2048` |  |
| `spec.instanceRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.healthCheck` | `AwsAppRunnerServiceHealthCheck` |  |  |  |
| `spec.healthCheck.protocol` | `string` |  | `TCP` |  |
| `spec.healthCheck.path` | `string` |  | `/` |  |
| `spec.healthCheck.interval` | `int32` |  | `5` |  |
| `spec.healthCheck.timeout` | `int32` |  | `2` |  |
| `spec.healthCheck.healthyThreshold` | `int32` |  | `1` |  |
| `spec.healthCheck.unhealthyThreshold` | `int32` |  | `5` |  |
| `spec.autoScalingConfigurationArn` | `string \| valueFrom` |  |  | AwsAppRunnerAutoScalingConfiguration (`status.outputs.configuration_arn`) |
| `spec.vpcConnectorArn` | `string \| valueFrom` |  |  | AwsAppRunnerVpcConnector (`status.outputs.vpc_connector_arn`) |
| `spec.observabilityConfigurationArn` | `string \| valueFrom` |  |  | AwsAppRunnerObservabilityConfiguration (`status.outputs.configuration_arn`) |
| `spec.isPubliclyAccessible` | `bool` |  | `true` |  |
| `spec.vpcIngressConnections` | `[]AwsAppRunnerServiceVpcIngressConnection` |  |  |  |
| `spec.vpcIngressConnections[].name` | `string` | yes |  |  |
| `spec.vpcIngressConnections[].vpcId` | `string \| valueFrom` | yes |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.vpcIngressConnections[].vpcEndpointId` | `string \| valueFrom` | yes |  | AwsVpcEndpoint (`status.outputs.vpc_endpoint_id`) |
| `spec.ipAddressType` | `string` |  | `IPV4` |  |
| `spec.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.autoDeploymentsEnabled` | `bool` |  |  |  |
| `spec.customDomains` | `[]AwsAppRunnerServiceCustomDomain` |  |  |  |
| `spec.customDomains[].domainName` | `string` | yes |  |  |
| `spec.customDomains[].enableWwwSubdomain` | `bool` |  | `true` |  |
| `spec.webAclArn` | `string \| valueFrom` |  |  | AwsWafWebAcl (`status.outputs.web_acl_arn`) |

## Field Details

### spec.region

`string` · required

The AWS region where the App Runner service will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.imageSource

`AwsAppRunnerServiceImageSource`

Image-based deployment source. Use this to deploy a container image
stored in Amazon ECR (private) or ECR Public Gallery.

- rule: image_repository_type must be 'ECR' (private registry) or 'ECR_PUBLIC' (public gallery)
- rule: access_role_arn is required when image_repository_type is 'ECR' -- App Runner needs an IAM role to pull from a private registry
- rule: image_identifier must be a private ECR path (ACCOUNT_ID.dkr.ecr.REGION.amazonaws.com/REPO:TAG) or an ECR Public Gallery path (public.ecr.aws/ALIAS/REPO:TAG)

### spec.imageSource.imageIdentifier

`string` · required

Full container image identifier including tag or digest.
ECR format: "ACCOUNT_ID.dkr.ecr.REGION.amazonaws.com/REPO:TAG"
ECR Public format: "public.ecr.aws/ALIAS/REPO:TAG"
This stays a literal string (not a reference) because it carries a
repository-plus-tag coordinate no single upstream output represents.
The format rule below is AWS's own ImageIdentifier pattern (mirrored
by the provider): a 12-digit-account private ECR host path, or a
public.ecr.aws gallery path.

- rule: {"string":{"minLen":"1"}}

### spec.imageSource.imageRepositoryType

`string` · required

Type of image repository.
"ECR": private Amazon ECR (requires access_role_arn for pull access).
"ECR_PUBLIC": public ECR Gallery (no pull authentication; automatic
deployments are not supported for public images).

- rule: {"string":{"minLen":"1"}}

### spec.imageSource.accessRoleArn

`string | valueFrom`

IAM role ARN that grants App Runner permission to pull images from
private ECR. Required when image_repository_type is "ECR"; not used for
"ECR_PUBLIC". The role must be assumable by build.apprunner.amazonaws.com
and carry ECR read permissions (the AWS-managed
AWSAppRunnerServicePolicyForECRAccess policy covers it).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.codeSource

`AwsAppRunnerServiceCodeSource`

Code-based deployment source. Use this to deploy from a source
repository through an App Runner connection; App Runner clones the
repository, builds the application with the selected managed runtime,
and deploys the resulting container.

- rule: configuration_source must be 'API' (configure the build in this spec) or 'REPOSITORY' (read apprunner.yaml from the repository)
- rule: runtime is required when configuration_source is 'API' -- pick the managed runtime App Runner should build with
- rule: runtime must be one of App Runner's managed runtimes: PYTHON_3, PYTHON_311, NODEJS_12, NODEJS_14, NODEJS_16, NODEJS_18, NODEJS_22, CORRETTO_8, CORRETTO_11, GO_1, DOTNET_6, PHP_81, RUBY_31

### spec.codeSource.repositoryUrl

`string` · required

Repository URL (e.g. "https://github.com/owner/repo").

- rule: {"string":{"minLen":"1"}}

### spec.codeSource.branch

`string` · required

Branch to deploy from (e.g. "main", "production"). App Runner tracks
this branch; with auto_deployments_enabled, every push deploys.

- rule: {"string":{"minLen":"1"}}

### spec.codeSource.sourceDirectory

`string`

Subdirectory within the repository containing the application source.
Defaults to the repository root. Useful for monorepos where the service
lives in a subfolder; with configuration_source="REPOSITORY", the
apprunner.yaml is read from this directory. One-way on a live service:
clearing this does not return the build to the repository root (the
provider keeps the last-applied value) -- set "/" explicitly instead.

### spec.codeSource.connectionArn

`string | valueFrom` · required

ARN of an App Runner connection that authorizes repository access
(GitHub or Bitbucket). Connections require a one-time OAuth handshake
completed in the AWS console, so they are created out-of-band and
referenced here by ARN; one connection is shared across services.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.codeSource.configurationSource

`string` · required

Where App Runner reads the build/runtime configuration from.
"API": configuration comes from this spec (runtime, build_command,
  port, start_command).
"REPOSITORY": configuration is read from an apprunner.yaml at the
  repository root (or source_directory); runtime and build_command in
  this spec are ignored.

- rule: {"string":{"minLen":"1"}}

### spec.codeSource.runtime

`string`

Managed runtime that builds and runs the application. Required when
configuration_source is "API", ignored for "REPOSITORY".

### spec.codeSource.buildCommand

`string`

Shell command that builds the application (e.g. "npm ci && npm run
build", "pip install -r requirements.txt"). Only used when
configuration_source is "API".

### spec.port

`string` · optional (explicit presence)

Port the application listens on inside the container. App Runner
terminates TLS on 443 and forwards requests to this port.

- default: `8080`

### spec.startCommand

`string`

Override the container start command. For image_source this overrides
the image ENTRYPOINT/CMD; for code_source with
configuration_source="API" it is the command that starts the built
application. One-way on a live service: the provider drops empty
values on update, so clearing this does not restore the image's own
ENTRYPOINT/CMD -- set the desired command explicitly instead.

### spec.environmentVariables

`map<string, string>`

Environment variables injected into every instance at runtime. Keys are
variable names, values are plaintext strings. Never put secret values
here -- use environment_secrets for anything sensitive. Keys prefixed
with "AWSAPPRUNNER" are reserved by the service.

### spec.environmentSecrets

`map<string, string>`

Environment secrets injected at runtime. Keys are variable names;
values are full ARNs of AWS Secrets Manager secrets or SSM Parameter
Store parameters -- App Runner resolves each ARN at deploy time and
injects the resolved value as an environment variable. The
instance_role_arn role must be allowed to read the referenced secrets
(secretsmanager:GetSecretValue / ssm:GetParameters).

### spec.cpu

`string` · optional (explicit presence)

CPU allocation per instance. Accepts millicore strings or
human-readable vCPU format.
Numeric: "256", "512", "1024", "2048", "4096"
Human-readable: "0.25 vCPU", "0.5 vCPU", "1 vCPU", "2 vCPU", "4 vCPU"

- default: `1024`

### spec.memory

`string` · optional (explicit presence)

Memory allocation per instance. Accepts megabyte strings or
human-readable GB format.
Numeric: "512", "1024", "2048", "3072", "4096", "6144", "8192",
"10240", "12288"
Human-readable: "0.5 GB", "1 GB", "2 GB", "3 GB", "4 GB", "6 GB",
"8 GB", "10 GB", "12 GB"
Not every CPU/memory pairing is valid -- App Runner accepts specific
combinations (e.g. 4 vCPU requires 8-12 GB); the create call rejects
invalid pairs.

- default: `2048`

### spec.instanceRoleArn

`string | valueFrom`

IAM role that service instances assume at runtime to call AWS APIs --
the role your application code uses (reading S3, writing DynamoDB,
resolving environment_secrets). This is NOT the image-pull role (that
is image_source.access_role_arn). One-way on a live service: the
provider drops empty values on update, so REMOVING the role does not
detach it from a running service -- attach a replacement role instead.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.healthCheck

`AwsAppRunnerServiceHealthCheck`

Health check configuration App Runner uses to monitor instance
readiness; unhealthy instances are replaced automatically. When
omitted, App Runner performs TCP checks on the configured port with
AWS defaults. Removing the block from a live service stops managing
the check but does not reset it -- to return to defaults, set the
default values explicitly (protocol TCP, interval 5, timeout 2,
thresholds 1/5).

- rule: health check protocol must be 'TCP' (port check) or 'HTTP' (GET request to path)

### spec.healthCheck.protocol

`string` · optional (explicit presence)

Health check protocol.
"TCP": checks the port accepts connections (default).
"HTTP": sends GET requests to path and expects a 2xx response.

- default: `TCP`

### spec.healthCheck.path

`string` · optional (explicit presence)

URL path for HTTP health checks (e.g. "/health", "/readyz"). Ignored
when protocol is "TCP".

- default: `/`

### spec.healthCheck.interval

`int32` · optional (explicit presence)

Seconds between consecutive health checks.

- default: `5`
- rule: {"int32":{"lte":20,"gte":1}}

### spec.healthCheck.timeout

`int32` · optional (explicit presence)

Seconds to wait for a health check response before counting the check
as failed.

- default: `2`
- rule: {"int32":{"lte":20,"gte":1}}

### spec.healthCheck.healthyThreshold

`int32` · optional (explicit presence)

Consecutive successful checks required to mark an instance healthy.

- default: `1`
- rule: {"int32":{"lte":20,"gte":1}}

### spec.healthCheck.unhealthyThreshold

`int32` · optional (explicit presence)

Consecutive failed checks before an instance is marked unhealthy and
replaced.

- default: `5`
- rule: {"int32":{"lte":20,"gte":1}}

### spec.autoScalingConfigurationArn

`string | valueFrom`

ARN of an AwsAppRunnerAutoScalingConfiguration revision that governs
concurrency-based scaling. When omitted, AWS applies the account's
default auto scaling configuration (1 min / 25 max / 100 concurrency
unless the account default was changed). The referenced ARN carries a
revision, so registering a new revision rolls this service on its next
deployment. One-way on a live service: once set, REMOVING this
reference does not return the service to the account default -- the
provider keeps the last-applied configuration (its attribute is
Optional+Computed, so removal produces no change). To move back to the
default, reference the default configuration's ARN explicitly.

- references: AwsAppRunnerAutoScalingConfiguration (`status.outputs.configuration_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsAppRunnerAutoScalingConfiguration, name: <that resource's name>, fieldPath: status.outputs.configuration_arn}} -- a bare string does not parse

### spec.vpcConnectorArn

`string | valueFrom`

ARN of an AwsAppRunnerVpcConnector for outbound VPC access -- lets the
service reach private resources (databases, caches, internal APIs)
inside a VPC. When omitted, egress uses App Runner's default public
path and the service can only reach public endpoints. One connector is
shared by any number of services.

- references: AwsAppRunnerVpcConnector (`status.outputs.vpc_connector_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsAppRunnerVpcConnector, name: <that resource's name>, fieldPath: status.outputs.vpc_connector_arn}} -- a bare string does not parse

### spec.observabilityConfigurationArn

`string | valueFrom`

ARN of an AwsAppRunnerObservabilityConfiguration revision. When set,
the service sends request traces to the configured vendor (AWS X-Ray);
when omitted at creation, tracing is off. Presence of the reference IS
the enable switch -- there is no separate toggle to keep in sync.
One-way on a live service (upstream provider gap): once tracing is
enabled, REMOVING this reference does not disable it -- the provider
sends nothing for an absent block, so AWS keeps tracing on and the
plan never converges. To genuinely disable tracing today, replace the
service (or disable it in the AWS console) until the provider gains a
disable path.

- references: AwsAppRunnerObservabilityConfiguration (`status.outputs.configuration_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsAppRunnerObservabilityConfiguration, name: <that resource's name>, fieldPath: status.outputs.configuration_arn}} -- a bare string does not parse

### spec.isPubliclyAccessible

`bool` · optional (explicit presence)

Whether the service endpoint is publicly reachable from the internet.
When true (default), the service gets a public HTTPS URL. When false,
the endpoint is reachable only from within VPCs that attach a VPC
Ingress Connection to this service -- declare those below in
vpc_ingress_connections.

- default: `true`

### spec.vpcIngressConnections

`[]AwsAppRunnerServiceVpcIngressConnection`

VPC Ingress Connections for this service -- each entry publishes the
service into one VPC through an interface VPC endpoint
(AWS PrivateLink), so clients inside that VPC reach the service
privately at the exported per-connection domain name. Pair with
is_publicly_accessible: false for a service reachable ONLY from inside
the named VPCs. Keyed by name: adding or removing entries updates in
place; changing an entry's VPC or endpoint replaces that one
connection.

### spec.vpcIngressConnections[].name

`string` · required

Name for this VPC Ingress Connection. Must be unique across all active
VPC Ingress Connections in the AWS account and region (it is the AWS
resource name, not a label). App Runner family names are 4-40
characters.

- rule: {"string":{"minLen":"4","maxLen":"40"}}

### spec.vpcIngressConnections[].vpcId

`string | valueFrom` · required

The VPC to publish the service into.

Containment-exempt: the connection lets clients INSIDE this VPC
reach the service over PrivateLink; the service itself runs on
App Runner's managed infrastructure and is never deployed into the
VPC (a VPC connector, not an ingress connection, is what places a
service's egress in a VPC). On a diagram the service stands outside
the VPC with a line to the network it is published into.

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.vpcIngressConnections[].vpcEndpointId

`string | valueFrom` · required

The interface VPC endpoint in that VPC that carries the traffic. AWS
requires both members -- an ingress connection cannot exist without its
endpoint (the provider leaves both optional and lets the create call
fail server-side; this spec enforces AWS's contract up front).

- references: AwsVpcEndpoint (`status.outputs.vpc_endpoint_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpcEndpoint, name: <that resource's name>, fieldPath: status.outputs.vpc_endpoint_id}} -- a bare string does not parse

### spec.ipAddressType

`string` · optional (explicit presence)

IP address type for the service endpoint.
"IPV4": IPv4-only (default). "DUAL_STACK": IPv4 + IPv6.

- default: `IPV4`

### spec.kmsKeyArn

`string | valueFrom`

Customer-managed KMS key ARN App Runner uses to encrypt its stored copy
of the deployment source (image or repository archive) and build logs.
When omitted, App Runner uses an AWS-managed key. ForceNew: changing
this replaces the service.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.autoDeploymentsEnabled

`bool`

Whether App Runner automatically starts a deployment when the source
changes (a new image pushed to the tracked tag, or a new commit on the
tracked branch). Disabled by default: deployments then happen only when
this resource is applied or a deployment is started explicitly, which
keeps rollouts deterministic and graph-driven. AWS supports automatic
deployments only for private same-account ECR and code repositories --
ECR Public images cannot enable this (AWS rejects the create call).
The modules always send this value explicitly: AWS's own default is
conditional on the source type (on for code repos and same-account
ECR, off otherwise), and the provider substitutes an unconditional
"on" for any omitted value -- explicit send is the only deterministic
path through both layers.

### spec.customDomains

`[]AwsAppRunnerServiceCustomDomain`

Custom domains associated with this service, keyed by domain name. App
Runner provisions and renews the TLS certificate for each domain; you
prove ownership by creating the CNAME records exported per domain in
status.outputs.custom_domains (compose them into AwsRoute53DnsRecord
resources), plus a CNAME/alias from the domain to the exported
dns_target. Adding or removing entries updates in place; changing an
entry replaces that one association.

### spec.customDomains[].domainName

`string` · required

The domain to associate (e.g. "app.example.com" or "example.com").

- rule: {"string":{"minLen":"1","maxLen":"255"}}

### spec.customDomains[].enableWwwSubdomain

`bool` · optional (explicit presence)

Whether to also associate the "www." subdomain of domain_name.
AWS defaults this to true; meaningful mainly for apex domains.

- default: `true`

### spec.webAclArn

`string | valueFrom`

ARN of a REGIONAL AwsWafWebAcl to associate with this service. All
requests pass WAF inspection before reaching the application. When
omitted, no WAF is attached.

- references: AwsWafWebAcl (`status.outputs.web_acl_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsWafWebAcl, name: <that resource's name>, fieldPath: status.outputs.web_acl_arn}} -- a bare string does not parse

## Validation Rules

- `exactly_one_source`: provide exactly one deployment source -- either image_source (deploy a container image) or code_source (build from a repository), not both
- `no_auto_deploy_for_ecr_public`: auto_deployments_enabled is not supported for ECR_PUBLIC images -- AWS can only watch private ECR repositories and code repositories for changes; deploy public images explicitly instead
- `valid_ip_address_type`: ip_address_type must be 'IPV4' or 'DUAL_STACK'
- `valid_cpu`: cpu must be one of: '256','512','1024','2048','4096','0.25 vCPU','0.5 vCPU','1 vCPU','2 vCPU','4 vCPU'
- `valid_memory`: memory must be one of: '512','1024','2048','3072','4096','6144','8192','10240','12288','0.5 GB','1 GB','2 GB','3 GB','4 GB','6 GB','8 GB','10 GB','12 GB'

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsAppRunnerService, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.service_arn` | `string` | The full ARN of the App Runner service (e.g. "arn:aws:apprunner: us-west-2:123456789012:service/my-api/abc123"). The handle IAM policies, VPC Ingress Connections, and deployment triggers reference. |
| `status.outputs.service_id` | `string` | The service ID -- the AWS-assigned identifier unique within the account and region (the trailing component of the ARN). |
| `status.outputs.service_url` | `string` | The default HTTPS endpoint of the service (e.g. "abc123.us-west-2.awsapprunner.com", without scheme). For private services this is the domain a VPC Ingress Connection resolves to. |
| `status.outputs.service_name` | `string` | The service name (metadata.name) the service was created under. |
| `status.outputs.service_status` | `string` | The service's lifecycle status at the end of the deployment ("RUNNING" when serving traffic). |
| `status.outputs.custom_domains` | `[]AwsAppRunnerServiceCustomDomainOutput` | Per-domain DNS material for the associated custom_domains -- one entry per spec entry. Create the certificate-validation CNAMEs (and a CNAME or alias from the domain to dns_target) in your DNS -- each record composes directly into an AwsRoute53DnsRecord resource. Empty when the spec associates no custom domains. |
| `status.outputs.custom_domains[].domain_name` | `string` | The associated domain (matches the spec entry's domain_name). |
| `status.outputs.custom_domains[].dns_target` | `string` | The App Runner subdomain to point the custom domain at -- create a CNAME (subdomains) or ALIAS (apex) from domain_name to this target. |
| `status.outputs.custom_domains[].status` | `string` | The association status at the end of the deployment. The resting state right after creation is "pending_certificate_dns_validation" -- the association completes on its own once the validation records below are resolvable in DNS. |
| `status.outputs.custom_domains[].certificate_validation_records` | `[]AwsAppRunnerServiceCertificateValidationRecord` | The certificate-validation CNAME records proving domain ownership. Keep them in place after validation so App Runner can renew the certificate automatically. |
| `status.outputs.custom_domains[].certificate_validation_records[].record_name` | `string` | The DNS record name to create (a "_<hash>.<domain>." CNAME). |
| `status.outputs.custom_domains[].certificate_validation_records[].record_type` | `string` | The DNS record type (always "CNAME" today). |
| `status.outputs.custom_domains[].certificate_validation_records[].record_value` | `string` | The DNS record value the name must resolve to. |
| `status.outputs.vpc_ingress_connections` | `[]AwsAppRunnerServiceVpcIngressConnectionOutput` | Per-connection identifiers for the spec's vpc_ingress_connections -- one entry per spec entry. The domain_name is what clients inside the connected VPC resolve to reach the service privately. Empty when the spec declares no ingress connections. |
| `status.outputs.vpc_ingress_connections[].name` | `string` | The connection name (matches the spec entry's name). |
| `status.outputs.vpc_ingress_connections[].arn` | `string` | The ARN of the VPC Ingress Connection resource. |
| `status.outputs.vpc_ingress_connections[].domain_name` | `string` | The domain name clients inside the connected VPC use to reach the service (resolves through the interface VPC endpoint). |
| `status.outputs.vpc_ingress_connections[].status` | `string` | The connection status at the end of the deployment ("AVAILABLE" when serving). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.imageSource.accessRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.instanceRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.autoScalingConfigurationArn` | AwsAppRunnerAutoScalingConfiguration | `status.outputs.configuration_arn` |
| `spec.vpcConnectorArn` | AwsAppRunnerVpcConnector | `status.outputs.vpc_connector_arn` |
| `spec.observabilityConfigurationArn` | AwsAppRunnerObservabilityConfiguration | `status.outputs.configuration_arn` |
| `spec.vpcIngressConnections[].vpcId` | AwsVpc | `status.outputs.vpc_id` |
| `spec.vpcIngressConnections[].vpcEndpointId` | AwsVpcEndpoint | `status.outputs.vpc_endpoint_id` |
| `spec.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.webAclArn` | AwsWafWebAcl | `status.outputs.web_acl_arn` |

## See Also

- [Overview](../README.md)
