# Spot Worker Instance

This preset runs a standalone instance on Spot capacity -- typically 60-90% cheaper than On-Demand -- configured as a persistent request that stops (rather than terminates) on interruption and resumes when capacity returns. The right shape for interruption-tolerant workers: build agents, batch processors, dev sandboxes.

## When to Use

- CI/build agents and batch workers that checkpoint their work
- Development or experimentation boxes where an occasional pause is acceptable
- Any workload where the Spot discount outweighs the two-minute interruption notice

## Key Configuration Choices

- **Persistent Spot request** (`spotOptions.spotInstanceType: persistent`) -- EC2 automatically re-requests capacity after an interruption; one-time requests would leave the worker gone
- **Stop on interruption** (`instanceInterruptionBehavior: stop`) -- The instance's EBS volumes survive the reclaim; work resumes where it left off when capacity returns
- **No max price** -- Leaving `maxPrice` unset caps the bid at the On-Demand price, the AWS recommendation (the discount comes from interruption risk, not bidding)
- **m7g.large** (`instanceType`) -- Graviton price-performance compounds the Spot discount

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | AWS region where the instance will be created | AWS region list |
| `ami-0123456789abcdef0` | Amazon Machine Image ID matching the instance type's architecture | AWS EC2 AMI catalog |
| `<private-subnet-id>` | Private subnet ID where the instance will launch | `AwsSubnet` status outputs |
| `<security-group-id>` | Security group ID controlling instance traffic | `AwsSecurityGroup` status outputs |
| `<instance-profile-name>` | NAME of the IAM instance profile the worker assumes | `AwsIamInstanceProfile` status outputs |

## Related Presets

- **01-ssm-managed** -- On-Demand baseline for instances that must never be interrupted
- **02-launch-template** -- Inherit the org's golden template instead of inline configuration
