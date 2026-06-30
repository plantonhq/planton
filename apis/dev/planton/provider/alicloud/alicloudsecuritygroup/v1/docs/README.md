# Alibaba Cloud Security Groups: From Console to Control Planes

## Introduction

Alibaba Cloud Security Groups are stateful virtual firewalls that control inbound and outbound traffic for VPC-based resources. Every ECS instance, ACK node, RDS instance, and most other VPC-aware resources must be associated with at least one security group. A security group without rules is effectively wide open -- the rules are what define the access boundary.

Despite the conceptual simplicity of "create a group, add rules," production security group management is riddled with operational pitfalls. Teams create security groups with no rules, producing resources with an undefined security posture. They mix ingress and egress rules without understanding evaluation order. They forget that all rule fields except description are immutable (ForceNew), causing unexpected resource recreation during updates. And they manage the group and its rules as independent lifecycle objects, leading to drift where the group exists but its rules have been modified outside of IaC.

This document examines the full deployment landscape and explains how Planton bundles the security group with its rules into a single validated API resource.

## The Alibaba Cloud Security Group Model

### Normal vs Enterprise Security Groups

Alibaba Cloud offers two types of security groups:

- **Normal**: The default type. Supports up to 200 rules per group and 5 groups per network interface. Intra-group traffic is controlled by `inner_access_policy`.
- **Enterprise**: Supports up to 2,000 rules per group and 2 groups per network interface. Does not support the `inner_access_policy` setting -- intra-group traffic follows normal rule evaluation.

Planton v1 creates normal security groups, which cover the 80% use case. Enterprise support can be added in v2 for high-rule-count scenarios.

### Rule Evaluation

Rules are evaluated per-direction (ingress or egress) in priority order:

1. Lower priority number = higher evaluation priority (range: 1-100)
2. First matching rule wins (accept or drop)
3. If no rule matches: intra-group traffic follows `inner_access_policy`; all other traffic is denied

### Immutability Constraint

All security group rule fields except `description` are **ForceNew** in the provider. This means changing any field (protocol, port range, CIDR, policy, priority) requires destroying and recreating the rule. This is an Alibaba Cloud API limitation, not a provider design choice -- the underlying ECS API operations `AuthorizeSecurityGroup` and `RevokeSecurityGroup` do not support in-place modification.

### NIC Type

Every security group rule has a `nic_type` field that can be `internet` or `intranet`. For VPC-based security groups, `nic_type` must always be `intranet`. Since Planton requires a `vpc_id` on every security group, the NIC type is hardcoded to `intranet` in the IaC modules, removing a common source of validation errors.

## Provider Resources

### Terraform

| Resource | Purpose |
|----------|---------|
| `alicloud_security_group` | The security group itself |
| `alicloud_security_group_rule` | Individual ingress/egress rules |

### Pulumi

| Type | Package |
|------|---------|
| `ecs.SecurityGroup` | `github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/ecs` |
| `ecs.SecurityGroupRule` | `github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/ecs` |

## Planton Design Decisions

### Composite Bundling (DD07)

The security group and its rules are bundled into a single component because:

1. A security group without rules has an undefined security posture
2. Rules are tightly coupled to the group -- they cannot exist independently
3. The lifecycle of group + rules should be atomic

### Rule Direction as a Field

Unlike AWS (which separates ingress and egress into structural lists), Alibaba Cloud treats direction as a field on the rule. Planton follows the provider's native model: a single `rules` list with a `type` field for direction. This matches both the Terraform/Pulumi resource model and the underlying ECS API.

### Fields Omitted from v1

| Field | Reason for Omission |
|-------|-------------------|
| `security_group_type` | Normal type covers 80% of use cases; enterprise adds complexity around inner_access_policy |
| `nic_type` | Hardcoded to "intranet" since vpc_id is required |
| `ipv6_cidr_ip` | Edge case for v1; can be added alongside VPC IPv6 support in v2 |
| `prefix_list_id` | Advanced feature; defer to v2 |
| `source_group_owner_account` | Cross-account feature; defer to v2 |

## Common Patterns

### Web Tier (Public Ingress)

Allow HTTP/HTTPS from anywhere, allow all outbound:

```
ingress tcp 80/80 0.0.0.0/0 accept priority=1
ingress tcp 443/443 0.0.0.0/0 accept priority=2
egress all -1/-1 0.0.0.0/0 accept priority=1
```

### Database Tier (VPC-Only Ingress)

Allow database ports from VPC CIDR only, no outbound rules needed (stateful return traffic is allowed):

```
ingress tcp 3306/3306 10.0.0.0/8 accept priority=1
ingress tcp 5432/5432 10.0.0.0/8 accept priority=2
ingress tcp 6379/6379 10.0.0.0/8 accept priority=3
```

### SG-to-SG References

Allow traffic from a specific security group (e.g., web tier) rather than a CIDR range:

```
ingress tcp 8080/8080 source_sg=sg-web-tier accept priority=1
```
