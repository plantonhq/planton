# AzureApplicationGateway

Azure Application Gateway is a Layer 7 (HTTP/HTTPS) load balancer and web traffic manager that provides SSL termination, host-based routing, cookie-based session affinity, custom health probes, and optional Web Application Firewall (WAF) protection.

## When to Use

Use AzureApplicationGateway when you need:
- **SSL/TLS termination** -- offload certificate management from backend servers
- **Host-based routing** -- route traffic to different backends based on domain name
- **Web Application Firewall** -- protect applications from common web exploits (SQL injection, XSS)
- **Layer 7 load balancing** -- route traffic based on HTTP attributes rather than TCP/UDP ports
- **Cookie-based session affinity** -- route repeat clients to the same backend

Use [AzureLoadBalancer](../azureloadbalancer/v1/) instead for Layer 4 (TCP/UDP) load balancing where HTTP awareness is not needed.

## Key Configuration

### SKU (Required)

Only v2 SKUs are supported:
- **Standard_v2** -- general L7 load balancing with autoscale and zone redundancy
- **WAF_v2** -- same as Standard_v2 plus Web Application Firewall

V1 SKUs (Standard, WAF) are legacy and not supported.

### Dedicated Subnet (Required)

Application Gateway v2 requires a dedicated subnet with no other resources. A /24 CIDR block is recommended (supports up to 125 instances + 5 Azure-reserved addresses).

### Capacity

Choose one:
- **Fixed capacity** (`capacity`): 1-125 instances (default: 2)
- **Autoscale** (`autoscale`): min/max bounds, scales based on traffic

### Backend Address Pools

Define backend targets by FQDN and/or IP address. At least one pool is required.

### Backend HTTP Settings

Control how the gateway communicates with backends: port, protocol, timeout, cookie affinity, and health probe association. At least one settings object is required.

Key fields:
- `port` / `protocol` -- backend communication (Http or Https)
- `cookie_based_affinity` -- Enabled or Disabled (default: Disabled)
- `request_timeout` -- seconds to wait for backend response (default: 30)
- `probe_name` -- link to a custom health probe
- `host_name` / `pick_host_name_from_backend_address` -- override the Host header

### HTTP Listeners

Define entry points for traffic. Each listener binds to a port and protocol (HTTP/HTTPS). For host-based routing, set `host_name` on listeners.

Frontend port names are auto-derived from listeners -- you only specify the port number.

### Request Routing Rules

Connect listeners to backend pools via HTTP settings. Only Basic rule type is supported (path-based routing is planned for v2). Each rule requires a unique `priority` (1-20000, lower = higher priority).

### SSL Certificates

For HTTPS listeners, certificates are sourced from Azure Key Vault via `key_vault_secret_id`. The gateway must have a user-assigned identity with GET permission on the Key Vault certificate.

### WAF Configuration

Enable WAF for the WAF_v2 SKU. Choose between Detection mode (log only) and Prevention mode (block attacks, default). Uses OWASP 3.2 rule set.

### HTTP/2 Support

Enable HTTP/2 for improved client-to-gateway performance via multiplexed streams and header compression. Backend connections always use HTTP/1.1.

## Outputs

| Output | Description |
|--------|-------------|
| `app_gateway_id` | Azure Resource Manager ID of the Application Gateway |
| `app_gateway_name` | Name of the Application Gateway |

Note: The public frontend IP address is not exported here because it comes from the separate AzurePublicIp resource. DNS records should reference the AzurePublicIp's `ip_address` output directly.

## Infra Chart Usage

AzureApplicationGateway is used in the **enterprise-network-foundation** infra chart as the L7 ingress point. The public IP is managed by a separate AzurePublicIp resource (referenced via `public_ip_id`), and DNS records point to the public IP output.

## Related Resources

- [AzureSubnet](../azuresubnet/v1/) -- dedicated subnet for the gateway
- [AzurePublicIp](../azurepublicip/v1/) -- public IP for the frontend
- [AzureLoadBalancer](../azureloadbalancer/v1/) -- Layer 4 alternative
- [AzureNetworkSecurityGroup](../azurenetworksecuritygroup/v1/) -- subnet security rules
- [AzureUserAssignedIdentity](../azureuserassignedidentity/v1/) -- identity for Key Vault SSL access
- [AzureKeyVault](../azurekeyvault/v1/) -- certificate storage for HTTPS
- [AzureResourceGroup](../azureresourcegroup/v1/) -- resource group for the gateway

## Deployment Notes

- Application Gateway provisioning typically takes **10-20 minutes**
- The dedicated subnet must be /24 or larger for production workloads
- Public IP must use Standard SKU with Static allocation
- When using SSL certificates from Key Vault, create the identity and Key Vault first
- Gateway IP configuration, frontend IP configuration, and frontend port names are auto-derived by IaC modules -- you never need to specify these internal Azure naming details

## References

- Azure Application Gateway Documentation: https://learn.microsoft.com/en-us/azure/application-gateway/
- Application Gateway v2 Overview: https://learn.microsoft.com/en-us/azure/application-gateway/overview-v2
- WAF on Application Gateway: https://learn.microsoft.com/en-us/azure/web-application-firewall/ag/ag-overview
- Terraform Provider: https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/application_gateway
- Pulumi Provider: https://www.pulumi.com/registry/packages/azure-native/api-docs/network/applicationgateway/
