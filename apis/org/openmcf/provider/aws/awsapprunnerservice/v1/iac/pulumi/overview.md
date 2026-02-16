# Pulumi Module Architecture: AWS App Runner Service

## Overview

This Pulumi module provisions AWS App Runner services through a declarative, protobuf-defined specification. App Runner is a fully managed container application service that deploys web applications and APIs from a container image or source code repository without managing infrastructure. App Runner handles build, deploy, scale, and load balancing automatically.

The module follows OpenMCF's standard pattern: **input transformation → resource provisioning → output extraction**, with special handling for App Runner's dual source types (image vs code), VPC Connector creation, and Auto Scaling Configuration management.

## Module Structure

```
iac/pulumi/
├── main.go              # Pulumi entrypoint (loads stack input, calls module)
├── Pulumi.yaml          # Pulumi project metadata
├── Makefile             # Build/test helpers
├── debug.sh             # Delve debugging wrapper
└── module/
    ├── main.go          # Orchestration logic (provider setup, output export)
    ├── locals.go        # Input transformation and tag construction
    ├── service.go       # App Runner service resource implementation
    ├── vpc_connector.go # VPC Connector resource (conditional)
    ├── auto_scaling.go  # Auto Scaling Configuration resource (conditional)
    └── outputs.go       # Output key constants
```

### File Responsibilities

#### `main.go` (entrypoint)
- Unmarshals `AwsAppRunnerServiceStackInput` from Pulumi stack configuration
- Delegates to `module.Resources()` for actual provisioning
- Minimal logic—purely a thin entrypoint wrapper

**Standard Pulumi Pattern**: This file is identical across all AWS components, only varying in the protobuf type imported.

#### `module/main.go` (orchestrator)
**Key Function:** `Resources(ctx, stackInput)`

Responsibilities:
1. **Locals Initialization**: Transform stack input into typed locals struct
2. **Provider Configuration**: Handle two provider scenarios:
   - **Default**: Create provider with ambient AWS credentials (IAM role, environment variables)
   - **Explicit**: Create provider with credentials from `stackInput.ProviderConfig` (access key, secret key, session token)
3. **VPC Connector Creation**: Conditionally create inline VPC Connector when `subnet_ids` are provided
4. **Auto Scaling Configuration**: Conditionally create Auto Scaling Configuration Version when `auto_scaling` is provided
5. **Service Creation**: Invoke `service()` with initialized locals, provider, and created sub-resources
6. **Error Propagation**: Wrap errors with context for debugging

**Design Decision**: Provider handling supports both CI/CD environments (IRSA, instance profiles) and local development (explicit credentials).

#### `module/locals.go` (input transformer)
**Key Function:** `initializeLocals(ctx, stackInput)`

Transforms the protobuf `AwsAppRunnerServiceStackInput` into a strongly-typed `Locals` struct and constructs AWS resource tags.

**Tag Construction**: Every App Runner service receives six tags:
- `resource=true`: Marks this as a OpenMCF managed resource
- `organization=<org>`: Organization ID from metadata
- `environment=<env>`: Environment (dev/staging/prod) from metadata
- `resource-kind=AwsAppRunnerService`: CloudResourceKind enum string
- `resource-id=<id>`: Unique resource identifier from metadata
- `name=<name>`: Service name from metadata

**Why Tags Matter**: These tags enable:
- Cost allocation reporting by org/env/resource-kind (App Runner costs can be significant)
- Policy enforcement (e.g., "only services tagged env=prod can access production secrets")
- Resource discovery and inventory management
- CloudWatch Logs Insights filtering by resource metadata

#### `module/service.go` (resource implementation)
**Key Function:** `service(ctx, locals, provider, createdVpcConnector, createdAutoScaling)`

This is the core implementation file containing all App Runner service provisioning logic. It translates the protobuf `AwsAppRunnerServiceSpec` into Pulumi's App Runner resources.

**Implementation Pattern**: The module follows a **multi-phase construction** approach:

1. **Phase 1: Source Configuration**
   - **Image Source**: Extract image identifier, repository type, access role ARN (for private ECR)
   - **Code Source**: Extract repository URL, branch, connection ARN, configuration source (API vs REPOSITORY), runtime, build command
   - Build `ServiceSourceConfigurationArgs` with appropriate source type

