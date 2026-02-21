# Organizational Domain Registration

This preset registers a domain in Alibaba Cloud DNS (Alidns) with resource group placement, remarks, and organizational tags. Suitable for production environments where governance, access control, and cost attribution are important.

## When to Use

- Production domains that need organizational tracking and access control
- Environments where resource groups are used for billing and permission boundaries
- Organizations managing multiple domains that need consistent tagging for cost attribution
- Domains that should carry descriptive metadata visible in the Alidns console

## Key Configuration Choices

- **Resource group** (`resourceGroupId`) -- places the domain in a specific resource group for access control and cost tracking. Resource group cannot be changed after creation.
- **Remark** (`remark`) -- a human-readable description visible in the Alidns console, useful for operations teams managing many domains.
- **Tags** (`team`, `costCenter`) -- organizational metadata for cost attribution and operational ownership. Replace placeholders with your organization's values.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`, `ap-southeast-1`) | Your deployment region strategy |
| `<your-domain-name>` | The domain name to register (e.g., `platform.example.com`) | Your domain ownership |
| `<domain-description>` | A brief description of the domain's purpose (e.g., "Primary platform domain for production services") | Your team's documentation |
| `<your-resource-group-id>` | Alibaba Cloud resource group ID (e.g., `rg-prod-123`) | Alibaba Cloud console > Resource Management |
| `<your-team>` | Team or business unit that owns this domain | Your organizational structure |
| `<your-cost-center>` | Cost center code for billing attribution | Your finance or cloud operations team |

## Post-Deployment Steps

1. Deploy the manifest to register the domain in Alidns
2. Retrieve the `dns_servers` output from `status.outputs`
3. Update your domain registrar's NS records to point to the Alibaba Cloud DNS servers
4. Create DNS records using the AlicloudDnsRecord component

## Related Presets

- **01-standard** -- use instead for simple domain registrations without organizational overhead
