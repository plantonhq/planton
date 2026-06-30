# Alibaba Cloud RAM Custom Policies: From JSON Editors to Control Planes

## Introduction

Alibaba Cloud Resource Access Management (RAM) custom policies are the fine-grained permission layer for every Alibaba Cloud account. While Alibaba Cloud ships hundreds of system-managed policies — broad grants like `AliyunOSSFullAccess` or `AliyunECSReadOnlyAccess` — production environments demand permissions scoped to specific buckets, specific API actions, and specific conditions. A custom policy is a JSON document that defines exactly what actions are allowed or denied on which resources, optionally gated by conditions like source IP, time window, or MFA status.

The gap between system policies and production security requirements is where custom policies live. A system policy like `AliyunOSSFullAccess` grants all OSS operations on all buckets across the entire account — useful for development, unacceptable for production. A custom policy restricting `oss:GetObject` and `oss:PutObject` to a single bucket (`acs:oss:*:*:my-app-data/*`) is what production actually needs. This is the least-privilege principle applied to cloud IAM: grant exactly the permissions required, nothing more.

Despite the conceptual simplicity of "write a JSON document, create a policy," production custom policy management is operationally complex. Policies have a versioning lifecycle with a hard limit of 5 versions per policy. Policy documents have a 6144-byte size limit that constrains complex multi-service permissions. Policies cannot be deleted while attached to roles, users, or groups — requiring careful lifecycle management. And the JSON document format, while flexible, has no schema validation at the API level — a syntactically valid JSON document with a misspelled action name (`oss:GetObjct`) creates a policy that silently grants nothing.

This document examines the full deployment landscape for RAM custom policies — from manual console editing to control-plane-based automation — and explains how Planton wraps the single `alicloud_ram_policy` resource into a validated, version-managed, tag-enriched component that makes the common case simple while respecting the operational complexity of policy lifecycle management.

## The RAM Policy Document: Anatomy of a Permission Grant

Before examining deployment methods, it's essential to understand what a RAM policy document actually is — because the document structure is the core of this component, not the deployment mechanism.

### Document Structure

Every Alibaba Cloud RAM policy document follows this structure:

```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["<service>:<action>"],
      "Resource": ["<arn>"],
      "Condition": {}
    }
  ]
}
```

**Version**: Always `"1"` for Alibaba Cloud RAM policies. This is not a user-configurable value — it identifies the policy language version. Unlike AWS IAM (which uses `"2012-10-17"`), Alibaba Cloud uses a simple integer version.

**Statement**: An array of permission statements. Each statement is an independent permission grant or denial. A policy can contain multiple statements for different services.

**Effect**: `"Allow"` or `"Deny"`. Deny always wins — if any statement denies an action, it is denied regardless of other Allow statements. This is the explicit deny principle.

**Action**: Service-specific API operations in the format `<service>:<action>`. Examples:
- `oss:GetObject` — read an object from OSS
- `ecs:DescribeInstances` — list ECS instances
- `rds:CreateDBInstance` — create an RDS database
- `log:PostLogStoreLogs` — write logs to SLS
- `cs:*` — all Container Service operations (wildcard)

**Resource**: Alibaba Cloud Resource Names (ARNs) identifying the target resources:
- `acs:oss:*:*:my-bucket/*` — all objects in a specific OSS bucket
- `acs:ecs:cn-hangzhou:1234567890:instance/*` — all ECS instances in cn-hangzhou
- `acs:rds:*:*:dbinstance/*` — all RDS instances across all regions
- `*` — all resources (discouraged in production)

**Condition** (optional): Context-based restrictions:
- `"IpAddress": {"acs:SourceIp": ["10.0.0.0/8"]}` — restrict to specific IP ranges
- `"DateLessThan": {"acs:CurrentTime": "2026-12-31T23:59:59Z"}` — time-bounded access
- `"Bool": {"acs:SecureTransport": "true"}` — require HTTPS
- `"StringEquals": {"acs:RequestedRegion": ["cn-hangzhou"]}` — restrict to specific regions

### The ARN Format

Alibaba Cloud ARNs follow this pattern:

```
acs:<service>:<region>:<account-id>:<resource-type>/<resource-id>
```

