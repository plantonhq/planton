# AzurePrivateDnsZone Pulumi Module Overview

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

1. **`main.go`** -- Pulumi entrypoint loads `AzurePrivateDnsZoneStackInput` from config
2. **`module.Resources()`** -- Creates Azure provider, initializes locals, provisions resources
3. **`initializeLocals()`** -- Extracts resource group name from StringValueOrRef, builds tag map
4. **Resource creation** -- Creates zone, then VNet link (with explicit dependency)
5. **Exports** -- Zone ID and zone name are exported as stack outputs

## Key Patterns

- **StringValueOrRef extraction**: `target.Spec.ResourceGroup.GetValue()` returns the resolved literal value
- **No region**: Private DNS zones are global -- no `Location` parameter in zone creation
- **VNet link dependency**: Explicit `pulumi.DependsOn` ensures link waits for zone creation
- **Tag convention**: Standard Azure tags (resource, resource_name, resource_kind, resource_id, organization, environment)
