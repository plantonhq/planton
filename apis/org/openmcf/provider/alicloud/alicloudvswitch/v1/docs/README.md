# Alibaba Cloud VSwitch Deployment: From Console Configuration to Declarative Control Planes

## Introduction

The VSwitch (Virtual Switch) is the fundamental subnet resource in Alibaba Cloud VPC networking. Every VPC-aware resource — ECS instances, RDS databases, Kubernetes node pools, NAT gateways, load balancers, serverless functions — must be placed into a VSwitch. The name "VSwitch" is Alibaba Cloud's proprietary term; it is functionally equivalent to an AWS Subnet, GCP Subnetwork, or Azure Subnet.

Unlike a VPC, which is a regional, logical isolation boundary, a VSwitch is zone-scoped and address-specific. It carves out a CIDR range from the parent VPC and pins it to a single Availability Zone. This makes VSwitch creation a high-stakes decision: three of its core properties (VPC, zone, CIDR block) are immutable after creation. A mistake in any of these requires destroying and recreating the VSwitch along with every resource deployed into it.

This document traces how VSwitch deployment has evolved on Alibaba Cloud, explains how OpenMCF abstracts VSwitch creation into a developer-friendly API, details the Pulumi and Terraform implementations, and documents the production best practices that make the difference between a well-planned network and a topology that blocks growth.

## Alibaba Cloud VSwitch: Key Concepts

### VSwitch Naming: Why Not "Subnet"?

Alibaba Cloud's decision to use "VSwitch" instead of "Subnet" reflects its networking model. In the Alibaba Cloud VPC architecture, the VPC contains a VRouter (virtual router) that manages routing between VSwitches. A VSwitch is conceptually a Layer 2 network segment attached to the VRouter — hence "Virtual Switch." The name is unique to Alibaba Cloud but the concept is universal:

| Cloud Provider | Term | Scope |
|----------------|------|-------|
| Alibaba Cloud | VSwitch | VPC + single AZ |
| AWS | Subnet | VPC + single AZ |
| Google Cloud | Subnetwork | VPC (regional, but auto mode creates per-AZ) |
| Azure | Subnet | VNet (no explicit AZ binding) |

### VPC-VSwitch Relationship

Every VSwitch belongs to exactly one VPC and one Availability Zone. The VSwitch's CIDR block must be a subset of the parent VPC's CIDR block. This creates a strict hierarchy:

```
VPC (10.0.0.0/16, cn-hangzhou)
├── VSwitch-A (10.0.0.0/24, cn-hangzhou-a)
├── VSwitch-B (10.0.1.0/24, cn-hangzhou-b)
├── VSwitch-C (10.0.2.0/24, cn-hangzhou-c)
└── ... up to 150 VSwitches per VPC (default quota)
```

The VPC provides the address space and the routing fabric. The VSwitch provides the actual network segment where resources are deployed. You cannot deploy a resource "in a VPC" without specifying a VSwitch — the VSwitch is the mandatory placement target.

### Immutability: Three ForceNew Fields

Three VSwitch fields are immutable in the Terraform provider (and consequently in Pulumi), meaning a change triggers destroy-and-recreate:

- **`vpc_id`** — cannot move a VSwitch between VPCs
- **`zone_id`** — cannot move a VSwitch between Availability Zones
- **`cidr_block`** — cannot resize or re-address a VSwitch after creation

This immutability is a direct consequence of the Alibaba Cloud API. The `ModifyVSwitchAttribute` API only supports changing the VSwitch name and description. Any topology change (VPC, zone, CIDR) requires `DeleteVSwitch` followed by `CreateVSwitch`. If the VSwitch contains resources (ECS instances, RDS databases, etc.), they must be removed first.

This makes VSwitch planning a critical upfront decision, particularly for CIDR allocation.

### CIDR Constraints

VSwitch CIDR blocks follow strict rules:

