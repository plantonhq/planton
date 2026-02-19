# Standard Domain Registration

This preset registers a domain in Alibaba Cloud DNS (Alidns) with only the required fields. After deployment, point your domain registrar's NS records to the DNS servers returned in the stack outputs.

## When to Use

- Registering a new domain for DNS hosting on Alibaba Cloud
- Simple DNS setups where organizational grouping, resource groups, and tags are unnecessary
- Development and testing environments
- Quick onboarding of a domain before creating DNS records via AlicloudDnsRecord

## Key Configuration Choices

- **Minimal fields** -- only `region` and `domainName` are specified, keeping the manifest as simple as possible
- **No tags** -- tags add operational overhead with no benefit for simple domain registrations. Add tags by customizing this preset if your organization requires them.
- **No resource group** -- the domain uses the account's default resource group

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`, `ap-southeast-1`). Alidns is global, but the provider requires a region. | Your deployment region strategy |
| `<your-domain-name>` | The domain name to register (e.g., `example.com`, `api.example.com`) | Your domain ownership |

## Post-Deployment Steps

1. Deploy the manifest to register the domain in Alidns
2. Retrieve the `dns_servers` output from `status.outputs`
3. Update your domain registrar's NS records to point to the Alibaba Cloud DNS servers
4. DNS propagation typically takes 24-48 hours for NS record changes
5. Create DNS records using the AlicloudDnsRecord component

## Related Presets

- **02-organizational** -- use instead when you need tags, resource group placement, and domain group assignment for governance
