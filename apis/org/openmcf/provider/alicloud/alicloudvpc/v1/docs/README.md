# Alibaba Cloud VPC Deployment: From Console Clicks to Declarative Control Planes

## Introduction

The Virtual Private Cloud (VPC) is the networking foundation for virtually every other resource on Alibaba Cloud. It provides an isolated virtual network with its own CIDR block, a virtual router (VRouter), and a system route table. VSwitches (Alibaba Cloud's equivalent of subnets), security groups, NAT gateways, load balancers, database instances, Kubernetes clusters, and serverless functions are all deployed into a VPC. Getting the VPC right is not optional — it is the first decision that shapes the security posture, scalability ceiling, and operational complexity of everything that follows.

Despite its foundational importance, VPC deployment is surprisingly easy to get wrong. Common mistakes — overlapping CIDR ranges that prevent future peering, IPv4-only networks that require painful migration later, missing resource group assignments that complicate cost attribution — are almost always the result of manual provisioning or ad-hoc scripting. These errors compound: a poorly planned VPC CIDR block cannot be changed after creation, forcing teams to tear down and rebuild entire environments.

This document examines how VPC deployment has evolved on Alibaba Cloud, from the console wizard to modern control-plane-based automation. It explains how OpenMCF abstracts VPC creation into a developer-friendly API designed for the 80% use case, what was intentionally left out (and why), and how the Pulumi and Terraform implementations work under the hood.

## Alibaba Cloud VPC: Key Concepts

Before examining deployment methods, it is worth understanding the Alibaba Cloud VPC model and how it differs from other cloud providers.

### VPC, VRouter, and System Route Table

When you create an Alibaba Cloud VPC, three things happen automatically:

1. **VPC** — the isolated virtual network with a primary IPv4 CIDR block
2. **VRouter** — a virtual router automatically created inside the VPC, responsible for routing traffic between VSwitches and managing route tables
3. **System Route Table** — a default route table associated with the VRouter, containing system routes for intra-VPC communication

This three-in-one creation is different from AWS, where the VPC, main route table, and router are treated as more distinct entities. On Alibaba Cloud, the VRouter is an explicit, first-class resource with its own ID, and the system route table is its child.

### VSwitches, Not Subnets

Alibaba Cloud uses the term **VSwitch** (Virtual Switch) where AWS uses "subnet" and Azure uses "subnet" within a VNet. A VSwitch:

- Belongs to exactly one VPC and one Availability Zone
- Has its own CIDR block (must be within the parent VPC's CIDR)
- Cannot span multiple Availability Zones
- Is the actual network segment where resources like ECS instances, RDS databases, and ACK node pools are placed

Critically, Alibaba Cloud does not have a built-in concept of "public" vs. "private" VSwitches at the VPC level. Whether a VSwitch provides internet access depends on whether a NAT Gateway or EIP is associated with it, or whether the resources inside it have public IP addresses. This is a routing-level concern, not a VSwitch property.

### CIDR Block Constraints

Alibaba Cloud VPCs support private IPv4 CIDR blocks from the RFC 1918 ranges:

| Range | Size | Typical Use |
|-------|------|-------------|
| `10.0.0.0/8` | 16,777,216 addresses | Large deployments, many VSwitches across multiple AZs |
| `172.16.0.0/12` | 1,048,576 addresses | Medium deployments |
| `192.168.0.0/16` | 65,536 addresses | Small deployments, dev/test environments |

The mask length must be between 8 and 28. Once created, the primary CIDR block cannot be changed. Alibaba Cloud supports adding secondary CIDR blocks to an existing VPC, but this is a separate operation not covered by this component (see "What Was Left Out" below).

### IPv6 Support

Alibaba Cloud supports dual-stack VPCs. When IPv6 is enabled on a VPC, the platform allocates a `/56` IPv6 CIDR block automatically — you do not choose the IPv6 range. VSwitches within the VPC can then be assigned IPv6 CIDR blocks from this pool. IPv6 is optional and off by default, which aligns with the reality that most workloads today are IPv4-only.

### Resource Groups

Alibaba Cloud uses resource groups for organizational grouping, access control, and cost attribution. Every resource belongs to a resource group — if you do not specify one, the resource is placed in the account's default resource group. For organizations managing many VPCs across teams, assigning each VPC to a specific resource group is a best practice for access isolation and billing clarity.

## The VPC Deployment Landscape

VPC management spans a spectrum from fully manual to continuously reconciled control planes.

### Level 0: Manual Provisioning via Alibaba Cloud Console

The Alibaba Cloud Console provides a wizard-driven workflow for creating VPCs at `vpc.console.aliyun.com`. The wizard prompts for region, VPC name, CIDR block, and optionally an IPv6 configuration and resource group.

**Common Mistakes**:

1. **CIDR Overlap** — The console does not warn if the chosen CIDR block overlaps with another VPC in the same account. This only becomes a problem later when attempting to set up VPC peering or Cloud Enterprise Network (CEN) transit routing. By then, the VPC is populated with resources and cannot have its CIDR changed.

2. **Forgetting Resource Group Assignment** — The console defaults to the account's default resource group. Teams that skip this step lose the ability to control access and attribute costs at the VPC level, creating operational headaches as the environment grows.

3. **No Version Control** — Console-created VPCs have no audit trail beyond Alibaba Cloud's ActionTrail (which captures API calls, not intent). There is no way to reproduce the exact VPC configuration in another region or account without manually repeating the process.

4. **IPv6 Planning** — Enabling IPv6 after VPC creation requires modifying the VPC, which can be disruptive. Teams that skip IPv6 at creation time often face a painful retrofit later when dual-stack requirements emerge.

**Verdict**: Acceptable for learning and one-off experimentation. Not acceptable for production environments that require reproducibility, consistency, or multi-environment parity.

### Level 1: Scripted Provisioning with Alibaba Cloud CLI

The `aliyun` CLI provides imperative commands for VPC management:

```bash
# Create a VPC
aliyun vpc CreateVpc \
  --RegionId cn-hangzhou \
  --VpcName my-vpc \
  --CidrBlock 10.0.0.0/16 \
  --Description "Production VPC" \
  --EnableIpv6 false

# Query the VPC
aliyun vpc DescribeVpcs \
  --RegionId cn-hangzhou \
  --VpcName my-vpc
```

The CLI is a thin wrapper around the Alibaba Cloud OpenAPI. It provides more control than the console and can be scripted, but it is fundamentally **imperative**: you issue commands one at a time, and the responsibility for idempotency, error handling, and state tracking falls entirely on the script author.

**Key Advantage**: The CLI makes the API parameters explicit. You see every field and its constraints directly. This forces you to confront decisions (like CIDR block selection) that the console wizard might gloss over.

**Key Limitation**: No state management. If you run `CreateVpc` twice with the same parameters, you get two VPCs. Cleanup, drift detection, and dependency management are all manual.

**Verdict**: Suitable for quick, one-off operations or embedding in CI scripts for simple use cases. Not suitable for managing VPC lifecycle across environments.

### Level 2: Infrastructure as Code with Terraform

Terraform (and OpenTofu) is the dominant IaC tool for Alibaba Cloud infrastructure, using the `aliyun/alicloud` provider.

```hcl
terraform {
  required_providers {
    alicloud = {
      source  = "aliyun/alicloud"
      version = "~> 1.200"
    }
  }
}

provider "alicloud" {
  region = "cn-hangzhou"
}

resource "alicloud_vpc" "main" {
  vpc_name    = "my-vpc"
  cidr_block  = "10.0.0.0/16"
  description = "Production VPC"
  enable_ipv6 = false

  tags = {
    environment = "production"
    team        = "platform"
  }
}

output "vpc_id" {
  value = alicloud_vpc.main.id
}

output "router_id" {
  value = alicloud_vpc.main.router_id
}

output "route_table_id" {
  value = alicloud_vpc.main.route_table_id
}
```

**Strengths**:
- **Declarative**: Define the desired state; Terraform calculates the diff
- **Stateful**: Tracks what exists vs. what is defined, enabling safe updates and deletions
- **Dependency Graphing**: Automatically understands that a VSwitch depends on a VPC
- **Plan Before Apply**: `terraform plan` shows exactly what will change before any action

**The `alicloud_vpc` Resource**: The Terraform provider exposes a comprehensive resource with all VPC parameters. The `router_id` and `route_table_id` are computed attributes — you do not set them; they are read back after creation. This matches the Alibaba Cloud API behavior where VRouter and system route table are auto-created.

**State Management**: Terraform's primary operational complexity. The state file must be stored in a remote backend (OSS + TableStore for locking on Alibaba Cloud) to prevent concurrent modifications and enable team collaboration.

**Multi-Environment**: Workspaces or separate state files per environment. Most teams use a directory-per-environment structure with shared modules.

**Verdict**: The standard for production VPC management. Reproducible, auditable, and safe.

### Level 3: Infrastructure as Code with Pulumi

Pulumi provides programmatic IaC using the `pulumi-alicloud` SDK, available in Go, TypeScript, Python, C#, and Java.

```go
package main

import (
    "github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/vpc"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        network, err := vpc.NewNetwork(ctx, "my-vpc", &vpc.NetworkArgs{
            VpcName:   pulumi.String("my-vpc"),
            CidrBlock: pulumi.String("10.0.0.0/16"),
            EnableIpv6: pulumi.Bool(false),
            Tags: pulumi.StringMap{
                "environment": pulumi.String("production"),
            },
        })
        if err != nil {
            return err
        }

        ctx.Export("vpc_id", network.ID())
        ctx.Export("router_id", network.RouterId)
        ctx.Export("route_table_id", network.RouteTableId)
        return nil
    })
}
```

**Key Differentiators from Terraform**:
- **General-Purpose Languages**: Use Go, TypeScript, Python, etc., with loops, conditionals, functions, and type checking
- **State Management**: Pulumi Service (managed), self-hosted backends, or local files
- **Automation API**: Embed the Pulumi engine inside applications for custom control planes
- **Strong Typing**: Type errors caught at compile time (in Go/TypeScript), not at `apply` time

**The `vpc.Network` Resource**: This is the Pulumi equivalent of `alicloud_vpc`. The naming convention (`Network` instead of `Vpc`) follows the Pulumi SDK's auto-generation from the Terraform provider schema. It exposes the same attributes: `VpcName`, `CidrBlock`, `EnableIpv6`, `ResourceGroupId`, `Tags`, and computed outputs `RouterId` and `RouteTableId`.

**Verdict**: Excellent for teams that prefer code over HCL, or that need to embed infrastructure provisioning into larger applications.

### Level 4: Control Planes (Crossplane, OpenMCF)

Control planes represent the most advanced deployment paradigm. Unlike CLI tools that run once and exit, a control plane **continuously observes** infrastructure state and reconciles it against the desired state.

**Crossplane**: Extends the Kubernetes API to manage cloud resources. A VPC can be defined as a Kubernetes Custom Resource and managed via `kubectl apply`. The Crossplane Alibaba Cloud provider watches for these resources and provisions/reconciles VPCs in the Alibaba Cloud account.

**OpenMCF**: Provides a protobuf-defined API for cloud resources, supporting both Pulumi and Terraform as IaC backends. The key innovation is a **unified resource model** (apiVersion, kind, metadata, spec, status) that works the same regardless of the underlying IaC engine. A VPC defined as an OpenMCF manifest can be deployed via Pulumi or Terraform with the same YAML.

**The Fundamental Difference**: A CLI-based tool runs, applies changes, and exits. A control plane continuously monitors and corrects drift. This is the GitOps paradigm applied to cloud infrastructure.

## Comparative Analysis

| Criterion | Console | CLI | Terraform | Pulumi | Control Plane |
|-----------|---------|-----|-----------|--------|---------------|
| Reproducibility | None | Script-dependent | Full | Full | Full |
| State Management | None | None | Remote backend | Managed/self-hosted | Continuous |
| Drift Detection | None | None | `plan` command | `preview` command | Continuous |
| Multi-Environment | Manual | Script-dependent | Workspaces/dirs | Stacks | Namespaced |
| Learning Curve | Low | Medium | Medium | Medium-High | High |
| Team Collaboration | Poor | Poor | Good (remote state) | Good (managed state) | Native |
| Audit Trail | ActionTrail only | ActionTrail only | State + VCS | State + VCS | State + VCS + reconciliation log |
| Rollback | Manual | Manual | State-based | State-based | Automatic reconciliation |

## What OpenMCF Supports

OpenMCF provides a Kubernetes-style API for deploying Alibaba Cloud VPCs that prioritizes the 80% use case while maintaining clear extensibility boundaries.

### The 80% Case: What Is Included

The `AlicloudVpcSpec` proto exposes seven fields:

| Field | Type | Required | Default | Rationale |
|-------|------|----------|---------|-----------|
| `region` | string | Yes | — | Every VPC must be in a region. No global VPCs on Alibaba Cloud. |
| `vpc_name` | string | Yes | — | Alibaba Cloud requires a VPC name (1-128 chars). Used as the Pulumi resource name. |
| `cidr_block` | string | Yes | — | The primary IPv4 CIDR. Cannot be changed after creation. |
| `description` | string | No | `""` | Human-readable description. 1-256 characters. |
| `enable_ipv6` | bool | No | `false` | Dual-stack networking. Off by default because most workloads are IPv4-only. |
| `resource_group_id` | string | No | `""` | Organizational grouping. Defers to account default if omitted. |
| `tags` | map | No | `{}` | User-defined key-value tags, merged with OpenMCF system tags. |

This covers the vast majority of VPC creation scenarios. A developer can deploy a production VPC with three lines of spec (region, name, CIDR) and optionally add tags and a resource group for organizational hygiene.

### What Was Left Out (and Why)

Several Alibaba Cloud VPC features are intentionally excluded from this component:

**Secondary CIDR Blocks**: Alibaba Cloud allows adding up to 5 secondary CIDR blocks to an existing VPC. This is a day-2 operation that is rarely needed at VPC creation time and adds complexity to the API surface. Teams that need secondary CIDRs can use the Alibaba Cloud console or CLI to add them after the initial deployment.

**VPC Flow Logs**: Flow logs capture information about IP traffic going to and from network interfaces in the VPC. While valuable for security and debugging, flow log configuration involves selecting log stores, defining capture filters, and managing storage costs — a separate concern that belongs in a dedicated component or is managed alongside the logging infrastructure (AlicloudLogProject).

**DHCP Options**: Alibaba Cloud VPCs support custom DHCP option sets for configuring DNS servers and NTP servers within the VPC. This is an advanced networking feature used primarily in hybrid cloud scenarios and does not belong in the foundational VPC component.

**VPC Peering / CEN Attachment**: Connecting VPCs together (via peering or Cloud Enterprise Network) is a networking topology concern managed by the AlicloudCenInstance component, not by the VPC itself.

**Custom Route Tables**: The system route table is auto-created. Custom route tables (for advanced routing scenarios like transit VPCs or overlapping CIDRs) are a separate concern.

The guiding principle: the VPC component creates the VPC. Everything else — subnets (VSwitches), security groups, NAT gateways, load balancers, route tables, peering — is a separate, composable component. This keeps the VPC API surface small and predictable, which is critical for a resource that nearly every other component depends on.

### Automatic Tag Management

OpenMCF adds system tags to every resource it manages. For AlicloudVpc, the Pulumi module merges user-defined tags from `spec.tags` with system tags:

| System Tag | Value | Purpose |
|------------|-------|---------|
| `resource` | `"true"` | Identifies OpenMCF-managed resources |
| `resource_name` | metadata.name | Links the cloud resource to its OpenMCF manifest |
| `resource_kind` | `"alicloud_vpc"` | Identifies the resource type |
| `resource_id` | metadata.id (if set) | Unique resource identifier |
| `organization` | metadata.org (if set) | Organization affiliation |
| `environment` | metadata.env (if set) | Environment designation (dev, staging, prod) |

User tags always win in the merge — if a user defines a tag key that conflicts with a system tag, the user's value takes precedence (because user tags are merged after system tags in `locals.go`).

### Foreign Key Pattern

AlicloudVpc is a **foundation resource** — it has no upstream dependencies. Its outputs (particularly `vpc_id`) are consumed by downstream components via the `StringValueOrRef` pattern:

```
AlicloudVpc.status.outputs.vpc_id
    ├── AlicloudVswitch.spec.vpc_id
    ├── AlicloudSecurityGroup.spec.vpc_id
    ├── AlicloudNatGateway.spec.vpc_id
    ├── AlicloudApplicationLoadBalancer.spec.vpc_id
    ├── AlicloudNetworkLoadBalancer.spec.vpc_id
    ├── AlicloudPrivateDnsZone.spec.vpc_id
    └── ... (nearly every networking and compute resource)
```

This makes the VPC the root node in Alibaba Cloud's resource dependency graph, which is why getting it right matters so much.

## Implementation Landscape

### Pulumi Module Architecture

The Pulumi implementation consists of three files under `iac/pulumi/module/`:

**`locals.go`** — Initializes the `Locals` struct from `StackInput`. Merges system tags (resource name, kind, ID, org, environment) with user-defined `spec.tags`. System tags are set first, then user tags overwrite, giving users the final say on tag values.

**`main.go`** — The controller function `Resources()` that:
1. Calls `initializeLocals()` to prepare transformed inputs
2. Creates the Alibaba Cloud provider with the specified region
3. Creates a single `vpc.NewNetwork` resource with all spec fields mapped
4. Exports five outputs: `vpc_id`, `vpc_name`, `cidr_block`, `router_id`, `route_table_id`

The `optionalString()` helper converts empty strings to `nil`, preventing Alibaba Cloud API errors from sending empty string values for optional fields.

**`outputs.go`** — Defines output constant names as Go constants, ensuring consistency between the Pulumi module and the `stack_outputs.proto` definition.

The entry point (`iac/pulumi/main.go`) loads the stack input from Pulumi config, deserializes it into the protobuf `AlicloudVpcStackInput`, and calls `module.Resources()`.

### Terraform Module Architecture

The Terraform implementation consists of four files under `iac/tf/`:

**`variables.tf`** — Defines `metadata` and `spec` input variables as typed objects. Includes validation rules for `vpc_name` length (1-128) and `cidr_block` non-emptiness. Optional fields have explicit defaults (`""` for strings, `false` for bools, `{}` for maps).

**`locals.tf`** — Computes the final tag map by merging base tags (resource, resource_id, resource_kind, resource_name) with optional organization and environment tags, then with user-defined tags. The merge order ensures user tags take precedence.

**`main.tf`** — Creates a single `alicloud_vpc` resource. Optional fields use conditional expressions (e.g., `var.spec.description != "" ? var.spec.description : null`) to avoid sending empty values to the API.

**`provider.tf`** — Configures the `aliyun/alicloud` provider with the region from the spec. The provider version is pinned to `~> 1.200`.

**`outputs.tf`** — Exposes five outputs matching the `stack_outputs.proto` definition: `vpc_id`, `vpc_name`, `cidr_block`, `router_id`, `route_table_id`.

Both implementations create exactly one cloud resource and produce exactly the same five outputs, maintaining parity between the Pulumi and Terraform IaC engines.

## Production Best Practices

### CIDR Block Planning

The single most impactful decision when creating a VPC is the CIDR block. Once set, it cannot be changed.

**Plan for Growth**: Start with a larger CIDR than you think you need. A `/16` gives 65,536 addresses, which accommodates hundreds of VSwitches across multiple AZs. A `/24` gives only 256 addresses — fine for a dev environment, but a bottleneck for production.

**Plan for Peering**: If you intend to connect VPCs via VPC peering or Cloud Enterprise Network (CEN), their CIDR blocks **must not overlap**. A common pattern is to allocate non-overlapping `/16` blocks:
- Production: `10.0.0.0/16`
- Staging: `10.1.0.0/16`
- Development: `10.2.0.0/16`

**Avoid `192.168.x.x` in Production**: The `192.168.0.0/16` range is commonly used by office networks, home routers, and VPN clients. Using it for production VPCs creates routing conflicts when establishing VPN connections from corporate networks.

### Resource Group Strategy

For organizations with multiple teams or business units:

- Assign each VPC to a resource group that matches its purpose (e.g., `rg-platform-prod`, `rg-data-staging`)
- Use resource group-level RAM policies to control who can create, modify, or delete VPCs
- Use resource group billing reports for cost attribution

### Tagging Strategy

Tags are the primary mechanism for organizing, filtering, and auditing cloud resources. Recommended tags for VPCs:

| Tag Key | Purpose | Example |
|---------|---------|---------|
| `team` | Owning team | `platform`, `data-engineering` |
| `costCenter` | Billing attribution | `eng-001`, `ops-002` |
| `environment` | Deployment environment | `production`, `staging`, `dev` |
| `project` | Project or service name | `payment-service`, `analytics` |

OpenMCF automatically adds system tags (resource name, kind, org, env), so user tags should focus on business context.

### IPv6 Considerations

Enable IPv6 at VPC creation time if there is any possibility of needing dual-stack networking in the future. Retrofitting IPv6 onto an existing VPC is possible but disruptive — it requires modifying the VPC, updating VSwitches, and reconfiguring security group rules. Enabling IPv6 at creation time costs nothing and keeps the option open.

### Monitoring and Alerting

While the VPC itself does not generate metrics in the same way that compute or database resources do, monitor these operational indicators:

- **VPC quota usage** — Alibaba Cloud limits the number of VPCs per region per account (default: 10). Monitor this to avoid hitting the ceiling during scaling events.
- **VSwitch CIDR exhaustion** — Track how many IP addresses are used vs. available in each VSwitch. Running out of IPs in a VSwitch prevents new resource launches.
- **Route table entry count** — The system route table has a maximum entry limit. Custom routes added by NAT gateways, VPN connections, or CEN transit routers consume entries.

### Security Considerations

A VPC is an isolation boundary, but isolation alone does not guarantee security:

- **Do not rely on VPC isolation as the sole security mechanism.** Security groups, network ACLs, and application-level authentication are all necessary layers.
- **Restrict VPC peering and CEN attachments** to the minimum necessary. Each peering connection widens the blast radius of a network compromise.
- **Use different VPCs for different trust levels.** Production workloads handling sensitive data should be in a separate VPC from development and staging environments.

## Conclusion

The Alibaba Cloud VPC is both simple and consequential. It creates a single resource (VPC + auto-created VRouter + system route table), but the decisions made at creation time — CIDR block, IPv6 enablement, resource group assignment — have lasting, difficult-to-reverse effects on everything built inside it.

OpenMCF's AlicloudVpc component reflects this reality with a deliberately small API surface: three required fields (region, name, CIDR) and four optional fields (description, IPv6, resource group, tags). Everything else — VSwitches, security groups, NAT gateways, load balancers, Kubernetes clusters — is a separate, composable component that references the VPC by its output ID.

This composability is the core design principle. A VPC is a building block, not a monolith. By keeping the VPC component focused on VPC creation and nothing else, OpenMCF ensures that:
- The API is predictable and easy to understand
- Changes to VPC configuration do not risk disrupting downstream resources
- Each networking concern (subnets, security, routing, DNS) can evolve independently
- The dependency graph is explicit, not implicit

For production deployments, plan your CIDR blocks carefully, use resource groups for organizational hygiene, tag consistently, and enable IPv6 at creation time if there is any future possibility of needing it. These decisions are easy to make at the start and painful to fix later.

## References

- [Alibaba Cloud VPC Documentation](https://www.alibabacloud.com/help/en/vpc/product-overview/what-is-a-vpc)
- [Alibaba Cloud VPC API Reference](https://www.alibabacloud.com/help/en/vpc/developer-reference/api-vpc-2016-04-28-createvpc)
- [Terraform alicloud_vpc Resource](https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/resources/vpc)
- [Pulumi Alicloud vpc.Network Resource](https://www.pulumi.com/registry/packages/alicloud/api-docs/vpc/network/)
- [RFC 1918 - Address Allocation for Private Internets](https://datatracker.ietf.org/doc/html/rfc1918)