| Component | Description | Examples |
|-----------|-------------|----------|
| `acs` | Alibaba Cloud Service (constant prefix) | Always `acs` |
| `service` | Service identifier | `oss`, `ecs`, `rds`, `log`, `ram`, `cs` |
| `region` | Region or `*` for global/all-region | `cn-hangzhou`, `us-west-1`, `*` |
| `account-id` | Alibaba Cloud account ID or `*` | `1234567890123456`, `*` |
| `resource-type/id` | Resource path | `instance/i-abc123`, `my-bucket/*` |

OSS is a special case: bucket ARNs omit the region and account because OSS bucket names are globally unique:
- `acs:oss:*:*:my-bucket` — the bucket itself
- `acs:oss:*:*:my-bucket/*` — all objects in the bucket

### The 6144-Byte Limit

Policy documents are limited to 6144 bytes. This constraint is meaningful for complex multi-service policies. A policy granting specific actions on specific resources across 10 services with conditions can easily approach this limit. When it does, the correct pattern is to split into multiple custom policies (each under 6144 bytes) and attach all of them to the target role via `AliCloudRamRole.policyAttachments`.

### Policy Versioning

Alibaba Cloud maintains up to 5 versions per custom policy. Each time a policy document is updated, a new version is created and set as the default. The version history provides an audit trail of permission changes.

**The 5-Version Limit Problem**: After 5 updates, the next update fails with an error unless an older version is deleted first. This is not a theoretical concern — policies that are updated frequently (e.g., policies granting access to dynamically-created resources) will hit this limit within a few deployment cycles.

**The `rotateStrategy` Solution**: The `rotateStrategy` field controls what happens at the limit:
- `"None"` (default): The update fails. An administrator must manually delete an old version before the next update can proceed. This is the safe choice for policies that rarely change.
- `"DeleteOldestNonDefaultVersionWhenLimitExceeded"`: Automatically deletes the oldest non-default version to make room. This is the practical choice for policies that are updated regularly through IaC.

For IaC-managed policies — which is the entire purpose of this component — `DeleteOldestNonDefaultVersionWhenLimitExceeded` is almost always the correct choice. A policy managed by Planton will be updated on every `apply` that changes the document, and hitting the version limit would break the deployment pipeline. The only reason `None` is the default (not `DeleteOldestNonDefaultVersionWhenLimitExceeded`) is to match the Alibaba Cloud API default and avoid surprising users who expect the upstream behavior.

## The RAM Policy Deployment Landscape

### Level 0: Manual Provisioning via Alibaba Cloud Console

The Alibaba Cloud console provides a JSON editor for creating custom policies:

1. Navigate to **RAM** → **Policies** → **Create Policy**
2. Select **Script** mode (the alternative "Visual Editor" mode supports only a subset of features)
3. Paste or type the JSON policy document
4. Enter a policy name and optional description
5. Click **OK**

**Common Mistakes**:

1. **Wildcard Everything**: The fastest way to "make it work" is `{"Version":"1","Statement":[{"Effect":"Allow","Action":["*"],"Resource":["*"]}]}`. This grants full administrative access — functionally equivalent to the root account. It appears in production more often than anyone admits, usually as a "temporary" fix that becomes permanent.

2. **Misspelled Actions**: The console does not validate action names against the actual API. `oss:GetObjct` (missing the 'e') is accepted as valid JSON and creates a policy that silently grants nothing. The user discovers this only when the consuming service gets `AccessDenied` errors — a debugging nightmare because the policy *exists* and *is attached*.

3. **Wrong Resource ARN Format**: Each service has its own ARN format. Using the ECS ARN format (`acs:ecs:cn-hangzhou:*:instance/*`) for an OSS resource silently grants nothing. The console provides no cross-reference between actions and their expected resource formats.

4. **Version Exhaustion**: After editing a policy 5 times through the console, the next edit fails. The console error message is unhelpful — it doesn't explain the 5-version limit or how to resolve it. Users often create a new policy with a different name instead of managing versions, leading to policy proliferation.

5. **No Lifecycle Awareness**: Deleting a policy through the console fails if it's attached to any role, user, or group. The console doesn't show *where* the policy is attached — the user must manually check all roles, users, and groups to find and detach it. This leads to orphaned policies that nobody dares delete.

**Verdict**: Acceptable for learning policy syntax and experimenting with permissions. Unacceptable for production environments where policies must be reproducible, auditable, and version-controlled.

### Level 1: Scripted Provisioning with Alibaba Cloud CLI

The `aliyun` CLI provides direct access to the RAM policy API:

```bash
# Create a custom policy
aliyun ram CreatePolicy \
  --PolicyName my-oss-reader \
  --PolicyDocument '{
    "Version": "1",
    "Statement": [{
      "Effect": "Allow",
      "Action": ["oss:GetObject", "oss:ListObjects"],
      "Resource": ["acs:oss:*:*:my-bucket/*"]
    }]
  }' \
  --Description "Read-only access to my-bucket"

# Update a policy (creates a new version)
aliyun ram CreatePolicyVersion \
  --PolicyName my-oss-reader \
  --PolicyDocument '{ ... new document ... }' \
  --SetAsDefault true \
  --RotateStrategy DeleteOldestNonDefaultVersionWhenLimitExceeded

# Delete a policy
aliyun ram DeletePolicy --PolicyName my-oss-reader
```

**The Create-vs-Update Split**: Creating a policy and updating it are different API calls (`CreatePolicy` vs `CreatePolicyVersion`). A script that runs on both first-time creation and subsequent updates must handle both paths — check if the policy exists, then decide which API to call. This is a classic impedance mismatch between imperative CLIs and declarative desired-state management.

**The Attachment Dependency**: `DeletePolicy` fails if the policy is still attached to any entity. The CLI provides no single command to detach-and-delete. The script must:

```bash
# List all entities the policy is attached to
aliyun ram ListEntitiesForPolicy --PolicyName my-oss-reader --PolicyType Custom

# Detach from each role
aliyun ram DetachPolicyFromRole --PolicyName my-oss-reader --PolicyType Custom --RoleName some-role

# Detach from each user
aliyun ram DetachPolicyFromUser --PolicyName my-oss-reader --PolicyType Custom --UserName some-user

# Detach from each group
aliyun ram DetachPolicyFromGroup --PolicyName my-oss-reader --PolicyType Custom --GroupName some-group

# Now delete
aliyun ram DeletePolicy --PolicyName my-oss-reader
```

This multi-step detach-then-delete sequence is exactly what the `force` parameter in IaC tools automates.

**The Version Management Burden**: Listing and deleting old versions requires additional API calls:

```bash
# List versions
aliyun ram ListPolicyVersions --PolicyName my-oss-reader --PolicyType Custom

# Delete a specific version
aliyun ram DeletePolicyVersion --PolicyName my-oss-reader --VersionId v3
```

**Verdict**: Suitable for one-off policy creation during account bootstrap or CI/CD pipeline steps where idempotency is handled externally. Not suitable for managing policies at scale across multiple environments.

### Level 2: Infrastructure as Code (Terraform / OpenTofu)

Terraform's `alicloud` provider models a custom policy as a single resource:

```hcl
resource "alicloud_ram_policy" "oss_reader" {
  policy_name     = "my-oss-reader"
  policy_document = jsonencode({
    Version = "1"
    Statement = [{
      Effect   = "Allow"
      Action   = ["oss:GetObject", "oss:ListObjects"]
      Resource = ["acs:oss:*:*:my-bucket/*"]
    }]
  })
  description     = "Read-only access to my-bucket"
  rotate_strategy = "DeleteOldestNonDefaultVersionWhenLimitExceeded"
  force           = true
  tags = {
    team        = "platform"
    environment = "production"
  }
}
```

**Strengths**:

- **Declarative**: Define the desired policy state; Terraform handles create-or-update logic automatically. No need to manually switch between `CreatePolicy` and `CreatePolicyVersion`.
- **Stateful**: Tracks the policy's existence, current document, and metadata in the state file. Detects drift if someone edits the policy through the console.
- **Version Management**: The `rotate_strategy` parameter handles the 5-version limit transparently. Each `terraform apply` that changes the document creates a new version, and the strategy controls what happens at the limit.
- **Force Deletion**: The `force` parameter handles the detach-then-delete sequence automatically.
- **Tags**: First-class tag support for organizational metadata.

**Weaknesses**:

- **No Document Validation**: Terraform validates that `policy_document` is valid JSON, but it does not validate that the actions, resources, or conditions are correct for Alibaba Cloud. A policy with `oss:GetObjct` (misspelled) passes `terraform plan` and `terraform apply` without errors.
- **State Management Overhead**: The state file must be stored remotely (OSS + TableStore for locking) in team environments.
- **JSON in HCL**: Embedding JSON policy documents in HCL is ergonomically poor. The `jsonencode()` function helps but nested JSON structures in HCL are still harder to read and maintain than raw JSON or YAML.

