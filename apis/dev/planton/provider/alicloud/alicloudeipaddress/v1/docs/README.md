# Alibaba Cloud Elastic IP Address: From Console Allocation to Control Plane Automation

## Introduction

An Elastic IP Address (EIP) on Alibaba Cloud is a standalone, static public IPv4 address that exists independently of any compute or networking resource. Unlike the ephemeral public IP automatically assigned to an ECS instance at launch — which is released the moment the instance stops — an EIP persists until explicitly freed. It can be associated with an ECS instance today, moved to a NAT gateway tomorrow, and re-pointed to a load balancer next week, all without changing the address itself.

This independence makes EIPs a foundational networking primitive. Every internet-facing architecture on Alibaba Cloud — whether it routes outbound traffic through a NAT gateway, terminates inbound traffic on an ALB, or connects on-premise networks through a VPN gateway — depends on at least one EIP. Despite this ubiquity, EIP management is rife with subtle decisions around bandwidth metering, ISP line selection, and immutable-after-creation constraints that catch operators off guard when misconfigured.

This document surveys the full landscape of EIP deployment methods on Alibaba Cloud, from manual console allocation through scripted CLI workflows to modern Infrastructure as Code approaches. It explains the design decisions behind the Planton `AliCloudEipAddress` component, documents the 80/20 scoping rationale, and provides production best practices for bandwidth planning, cost optimization, and lifecycle management.

## Evolution and Historical Context

### The Early Days: VPC-Bound Public IPs

When Alibaba Cloud's VPC product launched, the only way to give an ECS instance internet connectivity was to select "Assign Public IP" during creation. This address was tightly bound to the instance: stop the instance, lose the address. For stateless web servers behind a load balancer this was acceptable, but for any service requiring a stable public endpoint — DNS servers, VPN termination points, whitelisted API endpoints — it was untenable.

### Introduction of Elastic IP Addresses

Alibaba Cloud introduced EIPs to decouple the public address lifecycle from the resource lifecycle, mirroring the AWS Elastic IP and Azure Public IP concepts. The key innovation was the separation of three concerns:

1. **Address allocation** — requesting a public IP from the regional pool
2. **Address association** — binding the address to a specific resource (ECS, NAT, SLB, etc.)
3. **Address release** — returning the address to the pool

This three-phase lifecycle means an EIP can survive instance replacements, failovers, and blue-green deployments. The address remains constant even as the underlying infrastructure changes.

### Modern EIP: BGP Multi-Line and ISP Selection

The EIP product evolved significantly with the introduction of multiple ISP line types. In mainland China, internet traffic quality varies dramatically between carriers (China Telecom, China Unicom, China Mobile). Alibaba Cloud addressed this with:

- **BGP multi-line** — the default, routing traffic through multiple carriers dynamically for optimal performance
- **BGP_PRO** — a premium tier with Alibaba Cloud's optimized BGP routing, offering lower latency and fewer hops within mainland China
- **Single-carrier lines** — dedicated China Telecom, China Unicom, or China Mobile lines for operators who need carrier-specific routing

For international regions, `BGP_International` provides standard multi-carrier routing without the mainland-specific optimizations.

The critical detail that trips up new users: both `isp` and `internet_charge_type` are **immutable after creation**. Changing either requires destroying the EIP and creating a new one — which means a new IP address. Any downstream DNS records, firewall whitelists, or partner integrations pointing to the old address break immediately.

### Bandwidth Metering Evolution

Alibaba Cloud offers two fundamentally different charging models for EIP bandwidth:

- **PayByTraffic** — the bandwidth value acts as a ceiling, and you pay per GB of outbound data transfer. Inbound traffic is free. This model dominates for most workloads because it scales with actual usage.
- **PayByBandwidth** — you pay for the full bandwidth allocation 24/7 regardless of usage. This model is cost-effective only when utilization is consistently high (>30% of allocated bandwidth).

Understanding this distinction is essential for cost optimization. A 100 Mbps EIP on PayByTraffic costs nothing when idle; the same EIP on PayByBandwidth costs the full rate continuously.

## Deployment Methods Landscape

### Level 0: Manual Allocation via Alibaba Cloud Console

The Alibaba Cloud console provides a straightforward EIP creation wizard under **VPC > Elastic IP Addresses > Create EIP**.

**Workflow**:
1. Select region
2. Choose ISP line type (BGP, BGP_PRO, single-carrier)
3. Set bandwidth (Mbps)
4. Select metering method (PayByTraffic or PayByBandwidth)
5. Optionally assign to a resource group
6. Confirm and create

