---
title: "Private Standard Bucket"
description: "This preset creates a minimal private OSS bucket with default settings: Standard storage class, LRS redundancy, no versioning, no encryption, and no lifecycle rules. Ideal for getting started quickly..."
type: "preset"
rank: "01"
presetSlug: "01-private-standard"
componentSlug: "oss-bucket"
componentTitle: "OSS Bucket"
provider: "alicloud"
icon: "package"
order: 1
---

# Private Standard Bucket

This preset creates a minimal private OSS bucket with default settings: Standard storage class, LRS redundancy, no versioning, no encryption, and no lifecycle rules. Ideal for getting started quickly or for ephemeral workloads where data durability requirements are basic.

## When to Use

- Development and testing environments
- Temporary data storage that does not require versioning or encryption
- Application assets where the default LRS redundancy is sufficient
- Quick prototyping before configuring production-grade bucket features

## Key Configuration Choices

- **Private ACL** (default) -- objects are not publicly accessible; access is controlled via RAM policies and signed URLs.
- **Standard storage class** (default) -- frequent-access tier with the lowest latency and highest throughput. Objects can be transitioned to cheaper tiers later via lifecycle rules.
- **LRS redundancy** (default) -- locally redundant storage replicates data three times within a single availability zone. Suitable for non-critical data; upgrade to ZRS for production workloads requiring cross-zone durability.
- **No encryption** -- objects are stored without default server-side encryption. Individual objects can still be encrypted at upload time.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`, `ap-southeast-1`) | Your deployment region strategy |
| `<globally-unique-bucket-name>` | Bucket name (3-63 chars, lowercase + digits + hyphens, globally unique across all Alibaba Cloud accounts) | Choose a name with your org prefix (e.g., `myorg-dev-assets`) |

## Related Presets

- **02-versioned-encrypted** -- use instead for production data requiring versioning and encryption
- **03-archive-lifecycle** -- use instead for log archival or data with lifecycle tiering requirements