- Must be a subset of the parent VPC's CIDR block
- Mask length: 16-29 (giving between 65,536 and 8 addresses)
- No overlap between VSwitches in the same VPC
- Alibaba Cloud reserves 4 IP addresses per VSwitch:
  - First address (network address)
  - Second address (gateway)
  - Third address (reserved for DHCP)
  - Last address (broadcast)

A `/24` VSwitch therefore provides 252 usable addresses, not 256. A `/28` VSwitch provides 12 usable addresses. For Kubernetes workloads that require many pod IPs, these reservations matter at small CIDR sizes.

### IPv6 Support: A Two-Step Dependency

IPv6 on a VSwitch requires a prerequisite chain:

1. The parent VPC must have `enable_ipv6 = true`. Alibaba Cloud allocates a `/56` IPv6 CIDR block to the VPC automatically — you cannot choose the IPv6 range.
2. The VSwitch must set `enable_ipv6 = true` and provide `ipv6_cidr_block_mask` (0-255) to select a `/64` segment from the VPC's `/56` allocation.

The resulting IPv6 CIDR block is computed by Alibaba Cloud and returned as the `ipv6_cidr_block` output. This means IPv6 on a VSwitch is always a dependent decision — you must plan for it at the VPC level first.

### Default VSwitches

Alibaba Cloud supports a "default VSwitch" concept per Availability Zone, created automatically in the default VPC or via `CreateDefaultVSwitch`. Default VSwitches use a system-assigned CIDR block and exist to support quick-start workflows.

This component does not expose the default VSwitch feature. Managed infrastructure should use explicit CIDR planning rather than system-assigned defaults, as default CIDRs cannot be predicted and may conflict with carefully planned address spaces.

### Tags

VSwitches support key-value tags for organizational grouping, cost attribution, and automated management. Tags are the primary mechanism for identifying and filtering resources across large accounts.

## The VSwitch Deployment Landscape

VSwitch management spans a spectrum from fully manual to continuously reconciled control planes. Because a VSwitch is a dependent resource (it requires a VPC), the deployment story always begins with "the VPC already exists."

### Level 0: Manual Provisioning via Alibaba Cloud Console

The Alibaba Cloud Console provides a VSwitch creation form at `vpc.console.aliyun.com`. After selecting the parent VPC, the wizard prompts for name, zone, CIDR block, and an optional description.

**Common Mistakes**:

1. **CIDR Overlap** — The console validates that the CIDR falls within the VPC range but does not proactively warn about overlap with existing VSwitches until creation fails. In a VPC with many VSwitches, this requires manually tracking which CIDRs are already allocated.

2. **Wrong AZ Selection** — The console presents a dropdown of available zones. Selecting the wrong zone is easy and irreversible — you cannot move a VSwitch to a different zone after creation. If downstream resources (RDS, ACK nodes) require specific zones, the VSwitch must match.

3. **CIDR Sizing Misjudgment** — A common pattern is to start with `/28` or `/27` VSwitches to "conserve addresses," then discover later that Kubernetes pod networking or RDS multi-AZ deployments require hundreds or thousands of IPs. Since CIDR cannot be resized, the only fix is to create a new, larger VSwitch and migrate resources.

4. **No Audit Trail of Intent** — The console creates a VSwitch via API call, but there is no record of why that CIDR was chosen, which tier it serves, or how it relates to the broader network design. This context lives in the operator's head and is lost when they leave.

**Verdict**: Acceptable for learning and one-off experimentation. Not acceptable for production environments where VSwitch topology must be reproducible, auditable, and consistent across environments.

### Level 1: Scripted Provisioning with Alibaba Cloud CLI

The `aliyun` CLI provides imperative commands for VSwitch management:

```bash
# Create a VSwitch
aliyun vpc CreateVSwitch \
  --RegionId cn-hangzhou \
  --VpcId vpc-bp1234567890abcdef \
  --ZoneId cn-hangzhou-a \
  --CidrBlock 10.0.0.0/24 \
  --VSwitchName app-vswitch-a \
  --Description "Application tier in zone A"

# Query VSwitches in a VPC
aliyun vpc DescribeVSwitches \
  --RegionId cn-hangzhou \
  --VpcId vpc-bp1234567890abcdef

# Modify attributes (name and description only)
aliyun vpc ModifyVSwitchAttribute \
  --VSwitchId vsw-bp1234567890abcdef \
  --VSwitchName new-name

# Delete
aliyun vpc DeleteVSwitch \
  --VSwitchId vsw-bp1234567890abcdef
```

