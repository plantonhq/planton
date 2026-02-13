# AzurePostgresqlFlexibleServer Pulumi Module Overview

## Module Structure

```
module/
├── main.go      # Resource creation: server, databases, firewall rules
├── locals.go    # Local variable initialization, tags, resource group resolution
└── outputs.go   # Stack output constant definitions
```

## Resource Creation Flow

1. **Initialize locals** -- Resolve StringValueOrRef fields, build Azure tags
2. **Create Azure provider** -- Using credentials from ProviderConfig
3. **Build server args** -- Assemble FlexibleServerArgs with conditional networking
4. **Create server** -- `postgresql.NewFlexibleServer`
5. **Create databases** -- Loop over `spec.databases`, create each with `DependsOn` server
6. **Create firewall rules** -- Loop over `spec.firewall_rules`, create each with `DependsOn` server
7. **Export outputs** -- server_id, server_name, fqdn, administrator_login, database_ids

## Conditional Logic

### Network Mode
```
if delegated_subnet_id != nil && value != "" {
    → private access (PublicNetworkAccessEnabled = false)
} else {
    → public access (PublicNetworkAccessEnabled = true)
}
```

### Private DNS Zone
```
if private_dns_zone_id != nil && value != "" {
    → set PrivateDnsZoneId on server
}
```

### High Availability
```
if spec.HighAvailability != nil {
    → configure HA block with mode and optional standby zone
}
```

## Output Map Pattern

Database IDs are exported as `pulumi.StringMap` following the KeyVault `secret_id_map` pattern:
```go
dbIdMapOutput := pulumi.StringMap{}
for name, id := range databaseIdMap {
    dbIdMapOutput[name] = id
}
ctx.Export(OpDatabaseIds, dbIdMapOutput)
```