2. **Phase 2: Image/Code Configuration**
   - For image source: Map port, start command, environment variables, environment secrets
   - For code source: Map runtime, build command, port, start command, environment variables/secrets (when `configuration_source=API`)

3. **Phase 3: Instance Configuration**
   - Extract CPU, memory, instance role ARN
   - Build `ServiceInstanceConfigurationArgs`

4. **Phase 4: Health Check Configuration**
   - Extract protocol (TCP/HTTP), path, interval, timeout, healthy/unhealthy thresholds
   - Build `ServiceHealthCheckConfigurationArgs` if provided

5. **Phase 5: Network Configuration**
   - **Egress**: Attach VPC Connector (inline-created or externally referenced)
   - **Ingress**: Configure public accessibility and IP address type (IPv4/Dual Stack)

6. **Phase 6: Auto Scaling Configuration**
   - Attach Auto Scaling Configuration ARN (inline-created or externally referenced)

7. **Phase 7: Encryption Configuration**
   - Set KMS key ARN if provided (ForceNew: changing requires replacement)

8. **Phase 8: Observability Configuration**
   - Enable X-Ray tracing if `observability_enabled=true` and configuration ARN provided

9. **Phase 9: Service Creation**
   - Create `apprunner.Service` with all configured settings

10. **Phase 10: Output Exports**
    - Export service ARN, ID, URL, name, and status

#### `module/vpc_connector.go` (conditional sub-resource)
**Key Function:** `vpcConnector(ctx, locals, provider)`

Creates an inline VPC Connector when `subnet_ids` are provided and no `vpc_connector_arn` is referenced.

**When Created**: Only when `len(spec.GetSubnetIds()) > 0 && spec.GetVpcConnectorArn().GetValue() == ""`

**Purpose**: Enables App Runner service to reach resources in VPC (databases, caches, internal APIs).

**Requirements**:
- At least two subnet IDs (recommended for high availability across AZs)
- Security group IDs (controls what VPC resources the service can reach)
- VPC Connector is shared across multiple services if needed (can be referenced via ARN)

#### `module/auto_scaling.go` (conditional sub-resource)
**Key Function:** `autoScalingConfig(ctx, locals, provider)`

Creates an Auto Scaling Configuration Version when `auto_scaling` block is provided.

**When Created**: Only when `spec.GetAutoScaling() != nil`

**Purpose**: Controls how App Runner scales instances based on concurrent request load.

**Configuration**:
- `min_size`: Minimum instances (default: 1)
- `max_size`: Maximum instances (default: 25)
- `max_concurrency`: Max concurrent requests per instance before scaling (default: 100)

#### `module/outputs.go` (constants)
Defines string constants for output keys, preventing typos and providing a single source of truth.

**Exported Outputs**:
- `service_arn`: Full ARN of the App Runner service
- `service_id`: Unique service identifier
- `service_url`: Public HTTPS URL of the service
- `service_name`: Name of the service
- `service_status`: Current status (RUNNING, CREATE_FAILED, etc.)
- `vpc_connector_arn`: ARN of created VPC Connector (if inline-created)
- `auto_scaling_configuration_arn`: ARN of created Auto Scaling Configuration (if inline-created)

## Data Flow Diagram

