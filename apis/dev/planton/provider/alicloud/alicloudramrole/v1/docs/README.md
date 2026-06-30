# Alibaba Cloud RAM Roles: From Console Clicks to Control Planes

## Introduction

Alibaba Cloud Resource Access Management (RAM) roles are the linchpin of secure service integration on Alibaba Cloud. A RAM role is a virtual identity — it has no permanent credentials, no password, no access key. Instead, trusted entities (cloud services, other Alibaba Cloud accounts, or federated identities) *assume* the role via the Security Token Service (STS) to obtain temporary security tokens. This assume-role pattern is the standard mechanism for granting Alibaba Cloud services (ECS, ACK, FC, SAE) the permissions they need to operate on your behalf without embedding long-lived credentials in application code or configuration.

Despite the conceptual simplicity of "create a role, attach policies, let a service assume it," production RAM role management is riddled with subtle errors. Teams create roles with overly broad trust policies that allow any service to assume them. They create roles without attaching any policies — producing an identity that can authenticate but authorize nothing. They forget to set appropriate session durations for long-running batch jobs, causing STS tokens to expire mid-execution. And they manage the role and its policy attachments as independent lifecycle objects, leading to drift where the role exists but its policies have been detached or modified outside of IaC.

This document examines the full deployment landscape for RAM roles — from manual console provisioning to control-plane-based automation — and explains how Planton bundles the role and its policy attachments into a single, validated API resource that eliminates the most common misconfigurations while providing a clean abstraction for the 80% use case.

## Evolution of RAM Role Management

Alibaba Cloud RAM has evolved through several phases since its introduction. Early adoption was entirely console-driven: an administrator would navigate to the RAM service page, click through a multi-step wizard to create a role, then navigate to a separate page to find and attach policies. The trust policy (which defines who can assume the role) was presented as a raw JSON editor buried in the role's "Trust Policy Management" tab — a critical security configuration hidden behind multiple clicks.

The introduction of the `aliyun` CLI brought scriptability. Teams could create roles with `aliyun ram CreateRole` and attach policies with `aliyun ram AttachPolicyToRole`. But each was an independent, imperative API call with no built-in idempotency — running the same script twice would fail because the role already existed. Teams wrapped these calls in existence checks, turning a simple two-step process into a fragile 20-line bash script.

Terraform's `alicloud` provider introduced declarative state management with `alicloud_ram_role` and `alicloud_ram_role_policy_attachment` as separate resources. This was a significant improvement: state tracking, drift detection, and dependency ordering. But the two-resource model still allowed a common failure pattern: creating a role with no policy attachments. Terraform would happily apply a plan that produced a role with zero permissions — a security identity that could authenticate via STS but do nothing once authenticated.

Pulumi's Go SDK (`pulumi-alicloud`) offers the same granularity with type safety. The `ram.NewRole` and `ram.NewRolePolicyAttachment` functions mirror the Terraform resources but catch field name typos at compile time. The fundamental problem remains: role and attachments are separate lifecycle objects.

The pattern across this evolution is consistent: every tool treats the role and its policy attachments as independent resources. This is architecturally correct (they are separate API objects in the RAM API) but operationally dangerous. A role without policies is the identity equivalent of an empty shell — it exists, it can be assumed, but it grants zero permissions. Planton's contribution is bundling the role and its policy attachments into a single resource declaration, ensuring that the role is always provisioned with its intended permissions.

## The RAM Role Deployment Landscape

### Level 0: Manual Provisioning via Alibaba Cloud Console

The Alibaba Cloud console provides a wizard-driven workflow for RAM role creation:

1. Navigate to **RAM** → **Roles** in the console
2. Click **Create Role**, select the trusted entity type (Alibaba Cloud Service, Alibaba Cloud Account, or Identity Provider)
3. Choose the specific trusted entity (e.g., ECS, ACK, FC)
4. Enter a role name and optional description
5. After role creation, navigate to **Permissions** → **Add Permissions**
6. Search for and attach policies (system or custom)

**Common Mistakes**:

1. **Overly Broad Trust Policies**: The console's "Alibaba Cloud Service" wizard presents a list of services but doesn't explain the security implications. Selecting "All Services" creates a trust policy that allows *any* Alibaba Cloud service to assume the role — an unnecessarily wide blast radius. The correct pattern is to select only the specific service (e.g., "ECS" or "Function Compute") that needs to assume the role.

2. **Missing Policy Attachments**: Creating a role and leaving the permissions tab empty. The role exists and can be assumed via STS, but the resulting session has zero permissions. This is especially common when teams "prepare" roles in advance and forget to attach policies before the consuming service goes live.

3. **Wrong Trust Policy for Cross-Account Access**: When creating a role that another Alibaba Cloud account should assume, the trust policy must reference the specific account ID (e.g., `acs:ram::1234567890123456:root`). A common mistake is using `acs:ram::*:root`, which allows *any* Alibaba Cloud account to assume the role — a severe security vulnerability that the console doesn't warn about.

4. **Default Session Duration**: The console defaults to 3600 seconds (1 hour). For long-running batch jobs, CI/CD pipelines, or cross-account operations, this is often too short. The STS token expires mid-execution, causing cryptic "access denied" errors that are difficult to diagnose because the initial authentication succeeded.

5. **No Force-Delete Awareness**: Deleting a role with attached policies fails silently in the console — the delete button appears to work but the role persists. Teams end up with orphaned roles they believe were deleted.

**Verdict**: Acceptable for initial learning and ad-hoc exploration. Unacceptable for production environments where role configurations must be reproducible, auditable, and version-controlled.

### Level 1: Scripted Provisioning with Alibaba Cloud CLI

The `aliyun` CLI provides direct access to the RAM API:

```bash
# Create role with ECS service trust policy
aliyun ram CreateRole \
  --RoleName my-ecs-role \
  --AssumeRolePolicyDocument '{
    "Statement": [{
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {"Service": ["ecs.aliyuncs.com"]}
    }],
    "Version": "1"
  }' \
  --Description "Role for ECS instances"

# Attach system policy
aliyun ram AttachPolicyToRole \
  --RoleName my-ecs-role \
  --PolicyName AliyunOSSFullAccess \
  --PolicyType System

# Attach another system policy
aliyun ram AttachPolicyToRole \
  --RoleName my-ecs-role \
  --PolicyName AliyunLogFullAccess \
  --PolicyType System
```

**The Sequence Problem**: Three separate API calls must execute in order (role before attachments). Each call is imperative — there is no built-in idempotency. Running the script twice fails on the `CreateRole` call because the role already exists. Teams wrap each call in `if-not-exists` checks:

```bash
# Check if role exists before creating
aliyun ram GetRole --RoleName my-ecs-role 2>/dev/null || \
  aliyun ram CreateRole --RoleName my-ecs-role --AssumeRolePolicyDocument '...'
```

This doubles the script length and introduces race conditions in concurrent executions.

**The State Problem**: The CLI has no state file. There is no way to know whether the current role configuration (trust policy, attached policies, session duration) matches the desired configuration without querying the API and comparing field-by-field. Detecting drift — whether someone manually detached a policy via the console — requires a custom audit script.

**The Detach-Before-Delete Problem**: Deleting a role requires first detaching all policies. The CLI doesn't handle this automatically:

```bash
# Must detach all policies before deleting
aliyun ram DetachPolicyFromRole --RoleName my-ecs-role --PolicyName AliyunOSSFullAccess --PolicyType System
aliyun ram DetachPolicyFromRole --RoleName my-ecs-role --PolicyName AliyunLogFullAccess --PolicyType System
aliyun ram DeleteRole --RoleName my-ecs-role
```

Missing a single detach causes the delete to fail. The `force` parameter in the IaC tools handles this automatically.

**Verdict**: Suitable for one-off tasks, CI/CD pipeline steps where idempotency is handled externally, or debugging. Not suitable for managing RAM roles at scale.

### Level 2: Infrastructure as Code (Terraform / OpenTofu)

Terraform's `alicloud` provider models the role and its policy attachments as separate resources:

```hcl
resource "alicloud_ram_role" "ecs_role" {
  role_name                    = "my-ecs-role"
  description                  = "Role for ECS instances to access OSS and SLS"
  assume_role_policy_document = jsonencode({
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = ["ecs.aliyuncs.com"] }
    }]
    Version = "1"
  })
  max_session_duration = 7200
  force                = true
}

resource "alicloud_ram_role_policy_attachment" "oss_access" {
  role_name   = alicloud_ram_role.ecs_role.role_name
  policy_name = "AliyunOSSFullAccess"
  policy_type = "System"
}

resource "alicloud_ram_role_policy_attachment" "log_access" {
  role_name   = alicloud_ram_role.ecs_role.role_name
  policy_name = "AliyunLogFullAccess"
  policy_type = "System"
}
```

**Strengths**:

- **Declarative**: Define the desired end state; Terraform calculates the diff
- **Stateful**: Tracks role ID, ARN, and attachment state; detects drift
- **Dependency Graph**: Automatically creates role before attachments, detaches before deleting
- **Idempotent**: Running `terraform apply` twice produces the same result

**Weaknesses**:

- **Verbose**: A role with 5 policy attachments requires 6 resource blocks (1 role + 5 attachments). Each attachment is a separate resource with its own lifecycle.
- **No Bundling Semantics**: Terraform doesn't enforce that a role should have at least one policy attachment. A plan that creates a role with zero attachments applies cleanly — producing a non-functional identity.
- **State Management Overhead**: The state file must be stored remotely (OSS bucket + TableStore for locking) in team environments.

**The `for_each` Pattern**: Production Terraform modules use `for_each` to iterate over a list of policy attachments:

```hcl
variable "policy_attachments" {
  type = list(object({
    policy_name = string
    policy_type = optional(string, "System")
  }))
}

resource "alicloud_ram_role_policy_attachment" "attachments" {
  for_each = { for pa in var.policy_attachments : "${pa.policy_name}-${pa.policy_type}" => pa }

  role_name   = alicloud_ram_role.main.role_name
  policy_name = each.value.policy_name
  policy_type = each.value.policy_type
}
```

This is exactly the pattern Planton's Terraform module uses — but wrapped behind a validated API that bundles the role and attachments into a single declaration.

**Verdict**: The modern standard for managing RAM roles. Recommended for teams already using Terraform. The two-resource model is a trade-off: maximum flexibility at the cost of allowing non-functional configurations.

### Level 3: Infrastructure as Code (Pulumi)

Pulumi's Go SDK provides type-safe RAM role creation:

```go
role, err := ram.NewRole(ctx, "my-ecs-role", &ram.RoleArgs{
    RoleName:                  pulumi.String("my-ecs-role"),
    Description:               pulumi.String("Role for ECS instances"),
    AssumeRolePolicyDocument: pulumi.String(`{
        "Statement": [{
            "Action": "sts:AssumeRole",
            "Effect": "Allow",
            "Principal": {"Service": ["ecs.aliyuncs.com"]}
        }],
        "Version": "1"
    }`),
    MaxSessionDuration: pulumi.Int(7200),
    Force:              pulumi.Bool(true),
})

_, err = ram.NewRolePolicyAttachment(ctx, "oss-access", &ram.RolePolicyAttachmentArgs{
    RoleName:   pulumi.String("my-ecs-role"),
    PolicyName: pulumi.String("AliyunOSSFullAccess"),
    PolicyType: pulumi.String("System"),
}, pulumi.Parent(role))

_, err = ram.NewRolePolicyAttachment(ctx, "log-access", &ram.RolePolicyAttachmentArgs{
    RoleName:   pulumi.String("my-ecs-role"),
    PolicyName: pulumi.String("AliyunLogFullAccess"),
    PolicyType: pulumi.String("System"),
}, pulumi.Parent(role))
```

**Key Advantages Over Terraform**:

- **Type Safety**: Compile-time validation of field names and types. Misspelling `AssumeRolePolicyDocument` is a build error, not a runtime surprise.
- **Parent Chaining**: `pulumi.Parent(role)` creates an explicit dependency hierarchy where policy attachments are children of the role. Deleting the role automatically handles detachment.
- **Programmatic Composition**: Loops over policy lists are native Go — no HCL `for_each` workarounds.
- **Multi-Language**: Same logic can be expressed in TypeScript, Python, Java, or C#.

