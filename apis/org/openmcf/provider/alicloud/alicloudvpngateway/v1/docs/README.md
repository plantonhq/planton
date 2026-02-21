# AlicloudVpnGateway Research Documentation

## Provider Resource Analysis

### alicloud_vpn_gateway (Terraform) / vpn.Gateway (Pulumi)

The VPN Gateway is the core resource. Key findings from provider analysis:

- **Network type**: Supports "public" (internet-facing, default) and "private" (CEN-only). This component targets the public use case.
- **VPN type**: "Normal" (standard) and "NationalStandard" (Chinese national encryption). NationalStandard is a niche compliance requirement, so we default to Normal.
- **Bandwidth tiers**: Public gateways accept 5, 10, 20, 50, 100, 200, 500, 1000 Mbps. Private gateways accept 200 or 1000 Mbps only.
- **SSL VPN**: The `enable_ssl` flag enables a dedicated SSL VPN endpoint. `ssl_connections` defines the max client count. SSL VPN server and client cert resources are managed separately.
- **ForceNew fields**: `bandwidth`, `vpc_id`, `vswitch_id`, `network_type`, `payment_type`, `ssl_connections`, `vpn_type` are all immutable after creation.
- **Deprecated fields**: `name` (use `vpn_gateway_name`), `instance_charge_type` (use `payment_type`)
- **Computed outputs**: `internet_ip` (gateway's public IP), `ssl_vpn_internet_ip` (SSL VPN IP), `status`, `business_status`

### alicloud_vpn_customer_gateway (Terraform) / vpn.CustomerGateway (Pulumi)

Represents the remote VPN device:

- `ip_address` -- public IP of the remote device (Required, ForceNew)
- `asn` -- BGP ASN (Optional, ForceNew)
- `customer_gateway_name` -- display name
- Deprecated: `name` (use `customer_gateway_name`)

### alicloud_vpn_connection (Terraform) / vpn.Connection (Pulumi)

Establishes the IPsec tunnel:

- `vpn_gateway_id` + `customer_gateway_id` -- links the two endpoints (both ForceNew)
- `local_subnet` / `remote_subnet` -- Set of CIDRs (1-10 each), validated as CIDR addresses
- `ike_config` / `ipsec_config` -- Phase 1 and Phase 2 negotiation parameters (all Optional+Computed)
- `health_check_config` -- tunnel health monitoring
- `bgp_config` -- BGP routing (advanced, not included in this component)
- `tunnel_options_specification` -- dual-tunnel mode (advanced, not included)
- `enable_dpd` -- Dead Peer Detection
- `enable_nat_traversal` -- NAT-T for devices behind NAT

## Design Rationale

### Composite Bundling (DD07)

A VPN Gateway without connections is useless. We bundle:
1. VPN Gateway (1)
2. Customer Gateways (N, one per connection)
3. VPN Connections (N, one per connection)

This matches the NatGateway pattern (gateway + EIP association + SNAT entries).

### Connection Naming Strategy

Each connection's `name` field serves triple duty:
- `customer_gateway_name` = `{name}-cg`
- `vpn_connection_name` = `{name}`
- Map key in `connection_ids` output = `{name}`

### Multi-CIDR Subnets

The T02 spec design had `string local_subnet` (singular), but the actual provider supports `Set` of CIDRs (1-10). A single VPN connection commonly bridges multiple CIDRs. The proto uses `repeated string local_subnets`.

### Omitted Advanced Features

- **BGP routing**: Advanced networking feature for dynamic route exchange. Most site-to-site VPNs use static routing with `local_subnets`/`remote_subnets`.
- **Dual-tunnel mode** (`tunnel_options_specification`): HA feature requiring two tunnels per connection. Not in the 80/20 scope.
- **Private network type**: Used exclusively with CEN Transit Router. Niche use case.
- **NationalStandard VPN type**: China-specific national encryption compliance. Niche.

## Alibaba Cloud VPN Gateway Pricing

- **PayAsYouGo**: Hourly billing based on bandwidth and connection count
- **Subscription**: Monthly/yearly reserved pricing with discount
- SSL VPN connections are billed separately based on `ssl_connections` count
- Data transfer charges apply for all traffic through the VPN tunnels