```
┌──────────────────────────────────────┐
│ AwsAppRunnerServiceStackInput        │
│  ├─ target: AwsAppRunnerService      │
│  │   ├─ metadata                     │
│  │   └─ spec: AwsAppRunnerServiceSpec│
│  │       ├─ image_source OR          │
│  │       │   code_source             │
│  │       ├─ port                     │
│  │       ├─ start_command            │
│  │       ├─ cpu/memory               │
│  │       ├─ instance_role_arn        │
│  │       ├─ environment_variables    │
│  │       ├─ environment_secrets      │
│  │       ├─ health_check             │
│  │       ├─ auto_scaling             │
│  │       ├─ subnet_ids/              │
│  │       │   security_group_ids      │
│  │       ├─ vpc_connector_arn        │
│  │       ├─ is_publicly_accessible   │
│  │       ├─ ip_address_type          │
│  │       ├─ kms_key_arn              │
│  │       └─ observability_config     │
│  └─ provider_config (optional)       │
└──────────────┬───────────────────────┘
               │
               ▼
      ┌────────────────┐
      │initializeLocals│
      └────────┬────────┘
               │ Creates:
               │ - Locals.AwsAppRunnerService
               │ - Locals.AwsTags (6 tags)
               ▼
     ┌──────────────────────┐
     │ AWS Provider Setup   │
     │  ├─ Ambient creds OR │
     │  └─ Explicit creds   │
     └──────────┬───────────┘
                │
                ├─ IF subnet_ids provided AND
                │    vpc_connector_arn empty:
                │  ▼
                │  ┌──────────────────┐
                │  │ vpcConnector()   │
                │  │ Creates VPC      │
                │  │ Connector        │
                │  └────────┬─────────┘
                │           │
                ├─ IF auto_scaling provided:
                │  ▼
                │  ┌──────────────────┐
                │  │ autoScalingConfig│
                │  │ Creates Auto     │
                │  │ Scaling Config   │
                │  └────────┬─────────┘
                │           │
                ▼
       ┌──────────────────┐
       │ service()        │
       └────────┬─────────┘
                │
                ├─ Build source configuration:
                │    IF image_source:
                │      - Image identifier
                │      - Repository type (ECR/ECR_PUBLIC)
                │      - Access role ARN (if private ECR)
                │      - Image config (port, start_command, env vars/secrets)
                │    
                │    IF code_source:
                │      - Repository URL, branch
                │      - Connection ARN
                │      - Configuration source (API/REPOSITORY)
                │      - Runtime, build_command (if API)
                │      - Code config (port, start_command, env vars/secrets)
                │
                ├─ Build instance configuration:
                │    - CPU, memory
                │    - Instance role ARN
                │
                ├─ Build health check configuration:
                │    - Protocol (TCP/HTTP)
                │    - Path (if HTTP)
                │    - Interval, timeout, thresholds
                │
                ├─ Build network configuration:
                │    - Egress: VPC Connector ARN
                │    - Ingress: public accessibility, IP type
                │
                ├─ Attach auto scaling configuration ARN
                │
                ├─ Set encryption configuration (KMS key)
                │
                ├─ Set observability configuration (X-Ray)
                │
                ├─ Create App Runner Service:
                │    apprunner.NewService(ctx, name, args, provider)
                │
                └─ Export outputs:
                     ctx.Export("service_arn", svc.Arn)
                     ctx.Export("service_id", svc.ServiceId)
                     ctx.Export("service_url", svc.ServiceUrl)
                     ctx.Export("service_name", name)
                     ctx.Export("service_status", svc.Status)
                │
                ▼
       ┌─────────────────────────────┐
       │  AWS App Runner Service      │
       │   ├─ Service                 │
       │   │   ├─ Source (Image/Code) │
       │   │   ├─ Instance Config     │
       │   │   ├─ Health Check        │
       │   │   ├─ Network Config      │
       │   │   │   ├─ VPC Connector   │
       │   │   │   └─ Ingress Config  │
       │   │   ├─ Auto Scaling Config │
       │   │   ├─ Encryption Config   │
       │   │   └─ Observability Config│
       │   ├─ VPC Connector (optional) │
       │   └─ Auto Scaling Config     │
       │       (optional)             │
       └─────────────────────────────┘
```

## Resource Relationships

```
AwsAppRunnerServiceSpec
  │
  ├─ image_source OR code_source ────┐
  │                                  │
  ├─ port ───────────────────────────┼─┐
  ├─ start_command ──────────────────┼─┤
  ├─ cpu/memory ────────────────────┼─┤
  ├─ instance_role_arn ─────────────┼─┤
  ├─ environment_variables ─────────┼─┤
  ├─ environment_secrets ───────────┼─┤
  ├─ health_check ──────────────────┼─┤
  │                                  │ │
  ├─ auto_scaling ──────────────────┼─┼─ Auto Scaling Config Version
  │                                  │ │   ├─ Min/Max size
  │                                  │ │   └─ Max concurrency
  │                                  │ │
  ├─ subnet_ids ────────────────────┼─┼─ VPC Connector (inline)
  ├─ security_group_ids ────────────┼─┼─   ├─ Subnets
  ├─ vpc_connector_arn ─────────────┼─┼─   └─ Security Groups
  │                                  │ │
  ├─ is_publicly_accessible ────────┼─┤
  ├─ ip_address_type ───────────────┼─┤
  ├─ kms_key_arn ───────────────────┼─┤
  ├─ observability_enabled ─────────┼─┤
  └─ observability_configuration_arn┼─┤
                                     │ │
                                     ▼ ▼
                            App Runner Service
                             ├─ Service ARN
                             ├─ Service ID
                             ├─ Service URL (HTTPS)
                             ├─ Service Name
                             └─ Service Status
```