**The `jsonencode()` Pattern**: Production Terraform modules use `jsonencode()` to construct policy documents from HCL data structures. This provides HCL-native formatting, type checking for the outer structure, and avoids string escaping issues. However, the *content* of the policy (action names, resource ARNs) remains unvalidated strings.

**Verdict**: The modern standard for managing custom policies. The single-resource model is a good fit for policies (unlike roles, which require separate attachment resources). Recommended for teams already using Terraform.

### Level 3: Infrastructure as Code (Pulumi)

Pulumi's Go SDK provides type-safe policy creation:

```go
policy, err := ram.NewPolicy(ctx, "oss-reader", &ram.PolicyArgs{
    PolicyName: pulumi.String("my-oss-reader"),
    PolicyDocument: pulumi.String(`{
        "Version": "1",
        "Statement": [{
            "Effect": "Allow",
            "Action": ["oss:GetObject", "oss:ListObjects"],
            "Resource": ["acs:oss:*:*:my-bucket/*"]
        }]
    }`),
    Description:    pulumi.String("Read-only access to my-bucket"),
    RotateStrategy: pulumi.String("DeleteOldestNonDefaultVersionWhenLimitExceeded"),
    Force:          pulumi.Bool(true),
    Tags:           pulumi.ToStringMap(map[string]string{
        "team": "platform",
    }),
})
```

**Key Advantages Over Terraform**:

- **Type Safety**: Compile-time validation of field names. Misspelling `PolicyDocument` as `PolicyDocumet` is a build error, not a runtime surprise.
- **Programmatic Composition**: Policy documents can be constructed programmatically using Go structs, enabling dynamic policy generation based on configuration:

```go
type PolicyDocument struct {
    Version   string      `json:"Version"`
    Statement []Statement `json:"Statement"`
}

type Statement struct {
    Effect   string   `json:"Effect"`
    Action   []string `json:"Action"`
    Resource []string `json:"Resource"`
}

doc := PolicyDocument{
    Version: "1",
    Statement: []Statement{
        {Effect: "Allow", Action: actions, Resource: resources},
    },
}
docJSON, _ := json.Marshal(doc)
```

This approach catches structural errors (missing `Version`, wrong `Statement` type) at compile time.

- **Multi-Language**: Same logic can be expressed in TypeScript, Python, Java, or C#.

**Key Disadvantage**: Requires compiling Go code (or running a Node/Python runtime). Terraform's declarative HCL is simpler for teams that don't need programmatic composition.

**Verdict**: Preferred for teams using Go or TypeScript, especially when policies are generated dynamically from application configuration. The type safety is valuable for catching structural errors in policy documents.

### Level 4: Control Planes and Continuous Reconciliation

The most advanced deployment model treats policy configuration as a continuously reconciled desired state:

- **Crossplane**: Extends the Kubernetes API with custom resources for Alibaba Cloud. An operator watches for RAM policy custom resources and provisions/reconciles them.
- **Custom Operators**: Organizations build controllers that watch application deployments and automatically create corresponding least-privilege policies.

**Planton Context**: Planton's protobuf-defined API is designed for this model. The YAML manifest is a desired-state declaration that can be applied once (CLI mode) or continuously reconciled (control-plane mode). The `AliCloudRamPolicy` resource is a Kubernetes-native API object, not just a CLI input format.

**Verdict**: The future of policy management in cloud-native platforms. Planton's API design anticipates this model even when used in CLI mode today.

## Comparative Analysis

| Method | Idempotent | State Tracked | Version-Safe | Force Delete | Validated | Effort |
|--------|-----------|--------------|-------------|-------------|-----------|--------|
| Console | No | No | No | No | No | 2 min clicking, raw JSON editor |
| CLI (`aliyun ram`) | No | No | Manual | Manual | No | ~20 lines of bash |
| Terraform | Yes | Yes | Yes (rotate_strategy) | Yes (force) | JSON only | 1 resource block |
| Pulumi | Yes | Yes | Yes (RotateStrategy) | Yes (Force) | Compile-time | ~15 lines of Go |
| Planton | Yes | Yes | Yes (rotateStrategy) | Yes (force) | Proto-validated | 1 YAML manifest |

