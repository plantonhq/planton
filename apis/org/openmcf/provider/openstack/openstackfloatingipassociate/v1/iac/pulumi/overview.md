# OpenStackFloatingIpAssociate Pulumi Module Overview

## Module Structure

```
module/
├── main.go                  # Entry point: Resources() orchestrates association
├── locals.go                # Locals: resolves FKs (floating_ip, port_id)
├── outputs.go               # Output constants matching stack_outputs.proto
└── floating_ip_associate.go # Association resource creation
```

## Resource Flow

1. `main.go`: Load stack input, initialize locals, create provider
2. `locals.go`: Resolve `floating_ip` FK (targets address) and `port_id` FK
3. `floating_ip_associate.go`: Build FloatingIpAssociateArgs, create resource
4. Export outputs: id, floating_ip, port_id, fixed_ip, region

## FK Resolution Pattern

Both FKs are resolved identically via `GetValue()`:
```go
locals.FloatingIp = stackInput.Target.Spec.FloatingIp.GetValue()
locals.PortId = stackInput.Target.Spec.PortId.GetValue()
```

Note: `FloatingIp` resolves to an IP address (e.g., "203.0.113.42") rather than a UUID, which is unique among OpenMCF FK targets.
