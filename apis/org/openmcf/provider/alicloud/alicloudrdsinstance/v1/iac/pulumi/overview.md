# AliCloudRdsInstance Pulumi Module Overview

## Module Structure

```
module/
├── main.go            # Provider setup, instance creation, sub-resource orchestration, exports
├── locals.go          # Locals struct, tag initialization, helper functions for optional fields
├── outputs.go         # Output constant names
├── databases.go       # Database creation function with engine-specific charset defaults
└── accounts.go        # Account creation + privilege grant functions
```

## Resource Flow

1. **Provider** -- `alicloud.NewProvider` with the spec's region
2. **RDS Instance** -- `rds.NewInstance` with engine, class, storage, networking, and security config
3. **Databases** -- Loop creates `rds.NewDatabase` for each entry (parented to instance)
4. **Accounts** -- Loop creates `rds.NewRdsAccount` for each entry (parented to instance)
5. **Privileges** -- Nested loop creates `rds.NewAccountPrivilege` for each privilege (parented to account)

## Key Design Decisions

- **Engine-specific defaults**: The `defaultCharacterSet()` function maps engine names to their standard character sets (utf8mb4 for MySQL/MariaDB, UTF8 for PostgreSQL, Chinese_PRC_CI_AS for SQL Server).
- **Parent relationships**: Databases and accounts are parented to the instance; privileges are parented to their account for clean resource hierarchy.
- **Optional field handling**: Helper functions (`instanceChargeType()`, `category()`, `accountType()`, `privilege()`, `optionalString()`, `optionalBool()`, `optionalInt()`) provide defaults when proto optional fields are nil.
- **No public endpoint**: The module exports the instance's intranet connection string, not a public internet endpoint.
