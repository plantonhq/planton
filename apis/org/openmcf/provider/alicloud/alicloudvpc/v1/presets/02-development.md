# Development VPC

This preset creates a minimal VPC for development and testing environments. It uses a 192.168.x.x CIDR range (different from the 10.x production range) so development and production VPCs can coexist or be peered without address conflicts. Only the three required fields are specified, keeping the configuration as simple as possible.

## When to Use

- Development and testing environments where simplicity is prioritized
- Quick sandbox VPCs for prototyping or validating infrastructure code
- Environments that may be peered with production VPCs (non-overlapping CIDR range)
- Cost-sensitive workloads where organizational tagging overhead is unnecessary

## Key Configuration Choices

- **192.168.x range** (`cidrBlock: 192.168.0.0/16`) -- Uses a different RFC 1918 range than the production preset (10.x). This avoids CIDR conflicts if you need to peer development and production VPCs or connect them via CEN. The /16 mask provides 65,536 IPs, more than sufficient for development workloads.
- **No tags** -- Tags add operational overhead with no benefit in ephemeral or small-scale development environments. Add tags by customizing this preset if your organization requires them in all environments.
- **No description** -- Keeps the manifest minimal. The VPC name conveys the purpose.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`, `ap-southeast-1`) | Your deployment region strategy |
| `<your-vpc-name>` | VPC name (1-128 chars; cannot start with `http://` or `https://`) | Choose a name following your naming convention (e.g., `dev-sandbox-vpc`) |

## Related Presets

- **01-standard-production** -- Use instead for production environments with organizational tagging and a 10.x CIDR range
- **03-dual-stack-ipv6** -- Use instead when your development workloads need IPv6 connectivity
