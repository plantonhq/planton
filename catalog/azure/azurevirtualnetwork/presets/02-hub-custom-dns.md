# Hub Network with Custom DNS

This preset shapes the network as a hub for hybrid or hub-and-spoke
topologies: two address-space blocks (shared services grow fast in hubs),
custom DNS servers for on-premises integration, and a raised
connection-tracking timeout for the long-lived idle connections that
gateway and shared-services traffic tends to carry.

Custom DNS servers replace Azure's default resolver for every workload in
the network. The usual hub design points them at a DNS forwarder (or Azure
DNS Private Resolver) that conditionally forwards on-premises names to
corporate DNS and everything else back to Azure -- keeping private-zone
resolution working through the forwarder.

## When to Use

- The hub of a hub-and-spoke topology (spokes arrive via VNet peering)
- Hybrid environments where workloads must resolve on-premises names
- Networks carrying long-lived idle connections that must not be dropped
  between keepalives

## Key Configuration Choices

- **Two CIDR blocks up front** -- blocks can be added in place later, but
  planning both now keeps firewall and route summaries clean
- **Custom DNS is all-or-nothing** -- once set, ALL resolution flows
  through these servers; they must forward to 168.63.129.16 for Azure
  private zones to keep resolving
- **`flowTimeoutInMinutes: 15`** -- raise toward 30 for database-heavy
  intra-network traffic; the default is 4

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<resource-group-name>` | The resource group to create the network in | The resource group's `status.outputs.resource_group_name` |
| `<dns-server-ip-1>` / `<dns-server-ip-2>` | Your DNS forwarder or resolver IPs inside (or reachable from) the network | Your DNS architecture |
| `<cost-center>` | Your org's cost-attribution tag value | Your tagging convention |
