# Multi-VPC Same Region

Connects two VPCs in the same Alibaba Cloud region via a Cloud Enterprise Network (CEN) instance. This is the most common CEN pattern -- isolating workloads (production, staging, shared-services) into separate VPCs while maintaining private, low-latency connectivity between them without VPC peering limitations.

## When to Use

- You have separate VPCs for different environments or teams in the same region and need private inter-VPC routing
- You want centralized network management through a hub instead of point-to-point VPC peering
- You plan to add more VPCs or cross-region connectivity later (CEN scales incrementally)

## Key Configuration Choices

- **Same-region attachments** -- both VPCs reside in the same region, so cross-region bandwidth charges do not apply; traffic flows over Alibaba Cloud's backbone at no additional cost
- **Default CIDR protection** (`protectionLevel` omitted) -- strict mode rejects overlapping CIDR blocks between attached VPCs, preventing routing ambiguity. Use `REDUCED` only if you intentionally have overlapping address spaces and plan to manage routing via route maps
- **Two attachments** -- the minimum useful CEN configuration; add more attachments as your network topology grows (up to 20 per instance by default)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Alibaba Cloud region (e.g., `cn-hangzhou`) | Your deployment region |
| `<your-cen-name>` | CEN instance name (2-128 characters) | Choose a descriptive name |
| `<your-team>` | Team or business unit tag | Your organizational structure |
| `<first-vpc-id>` | VPC ID of the first network to connect | `AlicloudVpc` stack outputs |
| `<second-vpc-id>` | VPC ID of the second network to connect | `AlicloudVpc` stack outputs |

## Related Presets

- **02-cross-region-backbone** -- Connects VPCs across multiple regions for a global private backbone