**Key Disadvantage**: Requires compiling Go code (or running a Node/Python runtime). Terraform's declarative HCL is simpler for teams that don't need programmatic composition.

**Verdict**: Preferred for teams using Go or TypeScript, especially when RAM role provisioning is embedded in a larger orchestration workflow. The type safety is particularly valuable for trust policy documents, where a JSON syntax error at runtime is far more costly than a compile-time check.

### Level 4: Control Planes and Continuous Reconciliation

The most advanced deployment model treats RAM role configuration as a continuously reconciled desired state:

- **Crossplane**: Extends the Kubernetes API with custom resources for Alibaba Cloud. An operator watches for RAM role custom resources and provisions/reconciles them automatically.
- **Custom Operators**: Organizations build Kubernetes operators that watch for application deployments and automatically create corresponding RAM roles with least-privilege policies.

**Planton Context**: Planton's protobuf-defined API is designed for this model. The YAML manifest is a desired-state declaration that can be applied once (CLI mode) or continuously reconciled (control-plane mode). The `AliCloudRamRole` resource is a Kubernetes-native API object, not just a CLI input format.

**Verdict**: The future of identity management in cloud-native platforms. Planton's API design anticipates this model even when used in CLI mode today.

## Comparative Analysis

| Method | Idempotent | State Tracked | Bundled | Validated | Drift Detection | Effort for Role + 5 Policies |
|--------|-----------|--------------|---------|-----------|----------------|------------------------------|
| Console | No | No | No | No | No | ~5 min clicking, 6 separate actions |
| CLI (`aliyun ram`) | No | No | No | No | No | ~30 lines of bash |
| Terraform | Yes | Yes | No | Partial | Yes | 6 resource blocks |
| Pulumi | Yes | Yes | No | Compile-time | Yes | ~40 lines of Go |
| Planton | Yes | Yes | Yes | Proto-validated | Yes | 1 YAML resource, 5 list items |

The key differentiator is the **Bundled** column. Every other method treats the role and its policy attachments as independent resources. Planton is the only approach that bundles them into a single validated declaration, ensuring that the role is always provisioned with its intended permissions as a single atomic unit.

## The Planton Approach

### Design Philosophy: The DD07 Bundling Decision

The most important design decision for AliCloudRamRole is **DD07: composite bundling**. Instead of requiring users to create the role and then separately attach policies, the component bundles both into a single resource.

**Why bundle?**

1. **A role without policies is non-functional**: A RAM role with no attached policies can be assumed via STS, but the resulting session has zero permissions. Every API call made with that session token returns `AccessDenied`. There is no production use case for an intentionally permissionless role. Unbundling would create an API that allows a broken state.

2. **Attachments are the role's purpose**: The reason a role exists is to grant permissions to a trusted entity. The trust policy (who can assume) and the permission policies (what they can do) are inseparable in purpose, even though the RAM API models them as separate objects. Bundling reflects the operational reality.

3. **The 80% use case is simple**: Most teams need "a role that service X can assume, with policies Y and Z attached." The bundled API serves this case with a flat list of policy attachments. The 20% case (conditional policy attachments, policy creation, complex STS configurations) is not addressed — and intentionally so.

**Why not bundle policy *creation*?** Because policies are reusable across multiple roles. A custom policy like "read-only-oss-bucket-xyz" might be attached to an ECS role, an FC role, and a cross-account audit role. Embedding policy creation inside the role resource would force duplication. Instead, `AliCloudRamPolicy` exists as a separate component for custom policy creation, and `AliCloudRamRole` references those policies by name via the `policyAttachments` field.

### 80/20 Scoping: What's In and What's Out

**Included (the 80%)**:

- **Role creation** with name, description, trust policy document, and tags
- **Session duration configuration** for controlling STS token lifetime (3600-43200 seconds)
- **Force deletion** for clean teardown even when policies are attached
- **Policy attachments** — both system-managed and custom policies, specified by name and type
- **Tags** for organizational grouping, cost tracking, and resource filtering

