# OpenStackSecurityGroupRule Terraform Module

This Terraform module provisions a standalone OpenStack Neutron security group rule.

## Resources Created

- `openstack_networking_secgroup_rule_v2` -- A single security group rule attached to the specified security group

## Variables

| Variable | Description |
|----------|-------------|
| `metadata` | Resource metadata (name, org, env, labels) |
| `spec` | Rule specification (security_group_id, direction, ethertype, protocol, ports, remote source) |

## Outputs

| Output | Description |
|--------|-------------|
| `rule_id` | UUID of the created security group rule |
| `security_group_id` | UUID of the parent security group |
| `direction` | Direction of the rule (ingress/egress) |
| `protocol` | IP protocol of the rule |
| `port_range_min` | Lower bound of port range |
| `port_range_max` | Upper bound of port range |
| `region` | OpenStack region |

## Usage

```hcl
module "security_group_rule" {
  source = "."

  metadata = {
    name = "allow-ssh"
  }

  spec = {
    security_group_id = { value = "sg-uuid-here" }
    direction         = "ingress"
    ethertype         = "IPv4"
    protocol          = "tcp"
    port_range_min    = 22
    port_range_max    = 22
    remote_ip_prefix  = "0.0.0.0/0"
  }
}
```
