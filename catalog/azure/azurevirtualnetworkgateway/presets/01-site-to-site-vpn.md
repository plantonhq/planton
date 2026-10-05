# Site-to-Site VPN Gateway

This preset creates a route-based VpnGw1AZ VPN gateway with BGP -- the standard Azure-side anchor for datacenter-to-Azure connectivity. Pair it with an `AzureLocalNetworkGateway` per site and an `AzureVirtualNetworkGatewayConnection` per tunnel; the gateway itself is long-lived hub infrastructure that new sites never touch.

## When to Use

- Connecting one or more datacenters/branch offices to a VNet over IPsec
- The hub side of a hub-spoke network where spokes reach on-premises through gateway transit
- Any deployment that will grow beyond one site (BGP is on from day one, so adding BGP-routed sites later costs nothing)

## Key Configuration Choices

- **`sku: VPN_GW_1_AZ`** -- 650 Mbps aggregate, 30 tunnels: the production entry point. Azure retired new non-AZ VpnGw creates, and the AZ SKUs deploy in every region (zone-redundant where the region has availability zones, regional elsewhere). Resizing within the VpnGw family is in-place; switching to BASIC is a rebuild
- **GatewaySubnet reference** -- resolves to the subnet's `subnet_id` output; the subnet's ARM name must be EXACTLY `GatewaySubnet` (/27+, no NSG, no other workloads)
- **Public IP reference** -- a Standard static address the gateway binds exclusively (never share it with a NAT gateway or load balancer)
- **`bgpEnabled` + ASN 65515** -- Azure's default ASN; dynamic route exchange for every tunnel that opts in
- **Creation takes 25-45 minutes** and the gateway bills hourly from provisioning, tunnels or not; the verified figure lives in the kind's generated estimate at `catalog/_pricing/estimates/azurevirtualnetworkgateway.yaml`

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (must match the VNet) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-gateway-subnet-resource-name>` | Planton metadata name of the `AzureSubnet` named "GatewaySubnet" | Your subnet resource |
| `<your-public-ip-resource-name>` | Planton metadata name of the `AzurePublicIp` (Standard, static, zones `["1","2","3"]` -- AZ gateway SKUs require a zoned address) | Your public IP resource |
