# AlicloudVpnGateway Pulumi Module Overview

## Module Structure

```
module/
├── main.go            # Provider setup, VPN gateway creation, connection orchestration
├── locals.go          # Locals struct, tag initialization, helper functions
├── outputs.go         # Output constant names
└── connections.go     # Customer gateway + VPN connection creation per entry
```

## Resource Flow

1. **Provider** -- `alicloud.NewProvider` with the spec's region
2. **VPN Gateway** -- `vpn.NewGateway` with VPC, VSwitch, bandwidth, billing, and SSL configuration
3. **Per connection**:
   a. **Customer Gateway** -- `vpn.NewCustomerGateway` from the remote device's IP + optional ASN
   b. **VPN Connection** -- `vpn.NewConnection` linking the VPN gateway to the customer gateway, with IKE/IPsec config
4. **Outputs** -- Export gateway ID, internet IP, SSL VPN IP, and connection ID map

## Key Design Decisions

- **Bundled creation**: Customer gateways are created inline per connection rather than as separate components. A customer gateway without a connection is useless.
- **Parent relationships**: VPN connections are parented to the gateway for clean Pulumi state management.
- **Builder pattern for nested configs**: `buildIkeConfig()`, `buildIpsecConfig()`, and `buildHealthCheckConfig()` convert proto messages to Pulumi input types, only setting fields that are explicitly configured.
- **Optional field handling**: Proto optional fields are checked for nil before being passed to Pulumi args, letting the provider apply its own defaults for unset fields.