### Critical Relationships

**Source Type → Configuration**:
- **Image Source**: Requires `image_identifier`, `image_repository_type`; optional `access_role_arn` for private ECR
- **Code Source**: Requires `repository_url`, `branch`, `connection_arn`, `configuration_source`; optional `runtime`, `build_command` when `configuration_source=API`
- **Mutually Exclusive**: Exactly one of `image_source` or `code_source` must be set (enforced by proto validation)

**VPC Connector → Egress Networking**:
- If `subnet_ids` provided and `vpc_connector_arn` empty: Creates inline VPC Connector
- If `vpc_connector_arn` provided: Uses existing VPC Connector (shared across services)
- Mutually exclusive: Cannot provide both `subnet_ids` and `vpc_connector_arn`

**Auto Scaling Configuration → Scaling Behavior**:
- If `auto_scaling` block provided: Creates Auto Scaling Configuration Version
- Controls min/max instances and max concurrency per instance
- Can be shared across services if referenced by ARN (not implemented in this module)

**Instance Role ARN → Runtime Permissions**:
- Instance role is assumed by App Runner service instances at runtime
- Must have permissions for services the application accesses (S3, DynamoDB, etc.)
- Different from `access_role_arn` (used for pulling images from private ECR)

**Connection ARN → GitHub Access**:
- Required for code source deployments
- Created out-of-band via AWS Console or CLI (requires OAuth handshake)
- Can be shared across multiple App Runner services

## Key Design Decisions

### 1. Inline VPC Connector Creation

**Decision**: Automatically create a VPC Connector when `subnet_ids` are provided and no `vpc_connector_arn` is referenced.

**Implementation**:
```go
if len(spec.GetSubnetIds()) > 0 && spec.GetVpcConnectorArn().GetValue() == "" {
    createdVpcConnector, err = vpcConnector(ctx, locals, provider)
}
```

**Rationale**:
- **Simplified UX**: Users don't need to create VPC Connector separately
- **Resource Lifecycle**: VPC Connector lifecycle tied to service lifecycle
- **Flexibility**: Still supports referencing existing VPC Connector via ARN for shared scenarios

**Alternative Approach**: Always require explicit VPC Connector creation. This module creates inline for convenience.

### 2. Source Type Branching

**Decision**: Use conditional logic based on which source type is provided (`image_source` vs `code_source`) to build different `ServiceSourceConfigurationArgs`.

**Implementation**:
```go
if img := spec.GetImageSource(); img != nil {
    // Build ImageRepository configuration
    sourceConfig.ImageRepository = imageRepoArgs
}

if code := spec.GetCodeSource(); code != nil {
    // Build CodeRepository configuration
    sourceConfig.CodeRepository = codeRepoArgs
}
```

**Rationale**:
- **API Compatibility**: AWS App Runner API has mutually exclusive field requirements for image vs code
- **Clear Intent**: Branching makes it obvious which code path is taken
- **Validation Upstream**: Protobuf validation ensures exactly one source type is set

**Trade-off**: More verbose than reflection-based mapping, but explicit branching catches API changes at compile time.

### 3. Conditional Auto Scaling Configuration Creation

**Decision**: Create Auto Scaling Configuration Version only when `auto_scaling` block is provided.

**Implementation**:
```go
if spec.GetAutoScaling() != nil {
    createdAutoScaling, err = autoScalingConfig(ctx, locals, provider)
}
```

**Rationale**:
- **Auto Scaling is Optional**: App Runner has defaults (1 min, 25 max, 100 max concurrency)
- **Resource Lifecycle**: Auto Scaling Configuration Version lifecycle tied to service lifecycle
- **Cost Optimization**: Users can rely on defaults for simple use cases

