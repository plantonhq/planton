# OpenStackSecurityGroupRule Pulumi Module Architecture

## Module Structure

```
module/
├── main.go                    # Entry point: Resources()
├── locals.go                  # Locals struct + FK resolution
├── outputs.go                 # Output constant names
└── security_group_rule.go     # Resource creation + exports
```

## Data Flow

1. **main.go** receives `OpenStackSecurityGroupRuleStackInput` from the Planton CLI
2. **locals.go** extracts the target resource and resolves both `StringValueOrRef` FKs:
   - `security_group_id` (required) -- resolved via `GetValue()`
   - `remote_group_id` (optional) -- resolved via `GetValue()` if set, empty string otherwise
3. **security_group_rule.go** creates a single `networking.SecGroupRule` resource using the resolved IDs
4. Stack outputs are exported matching `stack_outputs.proto` fields

## Foreign Key Resolution

Both FKs point to `OpenStackSecurityGroup.status.outputs.security_group_id`:

- **`security_group_id`**: Required. Always resolved. The security group that owns this rule.
- **`remote_group_id`**: Optional. Only resolved if present. The security group used as a traffic source/destination filter.

At runtime, the FK resolver middleware replaces `value_from` references with resolved literal values before the Pulumi module executes.
