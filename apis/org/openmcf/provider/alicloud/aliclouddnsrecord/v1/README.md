# AlicloudDnsRecord

Manages an Alibaba Cloud DNS record in the Alidns service.

## Overview

A DNS record maps a host record (subdomain) to a value within a parent domain hosted in Alibaba Cloud Alidns. The parent domain must already exist in Alidns -- either managed by the AlicloudDnsDomain component or added manually via the console.

### What Gets Created

- **Alidns Record** -- a single DNS record (`alicloud_alidns_record`) of the specified type (A, AAAA, CNAME, MX, TXT, NS, SRV, CAA) within the parent domain

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | string | Alibaba Cloud region for provider initialization (e.g., `cn-hangzhou`, `cn-shanghai`) |
| `domainName` | string | Parent domain name (e.g., `example.com`). Cannot be changed after creation. |
| `rr` | string | Host record / subdomain (e.g., `www`, `@`, `*`, `mail`). "rr" stands for Resource Record. |
| `type` | string | Record type: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `NS`, `SRV`, `CAA`, `REDIRECT_URL`, `FORWORD_URL` |
| `value` | string | Record value (IP address, CNAME target, TXT content, etc.) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ttl` | int32 | `600` | Time-to-live in seconds |
| `priority` | int32 | - | MX record priority (1-10). Required when type is `MX`. |
| `line` | string | `"default"` | DNS resolution line for ISP/geo-based routing |
| `status` | string | `"ENABLE"` | `ENABLE` or `DISABLE` -- disable a record without deleting it |
| `remark` | string | `""` | Description visible in the Alidns console |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `record_id` | The record ID assigned by Alibaba Cloud |

## Related Components

- **AlicloudDnsDomain** -- registers the parent domain in Alidns (prerequisite for records)
- **AlicloudPrivateZone** -- for private DNS resolution within a VPC (separate from public Alidns)
