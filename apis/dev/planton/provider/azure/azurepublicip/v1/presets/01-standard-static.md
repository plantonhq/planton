# Standard Static Public IP

This preset creates a zone-redundant Azure Public IP with Standard SKU and static allocation. Standard SKU with static allocation is the only supported configuration (Azure retired Basic SKU in September 2025). Zone redundancy across all three availability zones provides the highest availability for production load balancers, application gateways, and NAT gateways.

## When to Use

- Attaching to Azure Load Balancers, Application Gateways, or NAT Gateways
- Any resource that needs a stable, internet-routable IPv4 address
- Production workloads requiring zone-redundant availability

## Key Configuration Choices

- **Zone-redundant** (`zones: ["1", "2", "3"]`) -- Survives the failure of any single availability zone. Use fewer zones only if the region does not support all three
- **Idle timeout** (`idleTimeoutInMinutes: 4`) -- Azure default. Increase for long-lived connections (WebSocket, gRPC streaming); maximum is 30 minutes
- **Standard SKU and static allocation** -- Hardcoded in the IaC module (not exposed in spec). This is the only production-grade option

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (must match the resource this IP attaches to) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-public-ip-name>` | Name for the public IP resource | Your naming convention |
