# AzurePrivateEndpoint Pulumi Module Overview

## Module Structure

```
iac/pulumi/
├── main.go              # Entrypoint: loads stack input, calls module.Resources
├── Pulumi.yaml          # Pulumi project config (name, runtime)
├── Makefile             # Build, test, clean commands
├── debug.sh             # Delve debugger launcher
├── README.md            # Module documentation
├── overview.md          # This file
└── module/
    ├── main.go          # Core resource creation logic
    ├── locals.go        # Locals initialization (tags, resource group extraction)
    └── outputs.go       # Output constant definitions
```

## Execution Flow

1. **`main.go`** -- Pulumi entrypoint loads `AzurePrivateEndpointStackInput` from config
2. **`module.Resources()`** -- Creates Azure provider, initializes locals, provisions resources
3. **`initializeLocals()`** -- Extracts resource group name from StringValueOrRef, builds tag map
4. **Resource creation** -- Creates private endpoint with PrivateServiceConnection, conditionally adds DNS zone group if `private_dns_zone_id` is provided
5. **Exports** -- Private endpoint ID, private IP address, and network interface ID are exported as stack outputs

## Key Patterns

- **StringValueOrRef extraction**: `target.Spec.ResourceGroup.GetValue()` returns the resolved literal value
- **Conditional DNS zone group**: Only created when `spec.PrivateDnsZoneId != nil`
- **Auto-derived names**: Connection name and DNS zone group name are auto-derived from `metadata.name`
- **PrivateServiceConnection**: Hardcoded `IsManualConnection: false` for auto-approved connections
- **Output extraction**: Uses `ApplyT` to extract private IP and network interface ID from endpoint outputs
- **Tag convention**: Standard Azure tags (resource, resource_name, resource_kind, resource_id, organization, environment)
