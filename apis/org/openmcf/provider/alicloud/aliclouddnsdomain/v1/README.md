# AlicloudDnsDomain

Manages an Alibaba Cloud DNS domain in the Alidns service.

## Overview

A DNS domain is the top-level container for DNS records in Alibaba Cloud's Alidns service. Registering a domain in Alidns does not purchase or transfer the domain -- it adds it to the Alidns hosted zone so that you can create and manage DNS records. After adding the domain, point your domain registrar's nameserver (NS) records to the DNS servers returned in the stack outputs.

### What Gets Created

- **Alidns Domain** -- a hosted zone in the Alidns service for the specified domain name
- **DNS Servers** -- Alibaba Cloud assigns a set of authoritative DNS servers for the domain
- **Tags** -- system metadata tags merged with user-defined tags

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | string | Alibaba Cloud region for provider initialization (e.g., `cn-hangzhou`, `cn-shanghai`) |
| `domainName` | string | The domain name to manage (e.g., `example.com`). Cannot be changed after creation. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `groupId` | string | `""` | Alidns domain group ID for organizational grouping |
| `remark` | string | `""` | Description or notes for the domain |
| `resourceGroupId` | string | `""` | Resource group for access control and cost attribution |
| `tags` | map | `{}` | Key-value tags applied to the domain |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `domain_id` | The domain ID assigned by Alibaba Cloud |
| `domain_name` | The domain name as registered |
| `dns_servers` | DNS server names -- point your registrar's NS records to these |
| `group_name` | Computed domain group name |
| `puny_code` | Punycode representation for internationalized domain names |

## Related Components

- **AlicloudDnsRecord** -- creates DNS records (A, AAAA, CNAME, MX, TXT, etc.) within this domain
- **AlicloudPrivateZone** -- for private DNS resolution within a VPC (separate from public Alidns)
