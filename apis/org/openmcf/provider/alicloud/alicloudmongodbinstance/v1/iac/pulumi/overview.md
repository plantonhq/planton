# AlicloudMongodbInstance Pulumi Module Overview

## Module Structure

```
module/
├── main.go            # Provider setup, instance creation, exports
├── locals.go          # Locals struct, tag initialization, helper functions
└── outputs.go         # Output constant names
```

## Resource Flow

1. **Provider** -- `alicloud.NewProvider` with the spec's region
2. **MongoDB Instance** -- `mongodb.NewInstance` with engine version, class, storage, networking, and security config

## Key Design Decisions

- **No sub-resources**: Unlike RDS (which bundles databases and accounts), MongoDB uses a single `mongodb.NewInstance` resource. Authentication is the instance-level `accountPassword`.
- **Optional field handling**: Helper functions (`replicationFactor()`, `storageEngine()`, `instanceChargeType()`, `optionalString()`, `optionalBool()`, `optionalInt()`, `optionalStringPtr()`) provide defaults when proto optional fields are nil.
- **Parameters as nested args**: MongoDB engine parameters are passed as `InstanceParameterArray` entries (name/value pairs), matching the provider's nested block schema.
- **Dual encryption paths**: TDE (`tdeStatus` + `encryptionKey`) and cloud disk encryption (`encrypted` + `cloudDiskEncryptionKey`) are mutually exclusive at the provider level -- both are supported but should not be used together.
