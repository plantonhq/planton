# Pulumi Module to Deploy AliCloudVpc

This module provisions an Alibaba Cloud Virtual Private Cloud (VPC) with configurable CIDR block, optional IPv6 support, resource group assignment, and automatic tag management. It creates a single `vpc.Network` resource and exports the VPC ID, name, CIDR block, router ID, and route table ID.

Generated from the proto schema for `AliCloudVpc`.

## CLI Usage (Planton Pulumi)

```bash
# Preview changes
planton pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Apply changes
planton pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Refresh state from cloud
planton pulumi refresh \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Tear down
planton pulumi destroy \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

**Note**: Alibaba Cloud credentials are provided via environment variables (`ALICLOUD_ACCESS_KEY`, `ALICLOUD_SECRET_KEY`) or through the Planton provider config — not in the manifest `spec`.

## What This Module Creates

- **VPC** (`vpc.Network`) — an isolated virtual network with a primary IPv4 CIDR block
- **VRouter** — automatically created by Alibaba Cloud with the VPC
- **System Route Table** — the default route table associated with the VRouter

The module merges user-defined `spec.tags` with system tags (resource name, kind, organization, environment) and applies them to the VPC. User tags take precedence when keys overlap.

## Prerequisites

- [Planton CLI](https://github.com/plantonhq/planton) installed
- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/) installed
- Alibaba Cloud credentials configured via environment variables or Planton provider config

## Usage

1. Create or edit a manifest:

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudVpc
metadata:
  name: my-vpc
spec:
  region: cn-hangzhou
  vpcName: my-vpc
  cidrBlock: "10.0.0.0/16"
```

2. Preview:

```bash
planton pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

3. Apply:

```bash
planton pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

## Stack Outputs

| Output | Description |
|--------|-------------|
| `vpc_id` | The VPC ID assigned by Alibaba Cloud |
| `vpc_name` | The VPC name as created |
| `cidr_block` | The primary IPv4 CIDR block |
| `router_id` | The VRouter ID automatically created with the VPC |
| `route_table_id` | The system route table ID |

## Further Reading

- [examples.md](./examples.md) — runnable manifest examples with CLI commands
- [overview.md](./overview.md) — module architecture and design decisions
- [hack/manifest.yaml](../hack/manifest.yaml) — minimal test manifest
