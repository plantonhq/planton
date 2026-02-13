# ScalewayInstanceSecurityGroup

A Scaleway Instance Security Group is a zonal, stateful firewall that controls inbound and outbound traffic to [Scaleway Instances](https://www.scaleway.com/en/virtual-instances/). Security groups operate at the Instance level -- they are assigned to individual Instances via the Instance's `security_group_id` field, not attached to VPCs or Private Networks.

## Overview

This OpenMCF resource kind wraps the Scaleway `scaleway_instance_security_group` Terraform resource (or the Pulumi `instance.SecurityGroup` resource) into a declarative, Kubernetes-style manifest.

### Key Scaleway Security Group Concepts

- **Default policies**: Each security group has an `inbound_default_policy` and `outbound_default_policy` that determine what happens to traffic matching no rule. Set to `"accept"` (denylist model) or `"drop"` (allowlist model).
- **Rule ordering**: Rules are evaluated in order. The first matching rule wins.
- **Actions**: Rules use `"accept"` or `"drop"` (not "allow"/"deny" as in some other providers).
- **Protocols**: Scaleway uses uppercase protocol names: `"TCP"`, `"UDP"`, `"ICMP"`, `"ANY"`.
- **Stateful by default**: Return traffic for accepted connections is automatically permitted. Disable only for advanced stateless use cases.
- **SMTP security**: The `enable_default_security` flag blocks SMTP ports (25, 465, 587) by default to prevent spam abuse.

### Zonal Resource

Security groups are zonal resources (e.g., `fr-par-1`), not regional. The zone must match the zone of the Instances that will use this security group.

## Features

- Declarative YAML-based firewall configuration
- Inbound and outbound rules with per-rule action, protocol, port range, and IP range
- Configurable default policies (allowlist or denylist model)
- Stateful/stateless mode toggle
- SMTP security control
- Standard OpenMCF tags auto-applied for resource tracking
- Pulumi and Terraform IaC modules included

## Bundled Terraform Resources

| Resource | Purpose |
|----------|---------|
| `scaleway_instance_security_group` | The security group with inline inbound and outbound rules |

This is a single-resource kind (not a composite). Rules are managed inline on the security group resource.

## Dependencies

### Upstream (what this resource needs)

None. Security groups are standalone resources with no dependencies on VPCs, Private Networks, or other resources.

### Downstream (what references this resource)

| Resource Kind | Field | Description |
|---------------|-------|-------------|
| `ScalewayInstance` | `security_group_id` | Assigns this security group to an Instance |

## Stack Outputs

| Output | Description | Referenced By |
|--------|-------------|---------------|
| `security_group_id` | UUID of the created security group | `ScalewayInstance` via `StringValueOrRef` |

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `zone` | `string` | Scaleway zone (e.g., `"fr-par-1"`) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Human-readable description |
| `stateful` | `bool` | `true` | Whether to track connection state |
| `inbound_default_policy` | `string` | `"accept"` | Default inbound policy (`"accept"` or `"drop"`) |
| `outbound_default_policy` | `string` | `"accept"` | Default outbound policy (`"accept"` or `"drop"`) |
| `enable_default_security` | `bool` | `true` | Block SMTP ports 25/465/587 |
| `inbound_rules` | `repeated InboundRule` | `[]` | Inbound firewall rules |
| `outbound_rules` | `repeated OutboundRule` | `[]` | Outbound firewall rules |

### Rule Fields (InboundRule / OutboundRule)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `action` | `string` | *(required)* | `"accept"` or `"drop"` |
| `protocol` | `string` | `"TCP"` | `"TCP"`, `"UDP"`, `"ICMP"`, `"ANY"` |
| `port_range` | `string` | *(all ports)* | Single port `"80"` or range `"8000-9000"` |
| `ip_range` | `string` | *(all IPs)* | CIDR notation `"0.0.0.0/0"` |

## Infra Chart Integration

This resource participates in infra charts as a Layer 0/1 standalone resource. The typical composition pattern is:

```
ScalewayInstanceSecurityGroup (Layer 0 -- standalone)
  └── ScalewayInstance (Layer 2 -- references via security_group_id)
```

In infra chart templates, the Instance resource references the security group via `valueFrom`:

```yaml
spec:
  securityGroupId:
    valueFrom:
      kind: ScalewayInstanceSecurityGroup
      name: "{{ values.env }}-web-sg"
      fieldPath: status.outputs.security_group_id
```

## Security Best Practices

1. **Use "drop" as inbound default for production** -- Start with deny-all and explicitly accept only known traffic patterns.
2. **Restrict SSH access** -- Never allow SSH (port 22) from `0.0.0.0/0`. Restrict to office IPs, VPN exit IPs, or bastion hosts.
3. **Keep SMTP security enabled** -- Unless your account is explicitly authorized for email sending.
4. **Use stateful mode** -- Unless you have a specific need for stateless packet filtering.
5. **Scope IP ranges tightly** -- Use `/32` for single hosts and the narrowest CIDR that covers your source/destination.

## Links

- [Scaleway Security Groups Documentation](https://www.scaleway.com/en/docs/compute/instances/how-to/use-security-groups/)
- [Terraform: scaleway_instance_security_group](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/instance_security_group)
- [Pulumi: instance.SecurityGroup](https://www.pulumi.com/registry/packages/scaleway/api-docs/instance/securitygroup/)