**Excluded (the 20%)**:

- **RAM policy creation**: Managed by the separate `AliCloudRamPolicy` component. Creating policies and attaching them are different lifecycle operations — a policy is authored once and attached to many roles.
- **STS configuration**: Advanced STS settings (external ID conditions, MFA requirements, source IP restrictions) are rare in the 80% case. Teams requiring these can extend the trust policy document directly.
- **SAML providers and OIDC federation**: Enterprise SSO integration is a separate infrastructure concern with its own lifecycle. Bundling it with role creation would create an unwieldy resource.
- **RAM users and groups**: User/group management is a different domain entirely. Roles are for service-to-service authentication; users/groups are for human authentication.
- **Role-based SSO**: Alibaba Cloud SSO uses roles as the landing identity for federated users. This is a cross-cutting concern that spans multiple roles and SSO configurations.
- **Conditional policy logic**: Policies with conditions (e.g., "allow only if source IP is in range X") are features of the policy document itself, not the role-to-policy attachment.

### API Design Decisions

**`roleName` vs `name`**: The spec uses `roleName` (not `name`) because RAM role names must be unique within the Alibaba Cloud account and follow specific naming rules (1-64 characters, letters/digits/periods/hyphens/underscores). This is distinct from the Planton `metadata.name`, which is the local resource identifier. The metadata name identifies the Planton resource; the role name identifies the Alibaba Cloud object.

**`assumeRolePolicyDocument` as a string**: The trust policy is a raw JSON string rather than a structured protobuf message. This mirrors the Alibaba Cloud RAM API exactly and provides maximum flexibility for trust policy composition. Structured parsing would require modeling every possible principal type (Service, RAM, Federated) as proto messages — added complexity for little value since the JSON format is well-documented and stable.

**`policyAttachments` as a repeated message**: Each attachment is a structured message with `policyName` and `policyType` rather than a simple string list. This allows distinguishing system policies from custom policies without requiring naming conventions. The `policyType` field defaults to `"System"` because system policies are the most common case.

**`maxSessionDuration` with proto default**: The field uses `optional int32` with a `(dev.planton.shared.options.default) = "3600"` annotation. This ensures the default is documented in the API contract and consistently applied by both the Pulumi and Terraform modules. The range validation (3600-43200) is enforced at the proto level via `buf.validate`.

**`force` defaults to `false`**: Force-deleting a role detaches all policies before deletion, which is destructive. The safe default is `false`, requiring explicit opt-in for force deletion. This prevents accidental policy detachment during `destroy` operations.

**`region` despite RAM being global**: RAM is an account-global service — roles are not region-scoped. However, the Alibaba Cloud provider (both Terraform and Pulumi) requires a region for API endpoint configuration. The `region` field configures the provider endpoint, not the role's scope. This is documented in the spec.proto comments to prevent confusion.

### Foreign Key References

AliCloudRamRole has no `StringValueOrRef` fields — all fields are direct values. This is architecturally correct because RAM roles are foundation resources with no upstream dependencies. The trust policy document is a JSON string (not a reference to another resource), and policy names are simple strings that refer to Alibaba Cloud managed policies or separately-created custom policies.

Downstream resources reference this role's outputs:
- `AliCloudAckManagedCluster` can reference this role for cluster service authentication
- `AliCloudFcFunction` references the role ARN for function execution permissions
- `AliCloudEcsInstance` references the role for instance profile attachment

## Implementation Landscape

### Pulumi Module Architecture

The Pulumi module consists of three files under `v1/iac/pulumi/module/`:

**`main.go`** — The controller. Entry point is `Resources(ctx, stackInput)`.
1. Initializes locals (tag computation, default resolution)
2. Creates the Alibaba Cloud provider with the specified region
3. Creates the `ram.Role` resource with all spec fields
4. Iterates over `spec.PolicyAttachments`, creating a `ram.RolePolicyAttachment` for each (parented to the role)
5. Exports outputs: role_id, role_name, arn

