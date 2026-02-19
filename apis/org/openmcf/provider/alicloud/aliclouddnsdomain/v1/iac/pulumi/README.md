# Pulumi Module to Deploy AlicloudDnsDomain

This module provisions an Alibaba Cloud DNS domain in the Alidns service with optional group assignment, resource group placement, remarks, and automatic tag management. It creates a single `dns.AlidnsDomain` resource and exports the domain ID, domain name, DNS servers, group name, and punycode.

Generated from the proto schema for `AlicloudDnsDomain`.

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

- **DNS Domain** (`dns.AlidnsDomain`) -- registers a domain in the Alibaba Cloud Alidns service so that DNS records can be created against it
- **Tags** -- system metadata tags (`resource`, `resource_name`, `resource_kind`, `organization`, `environment`) merged with user-defined `spec.tags`, with user values taking precedence on key conflict

The module does not create DNS records. Records are managed by the separate AlicloudDnsRecord component.

## Prerequisites

- [OpenMCF CLI](https://github.com/plantonhq/openmcf) installed
- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/) installed
- Alibaba Cloud credentials configured via environment variables or OpenMCF provider config

## Usage

1. Create or edit a manifest:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudDnsDomain
metadata:
  name: my-domain
spec:
  region: cn-hangzhou
  domainName: example.com
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
| `domain_id` | The domain ID assigned by Alibaba Cloud |
| `domain_name` | The domain name as registered |
| `dns_servers` | DNS server names assigned by Alibaba Cloud (point your registrar's NS records here) |
| `group_name` | Computed domain group name |
| `puny_code` | Punycode representation for internationalized domain names |

## Further Reading

- [examples.md](./examples.md) -- runnable manifest examples with CLI commands
- [overview.md](./overview.md) -- module architecture and design decisions
- [hack/manifest.yaml](../hack/manifest.yaml) -- minimal test manifest