**Common Mistakes**:

1. **Selecting the wrong ISP for the region**. BGP_PRO is available only in specific mainland China regions. Selecting it for an overseas region silently falls back to standard BGP in some API versions, or fails with a cryptic error in others.

2. **Misunderstanding bandwidth on PayByTraffic**. Operators set bandwidth to 1000 Mbps thinking it's free (since PayByTraffic bills per GB). While the bandwidth value itself doesn't increase the base cost, setting an unnecessarily high ceiling can lead to unexpected traffic bills if an application bursts to full capacity during a traffic spike or DDoS event.

3. **No tagging discipline**. EIPs created manually rarely get proper tags. In accounts with dozens of EIPs across multiple projects, untagged addresses become orphaned — still incurring costs but no longer associated with any active resource or identifiable owner.

4. **Forgetting that ISP and charge type are immutable**. An operator creates an EIP with PayByBandwidth, realizes it should be PayByTraffic, and discovers the console offers no "edit" option. The only path is to release the EIP and create a new one, which means a new IP address and downstream disruption.

**Verdict**: Acceptable for one-off experimentation. Unacceptable for production environments that require reproducibility, tagging discipline, or change tracking.

### Level 1: Scripted Allocation with Alibaba Cloud CLI (aliyun)

The `aliyun` CLI provides imperative EIP management through the VPC API:

```shell
# Allocate an EIP
aliyun vpc AllocateEipAddress \
  --RegionId cn-hangzhou \
  --Bandwidth 10 \
  --InternetChargeType PayByTraffic \
  --ISP BGP \
  --Description "NAT gateway EIP"

# Associate with a NAT gateway
aliyun vpc AssociateEipAddress \
  --AllocationId eip-abc123 \
  --InstanceId ngw-xyz789 \
  --InstanceType Nat

# Unassociate
aliyun vpc UnassociateEipAddress \
  --AllocationId eip-abc123 \
  --InstanceId ngw-xyz789 \
  --InstanceType Nat

# Release
aliyun vpc ReleaseEipAddress \
  --AllocationId eip-abc123
```

**Key Advantages**:
- Scriptable and repeatable
- Can be embedded in CI/CD pipelines
- API error messages are more explicit than console errors
- Supports JSON output for parsing in automation scripts

**Key Limitations**:
- **No state management**. The script doesn't track what was created previously. Re-running the allocation script creates a *second* EIP rather than being idempotent.
- **Manual dependency ordering**. The operator must ensure the EIP exists before attempting association. There is no dependency graph.
- **No drift detection**. If someone modifies the EIP through the console (changing bandwidth, for example), the script has no awareness of the change.

**OpenAPI SDK Alternative**: Alibaba Cloud provides SDKs for Go, Java, Python, and other languages. The Go SDK is used internally by both the Terraform and Pulumi providers.

**Verdict**: Suitable for simple automation scripts and one-off operations. Not appropriate for managing EIPs as part of a larger, stateful infrastructure stack.

### Level 2: Configuration Management (Ansible)

Ansible provides an `ali_eip` module through the `community.general` collection for managing Alibaba Cloud EIPs:

```yaml
- name: Allocate EIP
  ali_eip:
    alicloud_region: cn-hangzhou
    bandwidth: 10
    internet_charge_type: PayByTraffic
    state: present
  register: eip_result
```

**The Mismatch**: Ansible is designed primarily for configuring software on servers, not for provisioning standalone cloud resources. Its state management for cloud resources is primitive compared to purpose-built IaC tools:

- No remote state file — idempotency depends on fragile attribute matching
- No dependency graph between resources — playbook ordering is manual
- No plan/preview capability — changes are applied directly

**Verdict**: Not recommended for EIP management. Use IaC tools (Terraform/Pulumi) for cloud resource provisioning; use Ansible for server-level configuration.

### Level 3: Infrastructure as Code (Terraform and Pulumi)

This is the standard approach for managing EIPs as part of a larger infrastructure stack.

#### Terraform / OpenTofu

Terraform manages EIPs through the `alicloud_eip_address` resource in the `alicloud` provider:

```hcl
resource "alicloud_eip_address" "nat" {
  address_name         = "prod-nat-eip"
  bandwidth            = "10"
  internet_charge_type = "PayByTraffic"
  isp                  = "BGP"
  tags = {
    purpose = "nat"
    team    = "platform"
  }
}
```