**Best Practice**: Provide explicit auto scaling configuration for production workloads to control costs and performance.

### 4. StringValueOrRef Pattern for Cross-Resource References

**Decision**: Fields like `instance_role_arn`, `vpc_connector_arn`, `access_role_arn`, `connection_arn`, `kms_key_arn`, and `observability_configuration_arn` use `StringValueOrRef` to support both literal values and references to other resources.

**Implementation**:
```go
// Direct value
spec.InstanceRoleArn.GetValue() → "arn:aws:iam::123456789012:role/apprunner-instance-role"

// Or reference to another resource
spec.InstanceRoleArn → references AwsIamRole.status.outputs.role_arn
```

**Rationale**:
- **Flexibility**: Users can provide ARNs directly or reference other OpenMCF resources
- **Dependency Management**: References create implicit dependencies between resources
- **Simplified Workflow**: No need to manually extract and copy ARNs between resources

**Helper Function**: `valuefrom.ToStringArray()` resolves `StringValueOrRef[]` to plain string arrays.

### 5. Environment Secrets vs Environment Variables

**Decision**: Separate `environment_variables` (plaintext) from `environment_secrets` (AWS Secrets Manager/SSM Parameter Store ARNs).

**Implementation**:
```go
if len(spec.GetEnvironmentVariables()) > 0 {
    args.RuntimeEnvironmentVariables = pulumi.ToStringMap(spec.GetEnvironmentVariables())
}
if len(spec.GetEnvironmentSecrets()) > 0 {
    args.RuntimeEnvironmentSecrets = pulumi.ToStringMap(spec.GetEnvironmentSecrets())
}
```

**Rationale**:
- **Security Best Practice**: Secrets should not be stored as plaintext in manifests
- **AWS Integration**: App Runner retrieves secrets from Secrets Manager/SSM at deploy time
- **IAM Requirements**: Instance role must have permission to read secrets

**Best Practice**: Use `environment_secrets` for sensitive values (database passwords, API keys) and `environment_variables` for non-sensitive configuration.

### 6. Configuration Source Pattern for Code Deployments

**Decision**: Support both `API` (configuration in spec) and `REPOSITORY` (configuration in `apprunner.yaml` file) modes for code source deployments.

**Implementation**:
```go
if code.GetConfigurationSource() == "API" {
    // Set runtime, build_command, port, start_command from spec
    codeValues.Runtime = pulumi.String(code.GetRuntime())
    codeValues.BuildCommand = pulumi.String(code.GetBuildCommand())
}
// If "REPOSITORY", App Runner reads apprunner.yaml from repo
```

**Rationale**:
- **Flexibility**: Teams can choose between centralized (API) or decentralized (REPOSITORY) configuration
- **GitOps Compatibility**: REPOSITORY mode enables configuration changes via pull requests
- **Migration Path**: Easy to migrate from API to REPOSITORY mode

**Best Practice**: Use REPOSITORY mode for teams that want configuration changes to go through code review.

## App Runner-Specific Implementation Details

### Image Source Handling

For image-based deployments, the module supports both private ECR and public ECR Gallery.

**Private ECR**:
```go
imageRepoArgs.ImageRepositoryType = pulumi.String("ECR")
imageRepoArgs.ImageIdentifier = pulumi.String("123456789012.dkr.ecr.us-east-1.amazonaws.com/repo:tag")
// Requires access_role_arn for pull permissions
```

**Public ECR**:
```go
imageRepoArgs.ImageRepositoryType = pulumi.String("ECR_PUBLIC")
imageRepoArgs.ImageIdentifier = pulumi.String("public.ecr.aws/nginx/nginx:latest")
// No access_role_arn needed
```

**Image Configuration**: Port, start command, environment variables, and secrets are set in `ImageConfiguration` block.

### Code Source Handling

For code-based deployments, the module supports GitHub repositories with two configuration modes.

**API Configuration Mode**:
- Runtime, build command, port, start command specified in spec
- App Runner uses these values directly
- Requires `runtime` and optionally `build_command` fields

**REPOSITORY Configuration Mode**:
- Configuration read from `apprunner.yaml` file in repository root (or `source_directory`)
- Runtime and build_command in spec are ignored
- Enables GitOps workflows