**`locals.go`** — Transformations and defaults.
- Computes the tag map by merging standard tags (`resource`, `resource_name`, `resource_kind`, `resource_id`, `organization`, `environment`) with user-provided `spec.Tags`
- Provides helper functions for resolving optional fields:
  - `maxSessionDuration(spec)` — returns `*spec.MaxSessionDuration` if set, otherwise `3600`
  - `forceDelete(spec)` — returns `*spec.Force` if set, otherwise `false`
  - `policyType(pa)` — returns `*pa.PolicyType` if set, otherwise `"System"`
- Provides `optionalString(s)` helper that returns `nil` for empty strings

**`outputs.go`** — Output constant definitions.
- `OpRoleId = "role_id"`
- `OpRoleName = "role_name"`
- `OpArn = "arn"`

**Resource Hierarchy**:

```
Provider (alicloud, region-scoped)
  └── ram.Role
        ├── ram.RolePolicyAttachment ("rolename-policyname-System")
        ├── ram.RolePolicyAttachment ("rolename-policyname-Custom")
        └── ...
```

The `pulumi.Parent(role)` chaining in `policyAttachment()` ensures that deleting the role cascades to attachments, and that the Pulumi dependency graph reflects the actual RAM API constraints. The `pulumi.Provider(provider)` option ensures all resources use the same region-configured provider.

**Naming Convention**: Policy attachment resources are named `{roleName}-{policyName}-{policyType}` to ensure uniqueness within the Pulumi state. This allows multiple policies to be attached without name collisions.

### Terraform Module Architecture

The Terraform module consists of five files under `v1/iac/tf/`:

**`main.tf`** — Resource definitions.
- `alicloud_ram_role.main`: Single role resource with all spec fields
- `alicloud_ram_role_policy_attachment.attachments`: Uses `for_each` over `local.policy_attachments_map`

**`variables.tf`** — Input variables matching the proto schema.
- `metadata` object with `name`, `id`, `org`, `env`, `labels`, `tags`
- `spec` object mirroring `AliCloudRamRoleSpec` with all fields, defaults, and validations
- Includes inline validations for `role_name` length (1-64) and `max_session_duration` range (3600-43200)

