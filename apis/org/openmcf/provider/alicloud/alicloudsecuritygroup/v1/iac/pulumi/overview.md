# Pulumi Module Overview

## Module Architecture

The AlicloudSecurityGroup Pulumi module is organized into four files under `iac/pulumi/module/`:

| File | Responsibility |
|------|---------------|
| `main.go` | Controller -- creates the provider, security group, and delegates rule creation |
| `locals.go` | Transformations -- tag computation, default resolution for optional fields |
| `outputs.go` | Constants -- defines output key names exported to the stack |
| `rules.go` | Rule creator -- creates individual `ecs.SecurityGroupRule` resources |

The entry point binary at `iac/pulumi/main.go` loads the stack input (manifest YAML -> AlicloudSecurityGroupStackInput) and delegates to `module.Resources()`.

## Control Flow

```
LoadStackInput (manifest YAML -> AlicloudSecurityGroupStackInput)
    |
initializeLocals() -> Locals{Tags, AlicloudSecurityGroup}
    |
alicloud.NewProvider (region-scoped)
    |
ecs.NewSecurityGroup (vpc_id, name, inner_access_policy, tags)
    |
ecs.NewSecurityGroupRule x N (parented to SG, nic_type=intranet)
    |
ctx.Export (security_group_id, security_group_name)
```

## Key Implementation Details

### NIC Type Hardcoding

All security group rules have `nic_type` set to `"intranet"` because `vpc_id` is required. This is not exposed as a user-facing field.

### Rule Naming

Rules are named `{sg_name}-rule-{index}` where index is the zero-based position in the rules list. This follows the same pattern used by AlicloudLogProject for sub-resources.

### Rule Defaults

| Field | Default | Resolved In |
|-------|---------|-------------|
| `port_range` | `"-1/-1"` | `rulePortRange()` in locals.go |
| `priority` | `1` | `rulePriority()` in locals.go |
| `policy` | `"accept"` | `rulePolicy()` in locals.go |
| `inner_access_policy` | `"Accept"` | `innerAccessPolicy()` in locals.go |

### Dependency Chain

Each rule has `pulumi.Parent(sg)` set, establishing an explicit dependency on the security group. This ensures the SG exists before rules are created and that destroying the SG also destroys all rules.