**Key Behaviors**:
- `bandwidth` is a **string** in the provider (e.g., `"10"`), not a number, because the underlying API accepts a string
- `isp` and `internet_charge_type` are marked `ForceNew` — changing either triggers destroy-and-recreate
- The resource exports `id` (the allocation ID) and `ip_address` (the public IPv4 address)

**Association Pattern**: EIP association is handled by a separate resource, `alicloud_eip_association`, not by the EIP resource itself. This separation mirrors the Alibaba Cloud API's three-phase lifecycle (allocate, associate, release).

```hcl
resource "alicloud_eip_association" "nat" {
  allocation_id = alicloud_eip_address.nat.id
  instance_id   = alicloud_nat_gateway.main.id
  instance_type = "Nat"
}
```

**State Management**: Terraform tracks the EIP's allocation ID in its state file. Subsequent `terraform apply` runs are idempotent — no duplicate EIPs are created. If someone deletes the EIP outside Terraform, the next `apply` detects the drift and recreates it.

#### Pulumi

Pulumi uses the `pulumi-alicloud` SDK, which wraps the same underlying Go SDK:

```go
eip, err := ecs.NewEipAddress(ctx, "nat-eip", &ecs.EipAddressArgs{
    AddressName:        pulumi.String("prod-nat-eip"),
    Bandwidth:          pulumi.String("10"),
    InternetChargeType: pulumi.String("PayByTraffic"),
    Isp:                pulumi.String("BGP"),
    Tags: pulumi.StringMap{
        "purpose": pulumi.String("nat"),
    },
})
```

**Key Differentiators from Terraform**:
- Type safety through Go's compiler
- Programmatic logic (loops, conditionals) without HCL workarounds
- First-class stack concept for multi-environment management
- Automation API for embedding in custom control planes

**Shared Behavior**: Both Terraform and Pulumi interact with the same `alicloud_eip_address` provider resource. The ForceNew semantics, string bandwidth type, and two-resource association pattern apply identically.

### Level 4: Control Planes (Crossplane, Planton)

Control planes extend the IaC pattern from "run a CLI tool, apply changes, exit" to "continuously observe and reconcile desired state."

**Crossplane**: Defines Alibaba Cloud EIPs as Kubernetes Custom Resources. A cluster-resident controller watches for EIP resources and provisions/reconciles them through the Alibaba Cloud API. Drift is automatically corrected.

**Planton**: Takes the control plane approach further by defining a protobuf-native API schema for each resource kind. The schema includes validation rules (`buf.validate`), default annotations, and typed output contracts. The `AliCloudEipAddress` resource kind can be deployed through either the Pulumi or Terraform backend, with the same manifest YAML driving both.

**The Key Advantage**: A manifest-driven approach means the EIP's configuration is a versionable, auditable YAML document. Changes flow through code review. The control plane ensures the actual state matches the declared state continuously.

## Comparative Analysis

| Method | State Management | Idempotent | Drift Detection | Multi-Resource Dependencies | Reproducible | Audit Trail |
|--------|-----------------|------------|-----------------|----------------------------|--------------|-------------|
| Console | None | No | No | Manual | No | CloudTrail only |
| CLI (`aliyun`) | None | No | No | Manual | Partial | Script versioning |
| Ansible | Primitive | Fragile | No | Playbook ordering | Partial | Playbook versioning |
| Terraform | State file | Yes | Yes | Dependency graph | Yes | State + VCS |
| Pulumi | State backend | Yes | Yes | Dependency graph | Yes | State + VCS |
| Planton | Backend state | Yes | Yes (via backend) | Manifest references | Yes | Manifest VCS |

**Key Insight**: For a standalone resource like an EIP, the deployment method matters less in isolation. The real value of IaC and control planes emerges when the EIP is part of a larger stack — referenced by a NAT gateway, which is referenced by an ACK cluster, which is referenced by node pools. Manual or scripted approaches cannot manage these dependency chains reliably.

## The Planton Approach

### Design Philosophy: Standalone Allocation, Downstream Association

The `AliCloudEipAddress` component follows a deliberate single-responsibility design: it allocates an EIP and exports its `eip_id` and `ip_address`. It does **not** handle association with a target resource.

Association is the downstream component's responsibility:
- `AliCloudNatGateway` accepts an `eip_id` field (as `StringValueOrRef`) and creates the EIP association internally
- `AliCloudApplicationLoadBalancer` and `AliCloudVpnGateway` similarly accept EIP references

This separation mirrors the Alibaba Cloud API's own resource boundaries and enables flexible topologies — one EIP can be pre-allocated and associated with different resources over time without modifying the EIP component itself.

