# OpenStackNetworkPort Pulumi Module Overview

## Module Structure

```
module/
├── main.go       # Entry point: Resources() orchestrates port creation
├── locals.go     # Locals: resolves FKs (network_id, security_group_ids)
├── outputs.go    # Output constants matching stack_outputs.proto
└── port.go       # Port resource creation with fixed IPs and SGs
```

## Resource Flow

1. `main.go`: Load stack input, initialize locals, create provider
2. `locals.go`: Resolve `network_id` FK and `security_group_ids` repeated FK
3. `port.go`: Build PortArgs with fixed IPs (nested FK resolution), SGs, and optional fields
4. Export outputs: port_id, mac_address, all_fixed_ips, all_security_group_ids, region

## FK Resolution Patterns

### Required Singular FK (network_id)
```go
locals.NetworkId = stackInput.Target.Spec.NetworkId.GetValue()
```

### Repeated FK (security_group_ids) -- NEW PATTERN
```go
for _, sgRef := range stackInput.Target.Spec.SecurityGroupIds {
    locals.SecurityGroupIds = append(locals.SecurityGroupIds, sgRef.GetValue())
}
```

### FK Inside Nested Message (FixedIp.subnet_id) -- NEW PATTERN
```go
if fip.SubnetId != nil {
    entry.SubnetId = pulumi.StringPtr(fip.SubnetId.GetValue())
}
```
