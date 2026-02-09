# OpenStackSecurityGroup Pulumi Module Architecture

## Module Structure

```
module/
├── main.go              # Entry point: Resources() orchestrates the module
├── locals.go            # Locals: extracts references from stack input
├── outputs.go           # Output key constants
└── security_group.go    # Resource creation: SG + inline rules
```

## Key Components

### Controller (main.go)

The `Resources()` function orchestrates:
1. Initialize locals from stack input
2. Create Pulumi OpenStack provider from credentials
3. Invoke `securityGroup()` to create all resources

### Locals (locals.go)

Extracts commonly referenced data from the stack input:
- `OpenStackSecurityGroup` -- the full target resource
- `OpenStackProviderConfig` -- provider authentication config

### Resource: Security Group + Inline Rules (security_group.go)

This is the first OpenMCF Pulumi module that creates N+1 resources:

```
securityGroup()
├── networking.NewSecGroup()          # 1 security group
└── for each spec.Rules:
    └── networking.NewSecGroupRule()   # N inline rules
```

Each inline rule:
- Is named `{sg-name}-rule-{key}` for unique identification
- Depends on the security group via `pulumi.DependsOn`
- Receives the same region override as the parent SG

### Outputs (outputs.go)

Constants for stack output keys, matching `stack_outputs.proto` fields:
- `security_group_id` -- primary FK for downstream components
- `name` -- security group name
- `region` -- OpenStack region

## Data Flow

```
StackInput
  └─> Locals (extract target + provider config)
        └─> securityGroup()
              ├─> NewSecGroup (name, description, delete_default_rules, stateful, tags, region)
              ├─> NewSecGroupRule * N (direction, ethertype, protocol, ports, remote source)
              └─> ctx.Export (security_group_id, name, region)
```

## Design Decisions

1. **Inline rules as separate resources**: Each inline rule is a full `SecGroupRule` Pulumi resource, not a sub-resource. This matches the Terraform provider's model and provides granular state management.

2. **Key-based naming**: Rule resource names use the user-provided `key` field (`{sg-name}-rule-{key}`), ensuring stable Pulumi URNs across updates.

3. **Explicit DependsOn**: Each rule explicitly depends on the security group, even though the `SecurityGroupId` input creates an implicit dependency. This makes the dependency graph explicit and self-documenting.

4. **Region propagation**: The region override from the SG spec is propagated to all inline rules, ensuring they're created in the same region.
