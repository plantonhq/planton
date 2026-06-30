# OpenStackSecurityGroup: Research & Design Documentation

## Introduction

OpenStack Neutron security groups provide a virtual firewall mechanism for controlling traffic to and from instances and network ports. They are one of the most fundamental and frequently used networking primitives in any OpenStack deployment -- every instance, every port, and every network-connected workload relies on security groups for access control.

This document captures the research, design decisions, and implementation rationale behind the Planton `OpenStackSecurityGroup` deployment component.

## Historical Context

### Evolution of Security Groups in OpenStack

Security groups in OpenStack have evolved through several phases:

1. **Nova-API security groups** (pre-Icehouse): Originally managed through the Nova API. These are now deprecated and should not be used in new deployments.

2. **Neutron security groups** (Icehouse+): The modern implementation, managed through the Networking (Neutron) API. This is what Planton targets.

3. **Stateless security groups** (Yoga+): A relatively recent addition allowing stateless firewall rules for performance-sensitive workloads. Stateless mode eliminates connection tracking overhead but requires explicit egress rules for return traffic.

### OpenStack Default Behavior

When a new security group is created in OpenStack, the system automatically creates two default egress rules:
- Allow all outbound IPv4 traffic
- Allow all outbound IPv6 traffic

This "allow all egress by default" policy is convenient for most use cases but violates the zero-trust principle. The `delete_default_rules` flag in this component allows operators to start with a completely empty rule set.

## Terraform Provider Analysis

### openstack_networking_secgroup_v2

The Terraform OpenStack provider's security group resource has 8 schema fields:

| Field | Type | Required | Computed | ForceNew | Notes |
|-------|------|----------|----------|----------|-------|
| `name` | string | Yes | No | No | Derived from metadata.name |
| `description` | string | No | Yes | No | |
| `region` | string | No | Yes | Yes | |
| `tenant_id` | string | No | Yes | Yes | Admin-only, excluded |
| `delete_default_rules` | bool | No | No | Yes | Create-time only |
| `stateful` | bool | No | Yes | No | Updateable |
| `tags` | set(string) | No | No | No | |
| `all_tags` | set(string) | No | Yes | No | Computed only |

### openstack_networking_secgroup_rule_v2

The rule resource has 12 schema fields:

| Field | Type | Required | Computed | ForceNew | Constraints |
|-------|------|----------|----------|----------|-------------|
| `security_group_id` | string | Yes | No | Yes | |
| `direction` | string | Yes | No | Yes | "ingress" or "egress" |
| `ethertype` | string | Yes | No | Yes | "IPv4" or "IPv6" |
| `protocol` | string | No | No | Yes | Named protocols or 0-255 |
| `port_range_min` | int | No | No | Yes | 0-65535, RequiredWith protocol |
| `port_range_max` | int | No | No | Yes | 0-65535, RequiredWith protocol |
| `remote_ip_prefix` | string | No | Yes | Yes | ConflictsWith remote_group_id |
| `remote_group_id` | string | No | Yes | Yes | ConflictsWith remote_ip_prefix |
| `remote_address_group_id` | string | No | Yes | Yes | ConflictsWith both above |
| `description` | string | No | No | Yes | |
| `region` | string | No | Yes | Yes | |
| `tenant_id` | string | No | Yes | Yes | Admin-only |

**Critical observation**: All rule fields are ForceNew. Any change to a rule requires destroying and recreating it. This is why the TF provider locks the parent security group during rule create/delete operations.

## 80/20 Design Decisions

### Included Fields

**Security Group**: `description`, `delete_default_rules`, `stateful`, `tags`, `region` -- covers all tenant-level configuration options.

**Inline Rules**: `key`, `direction`, `ethertype`, `protocol`, `port_range_min`, `port_range_max`, `remote_ip_prefix`, `remote_group_id`, `description` -- covers all practical rule definitions.

### Excluded Fields

- **`tenant_id`** (SG and rules): Admin-only field for creating resources in other tenants. Consistent with all other OpenStack components.
- **`remote_address_group_id`** (rules): Address groups are a newer Neutron extension not widely available. Can be added later if ARM needs it.
- **`all_tags`** (SG): Computed-only field that includes tags from all sources. Not useful as an input.
- **`name`** (SG): Derived from `metadata.name` following the established pattern.

### The `key` Field on Inline Rules

This is a field we added that does not exist in the Terraform provider schema. It solves a critical production issue:

