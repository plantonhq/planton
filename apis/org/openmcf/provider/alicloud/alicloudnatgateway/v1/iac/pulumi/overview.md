# AlicloudNatGateway Pulumi Module Overview

## Module Structure

```
module/
├── main.go            # Provider setup, NAT gateway creation, EIP association, SNAT orchestration
├── locals.go          # Locals struct, tag initialization, helper functions for optional fields
├── outputs.go         # Output constant names
└── snat_entries.go    # Individual SNAT entry creation function
```

## Resource Flow

1. **Provider** -- `alicloud.NewProvider` with the spec's region
2. **EIP Lookup** -- `ecs.GetEipAddresses` resolves the EIP allocation ID to its public IP address
3. **NAT Gateway** -- `vpc.NewNatGateway` with VPC, VSwitch, billing, and tagging configuration
4. **EIP Association** -- `ecs.NewEipAssociation` binds the EIP to the NAT Gateway (parented)
5. **SNAT Entries** -- Loop creates `vpc.NewSnatEntry` for each entry (parented to NAT Gateway)

## Key Design Decisions

- **Data source for EIP IP**: The module looks up the EIP's IP address via `ecs.GetEipAddresses` rather than requiring the user to supply it as a separate field. This keeps the user-facing spec simpler.
- **Parent relationships**: EIP association and SNAT entries are parented to the NAT Gateway for clean Pulumi state management.
- **Optional field handling**: Helper functions (`natType()`, `paymentType()`, `internetChargeType()`, `optionalString()`, `optionalBool()`) provide defaults when proto optional fields are nil.
