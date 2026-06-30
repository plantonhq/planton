# Public DNS Zone

This preset creates an Azure DNS Zone for hosting public DNS records for a domain. The zone is created empty -- DNS records are managed separately via `AzureDnsRecord` resources or added inline via the `records` field. This zone-only approach gives you the most flexibility and is the standard starting point for DNS management in Azure.

## When to Use

- Hosting DNS for a public domain (e.g., `example.com`)
- Migrating DNS management from another provider to Azure DNS
- Setting up a DNS zone before creating individual records with `AzureDnsRecord`

## Key Configuration Choices

- **Zone only, no inline records** -- Records are managed separately via `AzureDnsRecord` resources. This keeps the zone definition stable while records can change independently
- **Public zone** -- For private DNS (internal VNet resolution), use `AzurePrivateDnsZone` instead

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-domain.com>` | Your DNS domain name (e.g., `example.com`) | Your domain registrar |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |

## Related Presets

- **AzureDnsRecord 01-a-record** -- Create an A record in this zone
- **AzureDnsRecord 02-cname-record** -- Create a CNAME record in this zone