### 80/20 Scoping Decisions

**Included (the 80%)**:

| Field | Rationale |
|-------|-----------|
| `region` | Mandatory — determines the IP pool and which resources can use the EIP |
| `address_name` | User-facing identifier; critical for managing accounts with many EIPs |
| `description` | Operational documentation embedded in the resource |
| `bandwidth` | Core billing and performance parameter |
| `internet_charge_type` | Determines the cost model; must be chosen at creation time |
| `isp` | Determines routing quality; must be chosen at creation time |
| `resource_group_id` | Enterprise organizational grouping (per design decision DD05) |
| `tags` | Operational metadata for cost allocation, ownership, and automation |

**Excluded (the 20%)**:

| Feature | Rationale |
|---------|-----------|
| EIP association | Handled by downstream components to maintain single responsibility |
| Bandwidth packages | Advanced cost optimization feature; requires separate `alicloud_common_bandwidth_package` resource |
| IPv6 EIP | IPv6 EIPs use a different API (`alicloud_eipv6_address`) with distinct semantics |
| Secondary EIP addresses | Niche feature for advanced multi-IP scenarios |
| Payment period (prepaid) | EIPs are typically pay-as-you-go; prepaid optimization is a billing concern |
| Auto-pay and renewal | Subscription management is outside the IaC scope |

### API Design Decisions

**Bandwidth as `int32`, not string**: The Alibaba Cloud provider accepts bandwidth as a string (`"10"`), but the Planton proto defines it as `optional int32` with `gte: 1, lte: 1000` validation. This provides:
- Type-safe validation at the API boundary (not just at apply time)
- Better developer experience (integer input, not quoted string)
- The Pulumi module handles the `int32` → `string` conversion internally via `fmt.Sprintf`

