# AlicloudRedisInstance Pulumi Module Overview

## Module Structure

```
module/
├── main.go            # Provider setup, instance creation, exports
├── locals.go          # Locals struct, tag initialization, helper functions
└── outputs.go         # Output constant names
```

## Resource Flow

1. **Provider** -- `alicloud.NewProvider` with the spec's region
2. **KVStore Instance** -- `kvstore.NewInstance` with instance class, password, networking, and security config

## Key Design Decisions

- **No sub-resources**: Unlike RDS (which bundles databases and accounts), Redis uses a single `kvstore.NewInstance` resource. Authentication is the instance-level password.
- **Optional field handling**: Helper functions (`engineVersion()`, `instanceType()`, `paymentType()`, `vpcAuthMode()`, `optionalString()`, `optionalBool()`, `optionalInt()`, `optionalStringPtr()`) provide defaults when proto optional fields are nil.
- **Config as map**: Redis configuration parameters are passed as `map[string]string` using the modern `config` field, not the deprecated `parameters` field.
- **No public endpoint**: The module exports the instance's intranet connection domain, not a public internet endpoint.
