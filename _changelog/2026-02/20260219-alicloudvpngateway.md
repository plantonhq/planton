# AlicloudVpnGateway

**Date**: 2026-02-19
**Type**: New Resource
**Enum**: 3027
**ID Prefix**: acvpn

## Summary

Added `AlicloudVpnGateway` -- an Alibaba Cloud VPN Gateway component that bundles VPN gateway creation with customer gateways and IPsec VPN connections into a single deployable unit.

## What's Included

- **Proto API** -- spec.proto with 5 messages (AlicloudVpnGatewaySpec, AlicloudVpnConnection, AlicloudIkeConfig, AlicloudIpsecConfig, AlicloudVpnHealthCheckConfig), full buf.validate + CEL validation
- **Pulumi module** -- Go implementation with provider setup, gateway creation, and per-connection customer gateway + VPN connection bundling
- **Terraform module** -- HCL implementation with `for_each` for connections, dynamic blocks for IKE/IPsec/health-check config
- **33 validation tests** -- comprehensive spec_test.go covering valid and invalid inputs
- **3 presets** -- basic site-to-site, production multi-site, SSL-enabled
- **Full documentation** -- catalog page, examples (4 YAML), research docs, module READMEs

## Provider Resources

- `alicloud_vpn_gateway` / `vpn.Gateway`
- `alicloud_vpn_customer_gateway` / `vpn.CustomerGateway`
- `alicloud_vpn_connection` / `vpn.Connection`

## Key Design Decisions

- Composite bundling (DD07): Gateway + N customer gateways + N connections as a single unit
- `local_subnets`/`remote_subnets` as repeated strings (1-10 CIDRs each) instead of T02's singular design
- Added DPD, NAT traversal, PFS, and health check config beyond T02 baseline
- Omitted BGP, dual-tunnel mode, network_type, vpn_type (not in 80/20 scope)