**Key Advantage**: The CLI makes every parameter explicit. The `CreateVSwitch` API requires `VpcId`, `ZoneId`, and `CidrBlock` — there is no ambiguity about what is being created. Scripts can be version-controlled and parameterized.

**Key Limitation**: No state management. Running `CreateVSwitch` twice with the same parameters creates two VSwitches (with different IDs but potentially overlapping CIDRs if the API allows it). Cleanup, drift detection, and dependency ordering are all manual responsibilities.

**Verdict**: Suitable for quick operations or simple automation scripts. Not suitable for managing VSwitch topology as part of a larger infrastructure stack.

### Level 2: Infrastructure as Code with Terraform

Terraform (and OpenTofu) provides declarative VSwitch management using the `aliyun/alicloud` provider:

```hcl
resource "alicloud_vswitch" "app" {
  vpc_id       = alicloud_vpc.main.id
  zone_id      = "cn-hangzhou-a"
  cidr_block   = "10.0.0.0/24"
  vswitch_name = "app-vswitch-a"
  description  = "Application tier in zone A"

  tags = {
    tier        = "application"
    environment = "production"
  }
}
```

**Strengths**:

- **Declarative**: Define the desired VSwitch state; Terraform calculates the minimal diff
- **Dependency Management**: `alicloud_vpc.main.id` creates an implicit dependency — the VPC is created before the VSwitch
- **State Tracking**: Terraform knows the VSwitch exists and its current configuration
- **Plan Before Apply**: `terraform plan` shows that changing `cidr_block` will force replacement, giving the operator a chance to abort

**The `alicloud_vswitch` Resource**: The Terraform provider exposes `vpc_id`, `zone_id`, `cidr_block`, `vswitch_name`, `description`, `enable_ipv6`, `ipv6_cidr_block_mask`, and `tags`. The deprecated fields `name` (replaced by `vswitch_name` in v1.119.0) and `availability_zone` (replaced by `zone_id`) are still accepted but should not be used.

**Multi-VSwitch Patterns**: Terraform excels at creating multiple VSwitches with `for_each` or `count`:

```hcl
variable "zones" {
  default = ["cn-hangzhou-a", "cn-hangzhou-b", "cn-hangzhou-c"]
}

resource "alicloud_vswitch" "app" {
  for_each   = toset(var.zones)
  vpc_id     = alicloud_vpc.main.id
  zone_id    = each.key
  cidr_block = cidrsubnet(alicloud_vpc.main.cidr_block, 8, index(var.zones, each.key))
  vswitch_name = "app-${each.key}"
}
```

This pattern creates one VSwitch per zone with automatically calculated, non-overlapping CIDRs. It demonstrates why IaC is essential for multi-AZ deployments — the manual equivalent is error-prone and tedious.

**Verdict**: The standard for production VSwitch management. Reproducible, auditable, and safe.

### Level 3: Infrastructure as Code with Pulumi

Pulumi provides programmatic VSwitch management using the `pulumi-alicloud` SDK:

```go
import (
    "github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/vpc"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

vswitch, err := vpc.NewSwitch(ctx, "app-vswitch-a", &vpc.SwitchArgs{
    VpcId:       network.ID(),
    ZoneId:      pulumi.String("cn-hangzhou-a"),
    CidrBlock:   pulumi.String("10.0.0.0/24"),
    VswitchName: pulumi.String("app-vswitch-a"),
    Description: pulumi.String("Application tier in zone A"),
    Tags: pulumi.StringMap{
        "tier":        pulumi.String("application"),
        "environment": pulumi.String("production"),
    },
})
```

