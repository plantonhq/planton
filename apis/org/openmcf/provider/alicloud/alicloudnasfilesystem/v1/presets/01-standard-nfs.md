# Standard NFS File System

This preset creates a minimal standard NAS file system with NFS protocol, Performance storage, and default VPC-wide access. The file system auto-scales capacity as data is written, with no pre-allocation needed.

## When to Use

- Development and staging environments needing shared file storage
- Applications requiring POSIX-compliant shared storage across multiple ECS instances
- Kubernetes clusters needing ReadWriteMany persistent volumes
- Quick prototyping before configuring production-grade access controls

## Key Configuration Choices

- **Standard file system** (default) -- auto-scaling capacity, no pre-allocation. You pay only for stored data.
- **NFS protocol** -- the standard Linux/Unix file system protocol. Supports NFS v3 and v4.0.
- **Performance storage** -- SSD-backed tier with low latency and up to 10 GiB/s throughput. Suitable for most workloads.
- **Default VPC access** -- no custom access rules means the mount target uses the built-in DEFAULT_VPC_GROUP_NAME, allowing full read-write access from all IP addresses within the VPC.
- **No encryption** -- data stored without encryption. Add an `encryption` block for production workloads.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`, `ap-southeast-1`) | Your deployment region strategy |
| `<vpc-id>` | VPC ID where the mount target will be created | AliCloudVpc outputs or Alibaba Cloud console |
| `<vswitch-id>` | VSwitch ID within the VPC | AliCloudVswitch outputs or Alibaba Cloud console |

## Related Presets

- **02-production-encrypted** -- use instead for production workloads requiring encryption and restrictive access rules