**Connection ARN**: Required for GitHub access. Created out-of-band via AWS Console or CLI (OAuth handshake required).

### VPC Connector and Egress Networking

When VPC Connector is attached, App Runner service instances can reach resources in VPC.

**Requirements**:
1. **Subnet IDs**: At least two subnets (recommended for HA across AZs)
2. **Security Group IDs**: Controls what VPC resources the service can reach
3. **IAM Permissions**: Instance role must have permissions for VPC resources (e.g., RDS, ElastiCache)

**Egress Behavior**:
- Traffic to VPC resources routes through VPC Connector
- Internet-bound traffic routes through VPC Connector if NAT Gateway configured
- Without VPC Connector: service can only reach public internet (no VPC access)

### Auto Scaling Configuration

App Runner uses a concurrency-based scaling model.

**Scaling Behavior**:
- When concurrent requests per instance exceed `max_concurrency`, new instance launched
- Scales up to `max_size` instances
- Scales down to `min_size` instances when traffic decreases

**Default Values** (if not provided):
- `min_size`: 1
- `max_size`: 25
- `max_concurrency`: 100

**Cost Impact**: Higher `min_size` reduces cold starts but increases baseline cost. Lower `max_concurrency` provides more headroom per instance but costs more (more instances for same traffic).

### Health Check Configuration

App Runner monitors instance health and replaces unhealthy instances.

**Protocols**:
- **TCP**: Checks that port is open and accepting connections (default)
- **HTTP**: Sends HTTP GET request to specified path, expects 200 response

**Default Values** (if not provided):
- `protocol`: TCP
- `path`: "/" (for HTTP)
- `interval_seconds`: 5
- `timeout_seconds`: 2
- `healthy_threshold`: 1
- `unhealthy_threshold`: 5

**Best Practice**: Use HTTP health checks with application-specific path (e.g., `/health`) for better readiness detection.

### Encryption Configuration

App Runner encrypts stored container images and data logs.

**AWS-Managed Key** (default):
- Automatic encryption with AWS-managed KMS key
- No additional configuration needed

**Customer-Managed Key**:
- Provide `kms_key_arn` for customer-managed KMS key
- **ForceNew**: Changing `kms_key_arn` requires replacing the service
- Instance role must have `kms:Decrypt` permission

### Observability Configuration

App Runner supports AWS X-Ray tracing for distributed tracing.

**Requirements**:
- Set `observability_enabled: true`
- Provide `observability_configuration_arn` (created separately via AWS Console or CLI)
- Observability configurations can be shared across multiple services

**Use Case**: Enable for microservices architectures to trace requests across services.

### Auto-Deployments

App Runner can automatically trigger new deployments when source changes.

**Image Source**:
- Redeploys when new image is pushed to the same tag
- Controlled by `auto_deployments_enabled` (default: true)

**Code Source**:
- Redeploys when new commit is pushed to configured branch
- Controlled by `auto_deployments_enabled` (default: true)

**Best Practice**: Enable for development environments, disable for production (use manual deployments for control).

## Error Handling Philosophy

The module follows a **fail-fast** approach:

1. **Validation at Protobuf Level**: The `spec.proto` validation rules (including CEL validations) catch configuration errors before Pulumi runs
2. **AWS API Errors Propagate**: If AWS rejects a configuration, the error propagates immediately (no silent fallbacks)
3. **Wrapped Errors**: All errors include context (`errors.Wrap()`) for easier debugging

**Rationale**: Infrastructure as code demands predictability. Silent defaults or error recovery can mask misconfiguration and create surprise behavior.

## Common Pitfalls and Gotchas

### Pitfall 1: Missing ECR Access Role Permissions
**Symptom**: App Runner service fails to pull image from private ECR.

**Cause**: `access_role_arn` lacks required ECR permissions.

**Solution**: Ensure access role has `ecr:GetDownloadUrlForLayer`, `ecr:BatchGetImage`, and `ecr:GetAuthorizationToken` permissions.

### Pitfall 2: Missing Instance Role Permissions
**Symptom**: Application fails at runtime with access denied errors.

**Cause**: Instance role doesn't have permissions for services the application accesses (S3, DynamoDB, etc.).

**Solution**: Update the instance role to grant required permissions. Use `instance_role_arn` reference to an `AwsIamRole` resource for centralized management.