**Key Differentiators from Terraform**:
- **General-Purpose Languages**: Use Go, TypeScript, Python with full language capabilities (loops, conditionals, functions, type checking)
- **Strong Typing**: The `vpc.SwitchArgs` struct enforces correct field types at compile time
- **Programmatic Dependencies**: `network.ID()` creates a typed reference to the VPC output

**The `vpc.Switch` Resource**: This is the Pulumi equivalent of `alicloud_vswitch`. The naming convention (`Switch` instead of `VSwitch`) follows the Pulumi SDK's auto-generation from the Terraform provider schema. It exposes the same fields and computed outputs.

**Verdict**: Excellent for teams that prefer code over HCL, or that need to create VSwitch topologies with complex logic (e.g., calculating CIDRs based on zone count and tier requirements).

### Level 4: Control Planes (Crossplane, OpenMCF)

Control planes bring continuous reconciliation to VSwitch management.

**Crossplane**: Extends the Kubernetes API to manage VSwitch resources via Custom Resources. A VSwitch is defined as a YAML manifest, applied via `kubectl`, and continuously reconciled by the Crossplane Alibaba Cloud provider.

**OpenMCF**: Provides a protobuf-defined API with a unified resource model (apiVersion, kind, metadata, spec, status). A VSwitch manifest can be deployed via Pulumi or Terraform with the same YAML. The `StringValueOrRef` pattern enables declarative cross-resource references (e.g., a VSwitch referencing a VPC by name rather than by hardcoded ID).

**The Fundamental Difference**: CLI tools and even IaC tools run, apply, and exit. A control plane continuously observes infrastructure state and corrects drift. If someone manually deletes a VSwitch or modifies its tags, the control plane detects the divergence and restores the declared state.

## Comparative Analysis

| Criterion | Console | CLI | Terraform | Pulumi | Control Plane |
|-----------|---------|-----|-----------|--------|---------------|
| Reproducibility | None | Script-dependent | Full | Full | Full |
| State Management | None | None | Remote backend | Managed/self-hosted | Continuous |
| Drift Detection | None | None | `plan` command | `preview` command | Continuous |
| Multi-AZ Patterns | Manual per AZ | Scriptable | `for_each` / modules | Loops / functions | Declarative |
| Dependency Tracking | None | None | Implicit (HCL refs) | Typed (SDK refs) | Foreign keys |
| ForceNew Warnings | None | None | Shown in plan | Shown in preview | Shown in diff |
| Audit Trail | ActionTrail only | ActionTrail only | State + VCS | State + VCS | State + VCS + reconciliation |

## What OpenMCF Supports

OpenMCF provides a Kubernetes-style API for deploying individual Alibaba Cloud VSwitches. Each `AliCloudVswitch` resource creates one VSwitch in one Availability Zone — the fundamental building block for multi-AZ network topologies.

### The 80% Case: What Is Included

The `AliCloudVswitchSpec` proto exposes nine fields:

| Field | Type | Required | Default | Rationale |
|-------|------|----------|---------|-----------|
| `region` | string | Yes | — | Required by the Alibaba Cloud provider for API routing. Must match the VPC's region. |
| `vpc_id` | StringValueOrRef | Yes | — | The parent VPC. Uses `StringValueOrRef` to support both literal IDs and cross-resource references. |
| `zone_id` | string | Yes | — | The Availability Zone. One VSwitch per AZ is the standard pattern. |
| `cidr_block` | string | Yes | — | The IPv4 CIDR range. Immutable after creation, so the API demands explicit specification. |
| `vswitch_name` | string | Yes | — | 1-128 characters. Used as the Pulumi resource name. |
| `description` | string | No | `""` | Optional human-readable description. |
| `enable_ipv6` | bool | No | `false` | Dual-stack networking. Off by default because most workloads are IPv4-only. |
| `ipv6_cidr_block_mask` | int32 | No | `0` | Selects a /64 from the VPC's /56 IPv6 allocation. Only meaningful when IPv6 is enabled on both VPC and VSwitch. |
| `tags` | map | No | `{}` | User-defined tags, merged with OpenMCF system tags. |

