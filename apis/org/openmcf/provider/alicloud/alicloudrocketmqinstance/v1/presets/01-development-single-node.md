# Development Single-Node RocketMQ

A minimal RocketMQ 5.x instance for development and testing. Uses the standard edition with a single-node deployment, which is the cheapest way to get a working message broker for local integration testing, prototyping, or CI pipelines. No topics or consumer groups are pre-created -- add them as your application takes shape.

## When to Use

- Development and testing environments where cost matters more than availability
- Proof-of-concept projects exploring RocketMQ messaging patterns
- CI/CD pipelines that need a disposable message broker
- Learning and experimentation with Alibaba Cloud RocketMQ 5.x

## Key Configuration Choices

- **Standard edition** (`seriesCode: standard`) -- lowest cost tier with basic throughput; sufficient for development workloads where message volume is low
- **Single-node deployment** (`subSeriesCode: single_node`) -- no replication or HA; keeps costs minimal and startup fast. Not suitable for production
- **PayAsYouGo billing** (default) -- no upfront commitment; delete the instance when you are done
- **No topics or consumer groups** -- start clean and add messaging resources as your application design solidifies
- **No internet access** (default) -- instance is accessible only within the VPC, which is appropriate for development

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Alibaba Cloud region (e.g., `cn-hangzhou`, `ap-southeast-1`) | Your deployment region |
| `<your-vpc-id>` | VPC ID where the instance is deployed | `AlicloudVpc` stack outputs |

## Related Presets

- **02-production-ha** -- Professional edition with HA clustering, topics, and consumer groups for production workloads
- **03-enterprise-encrypted** -- Ultimate edition with encryption at rest, internet access, and subscription billing for compliance-sensitive environments