**Problem**: Security group rules have no natural key. In Terraform, if you use `count` (list-index-based), inserting or removing a rule in the middle forces recreation of all subsequent rules. For 20+ rules, this creates unnecessary churn and potential downtime.

**Solution**: The `key` field provides a stable, user-defined identifier for each rule. In Terraform, we use `for_each` keyed by `key`, so changes to one rule only affect that rule. In Pulumi, the `key` becomes part of the resource name suffix.

### Inline Rules vs Standalone OpenStackSecurityGroupRule

The platform supports two modes for managing security group rules:

1. **Inline rules** (this component): Rules defined in the `OpenStackSecurityGroupSpec.rules` field. Best for self-contained security groups where all rules are known at definition time.

2. **Standalone rules** (`OpenStackSecurityGroupRule`, component 2525): Separate KRM resources with full `StringValueOrRef` FK support. Best for InfraCharts where rules need DAG visibility and cross-resource FK resolution.

This dual approach mirrors the Terraform community pattern where both `openstack_networking_secgroup_v2` with inline rules and standalone `openstack_networking_secgroup_rule_v2` resources are commonly used.

### Port Range Handling for ICMP

ICMP rules reuse the `port_range_min` and `port_range_max` fields with different semantics:
- `port_range_min` = ICMP type (0-255)
- `port_range_max` = ICMP code (0-255)

This means `port_range_min > port_range_max` is valid for ICMP (e.g., type 8 / code 0 for Echo Request). We intentionally do **not** enforce `min <= max` in our CEL validations.

## Multi-Resource IaC Pattern

This is the first Planton component that creates N+1 resources from a single spec:
- 1 `openstack_networking_secgroup_v2` resource
- N `openstack_networking_secgroup_rule_v2` resources (one per inline rule)

### Terraform Approach

```hcl
resource "openstack_networking_secgroup_rule_v2" "rules" {
  for_each          = { for rule in var.spec.rules : rule.key => rule }
  security_group_id = openstack_networking_secgroup_v2.main.id
  ...
}
```

The `for_each` is keyed by `rule.key`, providing stable state management.

### Pulumi Approach

```go
for _, rule := range spec.Rules {
    ruleName := fmt.Sprintf("%s-rule-%s", sgName, rule.Key)
    networking.NewSecGroupRule(ctx, ruleName, ruleArgs,
        pulumi.DependsOn([]pulumi.Resource{createdSG}),
    )
}
```

Each rule resource is explicitly dependent on the security group resource.

## Production Best Practices

### Zero-Trust Security Groups

For production workloads, use `delete_default_rules: true` and explicitly define all egress rules:

```yaml
spec:
  delete_default_rules: true
  rules:
    - key: egress-dns
      direction: egress
      ethertype: IPv4
      protocol: udp
      port_range_min: 53
      port_range_max: 53
      remote_ip_prefix: "0.0.0.0/0"
    - key: egress-https
      direction: egress
      ethertype: IPv4
      protocol: tcp
      port_range_min: 443
      port_range_max: 443
      remote_ip_prefix: "0.0.0.0/0"
```

### Naming Conventions for Rule Keys

Use descriptive, kebab-case keys that encode the intent:
- `allow-ssh-from-bastion`
- `egress-all-ipv4`
- `allow-postgres-from-app-subnet`
- `allow-icmp-echo-request`

### When to Use Stateless Security Groups

Stateless security groups (`stateful: false`) are appropriate when:
- The workload handles millions of connections per second
- Connection tracking overhead is measurable
- The team can manage bidirectional rules (explicit egress for return traffic)
- The OpenStack deployment supports stateless SGs (Yoga+)

## Relationship to Other Components

```
OpenStackSecurityGroup (this component)
  ├── Referenced by: OpenStackSecurityGroupRule (security_group_id FK)
  ├── Referenced by: OpenStackSecurityGroupRule (remote_group_id FK)
  ├── Referenced by: OpenStackNetworkPort (security_group_ids[] FK)
  ├── Referenced by: OpenStackInstance (security_groups[])
  └── InfraChart role: Layer 2 (created after Network, before Port/Instance)
```

## Conclusion

The `OpenStackSecurityGroup` component delivers a production-ready virtual firewall abstraction that balances convenience (inline rules) with power (standalone rule component for InfraCharts). The `key` field innovation provides stable IaC state management, and the careful CEL validation suite catches invalid configurations before they reach the OpenStack API.
