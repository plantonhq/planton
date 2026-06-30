# AliCloud CdnDomain

Deploys an Alibaba Cloud CDN accelerated domain. The component registers a domain name in the CDN service, configures one or more origin sources with priority-based failover and weighted load balancing, and optionally enables HTTPS with certificate management. After deployment, create a DNS CNAME record pointing the accelerated domain to the `cname` stack output for edge acceleration to take effect.

## What Gets Created

When you deploy an AliCloudCdnDomain resource, Planton provisions:

- **CDN Domain** -- an `alicloud_cdn_domain_new` resource registered in the Alibaba Cloud CDN service with the specified content type and geographic scope
- **Origin Sources** -- one or more origin server configurations with type, address, port, priority, and weight for failover and load distribution
- **HTTPS Certificate** -- optional TLS certificate configuration (CAS-managed, uploaded, or free DV) for HTTPS acceleration on edge nodes
- **Tags** -- system metadata tags (`resource_name`, `resource_kind`, `organization`, `environment`) merged with user-defined tags

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables or Planton provider config
- **Domain name** with DNS access to create the required CNAME record after deployment
- **ICP filing** if the domain uses `domestic` or `global` scope (mainland China requirement)
- **CAS certificate** if using `certType: cas` for HTTPS (certificate must exist in Certificate Management Service)

## Quick Start

Create a file `cdn-domain.yaml`:

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudCdnDomain
metadata:
  name: my-cdn
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: cdn-project
    pulumi.planton.dev/stack.name: dev.AliCloudCdnDomain.my-cdn
spec:
  region: cn-hangzhou
  domainName: cdn.example.com
  cdnType: web
  sources:
    - type: ipaddr
      content: "203.0.113.10"
```

Deploy:

```shell
planton apply -f cdn-domain.yaml
```

This creates a CDN domain accelerating web content from a single IP origin in the `cn-hangzhou` region.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region for provider API calls (e.g., `cn-hangzhou`). | Required; non-empty |
| `domainName` | `string` | The accelerated domain name (e.g., `cdn.example.com`). Immutable after creation. | Required; 1-63 chars |
| `cdnType` | `string` | Content type: `web`, `download`, or `video`. Immutable after creation. | Required; must be one of the listed values |
| `sources` | `list` | Origin server sources (see below). | At least one required |
| `sources[].type` | `string` | Origin type: `ipaddr`, `domain`, `oss`, or `common`. | Required; must be one of the listed values |
| `sources[].content` | `string` | Origin address (IP, domain name, or OSS bucket domain). | Required; non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `scope` | `string` | `domestic` | Geographic scope: `domestic`, `overseas`, or `global`. |
| `sources[].port` | `int32` | `80` | Origin port (typically 80 or 443). |
| `sources[].priority` | `int32` | `20` | Source priority (0-100, lower = higher priority). Use 20 for primary, 30 for standby. |
| `sources[].weight` | `int32` | `10` | Load balancing weight (0-100). Only effective when sources share the same priority. |
| `certificateConfig` | `object` | `null` | HTTPS certificate configuration. If omitted, CDN serves HTTP only. |
| `certificateConfig.certName` | `string` | `""` | Certificate display name. |
| `certificateConfig.certType` | `string` | `""` | Certificate type: `upload`, `cas`, or `free`. |
| `certificateConfig.certId` | `string` | `""` | CAS certificate ID. Required when `certType` is `cas`. |
| `certificateConfig.certRegion` | `string` | `cn-hangzhou` | CAS certificate region: `cn-hangzhou` (domestic) or `ap-southeast-1` (international). |
| `certificateConfig.serverCertificate` | `string` | `""` | PEM-encoded certificate content. Required when `certType` is `upload`. |
| `certificateConfig.privateKey` | `string` | `""` | PEM-encoded private key. Required when `certType` is `upload`. |
| `certificateConfig.serverCertificateStatus` | `string` | `on` | HTTPS status: `on` or `off`. |
| `checkUrl` | `string` | `""` | URL for origin health check during domain creation. |
| `resourceGroupId` | `string` | `""` | Resource group ID for access control and cost attribution. |
| `tags` | `map<string, string>` | `{}` | Key-value tags merged with system-generated tags. |

## Examples

### Minimal Web CDN

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudCdnDomain
metadata:
  name: my-cdn
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: cdn-project
    pulumi.planton.dev/stack.name: dev.AliCloudCdnDomain.my-cdn
spec:
  region: cn-hangzhou
  domainName: cdn.example.com
  cdnType: web
  sources:
    - type: ipaddr
      content: "203.0.113.10"
```

### OSS Static Assets with Failover

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudCdnDomain
metadata:
  name: assets-cdn
  org: platform-team
  env: production
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: platform-team
    pulumi.planton.dev/project: cdn-project
    pulumi.planton.dev/stack.name: prod.AliCloudCdnDomain.assets-cdn
spec:
  region: cn-shanghai
  domainName: assets.example.com
  cdnType: web
  scope: global
  sources:
    - type: oss
      content: my-assets.oss-cn-shanghai.aliyuncs.com
      priority: 20
    - type: domain
      content: origin-standby.example.com
      port: 443
      priority: 30
  checkUrl: http://my-assets.oss-cn-shanghai.aliyuncs.com/health.txt
  resourceGroupId: rg-prod-456
  tags:
    team: platform
    costCenter: engineering
```

### HTTPS with CAS Certificate

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudCdnDomain
metadata:
  name: secure-cdn
  org: my-org
  env: production
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: cdn-project
    pulumi.planton.dev/stack.name: prod.AliCloudCdnDomain.secure-cdn
spec:
  region: cn-hangzhou
  domainName: secure.example.com
  cdnType: web
  scope: domestic
  sources:
    - type: domain
      content: origin-a.example.com
      port: 443
      priority: 20
      weight: 60
    - type: domain
      content: origin-b.example.com
      port: 443
      priority: 20
      weight: 40
  certificateConfig:
    certType: cas
    certId: cas-cn-abc123
    certRegion: cn-hangzhou
    serverCertificateStatus: "on"
  tags:
    team: security
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `domain_name` | `string` | The accelerated domain name as registered in CDN |
| `cname` | `string` | The CNAME assigned by CDN -- create a DNS CNAME record pointing your domain to this value |
| `status` | `string` | Current domain status (`online`, `offline`, `configuring`, `checking`, `check_failed`) |

## Related Components

- [AliCloudDnsRecord](/docs/catalog/alicloud/aliclouddnsrecord) -- create the CNAME record pointing to the `cname` output
- [AliCloudStorageBucket](/docs/catalog/alicloud/alicloudstoragebucket) -- OSS bucket to use as an origin source
- [AliCloudCertificate](/docs/catalog/alicloud/alicloudcertificate) -- manage certificates in CAS for HTTPS configuration
