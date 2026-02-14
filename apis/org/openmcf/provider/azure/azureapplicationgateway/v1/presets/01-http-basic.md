# Basic HTTP Application Gateway

This preset creates an Azure Application Gateway v2 with Standard_v2 SKU, a single HTTP listener on port 80, one backend pool with FQDN-based targets, and a custom health probe. This is the simplest working Application Gateway configuration for Layer 7 HTTP load balancing without SSL termination or WAF.

## When to Use

- Development or staging environments where HTTPS is not yet configured
- Internal-facing HTTP services behind a public entry point (SSL handled upstream or not required)
- Quick proof-of-concept for Application Gateway routing before adding SSL certificates
- Layer 7 load balancing with health probes when Azure Load Balancer (Layer 4) is insufficient

## Key Configuration Choices

- **Standard_v2 SKU** (`sku: Standard_v2`) -- General-purpose L7 load balancing with autoscale support and zone redundancy. No WAF
- **Fixed capacity** (`capacity: 2`) -- Two instances for basic availability. Switch to `autoscale` for production traffic variability
- **HTTP listener on port 80** (`httpListeners: Http on 80`) -- Plain HTTP entry point. No SSL certificates needed
- **FQDN backend pool** (`backendAddressPools: fqdns`) -- Backends identified by DNS name; App Gateway resolves and routes accordingly
- **Custom health probe** (`probes: Http /health`) -- Checks backend health every 30 seconds on `/health` path; marks unhealthy after 3 consecutive failures
- **No WAF** -- Web Application Firewall is disabled. Use the `02-https-waf` preset for WAF protection

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (must match subnet and public IP) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-appgw-name>` | Name for the Application Gateway (unique within resource group) | Your naming convention |
| `<dedicated-subnet-resource-id>` | Full ARM resource ID of a dedicated subnet (no other resources; /24 recommended) | Azure portal or `AzureSubnet` status outputs |
| `<public-ip-resource-id>` | Full ARM resource ID of a Standard SKU static public IP | Azure portal or `AzurePublicIp` status outputs |
| `<your-backend-fqdn>` | FQDN of the backend server (e.g., `api.contoso.com`) | Your application DNS configuration |

## Related Presets

- **02-https-waf** -- Use instead for production with HTTPS termination and Web Application Firewall protection
