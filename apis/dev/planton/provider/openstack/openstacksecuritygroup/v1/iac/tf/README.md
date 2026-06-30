# OpenStackSecurityGroup Terraform Module

Terraform HCL module for provisioning an OpenStack Neutron security group with optional inline rules.

## Usage

This module is invoked by the Planton CLI with variables auto-populated from the YAML manifest.

```bash
# Preview changes
planton terraform plan --manifest security-group.yaml --module-dir ./iac/tf

# Apply changes
planton terraform apply --manifest security-group.yaml --module-dir ./iac/tf --yes
```

## Resources Created

- `openstack_networking_secgroup_v2.main` -- The security group
- `openstack_networking_secgroup_rule_v2.rules["<key>"]` -- One per inline rule, keyed by `key`

## Variables

| Variable | Description |
|----------|-------------|
| `metadata` | Resource metadata (name, org, env, labels) |
| `spec` | Security group specification (description, rules, tags, etc.) |

## Outputs

| Output | Description |
|--------|-------------|
| `security_group_id` | UUID of the security group |
| `name` | Name of the security group |
| `region` | OpenStack region |

## State Management

Inline rules use `for_each` keyed by the `key` field from each rule. This means:
- Adding a rule only creates that rule
- Removing a rule only destroys that rule
- Reordering rules has no effect on state
- Changing a rule's configuration destroys and recreates only that rule (all rule fields are ForceNew)
