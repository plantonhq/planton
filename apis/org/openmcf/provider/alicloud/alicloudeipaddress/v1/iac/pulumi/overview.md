# Pulumi Module Overview

## Module Architecture

The AliCloudEipAddress Pulumi module is organized into three files under `iac/pulumi/module/`:

| File | Responsibility |
|------|---------------|
| `main.go` | Controller -- creates the provider and EIP resource, exports outputs |
| `locals.go` | Transformations -- tag computation, default resolution for optional fields |
| `outputs.go` | Constants -- defines output key names exported to the stack |

The entry point binary at `iac/pulumi/main.go` loads the stack input (manifest YAML -> AliCloudEipAddressStackInput) and delegates to `module.Resources()`.

## Control Flow

```
LoadStackInput (manifest YAML -> AliCloudEipAddressStackInput)
    |
initializeLocals() -> Locals{Tags, AliCloudEipAddress}
    |
alicloud.NewProvider (region-scoped)
    |
ecs.NewEipAddress (address_name, bandwidth, isp, internet_charge_type, tags)
    |
ctx.Export (eip_id, ip_address)
```

## Key Implementation Details

### Resource Naming

The Pulumi resource name is derived from `spec.AddressName` when set, falling back to `metadata.Name` when the address name is empty. This ensures the Pulumi resource URN is stable across updates.

### Bandwidth Type Conversion

The Alibaba Cloud provider represents bandwidth as a string, but the proto spec uses `int32` for better user experience. The conversion happens in `main.go` via `fmt.Sprintf("%d", bandwidth(spec))`.

### Default Resolution

| Field | Default | Resolved In |
|-------|---------|-------------|
| `bandwidth` | `5` | `bandwidth()` in locals.go |
| `internet_charge_type` | `"PayByTraffic"` | `internetChargeType()` in locals.go |
| `isp` | `"BGP"` | `isp()` in locals.go |

Each helper checks whether the optional proto field pointer is non-nil. If set, it returns the user's value; otherwise it returns the hardcoded default matching the `(org.openmcf.shared.options.default)` annotation in `spec.proto`.

### Tag Merging

The `initializeLocals()` function builds a tag map in this order:

1. Base tags: `resource=true`, `resource_name`, `resource_kind`
2. Conditional tags: `resource_id` (if metadata.id is set), `organization` (if metadata.org is set), `environment` (if metadata.env is set)
3. User tags: merged last from `spec.Tags`, so user-provided tags override base tags on key collision

### Optional String Handling

The `optionalString()` helper returns `nil` (Pulumi null) for empty strings, preventing the provider from sending empty-string values to the Alibaba Cloud API. Without this, fields like `address_name` and `description` would be set to `""` instead of omitted.

### Immutable Fields

`internet_charge_type` and `isp` are ForceNew in the Alibaba Cloud provider. Changing these values after creation will trigger resource replacement (destroy + recreate), which means a new IP address is allocated.

## Design Decisions

**Single resource, no association**: The module creates only `ecs.NewEipAddress`. EIP-to-resource association (NAT gateway, ALB, VPN) is handled by downstream components. This keeps the module focused and avoids circular dependencies.

**Region-scoped provider**: A dedicated `alicloud.NewProvider` is created with the spec's region. This isolates the EIP from any ambient provider configuration and ensures the EIP is allocated in the intended region.