This covers the vast majority of VSwitch creation scenarios. A developer can deploy a production VSwitch with five lines of spec (region, vpc_id, zone_id, cidr_block, vswitch_name) and optionally add IPv6, description, and tags.

### Design Decision: One VSwitch Per Resource

OpenMCF models each VSwitch as a separate resource (`AliCloudVswitch`), rather than embedding VSwitches as a repeated field inside `AliCloudVpc`. This is a deliberate choice:

- **Independent Lifecycle**: VSwitches can be created, updated, and deleted independently of their parent VPC
- **Separate State Management**: Each VSwitch has its own stack state, preventing a single large state file from becoming a bottleneck
- **Composability**: Different teams can own different VSwitches within the same VPC
- **Clarity**: A single resource creates a single VSwitch. No hidden iteration or implicit looping.

The trade-off is verbosity — a three-AZ deployment requires three separate `AliCloudVswitch` manifests. This is intentional. Explicit is better than implicit for infrastructure that is immutable after creation.

### What Was Left Out (and Why)

Several VSwitch-related features are intentionally excluded:

**Multi-VSwitch Batch Creation**: Some tools (like Terraform modules) create multiple VSwitches in a single declaration using `for_each` over a list of zones. OpenMCF keeps each VSwitch as an individual resource for lifecycle isolation. Infra charts (composed resource graphs) handle multi-VSwitch patterns at a higher level of abstraction.

**VSwitch Route Table Association**: VSwitches use the VPC's system route table by default. Custom route table associations are an advanced networking feature for transit VPC patterns and overlapping CIDRs. This belongs in a separate route table component, not in the VSwitch resource.

**Network ACLs**: Alibaba Cloud supports Network ACLs that can be associated with VSwitches for stateless packet filtering. This is a security feature orthogonal to VSwitch creation and is better managed as a separate resource with its own lifecycle.

**Default VSwitch Creation**: The `CreateDefaultVSwitch` API creates a VSwitch with a system-assigned CIDR in the default VPC. Managed infrastructure should use explicit CIDR planning, so this feature is excluded.

The guiding principle: the VSwitch component creates one VSwitch. Route tables, ACLs, and multi-AZ orchestration are separate, composable concerns.

### Automatic Tag Management

OpenMCF adds system tags to every VSwitch it manages. The Pulumi module merges user-defined tags from `spec.tags` with system tags:

| System Tag | Value | Purpose |
|------------|-------|---------|
| `resource` | `"true"` | Identifies OpenMCF-managed resources |
| `resource_name` | metadata.name | Links the cloud resource to its OpenMCF manifest |
| `resource_kind` | `"alicloud_vswitch"` | Identifies the resource type |
| `resource_id` | metadata.id (if set) | Unique resource identifier |
| `organization` | metadata.org (if set) | Organization affiliation |
| `environment` | metadata.env (if set) | Environment designation |

User tags always win in the merge — if a user defines a tag key that conflicts with a system tag, the user's value takes precedence (because user tags are merged after system tags in both `locals.go` and `locals.tf`).

### Foreign Key Pattern: StringValueOrRef

The `vpc_id` field uses the `StringValueOrRef` pattern, which allows two ways to specify the parent VPC:

**Direct Value** — for existing VPCs or cross-tool references:
```yaml
spec:
  vpcId: vpc-bp1234567890abcdef
```

**Cross-Resource Reference** — for VPCs managed by OpenMCF:
```yaml
spec:
  vpcId:
    valueFrom:
      name: my-vpc
```

The `valueFrom` reference is resolved by the OpenMCF platform before the IaC module executes. At runtime, `spec.vpc_id.GetValue()` always returns a literal VPC ID string. This pattern enables declarative dependency graphs without requiring the VSwitch module to know how to look up VPC IDs.

The `default_kind` annotation on the proto field (`AliCloudVpc`) and `default_kind_field_path` (`status.outputs.vpc_id`) tell the platform which resource type and output field to resolve against.