**Defaults via proto annotations**: The three optional fields with defaults use `(dev.planton.shared.options.default)` annotations:
- `bandwidth` defaults to `5` (Alibaba Cloud's standard minimum for PayByTraffic)
- `internet_charge_type` defaults to `"PayByTraffic"` (lowest-risk metering for most workloads)
- `isp` defaults to `"BGP"` (available in all regions, no carrier lock-in)

These defaults mean a minimal manifest needs only `region` to produce a working EIP.

**CEL validation for enum-like strings**: Rather than using a proto `enum` for `internet_charge_type` and `isp`, the API uses `string` fields with CEL validation expressions. This approach:
- Avoids proto enum versioning problems when Alibaba Cloud adds new ISP types
- Allows empty string (field not set) to trigger the default
- Provides human-readable validation error messages

### Foreign Key Pattern

The `AliCloudEipAddress` stack outputs (`eip_id`, `ip_address`) serve as foreign keys for downstream components. The `StringValueOrRef` pattern in downstream specs (like `AliCloudNatGateway.spec.eip_id`) enables two reference modes:

1. **Direct value**: `eip_id: { value: "eip-abc123" }` — for pre-existing EIPs or cross-stack references
2. **Resource reference**: `eip_id: { ref: { kind: "AliCloudEipAddress", name: "my-eip", output: "eip_id" } }` — for control-plane-managed dependency resolution

## Implementation Landscape

### Pulumi Module Architecture

The Pulumi implementation is organized into three files under `iac/pulumi/module/`:

| File | Responsibility |
|------|---------------|
| `main.go` | Controller — creates the Alibaba Cloud provider and `ecs.NewEipAddress` resource, exports outputs |
| `locals.go` | Transformations — tag computation (base tags + org/env + user tags), default resolution for bandwidth/internetChargeType/isp |
| `outputs.go` | Constants — defines output key names (`eip_id`, `ip_address`) as string constants |

**Control Flow**:
1. `main.go:Resources()` receives the stack input (deserialized from the manifest YAML)
2. `initializeLocals()` computes the merged tag map and stores the target resource reference
3. A region-scoped `alicloud.NewProvider` is created
4. `ecs.NewEipAddress` is called with resolved field values
5. `ctx.Export` publishes `eip_id` (from `eip.ID()`) and `ip_address` (from `eip.IpAddress`)

**Default Resolution**: Each optional field has a corresponding helper function in `locals.go`:
- `bandwidth(spec)` returns `*spec.Bandwidth` if set, otherwise `5`
- `internetChargeType(spec)` returns `*spec.InternetChargeType` if set, otherwise `"PayByTraffic"`
- `isp(spec)` returns `*spec.Isp` if set, otherwise `"BGP"`

**Tag Merging**: The module computes a base tag set (`resource`, `resource_name`, `resource_kind`, optionally `resource_id`, `organization`, `environment`) and merges in user-provided tags. User tags take precedence over base tags on collision.

### Terraform Module Architecture

The Terraform implementation follows the same pattern:

| File | Responsibility |
|------|---------------|
| `main.tf` | Single `alicloud_eip_address` resource with conditional null handling for optional fields |
| `variables.tf` | Input variable definitions mirroring the proto spec, with Terraform-native validation blocks |
| `outputs.tf` | Output definitions for `eip_id` and `ip_address` |
| `locals.tf` | Tag merging logic (base_tags + org_tag + env_tag + user tags) |
| `provider.tf` | Alibaba Cloud provider configuration (region from spec) |

**Notable Implementation Details**:
- Optional string fields use the pattern `var.spec.address_name != "" ? var.spec.address_name : null` to avoid sending empty strings to the API
- Bandwidth is converted to string with `tostring(var.spec.bandwidth)` to match the provider's string expectation
- Three validation blocks enforce bandwidth range, internet_charge_type enum, and isp enum

### Resource Inventory

Both modules create exactly one cloud resource:

| Resource | Terraform Type | Pulumi Type | Purpose |
|----------|---------------|-------------|---------|
| EIP | `alicloud_eip_address` | `ecs.EipAddress` | Allocates a static public IPv4 address |

This is intentionally minimal. The component's value is not in complexity but in providing a standardized, validated, tagged EIP allocation that integrates cleanly with the rest of the Planton resource graph.

## Production Best Practices

### Bandwidth Planning

**Start with PayByTraffic**: For most workloads, PayByTraffic with a reasonable bandwidth ceiling (10-50 Mbps) is the safest starting point. You pay only for actual outbound data, and the ceiling prevents runaway costs during traffic spikes.

**Switch to PayByBandwidth when**: Your EIP consistently uses >30% of its allocated bandwidth and the traffic pattern is steady (not bursty). At high sustained utilization, PayByBandwidth is cheaper per GB than PayByTraffic.

**Monitor before committing**: Use Alibaba Cloud's EIP monitoring metrics (`EipRxRate`, `EipTxRate`) for at least two weeks before choosing PayByBandwidth. Many workloads that *feel* bandwidth-heavy are actually bursty, making PayByTraffic cheaper overall.

### ISP Line Selection

| Scenario | Recommended ISP | Rationale |
|----------|----------------|-----------|
| General mainland China workloads | `BGP` | Multi-carrier, available everywhere, best default |
| Latency-sensitive China mainland apps | `BGP_PRO` | Optimized routing, fewer hops |
| Users primarily on one carrier | `ChinaTelecom` / `ChinaUnicom` / `ChinaMobile` | Direct carrier path, potentially lower latency for that carrier's users |
| International-only workloads | `BGP_International` | Standard international routing |
| Finance Cloud regions | `BGP_FinanceCloud` | Required for compliance in finance cloud |

**Critical Rule**: ISP is immutable. If you need to change ISP, you must allocate a new EIP with the new ISP, update all downstream references (DNS, NAT, LB), and release the old EIP. Plan ISP selection carefully at initial allocation.

### Immutable Field Awareness

Two fields on the EIP are ForceNew in both Terraform and Pulumi providers:

| Field | Consequence of Change |
|-------|----------------------|
| `internet_charge_type` | EIP is destroyed and recreated — **new IP address** |
| `isp` | EIP is destroyed and recreated — **new IP address** |

**Mitigation**: If you need to change these values, pre-allocate the new EIP, update DNS/references to point to the new address, verify connectivity, then release the old EIP. Never change these fields on a production EIP without a migration plan.

### Tagging Strategy

EIPs are one of the resources most prone to becoming "orphaned" — allocated but not associated with any active resource, still incurring costs. A disciplined tagging strategy prevents this:

| Tag | Purpose | Example |
|-----|---------|---------|
| `purpose` | What the EIP is used for | `nat`, `alb`, `vpn` |
| `team` | Owning team | `platform`, `networking` |
| `environment` | Deployment environment | `production`, `staging` |
| `associated_resource` | The resource this EIP should be bound to | `ngw-prod-hangzhou` |

**Regular Audits**: Query for EIPs that are allocated but not associated with any resource. These are common cost leaks, especially after infrastructure teardowns that release the associated resource but forget to release the EIP.

### Cost Optimization

**Avoid over-provisioning bandwidth**: On PayByTraffic, the bandwidth ceiling doesn't affect cost directly, but it allows higher burst throughput which generates higher traffic bills. Set the ceiling to your actual peak requirement, not an aspirational maximum.

**Release unused EIPs**: An unassociated EIP still incurs a small hourly charge (the "idle EIP fee"). In accounts with many development/testing EIPs, these charges accumulate.

**Consider bandwidth packages**: For accounts with multiple EIPs (e.g., one per NAT gateway across multiple VPCs), Alibaba Cloud's Common Bandwidth Package lets multiple EIPs share a single bandwidth pool. This is outside the Planton `AliCloudEipAddress` scope but is worth evaluating at the account level.

### Security Considerations

**Minimize EIP exposure**: An EIP is a public internet endpoint. Only allocate EIPs for resources that genuinely need internet connectivity:
- NAT gateways (outbound-only, relatively low risk)
- Load balancers (inbound, protected by security groups and WAF)
- VPN gateways (encrypted tunnel, acceptable exposure)

**Avoid assigning EIPs directly to ECS instances** when the instance only needs outbound internet access. Use a NAT gateway with an EIP instead — this keeps the instance on a private IP while providing outbound connectivity through the NAT's SNAT rules.

**DDoS protection**: Every Alibaba Cloud EIP includes basic Anti-DDoS protection (5 Gbps scrubbing). For production workloads, consider upgrading to Anti-DDoS Premium for higher scrubbing capacity and traffic diversion capabilities.

### Lifecycle Management

**Pre-allocate before dependent resources**: When building a stack (VPC → EIP → NAT → ACK), allocate the EIP before the NAT gateway. This ensures the EIP exists when the NAT gateway's `eip_id` reference resolves. Planton's dependency graph handles this automatically, but manual operators should be aware of the ordering.

**Document the IP address**: After allocation, record the `ip_address` output in your operational documentation. This address will appear in DNS records, firewall rules, partner whitelists, and monitoring dashboards. Losing track of which EIP serves which purpose is a common operational failure.

**Plan for IP changes**: Even with careful management, situations arise where an EIP must be replaced (ISP change, region migration, account restructuring). Document a runbook for EIP replacement that includes:
1. Allocate new EIP
2. Update DNS records (with appropriate TTL consideration)
3. Update firewall whitelists (both your own and partners')
4. Swap association from old EIP to new EIP
5. Monitor for connectivity issues
6. Release old EIP after confirmation period

## Conclusion

The Alibaba Cloud Elastic IP Address is a deceptively simple resource — allocate a public IP, associate it with something — but production management involves nuanced decisions around ISP selection, bandwidth metering, immutability constraints, and lifecycle planning. The gap between "working in dev" and "reliable in production" lies entirely in these operational details.

Planton's `AliCloudEipAddress` component addresses this gap by:

- **Enforcing validation at the API boundary** — bandwidth range, ISP enum, and charge type enum are validated before any cloud API call
- **Providing safe defaults** — 5 Mbps, PayByTraffic, BGP is the lowest-risk starting configuration
- **Standardizing tagging** — every EIP automatically receives resource metadata tags for traceability
- **Exposing typed outputs** — `eip_id` and `ip_address` are published as stack outputs for downstream consumption via `StringValueOrRef`
- **Maintaining single responsibility** — allocation is separate from association, keeping the component focused and composable

The component deliberately excludes bandwidth packages, IPv6 EIPs, prepaid billing, and association management. These are either advanced features serving the 20% use case, handled by dedicated components, or billing concerns outside the IaC scope. This scoping keeps the API surface minimal and the implementation straightforward — a single cloud resource with well-understood behavior.

### References

- [Alibaba Cloud EIP Product Overview](https://www.alibabacloud.com/help/en/vpc/product-overview/overview-3)
- [EIP API Reference — AllocateEipAddress](https://www.alibabacloud.com/help/en/vpc/developer-reference/api-vpc-2016-04-28-allocateeipaddress)
- [Terraform alicloud_eip_address Resource](https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/resources/eip_address)
- [Pulumi AliCloud ecs.EipAddress](https://www.pulumi.com/registry/packages/alicloud/api-docs/ecs/eipaddress/)
- [Common Bandwidth Package](https://www.alibabacloud.com/help/en/vpc/product-overview/overview-2)
- [Anti-DDoS Basic](https://www.alibabacloud.com/help/en/ddos/anti-ddos-basic/product-overview/what-is-anti-ddos-origin-basic)
