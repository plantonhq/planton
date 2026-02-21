# Multi-VPC Database Zone

This preset creates a private DNS zone for database endpoint discovery, shared across multiple VPCs including cross-region. Applications in any attached VPC can resolve database hostnames without knowing the underlying IP addresses or instance endpoints.

## When to Use

- Database instances that need to be reachable from multiple application VPCs
- Cross-region deployments where databases in one region serve applications in another
- Centralizing database endpoint management (rename, failover) without updating application configs
- Production environments requiring organizational tags and resource group placement

## Key Configuration Choices

- **Multi-VPC attachment** -- the zone is shared across two or more VPCs, including cross-region via `regionId`
- **Resource group** -- placed in a specific resource group for access control and cost attribution
- **Organizational metadata** -- `org`, `env`, tags for governance and cost tracking
- **A records for database endpoints** -- map friendly names to private IPs. Use CNAME for managed service endpoints.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<primary-region>` | Provider region (e.g., `cn-hangzhou`) | Your primary deployment region |
| `<zone-name>` | Private zone name (e.g., `db.corp`, `data.internal`) | Your naming convention |
| `<description>` | Human-readable remark for the zone | -- |
| `<resource-group-id>` | Resource group ID (e.g., `rg-prod-123`) | Resource Group console |
| `<primary-vpc-id>` | First VPC to attach | VPC console or AlicloudVpc output |
| `<secondary-vpc-id>` | Second VPC to attach (can be in a different region) | VPC console |
| `<secondary-region>` | Region of the secondary VPC (e.g., `cn-shanghai`) | Your DR/multi-region strategy |
| `<database-name>` | Database hostname prefix (e.g., `mysql`, `redis`, `mongo`) | Your database inventory |
| `<database-private-ip>` | Private IP of the database instance | RDS/Redis/MongoDB console |
| `<organization>` | Organization name for metadata | Your org |
| `<environment>` | Environment name (e.g., `production`, `staging`) | Your env strategy |
| `<team-name>` | Team responsible for this zone | Your team structure |

## Post-Deployment Steps

1. Deploy the manifest to create the zone with VPC attachments
2. Applications in any attached VPC can resolve `<database-name>.<zone-name>` immediately
3. During database failover, update the record value and redeploy -- no application config changes needed
4. Add more VPC attachments or records by updating the manifest

## Related Presets

- **01-internal-service-discovery** -- use for simpler single-VPC service discovery scenarios