## Implementation Landscape

### Provider Resource Mapping

| Layer | Resource | Notes |
|-------|----------|-------|
| Terraform | `alicloud_vswitch` | Current name; deprecated aliases: `alicloud_subnet` |
| Pulumi | `vpc.Switch` | Auto-generated from TF schema; `Switch` not `VSwitch` |
| Alibaba Cloud API | `CreateVSwitch` / `DescribeVSwitches` / `ModifyVSwitchAttribute` / `DeleteVSwitch` | Only name and description are mutable |

### Deprecated Provider Fields

The Terraform provider has two deprecated fields (since v1.119.0):

| Deprecated | Replacement | Version |
|-----------|-------------|---------|
| `name` | `vswitch_name` | v1.119.0+ |
| `availability_zone` | `zone_id` | v1.119.0+ |

This component uses only the current field names.

### Pulumi Module Architecture

The Pulumi implementation consists of three files under `iac/pulumi/module/`:

**`locals.go`** — Initializes the `Locals` struct from `StackInput`. Resolves `vpc_id` from `StringValueOrRef` via `GetValue()`. Merges system tags (resource name, kind, ID, org, environment) with user-defined `spec.tags`. System tags are set first, then user tags overwrite, giving users the final say on tag values.

**`main.go`** — The controller function `Resources()` that:
1. Calls `initializeLocals()` to prepare transformed inputs
2. Creates the Alibaba Cloud provider with the specified region
3. Builds `vpc.SwitchArgs` from spec fields, conditionally including `Ipv6CidrBlockMask` only when non-zero
4. Creates a single `vpc.NewSwitch` resource
5. Exports five outputs: `vswitch_id`, `vswitch_name`, `cidr_block`, `zone_id`, `ipv6_cidr_block`

The `optionalString()` helper converts empty strings to `nil`, preventing Alibaba Cloud API errors from sending empty string values for optional fields like `description`.

**`outputs.go`** — Defines output constant names as Go constants (`OpVswitchId`, `OpVswitchName`, `OpCidrBlock`, `OpZoneId`, `OpIpv6CidrBlock`), ensuring consistency between the module and the `stack_outputs.proto` definition.

The entry point (`iac/pulumi/main.go`) loads the stack input from Pulumi config, deserializes it into the protobuf `AliCloudVswitchStackInput`, and calls `module.Resources()`.

### Terraform Module Architecture

The Terraform implementation consists of five files under `iac/tf/`:

**`provider.tf`** — Configures the `aliyun/alicloud` provider with the region from `var.spec.region`. Provider version pinned to `~> 1.200`.

**`variables.tf`** — Defines `metadata` and `spec` input variables as typed objects. Includes validation rules for `vswitch_name` length (1-128), `cidr_block` non-emptiness, `zone_id` non-emptiness, and `vpc_id` non-emptiness. Optional fields have explicit defaults (`""` for strings, `false` for bools, `0` for numbers, `{}` for maps).

**`locals.tf`** — Computes the final tag map by merging base tags (`resource`, `resource_id`, `resource_kind`, `resource_name`) with optional organization and environment tags, then with user-defined tags. The merge order ensures user tags take precedence.

**`main.tf`** — Creates a single `alicloud_vswitch` resource. Optional fields use conditional expressions (e.g., `var.spec.description != "" ? var.spec.description : null`) to avoid sending empty values to the API. The `ipv6_cidr_block_mask` is only set when non-zero.

**`outputs.tf`** — Exposes five outputs matching the `stack_outputs.proto` definition: `vswitch_id`, `vswitch_name`, `cidr_block`, `zone_id`, `ipv6_cidr_block`.

Both implementations create exactly one cloud resource and produce exactly the same five outputs, maintaining parity between the Pulumi and Terraform IaC engines.

## Production Best Practices

### CIDR Planning Within the VPC

