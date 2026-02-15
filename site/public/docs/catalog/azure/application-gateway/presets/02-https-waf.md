---
title: "HTTPS Application Gateway with WAF"
description: "This preset creates an Azure Application Gateway v2 with WAF_v2 SKU, HTTPS termination using a Key Vault certificate, Web Application Firewall in Prevention mode, and an HTTP listener for..."
type: "preset"
rank: "02"
presetSlug: "02-https-waf"
componentSlug: "application-gateway"
componentTitle: "Application Gateway"
provider: "azure"
icon: "package"
order: 2
---

# HTTPS Application Gateway with WAF

This preset creates an Azure Application Gateway v2 with WAF_v2 SKU, HTTPS termination using a Key Vault certificate, Web Application Firewall in Prevention mode, and an HTTP listener for redirect-to-HTTPS. This is the production-recommended configuration for internet-facing web applications that need Layer 7 load balancing with security protections.

## When to Use

- Production web applications that need HTTPS termination with SSL certificates from Key Vault
- Internet-facing services requiring Web Application Firewall protection against OWASP top-10 threats
- Applications that need both HTTPS and HTTP-to-HTTPS redirect on the same gateway
- Compliance-driven environments that mandate WAF for web traffic inspection

## Key Configuration Choices

- **WAF_v2 SKU** (`sku: WAF_v2`) -- Includes Web Application Firewall capability on top of Standard_v2 features
- **WAF Prevention mode** (`wafMode: Prevention`) -- Actively blocks detected attacks. Use `Detection` mode for monitoring-only during initial rollout
- **HTTPS listener with Key Vault cert** (`sslCertificates: keyVaultSecretId`) -- SSL certificate sourced from Azure Key Vault. Requires a user-assigned identity with GET permission on the Key Vault
- **HTTP redirect listener** (`httpListeners: Http on 80`) -- Catches HTTP traffic for redirect-to-HTTPS at the application or rule level
- **HTTPS backend settings** (`backendHttpSettings: Https on 443`) -- End-to-end encryption with backend hostname picked from the backend address
- **HTTP/2 enabled** (`enableHttp2: true`) -- Improved client-to-gateway performance via multiplexed streams and header compression
- **User-assigned identity** (`identityIds`) -- Required for Key Vault certificate access

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (must match subnet and public IP) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-appgw-name>` | Name for the Application Gateway (unique within resource group) | Your naming convention |
| `<dedicated-subnet-resource-id>` | Full ARM resource ID of a dedicated subnet (no other resources; /24 recommended) | Azure portal or `AzureSubnet` status outputs |
| `<public-ip-resource-id>` | Full ARM resource ID of a Standard SKU static public IP | Azure portal or `AzurePublicIp` status outputs |
| `<user-assigned-identity-resource-id>` | Full ARM resource ID of a user-assigned managed identity with Key Vault access | Azure portal or `AzureUserAssignedIdentity` status outputs |
| `<key-vault-certificate-secret-uri>` | Key Vault secret URI for the SSL certificate (e.g., `https://myvault.vault.azure.net/secrets/my-cert`) | Azure Key Vault portal or `AzureKeyVault` status outputs |
| `<your-backend-fqdn>` | FQDN of the backend server (e.g., `api.contoso.com`) | Your application DNS configuration |

## Related Presets

- **01-http-basic** -- Use instead for development/staging without SSL or WAF requirements
