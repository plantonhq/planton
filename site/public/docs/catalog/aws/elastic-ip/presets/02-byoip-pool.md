---
title: "Preset: BYOIP Pool Elastic IP"
description: "**Use case:** Allocate a static public IPv4 address from your own registered IP address range."
type: "preset"
rank: "02"
presetSlug: "02-byoip-pool"
componentSlug: "elastic-ip"
componentTitle: "Elastic IP"
provider: "aws"
icon: "package"
order: 2
---

# Preset: BYOIP Pool Elastic IP

**Use case:** Allocate a static public IPv4 address from your own registered IP address range.

Use this preset when your organization has registered a Bring-Your-Own-IP (BYOIP) address range with AWS and you need EIPs from that specific pool.

## What You Get

- A VPC Elastic IP allocated from your BYOIP pool
- The IP comes from your organization's registered address range
- Same outputs as a standard EIP: `allocation_id`, `public_ip`, `arn`, `public_dns`

## When to Use

- Migrating on-premises services to AWS while keeping existing public IPs
- Organization policy requires using corporate-owned IP address ranges
- External partners have already whitelisted your IP ranges

## Prerequisites

- A BYOIP address range registered with AWS (done outside Planton)
- The pool ID from the registered range (format: `ipv4pool-ec2-xxx`)

## Customization

To request a **specific IP** from the pool, add the `address` field:

```yaml
spec:
  publicIpv4Pool: ipv4pool-ec2-0123456789abcdef0
  address: "198.51.100.10"
```