The most consequential decision when creating VSwitches is CIDR allocation within the parent VPC's address space. Since VSwitch CIDRs cannot be changed after creation, poor planning creates permanent constraints.

**Strategy 1: Fixed Allocation per Zone and Tier**

Divide the VPC CIDR into predictable blocks per Availability Zone and network tier:

```
VPC: 10.0.0.0/16 (65,536 addresses)
├── Zone A (10.0.0.0/18, 16,384 addresses)
│   ├── app-a:  10.0.0.0/20  (4,096 addresses — Kubernetes pods, ECS instances)
│   ├── db-a:   10.0.16.0/24 (256 addresses — RDS, PolarDB, Redis)
│   └── mgmt-a: 10.0.17.0/24 (256 addresses — bastion hosts, monitoring)
├── Zone B (10.0.64.0/18, 16,384 addresses)
│   ├── app-b:  10.0.64.0/20
│   ├── db-b:   10.0.80.0/24
│   └── mgmt-b: 10.0.81.0/24
└── Zone C (10.0.128.0/18, 16,384 addresses)
    ├── app-c:  10.0.128.0/20
    ├── db-c:   10.0.144.0/24
    └── mgmt-c: 10.0.145.0/24
```

This pattern provides predictable addressing, easy firewall rules (all DB traffic is in `10.0.x.16.0/24` ranges), and clear tier separation.

**Strategy 2: Simple Sequential Allocation**

For less complex environments, allocate `/24` blocks sequentially:

```
VPC: 192.168.0.0/16
├── vswitch-a: 192.168.0.0/24 (zone A)
├── vswitch-b: 192.168.1.0/24 (zone B)
└── vswitch-c: 192.168.2.0/24 (zone C)
```

**Common Pitfall: Undersizing Application Tier VSwitches**

Kubernetes workloads (ACK) require IP addresses for both nodes and pods. With Alibaba Cloud's Terway networking plugin in ENIIP mode, each pod consumes a VPC IP address. An ACK cluster with 10 nodes running 50 pods each needs 500+ IP addresses for pods alone. A `/24` VSwitch (252 usable addresses) would be exhausted immediately. Use `/20` or larger for Kubernetes VSwitches.

### IP Address Reservation

Alibaba Cloud reserves 4 IP addresses per VSwitch. When planning capacity, account for these reservations:

| VSwitch CIDR | Total IPs | Reserved | Usable |
|-------------|-----------|----------|--------|
| /24 | 256 | 4 | 252 |
| /22 | 1,024 | 4 | 1,020 |
| /20 | 4,096 | 4 | 4,092 |
| /16 | 65,536 | 4 | 65,532 |
| /28 | 16 | 4 | 12 |
| /29 | 8 | 4 | 4 |

For VSwitches smaller than `/26` (64 addresses), the 4-address reservation becomes a significant percentage of the total capacity. Avoid `/28` and `/29` VSwitches except for highly specialized use cases (e.g., a management VSwitch with a single bastion host).

### Multi-AZ Strategy

For production environments, deploy VSwitches in at least two Availability Zones:

- **ACK (Kubernetes)** — requires at least 2 VSwitches in different AZs for control plane HA
- **ALB / NLB (Load Balancers)** — require zone mappings with VSwitches in at least 2 AZs
- **RDS (Databases)** — HA configurations deploy primary and standby instances in different AZs, each requiring a VSwitch
- **NAT Gateway** — deployed in one VSwitch but provides SNAT for other VSwitches in the same VPC

A common mistake is creating all VSwitches in a single AZ for simplicity, then discovering that HA services cannot be deployed. Start with at least two AZs, even for non-production environments, to avoid topology changes later.

### Naming Conventions

Consistent VSwitch naming enables filtering, automation, and troubleshooting:

```
{environment}-{tier}-{zone}
```

Examples: `prod-app-cn-hangzhou-a`, `staging-db-cn-shanghai-b`, `dev-mgmt-us-west-1a`

This convention makes it immediately clear from the VSwitch name what it is used for, which environment it belongs to, and which zone it is in.

### Tags as First-Class Metadata

