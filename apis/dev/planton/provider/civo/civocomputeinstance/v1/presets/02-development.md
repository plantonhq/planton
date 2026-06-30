# Development Instance

This preset creates a minimal, cost-effective compute instance for development and testing. Uses the smallest instance size with VPC networking but no explicit firewall or cloud-init, keeping configuration simple and boot time fast.

## When to Use

- Development and testing environments
- Proof-of-concept deployments and experimentation
- Temporary instances for CI/CD build agents or one-off tasks

## Key Configuration Choices

- **Small instance** (`size: g3.small`) -- lowest cost; sufficient for most dev/test workloads
- **Ubuntu 22.04 LTS** (`image: ubuntu-jammy`) -- consistent with production image for parity
- **VPC networking** (`network`) -- even dev instances should use private networking for security
- **No firewall** -- simplifies access during development; add a firewall for staging environments
- **No cloud-init** -- faster boot; install tools manually or add `userData` as needed
- **No volumes** -- use the instance's local disk; attach a `CivoVolume` if persistent storage is needed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<vpc-network-id>` | Network ID of the target CivoVpc | `CivoVpc` status outputs |

## Related Presets

- **01-production-web** -- Use instead for production deployments with firewall, cloud-init, and proper tagging
