# AwsWafWebAcl Pulumi Module Overview

## Architecture

The module creates a WAFv2 Web ACL with rules and optional logging:

```
AwsWafWebAclStackInput
    │
    ├── Web ACL (aws_wafv2_web_acl)
    │   ├── Default Action (allow/block)
    │   ├── Rules (via RuleJson)
    │   │   ├── Managed Rule Groups
    │   │   ├── Rate-Based Rules
    │   │   ├── Geo Match Rules
    │   │   ├── IP Set Reference Rules
    │   │   └── Custom Statement Rules
    │   ├── Custom Response Bodies
    │   └── Visibility Config
    │
    └── Logging Configuration (aws_wafv2_web_acl_logging_configuration)
        ├── Destination ARN
        └── Redacted Fields
```

## File Structure

| File | Purpose |
|------|---------|
| `module/main.go` | Entry point, provider setup, orchestration |
| `module/locals.go` | Locals initialization, AWS tags |
| `module/outputs.go` | Output key constants |
| `module/web_acl.go` | Web ACL resource creation |
| `module/rules.go` | Rule JSON construction from typed and custom statements |
| `module/logging.go` | Conditional logging configuration |

## Rule JSON Approach

Rules are constructed as JSON using the AWS WAFv2 API format (PascalCase keys) and passed via the `RuleJson` field. This approach:

1. Handles both typed first-class statements and custom_statement Struct escape hatches uniformly
2. Avoids the complexity of mapping arbitrary Struct contents to deeply nested Pulumi types
3. Provides full WAFv2 API coverage

The JSON is constructed in `rules.go` with builder functions for each statement type.

## Outputs

| Key | Source | Description |
|-----|--------|-------------|
| `web_acl_arn` | Web ACL ARN | Primary output for associations |
| `web_acl_id` | Web ACL ID | Unique identifier |
| `web_acl_name` | Web ACL Name | Resource name |
| `capacity` | Web ACL Capacity | WCUs consumed |