**`locals.tf`** — Computed values.
- `policy_attachments_map`: Converts the list of attachments to a map keyed by `{policy_name}-{policy_type}` (required for `for_each`)
- `final_tags`: Merges standard tags with user tags (same logic as Pulumi's `locals.go`)

**`outputs.tf`** — Three outputs matching the Pulumi module.
- `role_id`, `role_name`, `arn` — all sourced from `alicloud_ram_role.main`

**`provider.tf`** — Alibaba Cloud provider configuration with region from `var.spec.region`.

The Terraform and Pulumi modules create identical resources with identical outputs, ensuring that switching between IaC engines produces the same RAM role configuration.

## Production Best Practices

### Trust Policy Patterns

The trust policy document (`assumeRolePolicyDocument`) is the most security-critical field. It defines who can assume the role — incorrect configuration can expose the role to unintended entities.

#### Pattern 1: Single Service Trust

The most common pattern — allowing a specific Alibaba Cloud service to assume the role:

```json
{
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Effect": "Allow",
    "Principal": {"Service": ["ecs.aliyuncs.com"]}
  }],
  "Version": "1"
}
```

Common service principals:
- `ecs.aliyuncs.com` — ECS instances (instance profiles)
- `fc.aliyuncs.com` — Function Compute functions
- `cs.aliyuncs.com` — Container Service (ACK) clusters
- `sae.aliyuncs.com` — Serverless App Engine applications
- `oss.aliyuncs.com` — OSS event triggers
- `log.aliyuncs.com` — Log Service (SLS) data processing

**Anti-Pattern**: Using `{"Service": ["*"]}` or listing all services. This allows any Alibaba Cloud service to assume the role — an unnecessarily wide attack surface.

#### Pattern 2: Cross-Account Trust

Allowing another Alibaba Cloud account to assume the role:

```json
{
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Effect": "Allow",
    "Principal": {"RAM": ["acs:ram::1234567890123456:root"]}
  }],
  "Version": "1"
}
```

**Security Note**: The `root` in the principal means the target account's root identity — it does not grant access to a specific user. The target account must still have a RAM user or role with `sts:AssumeRole` permission to actually assume this role. This is a two-sided trust model.

**Anti-Pattern**: Using `{"RAM": ["acs:ram::*:root"]}` which allows *any* Alibaba Cloud account to assume the role. This is almost never intentional and is a critical security vulnerability.

#### Pattern 3: RRSA (RAM Roles for Service Accounts) in ACK

For Kubernetes workloads running on ACK, RRSA allows Kubernetes service accounts to assume RAM roles without embedding credentials:

```json
{
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Effect": "Allow",
    "Principal": {
      "Federated": ["acs:ram::ACCOUNT_ID:oidc-provider/ack-rrsa-CLUSTER_ID"]
    },
    "Condition": {
      "StringEquals": {
        "oidc:sub": "system:serviceaccount:NAMESPACE:SERVICE_ACCOUNT"
      }
    }
  }],
  "Version": "1"
}
```

This is the Alibaba Cloud equivalent of AWS IRSA (IAM Roles for Service Accounts). The condition restricts which Kubernetes service account can assume the role, providing pod-level IAM isolation.

### Least-Privilege Policy Selection

**System Policies**: Alibaba Cloud provides hundreds of pre-built system policies (e.g., `AliyunOSSFullAccess`, `AliyunOSSReadOnlyAccess`, `AliyunECSFullAccess`). System policies are managed by Alibaba Cloud and automatically updated when new API operations are added to a service.

**The Granularity Trade-off**:
- `AliyunOSSFullAccess` grants all OSS operations on all buckets — simple but overly broad
- A custom policy restricting to specific buckets and operations is more secure but requires creating and maintaining an `AliCloudRamPolicy` resource

**Recommendation**: Start with system policies for development and staging. Create custom policies (via `AliCloudRamPolicy`) for production environments where least-privilege is a compliance requirement.

**Common System Policy Combinations by Use Case**:

| Use Case | Policies |
|----------|----------|
| ECS instance accessing OSS and logs | `AliyunOSSFullAccess`, `AliyunLogFullAccess` |
| ACK cluster management | `AliyunCSManagedKubernetesRole`, `AliyunCSManagedLogRole` |
| FC function accessing VPC resources | `AliyunVPCFullAccess`, `AliyunECSNetworkInterfaceManagement` |
| Cross-account read-only audit | `AliyunBSSReadOnlyAccess`, `AliyunLogReadOnlyAccess`, `AliyunActionTrailFullAccess` |
| CI/CD pipeline deploying to ECS | `AliyunECSFullAccess`, `AliyunVPCFullAccess`, `AliyunSLBFullAccess` |

### Session Duration Tuning

The `maxSessionDuration` field (3600-43200 seconds) controls how long an STS session token remains valid after assuming the role.

| Use Case | Recommended Duration | Rationale |
|----------|---------------------|-----------|
| Interactive console access | 3600 (1 hour) | Short sessions for human operators |
| API service integration | 3600 (1 hour) | Services should refresh tokens regularly |
| CI/CD pipeline | 7200-14400 (2-4 hours) | Long-running builds and deployments |
| Data migration / batch job | 28800-43200 (8-12 hours) | Long-running data processing |
| Cross-account audit | 3600 (1 hour) | Short sessions for security-sensitive operations |

**Anti-Pattern**: Setting `maxSessionDuration` to the maximum (43200) for all roles. Longer sessions mean longer exposure windows if a token is compromised. Use the shortest duration that covers the use case.

### Tag Strategy

The Pulumi and Terraform modules automatically apply standard tags:

| Tag Key | Source | Purpose |
|---------|--------|---------|
| `resource` | `"true"` | Identifies Planton-managed resources |
| `resource_name` | `metadata.name` | Links back to the manifest |
| `resource_kind` | `"alicloudramrole"` | Resource type for filtering |
| `resource_id` | `metadata.id` | Unique resource instance ID |
| `organization` | `metadata.org` | Organizational grouping |
| `environment` | `metadata.env` | Environment isolation |

User-provided tags from `spec.tags` are merged with these standard tags. User tags take precedence on conflict. Tags on RAM roles are particularly useful for:
- **Cost attribution**: Identifying which team or project owns a role
- **Audit compliance**: Filtering roles by environment (production vs. staging)
- **Automated cleanup**: Identifying roles that belong to decommissioned environments

### Security Considerations

- **No credentials in the manifest**: Alibaba Cloud credentials are injected via environment variables (`ALIBABA_CLOUD_ACCESS_KEY_ID`, `ALIBABA_CLOUD_ACCESS_KEY_SECRET`) by the runner. The manifest `spec` never contains secrets.
- **Trust policy is the security boundary**: The `assumeRolePolicyDocument` field is the most security-sensitive configuration. Review trust policies carefully — an overly permissive trust policy is the RAM equivalent of a public S3 bucket.
- **Force deletion trade-off**: Setting `force: true` enables clean `destroy` operations but also means an accidental `destroy` will detach all policies and delete the role immediately. Use `force: true` for development/staging and `force: false` for production roles where accidental deletion should be prevented.
- **Role naming conventions**: RAM role names are account-global. Use a naming convention that includes the environment and purpose (e.g., `prod-ecs-worker-role`, `staging-fc-processor-role`) to prevent collisions and enable audit trail.

## Common Anti-Patterns

| Anti-Pattern | Consequence | Planton Mitigation |
|-------------|-------------|-------------------|
| Role with no policies | Identity can authenticate but authorize nothing | `policyAttachments` field encourages bundling at creation |
| Wildcard service principal | Any service can assume the role | Trust policy is explicit JSON, encouraging specific principals |
| Wildcard account principal | Any account can assume the role | Trust policy is explicit JSON, requiring specific account IDs |
| Default session duration for all roles | Long-running jobs fail; short jobs over-expose tokens | `maxSessionDuration` is configurable per role |
| Force-delete in production | Accidental destroy deletes role and detaches all policies | `force` defaults to `false` |
| Manual policy attachment via console | Drift between desired and actual state | Bundled attachments in the manifest are the source of truth |
| Non-unique role names | Creation failure (account-global unique) | Validated with min/max length constraints |

## Conclusion

RAM role management is a solved problem at every level of the tooling spectrum — from console to control plane. What Planton adds is not a new deployment mechanism but a **bundled, validated abstraction** that eliminates the most common identity misconfigurations:

- Roles are created with policy attachments (no permissionless shells)
- Trust policies are explicit JSON (no hidden wizard defaults)
- Session duration is configurable (no one-size-fits-all default)
- Force deletion is opt-in (no accidental production teardown)
- Tags are standardized (organization, environment, resource kind)
- Credentials are externalized (never in the manifest)

The DD07 bundling decision is the architectural cornerstone: by treating role + policy attachments as a single resource, Planton makes the common case simple (one YAML resource for a complete identity setup) while leaving the advanced case possible (custom policies via `AliCloudRamPolicy`, RRSA federation via trust policy JSON, cross-account trust via explicit principal ARNs).

For teams adopting Alibaba Cloud, `AliCloudRamRole` is typically one of the first resources deployed alongside `AliCloudLogProject` — it provides the identity foundation that `AliCloudAckManagedCluster`, `AliCloudFcFunction`, and `AliCloudEcsInstance` reference for service authentication.

### References

- [Alibaba Cloud RAM Overview](https://www.alibabacloud.com/help/en/ram/product-overview/)
- [RAM Role Concepts](https://www.alibabacloud.com/help/en/ram/user-guide/overview-of-ram-roles/)
- [STS AssumeRole API](https://www.alibabacloud.com/help/en/ram/developer-reference/api-sts-2015-04-01-assumerole/)
- [Trust Policy Examples](https://www.alibabacloud.com/help/en/ram/user-guide/edit-the-trust-policy-of-a-ram-role/)
- [RRSA for ACK](https://www.alibabacloud.com/help/en/ack/ack-managed-and-ack-dedicated/user-guide/use-rrsa-to-authorize-pods-to-access-different-cloud-services/)
- [Terraform alicloud_ram_role Resource](https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/resources/ram_role)
- [Pulumi Alibaba Cloud RAM Package](https://www.pulumi.com/registry/packages/alicloud/api-docs/ram/)
