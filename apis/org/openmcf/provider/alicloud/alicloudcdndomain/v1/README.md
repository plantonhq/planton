# AlicloudCdnDomain

Manages an Alibaba Cloud CDN accelerated domain.

## Overview

A CDN domain maps a user-facing domain name to one or more origin servers. Alibaba Cloud's globally distributed edge nodes cache and serve content from the origins, reducing latency for end users. After creating the CDN domain, create a CNAME record at your DNS provider pointing the domain name to the `cname` value returned in the stack outputs.

### What Gets Created

- **CDN Domain** -- an accelerated domain registered in the Alibaba Cloud CDN service
- **Origin Sources** -- one or more origin server configurations with load balancing via priority and weight
- **HTTPS Certificate** -- optional TLS certificate for HTTPS acceleration
- **Tags** -- system metadata tags merged with user-defined tags

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | string | Alibaba Cloud region for provider initialization (e.g., `cn-hangzhou`, `cn-shanghai`) |
| `domainName` | string | The accelerated domain name (e.g., `cdn.example.com`). Cannot be changed after creation. |
| `cdnType` | string | Content type: `web`, `download`, or `video` |
| `sources` | list | At least one origin server source (see below) |

### Source Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `type` | string | required | Origin type: `ipaddr`, `domain`, `oss`, or `common` |
| `content` | string | required | Origin address (IP, domain, or OSS bucket domain) |
| `port` | int | `80` | Origin port (typically 80 or 443) |
| `priority` | int | `20` | Source priority (0-100, lower = higher priority) |
| `weight` | int | `10` | Load balancing weight (0-100) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `scope` | string | `domestic` | Geographic scope: `domestic`, `overseas`, or `global` |
| `certificateConfig` | object | none | HTTPS certificate configuration (see below) |
| `checkUrl` | string | `""` | URL for origin health check during creation |
| `resourceGroupId` | string | `""` | Resource group for access control and cost attribution |
| `tags` | map | `{}` | Key-value tags applied to the domain |

### Certificate Config Fields

| Field | Type | Description |
|-------|------|-------------|
| `certName` | string | Certificate display name |
| `certType` | string | Certificate type: `upload`, `cas`, or `free` |
| `certId` | string | CAS certificate ID (when certType=cas) |
| `certRegion` | string | CAS certificate region (`cn-hangzhou` or `ap-southeast-1`) |
| `serverCertificate` | string | PEM certificate content (when certType=upload) |
| `privateKey` | string | PEM private key content (when certType=upload) |
| `serverCertificateStatus` | string | HTTPS enabled: `on` (default) or `off` |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `domain_name` | The accelerated domain name as registered |
| `cname` | CNAME value -- create a DNS CNAME record pointing your domain here |
| `status` | Current domain status (e.g., `online`, `offline`, `configuring`) |

## Related Components

- **AlicloudDnsRecord** -- create a CNAME record pointing to this domain's `cname` output
- **AlicloudStorageBucket** -- use as an OSS origin source for static content