The key differentiator is the **Validated** column. Terraform validates JSON syntax; Pulumi validates struct field names at compile time; Planton validates the entire resource envelope (apiVersion, kind, metadata, spec field types, string lengths, enum values) via protobuf before any cloud API call is made. A manifest with `policyName` exceeding 128 characters or `rotateStrategy` set to an invalid value is rejected at validation time, not at the Alibaba Cloud API layer.

## The Planton Approach

### Design Philosophy: Single Resource, Full Lifecycle

AliCloudRamPolicy is architecturally simple — it wraps a single `alicloud_ram_policy` resource. The value is not in resource orchestration (there's only one resource) but in four areas:

1. **Protobuf Validation**: Field types, lengths, and enum values are validated at the API layer before any IaC engine runs.
2. **Version Lifecycle Management**: The `rotateStrategy` field is surfaced as a first-class API field, not buried in provider documentation.
3. **Tag Standardization**: Standard Planton tags (`resource_name`, `resource_kind`, `organization`, `environment`) are automatically applied alongside user tags.
4. **Unified API**: The same YAML manifest format works with both Pulumi and Terraform, and the policy integrates with other Planton resources through the `status.outputs` contract.

### 80/20 Scoping: What's In and What's Out

**Included (the 80%)**:

- **Policy creation** with name, description, and JSON document
- **Version rotation** via `rotateStrategy` for policies updated through IaC
- **Force deletion** for clean teardown even when the policy is still attached
- **Tags** for organizational grouping, cost tracking, and resource filtering

**Excluded (the 20%)**:

- **Policy document validation**: Validating that action names (`oss:GetObject`) and resource ARNs (`acs:oss:*:*:bucket/*`) are correct for the target service would require an up-to-date catalog of all Alibaba Cloud service APIs. This is impractical to maintain and better handled by the Alibaba Cloud API itself at apply time.
- **Structured policy composition**: Modeling the policy document as a protobuf message (with `Statement`, `Effect`, `Action`, `Resource`, `Condition` as typed fields) would provide richer validation but would also limit flexibility. The JSON string format allows any valid policy structure, including advanced conditions and cross-service grants that a structured API might not anticipate.
- **Policy attachment management**: Attaching policies to roles is handled by `AliCloudRamRole.policyAttachments`. Attaching to users or groups is not yet in scope for Planton (users and groups are a different lifecycle domain).
- **Version history management**: Listing, inspecting, and deleting specific policy versions is an operational concern handled by the `rotateStrategy` field. Manual version management is available through the Alibaba Cloud CLI or console.
- **Policy simulator**: Alibaba Cloud provides a policy simulator tool for testing policies before applying them. Integrating this into the Planton workflow is out of scope.

### API Design Decisions

**`policyName` vs `name`**: The spec uses `policyName` (not `name`) because RAM policy names must be unique within the Alibaba Cloud account and follow specific naming rules (1-128 characters, English letters, digits, and hyphens). This is distinct from the Planton `metadata.name`, which is the local resource identifier. The metadata name identifies the Planton resource; the policy name identifies the Alibaba Cloud object.

**`policyDocument` as a string**: The policy document is a raw JSON string rather than a structured protobuf message. This mirrors the Alibaba Cloud RAM API exactly and provides maximum flexibility. Structured parsing would require modeling every possible action, resource ARN format, and condition operator — complexity that would constrain advanced use cases without providing proportional value. The trade-off is that document *content* is not validated at the proto level (only JSON syntax is validated by the IaC engine).

**`rotateStrategy` as optional string with enum validation**: The field uses `optional string` with `buf.validate` enum constraints (`in: ["None", "DeleteOldestNonDefaultVersionWhenLimitExceeded"]`). An enum proto type was considered but rejected because the Alibaba Cloud API uses string values, and mapping between proto enum names and API string values would add unnecessary complexity. The `in` validator achieves the same validation without the mapping burden.

**`force` defaults to `false`**: Force-deleting a policy detaches it from all roles, users, and groups before deletion — a destructive operation. The safe default is `false`, requiring explicit opt-in. This prevents accidental detachment during `destroy` operations. Teams should use `force: true` for development/staging policies and `force: false` for production policies where accidental deletion should be blocked.

**`region` despite RAM being global**: RAM is an account-global service — policies are not region-scoped. However, the Alibaba Cloud provider (both Terraform and Pulumi) requires a region for API endpoint configuration. The `region` field configures the provider endpoint, not the policy's scope. This is documented in the `spec.proto` comments to prevent confusion.

**`tags` as a map, not the `tags` proto annotation**: The `tags` field is a `map<string, string>` on the spec, not metadata-level tags. This is because RAM policy tags are applied to the Alibaba Cloud resource and are distinct from Planton metadata labels. The module merges user-provided tags with standard Planton tags (`resource`, `resource_name`, `resource_kind`, etc.), with user tags taking precedence on conflict.

### The Policy + Role Relationship

`AliCloudRamPolicy` and `AliCloudRamRole` form a deliberate pair:

1. **AliCloudRamPolicy** creates a custom policy document (this component)
2. **AliCloudRamRole** creates a role and attaches policies (system or custom) via `policyAttachments`

The connection point is the policy name. `AliCloudRamPolicy` outputs `policy_name` and `policy_type` (always `"Custom"`). `AliCloudRamRole` consumes these in its `policyAttachments` list:

```yaml
# Step 1: Create the custom policy
apiVersion: alicloud.planton.dev/v1
kind: AliCloudRamPolicy
metadata:
  name: oss-reader
spec:
  policyName: oss-read-only
  policyDocument: '...'
---
# Step 2: Attach it to a role
apiVersion: alicloud.planton.dev/v1
kind: AliCloudRamRole
metadata:
  name: my-ecs-role
spec:
  roleName: my-ecs-service-role
  policyAttachments:
    - policyName: oss-read-only
      policyType: Custom
```

This separation of concerns — policy authoring vs. policy attachment — mirrors the Alibaba Cloud RAM API and enables policy reuse across multiple roles.

## Implementation Landscape

### Pulumi Module Architecture

The Pulumi module consists of three files under `v1/iac/pulumi/module/`:

**`main.go`** — The controller. Entry point is `Resources(ctx, stackInput)`.
1. Initializes locals (tag computation)
2. Creates the Alibaba Cloud provider with the specified region
3. Creates a single `ram.NewPolicy` resource with all spec fields
4. Exports outputs: `policy_name`, `policy_type`

The module is minimal — no iteration, no sub-resources, no parent chaining. The single `ram.NewPolicy` call maps directly to the `alicloud_ram_policy` cloud resource.

Two helper functions handle optional proto fields:
- `optionalString(s string)` — returns `nil` for empty strings, allowing the Pulumi SDK to skip optional API parameters
- `optionalStringPtr(s *string)` — unwraps optional proto pointer fields, returning `nil` if the pointer is nil

A third helper resolves the `force` default:
- `forceDelete(spec)` — returns `*spec.Force` if set, otherwise `false`

**`locals.go`** — Tag computation.
- Computes the tag map by merging standard tags (`resource`, `resource_name`, `resource_kind`, `resource_id`, `organization`, `environment`) with user-provided `spec.Tags`
- User tags take precedence on key conflict

**`outputs.go`** — Output constant definitions.
- `OpPolicyName = "policy_name"`
- `OpPolicyType = "policy_type"`

**Resource Hierarchy**:

```
Provider (alicloud, region from spec.region)
  └── ram.Policy (spec.policyName)
```

### Terraform Module Architecture

The Terraform module consists of five files under `v1/iac/tf/`:

**`main.tf`** — Single resource:
```hcl
resource "alicloud_ram_policy" "main" {
  policy_name     = var.spec.policy_name
  policy_document = var.spec.policy_document
  description     = var.spec.description != "" ? var.spec.description : null
  rotate_strategy = var.spec.rotate_strategy
  force           = var.spec.force
  tags            = local.final_tags
}
```

**`variables.tf`** — Input variables mirroring the proto schema with inline validations:
- `policy_name` length validation (1-128)
- `rotate_strategy` enum validation ("None" or "DeleteOldestNonDefaultVersionWhenLimitExceeded")

**`locals.tf`** — Tag merging (same logic as the Pulumi module).

**`outputs.tf`** — Two outputs: `policy_name` and `policy_type`.

**`provider.tf`** — Alibaba Cloud provider with region from `var.spec.region`.

The Terraform and Pulumi modules create identical resources with identical outputs, ensuring that switching between IaC engines produces the same RAM policy.

## Production Best Practices

### Least-Privilege Policy Patterns

The most important principle in policy authoring is least privilege: grant exactly the permissions required, nothing more. Here are common patterns for Alibaba Cloud services:

#### Pattern 1: Scoped OSS Access

```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "oss:GetObject",
        "oss:PutObject",
        "oss:DeleteObject",
        "oss:ListObjects"
      ],
      "Resource": [
        "acs:oss:*:*:my-app-bucket",
        "acs:oss:*:*:my-app-bucket/*"
      ]
    }
  ]
}
```

Two resource ARNs are needed: one for the bucket itself (for `ListObjects`) and one for objects within the bucket (for `GetObject`, `PutObject`, `DeleteObject`).

#### Pattern 2: Read-Only Database Access

```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "rds:DescribeDBInstances",
        "rds:DescribeDBInstanceAttribute",
        "rds:DescribeDatabases",
        "rds:DescribeAccounts"
      ],
      "Resource": ["*"]
    }
  ]
}
```

RDS read-only access is typically granted at the account level (`"Resource": ["*"]`) because RDS instance ARNs are not known at policy creation time. The `Describe*` actions are read-only and safe to grant broadly.

#### Pattern 3: CI/CD Pipeline Permissions

```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "cr:GetRepository",
        "cr:PushRepository",
        "cr:PullRepository"
      ],
      "Resource": ["acs:cr:*:*:repository/my-org/*"]
    },
    {
      "Effect": "Allow",
      "Action": [
        "cs:DescribeClusterDetail",
        "cs:GetClusterKubeconfig",
        "cs:DescribeClusterNodes"
      ],
      "Resource": ["acs:cs:*:*:cluster/*"]
    },
    {
      "Effect": "Allow",
      "Action": [
        "log:PostLogStoreLogs",
        "log:GetLogStore"
      ],
      "Resource": ["acs:log:*:*:project/cicd-logs/*"]
    }
  ]
}
```

Multi-service policies use multiple statements, each scoped to a specific service and resource path.

#### Pattern 4: Condition-Based Access

```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["oss:*"],
      "Resource": ["acs:oss:*:*:sensitive-data/*"],
      "Condition": {
        "IpAddress": {
          "acs:SourceIp": ["10.0.0.0/8", "172.16.0.0/12"]
        },
        "Bool": {
          "acs:SecureTransport": "true"
        }
      }
    }
  ]
}
```

Conditions add defense-in-depth: even if the policy allows `oss:*`, access is restricted to internal IP ranges over HTTPS.

### Version Management Strategy

For policies managed through IaC (the primary use case for this component):

| Scenario | Recommended `rotateStrategy` | Rationale |
|----------|------------------------------|-----------|
| Rarely changing policy | `None` | Preserve version history for audit |
| Frequently updated policy | `DeleteOldestNonDefaultVersionWhenLimitExceeded` | Prevent version exhaustion |
| CI/CD pipeline policy | `DeleteOldestNonDefaultVersionWhenLimitExceeded` | Updated on every pipeline change |
| Compliance-sensitive policy | `None` | Require explicit version management |

**Recommendation for IaC**: Always set `rotateStrategy: DeleteOldestNonDefaultVersionWhenLimitExceeded` for policies managed through Planton. The 5-version limit will be hit after 5 deployments that change the policy document, breaking subsequent deployments. The automatic rotation prevents this silently.

### Tag Strategy

The Pulumi and Terraform modules automatically apply standard tags:

| Tag Key | Source | Purpose |
|---------|--------|---------|
| `resource` | `"true"` | Identifies Planton-managed resources |
| `resource_name` | `metadata.name` | Links back to the manifest |
| `resource_kind` | `"alicloudrampolicy"` | Resource type for filtering |
| `resource_id` | `metadata.id` | Unique resource instance ID |
| `organization` | `metadata.org` | Organizational grouping |
| `environment` | `metadata.env` | Environment isolation |

User-provided tags from `spec.tags` are merged with these standard tags, with user tags taking precedence on conflict. Tags on RAM policies are useful for:
- **Cost attribution**: Identifying which team owns a policy
- **Environment isolation**: Filtering policies by environment (production vs. staging)
- **Automated cleanup**: Identifying policies that belong to decommissioned environments
- **Audit compliance**: Grouping policies by security domain or compliance requirement

### Force Deletion Trade-offs

| Setting | Behavior on Destroy | Recommended For |
|---------|-------------------|-----------------|
| `force: false` (default) | Fails if policy is attached to any entity | Production policies where accidental deletion is dangerous |
| `force: true` | Detaches from all entities and deletes all non-default versions before deleting | Development, staging, and ephemeral environments |

**Production Guidance**: Use `force: false` for production policies and `force: true` for development/staging. A production policy that is accidentally destroyed while still attached to roles would break all services using those roles. The `force: false` default creates a safety net.

### Security Considerations

- **No credentials in the manifest**: Alibaba Cloud credentials are injected via environment variables (`ALIBABA_CLOUD_ACCESS_KEY_ID`, `ALIBABA_CLOUD_ACCESS_KEY_SECRET`) by the runner. The manifest `spec` never contains secrets.
- **Policy document is the security surface**: The `policyDocument` field defines what permissions exist. Review policy documents with the same rigor as code changes — an overly permissive policy is the RAM equivalent of a public-facing database.
- **Naming conventions**: RAM policy names are account-global. Use a naming convention that includes the environment and purpose (e.g., `prod-oss-reader-app-data`, `staging-cicd-deploy-policy`) to prevent collisions and enable audit trails.
- **Separate policies for separate concerns**: Prefer multiple small, focused policies over one large policy. Each policy should address a single service or use case. This makes permissions auditable and individually revocable.

## Common Anti-Patterns

| Anti-Pattern | Consequence | Planton Mitigation |
|-------------|-------------|-------------------|
| `"Action": ["*"]` on all resources | Administrative access through a "custom" policy | Proto validation catches the 80% case; JSON content validation is the user's responsibility |
| `"Resource": ["*"]` on sensitive actions | Unscoped access to all instances of a service | Best practice guidance in docs; ARN specificity encouraged in examples |
| No `rotateStrategy` with frequent updates | Version exhaustion after 5 deploys | Field surfaced as first-class API; docs recommend auto-rotation for IaC |
| `force: true` in production | Accidental `destroy` detaches policy from all entities | Default is `false`; docs emphasize production safety |
| Duplicating system policy content | Custom policy that mirrors `AliyunOSSFullAccess` exactly | Docs explain when to use system policies vs. custom |
| Single mega-policy for all services | 6144-byte limit hit; unauditable permissions | Docs recommend one policy per service/concern |
| Policy names without environment prefix | Name collisions across dev/staging/prod | Naming convention guidance in docs |
| Embedded secrets in policy conditions | Credentials in the policy document (e.g., S3 keys in conditions) | Best practice: conditions should use IP ranges, VPC IDs, or time constraints, never secrets |

## Conclusion

RAM custom policy management is architecturally simple — it's a single API resource — but operationally nuanced. The policy document JSON format, the 5-version lifecycle limit, the attachment dependency graph, and the force-deletion semantics all create sharp edges that catch teams during deployment.

Planton's `AliCloudRamPolicy` component doesn't add resource orchestration complexity (there's only one resource). Its value is in:

1. **Protobuf-validated API contract**: Field types, lengths, and enum values are validated before any cloud API call.
2. **Version lifecycle management**: `rotateStrategy` is a first-class field, preventing the most common deployment failure (version exhaustion).
3. **Tag standardization**: Automatic Planton tags alongside user tags.
4. **Unified IaC**: Same manifest works with Pulumi and Terraform.
5. **Integration contract**: `status.outputs.policy_name` and `status.outputs.policy_type` are the connection point for `AliCloudRamRole.policyAttachments`.

For teams adopting Alibaba Cloud, `AliCloudRamPolicy` is typically deployed alongside `AliCloudRamRole` — policies provide the permission definitions, roles provide the identities that use them. Together, they form the identity foundation that downstream resources (`AliCloudAckManagedCluster`, `AliCloudFcFunction`, `AliCloudEcsInstance`) depend on for service authentication and authorization.

### References

- [Alibaba Cloud RAM Policy Overview](https://www.alibabacloud.com/help/en/ram/user-guide/policy-overview/)
- [RAM Policy Structure and Syntax](https://www.alibabacloud.com/help/en/ram/user-guide/policy-structure-and-syntax/)
- [RAM Policy Elements](https://www.alibabacloud.com/help/en/ram/user-guide/policy-elements/)
- [RAM Condition Keys](https://www.alibabacloud.com/help/en/ram/user-guide/policy-elements-condition/)
- [Custom Policy Version Management](https://www.alibabacloud.com/help/en/ram/user-guide/manage-custom-policy-versions/)
- [Terraform alicloud_ram_policy Resource](https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/resources/ram_policy)
- [Pulumi Alibaba Cloud RAM Package](https://www.pulumi.com/registry/packages/alicloud/api-docs/ram/)