### Pitfall 3: VPC Connector Subnet Selection
**Symptom**: App Runner service cannot reach VPC resources.

**Cause**: Subnets selected don't have routes to target VPC resources, or security groups are too restrictive.

**Solution**: Ensure subnets have routes to target resources (via route tables) and security groups allow required traffic.

### Pitfall 4: Missing GitHub Connection
**Symptom**: Code source deployment fails with connection error.

**Cause**: `connection_arn` references non-existent or unauthorized connection.

**Solution**: Create App Runner Connection via AWS Console or CLI first, then reference ARN in spec.

### Pitfall 5: Configuration Source Mismatch
**Symptom**: Code source deployment ignores runtime/build_command settings.

**Cause**: `configuration_source` is set to `REPOSITORY` but expecting API-provided values.

**Solution**: Either set `configuration_source: API` or provide `apprunner.yaml` in repository root.

### Pitfall 6: KMS Key Permissions
**Symptom**: Service creation fails with KMS decrypt error.

**Cause**: Instance role lacks `kms:Decrypt` permission for customer-managed KMS key.

**Solution**: Add KMS decrypt permission to instance role for the KMS key ARN.

### Pitfall 7: Health Check Path Not Found
**Symptom**: Instances marked unhealthy despite application running.

**Cause**: HTTP health check path doesn't exist or returns non-200 status.

**Solution**: Ensure health check path exists and returns 200 status code, or use TCP health checks.

## Testing and Debugging

### Debugging with Delve

The `debug.sh` script enables step-through debugging:

1. Uncomment the binary option in `Pulumi.yaml`:
   ```yaml
   runtime:
     options:
       binary: ./debug.sh
   ```
2. Run Pulumi CLI commands normally
3. The debug script launches Delve, allowing breakpoints in any module file

**Use Case**: Debugging source configuration branching, VPC Connector creation, or AWS API errors.

### Manual Testing with Sample Manifest

Use the sample manifest in `iac/hack/manifest.yaml` for local testing:

```bash
cd iac/pulumi
pulumi stack init dev
pulumi config set aws:region us-east-1
# Set AWS credentials via environment variables or AWS_PROFILE
pulumi up
```

**Verification**:
```bash
# Get outputs
pulumi stack output service_arn
pulumi stack output service_url
pulumi stack output service_status

# Test service endpoint
curl $(pulumi stack output service_url)

# View service logs (via CloudWatch)
aws apprunner describe-service --service-arn $(pulumi stack output service_arn)
```

## Performance Considerations

### Resource Creation Time

App Runner service creation time varies by configuration:
- **Basic service (no VPC)**: 2-5 minutes
- **Service with VPC Connector**: 3-6 minutes (VPC Connector creation adds time)
- **Service with code source**: 5-10 minutes (includes build time)

**Total**: Expect 2-10 minutes depending on complexity and source type.

### Cold Start Implications

App Runner instances have cold start latency:
- **Image source**: 10-30 seconds (image pull + container start)
- **Code source**: 30-60 seconds (build + image creation + container start)

**Optimization Strategies**:
- Use `min_size > 1` to keep instances warm
- Optimize container image size (smaller images = faster pulls)
- Use multi-stage Docker builds to reduce image size
- Consider ARM64 architecture for better price/performance

### Pulumi State Tracking

The module creates up to three Pulumi resources:
1. One `apprunner.VpcConnector` resource (if inline-created)
2. One `apprunner.AutoScalingConfigurationVersion` resource (if provided)
3. One `apprunner.Service` resource

**State Size Impact**: Minimal for individual services. Large-scale deployments (100+ services) should monitor state file size.

## App Runner-Specific Best Practices

### Image Source Best Practices

**Use Specific Tags or Digests**:
- Avoid `latest` tag (unpredictable deployments)
- Use semantic versioning tags (e.g., `v1.2.3`)
- Consider using image digests for immutable deployments

**Optimize Image Size**:
- Use multi-stage Docker builds
- Exclude unnecessary files (dev dependencies, build artifacts)
- Use Alpine-based images when possible

**ECR Access Role**:
- Create dedicated IAM role for App Runner ECR access
- Scope permissions to specific repositories
- Use least privilege principle

