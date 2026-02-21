# AliCloudVpnGateway Pulumi Module

This Pulumi module provisions an Alibaba Cloud VPN Gateway with customer gateways and IPsec VPN connections.

## Resources Created

- `alicloud:vpn/gateway:Gateway` -- the VPN Gateway
- `alicloud:vpn/customerGateway:CustomerGateway` -- one per connection, representing the remote device
- `alicloud:vpn/connection:Connection` -- one per connection, with IKE/IPsec tunnel configuration

## Architecture

The module creates the VPN gateway first, then iterates over connections to create a customer gateway and VPN connection pair for each. Customer gateways are named `{connection-name}-cg`. VPN connections are parented to the gateway for clean resource hierarchy.

## Local Development

```bash
cd apis/org/openmcf/provider/alicloud/alicloudvpngateway/v1/iac/pulumi
go build ./...
go vet ./...
```

## Stack Outputs

| Name | Description |
| --- | --- |
| `vpn_gateway_id` | VPN Gateway resource ID |
| `internet_ip` | VPN Gateway's public IP address |
| `ssl_vpn_internet_ip` | SSL VPN IP (when SSL is enabled) |
| `connection_ids` | Map of connection name to VPN connection ID |
