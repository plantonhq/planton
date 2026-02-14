# Standard NAT Gateway

This preset creates an Azure NAT Gateway attached to a subnet, providing reliable SNAT for outbound internet connectivity. The NAT Gateway automatically provisions a public IP and associates it with the specified subnet. This is the standard configuration for subnets that need outbound internet access with a predictable, static source IP.

## When to Use

- Subnets running AKS nodes, VMs, or containers that need outbound internet access
- Replacing Azure default outbound access (which Microsoft is deprecating) with a dedicated NAT Gateway
- Workloads that require a known source IP for firewall allowlisting or compliance

## Key Configuration Choices

- **Idle timeout** (`idleTimeoutMinutes: 10`) -- 10 minutes balances connection reuse for long-lived connections (API calls, database connections) with timely resource cleanup. Azure default is 4 minutes; maximum is 120
- **No public IP prefix** -- Uses a single public IP (auto-provisioned). Add `publicIpPrefixLength` (28-31) if you need multiple outbound IPs for high-throughput scenarios

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (must match the subnet's VNet region) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<subnet-resource-id>` | Full ARM resource ID of the target subnet | Azure portal or `AzureSubnet` / `AzureVpc` status outputs |