### Code Source Best Practices

**Use REPOSITORY Configuration Mode**:
- Store `apprunner.yaml` in repository root
- Enables configuration changes via pull requests
- Better GitOps workflow integration

**Connection Management**:
- Create connections at organization level (shared across services)
- Use OAuth connections for GitHub (more secure than personal access tokens)
- Rotate connection credentials periodically

**Build Optimization**:
- Use `.dockerignore` to exclude unnecessary files
- Cache dependencies (npm, pip, etc.) in build process
- Minimize build time (faster deployments)

### Auto Scaling Best Practices

**Right-Size Min/Max**:
- Set `min_size` based on baseline traffic (reduces cold starts)
- Set `max_size` based on peak traffic expectations
- Monitor CloudWatch metrics to adjust over time

**Concurrency Tuning**:
- Start with default `max_concurrency: 100`
- Increase if instances are underutilized (cost optimization)
- Decrease if instances are overloaded (better performance)

**Cost Optimization**:
- Higher `min_size` = lower cold starts but higher baseline cost
- Lower `max_concurrency` = more instances for same traffic (higher cost)
- Balance performance requirements with cost constraints

### Health Check Best Practices

**Use HTTP Health Checks**:
- More accurate readiness detection than TCP
- Implement `/health` endpoint in application
- Return 200 when ready, 503 when not ready

**Tune Thresholds**:
- `healthy_threshold: 1` for fast recovery
- `unhealthy_threshold: 3-5` to avoid false positives
- Adjust `interval_seconds` and `timeout_seconds` based on application response time

### Security Best Practices

**Environment Secrets**:
- Never store secrets in `environment_variables` (plaintext)
- Use `environment_secrets` with Secrets Manager or SSM Parameter Store
- Rotate secrets regularly

**Instance Role Permissions**:
- Follow least privilege principle
- Grant only permissions needed by application
- Use resource-level permissions when possible

**VPC Security**:
- Use security groups to restrict VPC access
- Place App Runner service in private subnets if possible
- Use VPC endpoints for AWS services (reduces internet egress)

**Encryption**:
- Use customer-managed KMS keys for compliance requirements
- Ensure instance role has KMS decrypt permissions
- Enable encryption at rest and in transit

## Cost Optimization

App Runner pricing has three components:
1. **Compute**: $0.007 per vCPU-hour
2. **Memory**: $0.0008 per GB-hour
3. **Requests**: $0.0000000083 per request

**Optimization Strategies**:

**Right-Size CPU/Memory**:
- Start with default (1 vCPU, 2 GB)
- Monitor CloudWatch metrics (CPU utilization, memory usage)
- Reduce if underutilized, increase if overloaded

**Optimize Auto Scaling**:
- Set `min_size` based on baseline traffic
- Set `max_size` to cap peak costs
- Tune `max_concurrency` to balance performance and cost

**Reduce Cold Starts**:
- Use `min_size > 1` for latency-critical services
- Optimize container image size (faster pulls)
- Use ARM64 architecture (better price/performance)

**Monitor with CloudWatch**:
- Track service costs via Cost Explorer (filtered by tags)
- Set up billing alarms for unexpected cost spikes
- Review auto scaling metrics regularly

## Conclusion

This Pulumi module demonstrates OpenMCF's philosophy: **support both simple and complex use cases through a single, flexible API**.

The architecture accommodates:
- **Simple services**: Public ECR image, default settings, no VPC
- **Complex services**: Private ECR, VPC Connector, custom auto scaling, health checks, encryption, observability

Key design principles:
- **Explicit branching** for image vs code source
- **Inline sub-resource creation** for VPC Connector and Auto Scaling Configuration
- **Conditional configuration** to avoid unnecessary complexity
- **StringValueOrRef pattern** for cross-resource references
- **Separate environment variables and secrets** for security best practices

For teams familiar with Terraform or CloudFormation, the mapping should feel natural. For teams new to App Runner, the research documentation (`docs/README.md`) provides critical context about container deployments, auto scaling, and VPC networking.

The result is a production-ready module that provisions App Runner services with proper IAM roles, VPC networking (when needed), auto scaling, health checks, encryption, and comprehensive output exports—handling the complexity of App Runner's dual source types while keeping the API surface clean.
