# Production Encrypted NFS

This preset creates a production-grade NAS file system with NAS-managed encryption at rest and a custom access group restricting mount access to a specific application subnet. Root squashing is enabled for defense-in-depth.

## When to Use

- Production workloads requiring encryption at rest
- Environments with compliance requirements (e.g., data residency, audit)
- Multi-tenant VPCs where not all subnets should access the file system
- Kubernetes clusters requiring controlled access to shared persistent storage

## Key Configuration Choices

- **NAS-managed encryption** (`encryptType: 1`) -- all data encrypted at rest using Alibaba Cloud's managed key. No KMS key management needed.
- **Custom access rules** -- restricts mount access to a specific CIDR block instead of allowing all VPC IPs. Add more `accessRules` entries for additional subnets.
- **Root squash** -- maps root user (uid 0) to anonymous user on the NAS side, preventing containers or instances running as root from having NAS root-level access.
- **Performance storage** -- SSD-backed with low latency. Switch to `Capacity` for cost-sensitive warm/cold data workloads.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`) | Your deployment region strategy |
| `<organization>` | Organization name for tag-based resource grouping | Your OpenMCF org configuration |
| `<purpose-description>` | Human-readable purpose (e.g., "Shared config for payment service") | Your service catalog |
| `<vpc-id>` | VPC ID where the mount target will be created | AlicloudVpc outputs |
| `<vswitch-id>` | VSwitch ID within the VPC | AlicloudVswitch outputs |
| `<application-subnet-cidr>` | CIDR block of the subnet that should access NAS (e.g., `10.0.1.0/24`) | Your VSwitch CIDR allocation |

## Related Presets

- **01-standard-nfs** -- use instead for development/staging environments where encryption and access control are not critical
