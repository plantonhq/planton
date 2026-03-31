# Pulumi Module to Deploy AliCloudVswitch

This module provisions an Alibaba Cloud VSwitch (subnet) within an existing VPC. It creates a single `vpc.Switch` resource bound to a specific Availability Zone with a configured IPv4 CIDR block, optional IPv6 support, and automatic tag management. The module exports the VSwitch ID, name, CIDR block, zone ID, and IPv6 CIDR block.

Generated from the proto schema for `AliCloudVswitch`.

## CLI Usage (OpenMCF Pulumi)

```bash
# Preview changes
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Apply changes
openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Refresh state from cloud
openmcf pulumi refresh \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Tear down
openmcf pulumi destroy \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

**Note**: Alibaba Cloud credentials are provided via environment variables (`ALICLOUD_ACCESS_KEY`, `ALICLOUD_SECRET_KEY`) or through the OpenMCF provider config — not in the manifest `spec`.

## What This Module Creates

- **VSwitch** (`vpc.Switch`) — an isolated subnet within a VPC, bound to a single Availability Zone with a dedicated IPv4 CIDR block

The module resolves the `vpc_id` field from `StringValueOrRef` (supporting both literal VPC IDs and cross-resource references), merges user-defined `spec.tags` with system tags (resource name, kind, organization, environment), and conditionally configures IPv6 when `ipv6_cidr_block_mask` is non-zero. User tags take precedence when keys overlap.

## Prerequisites

- [OpenMCF CLI](https://github.com/plantonhq/openmcf) installed
- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/) installed
- Alibaba Cloud credentials configured via environment variables or OpenMCF provider config
- An existing VPC (the VSwitch's `vpcId` must reference a valid VPC ID)

## Usage

1. Create or edit a manifest:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudVswitch
metadata:
  name: my-vswitch
spec:
  region: cn-hangzhou
  vpcId: vpc-bp1234567890abcdef
  zoneId: cn-hangzhou-a
  cidrBlock: "10.0.0.0/24"
  vswitchName: my-vswitch
```

2. Preview:

```bash
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

3. Apply:

```bash
openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

## Stack Outputs

| Output | Description |
|--------|-------------|
| `vswitch_id` | The VSwitch ID assigned by Alibaba Cloud |
| `vswitch_name` | The VSwitch name as created |
| `cidr_block` | The IPv4 CIDR block of the VSwitch |
| `zone_id` | The Availability Zone in which the VSwitch resides |
| `ipv6_cidr_block` | The IPv6 CIDR block (empty if IPv6 is not enabled) |

## Further Reading

- [examples.md](./examples.md) — runnable manifest examples with CLI commands
- [overview.md](./overview.md) — module architecture and design decisions
- [hack/manifest.yaml](../hack/manifest.yaml) — minimal test manifest
