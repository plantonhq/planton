# CNAME Record

This preset creates a DNS CNAME record that aliases a subdomain to another domain name. CNAME records are used when you want a domain name to resolve to the same IP as another hostname, commonly for CDN endpoints, Traffic Manager profiles, and Azure App Service custom domains.

## When to Use

- Aliasing a subdomain to an Azure service endpoint (e.g., `www` to `myapp.azurewebsites.net`)
- Pointing to CDN endpoints (e.g., `cdn.example.com` to `endpoint.azureedge.net`)
- Creating vanity domains that track an underlying service's changing IPs

## Key Configuration Choices

- **Record type** (`type: CNAME`) -- Aliases a name to another domain name (not an IP address)
- **TTL** (`ttlSeconds: 300`) -- 5-minute cache; standard for CNAME records
- **Cannot be used at zone apex** -- DNS standards prohibit CNAME at the root domain (`@`). For apex records, use an A record or Azure Alias record

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-resource-group-name>` | Resource group containing the DNS zone | Azure portal or `AzureResourceGroup` status outputs |
| `<your-domain.com>` | The DNS zone name | Azure portal or `AzureDnsZone` status outputs |
| `<subdomain>` | Record name (e.g., `www`, `cdn`, `api`). Cannot be `@` for CNAME | Your DNS design |
| `<target-hostname>` | Target domain name (e.g., `myapp.azurewebsites.net`) | The service you are aliasing to |

## Related Presets

- **01-a-record** -- Use instead when pointing directly to an IPv4 address