Beyond the system tags that OpenMCF adds automatically, apply business-context tags:

| Tag Key | Purpose | Example |
|---------|---------|---------|
| `tier` | Network tier | `application`, `database`, `management` |
| `team` | Owning team | `platform`, `data-engineering` |
| `costCenter` | Billing attribution | `eng-001`, `ops-002` |
| `purpose` | Specific use case | `ack-nodes`, `rds-primary`, `bastion` |

Tags enable filtering in the Alibaba Cloud console (`aliyun vpc DescribeVSwitches --Tags '[{"Key":"tier","Value":"application"}]'`) and are essential for cost allocation when multiple teams share a VPC.

### IPv6 Planning

If there is any possibility of needing dual-stack networking, plan for IPv6 at VSwitch creation time:

1. Enable IPv6 on the parent VPC first (this allocates a `/56` block)
2. Assign non-overlapping `ipv6_cidr_block_mask` values to each VSwitch (0-255)
3. Document the mask-to-VSwitch mapping to avoid collisions when adding VSwitches later

Retrofitting IPv6 onto existing VSwitches requires modifying each VSwitch, which triggers resource updates and may require security group rule changes to allow IPv6 traffic.

### Monitoring and Capacity Management

While VSwitches do not generate metrics directly, monitor these operational indicators:

- **Available IP Count** — Alibaba Cloud provides `AvailableIpAddressCount` in the `DescribeVSwitches` API response. Alert when a VSwitch drops below 20% available IPs.
- **VSwitch Quota** — Default limit is 150 VSwitches per VPC. Monitor usage to avoid hitting the ceiling during rapid scaling.
- **Resource Count per VSwitch** — Track how many ECS instances, ENIs, and other resources are deployed in each VSwitch. Alibaba Cloud limits the number of ENIs per VSwitch.

### Security Considerations

- **VSwitch isolation is not security isolation.** Resources in different VSwitches within the same VPC can communicate freely by default. Use security groups for access control between tiers.
- **Database tier VSwitches** should be used exclusively for database resources. Avoid deploying application instances and database instances in the same VSwitch — this makes it harder to write precise security group rules.
- **Do not expose VSwitches directly to the internet.** Use NAT Gateways (AliCloudNatGateway) for outbound access and load balancers (AliCloudApplicationLoadBalancer, AliCloudNetworkLoadBalancer) for inbound access.

## Conclusion

The Alibaba Cloud VSwitch is architecturally simple — a single resource with a handful of fields — but operationally consequential. Its immutable properties (VPC, zone, CIDR) mean that mistakes in planning are expensive to fix. Its position as the mandatory placement target for every VPC-aware resource means that VSwitch topology shapes the entire deployment architecture.

OpenMCF's AliCloudVswitch component reflects this with a focused API: five required fields that force explicit decisions about VPC, zone, CIDR, and naming, plus four optional fields for IPv6, description, and tags. The `StringValueOrRef` pattern for `vpc_id` enables declarative dependency management without sacrificing simplicity.

For production deployments: plan CIDRs carefully with growth in mind, deploy in at least two Availability Zones, size application-tier VSwitches generously for Kubernetes workloads, and use consistent naming and tagging for operational hygiene. These decisions are trivial to make at creation time and painful to fix later.

## References

- [Alibaba Cloud VSwitch Documentation](https://www.alibabacloud.com/help/en/vpc/user-guide/create-a-vswitch)
- [Alibaba Cloud VSwitch API Reference](https://www.alibabacloud.com/help/en/vpc/developer-reference/api-vpc-2016-04-28-createvswitch)
- [Terraform alicloud_vswitch Resource](https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/resources/vswitch)
- [Pulumi AliCloud vpc.Switch Resource](https://www.pulumi.com/registry/packages/alicloud/api-docs/vpc/switch/)
- [Alibaba Cloud VPC CIDR Block Planning](https://www.alibabacloud.com/help/en/vpc/user-guide/plan-cidr-blocks-for-a-vpc)
