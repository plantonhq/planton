# Pulumi Module to Deploy AlicloudDnsRecord

This module creates an Alibaba Cloud DNS record in the Alidns service. It supports all standard record types (A, AAAA, CNAME, MX, TXT, NS, SRV, CAA) with configurable TTL, priority, resolution lines, and record status.

Generated from the proto schema for `AlicloudDnsRecord`.

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

**Note**: Alibaba Cloud credentials are provided via environment variables (`ALICLOUD_ACCESS_KEY`, `ALICLOUD_SECRET_KEY`) or through the OpenMCF provider config -- not in the manifest `spec`.

## What This Module Creates

- **DNS Record** (`dns.AlidnsRecord`) -- a single DNS record within the specified parent domain

The module does not manage the parent domain. The domain must already exist in Alidns, managed by the AlicloudDnsDomain component or added manually.

## Prerequisites

- [OpenMCF CLI](https://github.com/plantonhq/openmcf) installed
- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/) installed
- Alibaba Cloud credentials configured via environment variables or OpenMCF provider config
- Parent domain registered in Alidns

## Usage

1. Create or edit a manifest:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudDnsRecord
metadata:
  name: my-record
spec:
  region: cn-hangzhou
  domainName: example.com
  rr: www
  type: A
  value: "203.0.113.10"
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
| `record_id` | The record ID assigned by Alibaba Cloud |

## Further Reading

- [examples.md](./examples.md) -- runnable manifest examples with CLI commands
- [overview.md](./overview.md) -- module architecture and design decisions
- [hack/manifest.yaml](../hack/manifest.yaml) -- minimal test manifest
