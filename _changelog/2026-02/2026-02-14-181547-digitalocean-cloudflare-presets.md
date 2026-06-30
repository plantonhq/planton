# DigitalOcean and Cloudflare Presets -- 43 Presets Across 23 Components

**Date**: February 14, 2026
**Type**: Feature
**Components**: Presets System, DigitalOcean Provider, Cloudflare Provider

## Summary

Created production-quality presets for all 15 DigitalOcean and all 8 Cloudflare deployment components, adding 43 presets (86 files) to the Planton presets system. This brings the total presets coverage from 169/213 components (79%) to 192/213 components (90%), with only 21 components remaining across Civo, Snowflake, and OpenFGA.

## Problem Statement / Motivation

DigitalOcean and Cloudflare were the two largest remaining providers without presets. Together they represent 23 components (11% of the total 213) and cover a significant portion of the "small cloud" deployment patterns that Planton users encounter.

### Pain Points

- Users deploying to DigitalOcean had no ready-made starting points for any of the 15 resource types
- Cloudflare's edge platform (Workers, R2, D1, KV, Load Balancer) lacked configuration templates despite being popular for modern web architectures
- Zero Trust Access Application configurations require understanding policy types, email patterns, and MFA settings that benefit from curated examples

## Solution / What's New

### DigitalOcean (15 components, 29 presets)

Organized into 4 functional groups:

- **Network Foundation** (4 components, 7 presets): VPC, Droplet, Volume, Firewall -- covering production/dev tiers, ext4/XFS filesystem choices, web/database firewall patterns with tag-based targeting
- **Kubernetes & Compute** (4 components, 8 presets): DOKS Cluster, Node Pool, App Platform Service, Function -- covering HA production clusters, autoscaling node pools, git-source vs container-image deployment, web API vs scheduled job functions
- **Data, DNS & Storage** (4 components, 8 presets): Database Cluster, Bucket, Container Registry, DNS Zone -- covering engine-specific presets (PostgreSQL HA, PostgreSQL dev, Redis), private/public bucket access, zone-with-records patterns
- **Load Balancing, DNS Records & Certificates** (3 components, 6 presets): Load Balancer, DNS Record, Certificate -- covering HTTPS SSL termination, Let's Encrypt vs custom certificate variants

### Cloudflare (8 components, 14 presets)

Organized into 2 functional groups:

- **DNS & Security** (3 components, 5 presets): DNS Zone, DNS Record, Zero Trust Access Application -- covering free-plan zones, proxied A records, MX email records, company-wide email policies, Google Workspace group policies with MFA
- **Edge Platform** (5 components, 9 presets): D1 Database, KV Namespace, R2 Bucket, Worker, Load Balancer -- covering private/public CDN buckets, API workers with KV bindings, active-passive/geographic/weighted load balancing

## Implementation Details

### Key Conventions Applied

- **apiVersion**: `digital-ocean.planton.dev/v1` (all 15 DO components), `cloudflare.planton.dev/v1` (all 8 CF components) -- verified from `api.proto`
- **camelCase field names**: consistent with all prior providers (AWS, GCP, Azure, Kubernetes, OpenStack, Scaleway)
- **StringValueOrRef `value:` wrapper**: correctly applied to VPC, cluster, domain, zone ID, droplet ID, and KV binding references
- **Plain string foreign keys**: `zone_id` in CloudflareWorker DNS config and CloudflareZeroTrustAccessApplication are plain `string` (not `StringValueOrRef`), correctly rendered without `value:` wrapper
- **Angle-bracket placeholders**: all user-specific values use `<lowercase-hyphenated-description>` format
- **No hardcoded domains**: all `example.com` references replaced with proper placeholders

### Notable Design Decisions

- **DigitalOcean Database Cluster**: 3 engine-specific presets (PostgreSQL HA, PostgreSQL dev, Redis) -- consistent with AWS RDS and GCP CloudSQL approach
- **DigitalOcean DNS Zone**: includes inline records (A record in website preset, A+MX+TXT in email preset) since DO's zone model supports inline records as the primary pattern
- **DigitalOcean Firewall**: tag-based targeting (`sourceTags`, `tags`) preferred over explicit Droplet IDs -- follows DO's recommended "tag-first" philosophy
- **Cloudflare Load Balancer**: 3 presets for 3 distinct steering policies (off=failover, geo, random=weighted) -- genuinely different deployment patterns
- **Cloudflare DNS Zone**: zone-only (no inline records) for rank 01 -- records are better expressed in standalone CloudflareDnsRecord presets

### Reference Material Coverage

- All 15 DO components have `spec.proto`, `examples.md`, and `docs/README.md`
- Only 1 DO component (digitaloceandnsrecord) has a `hack/manifest.yaml`
- All 8 CF components have comprehensive `examples.md` and `docs/README.md`; most CF `hack/manifest.yaml` files contain only metadata

## Benefits

- **90% preset coverage**: 192/213 components now have presets (up from 79%)
- **43 new ready-made starting points**: users can deploy DigitalOcean and Cloudflare resources with minimal configuration
- **Consistent quality**: all presets follow the same conventions established in T01 and validated across 6 prior providers
- **Cross-provider consistency**: DNS zones, database clusters, container registries, and firewalls follow the same patterns as their AWS/GCP/Azure equivalents

## Impact

- DigitalOcean users: 15 new components with presets, covering the full DO infrastructure stack from VPC to Kubernetes to managed databases
- Cloudflare users: 8 new components with presets, covering DNS, edge computing, storage, load balancing, and zero-trust security
- Only 21 components remain without presets: Civo (12), OpenFGA (3), Snowflake (1), plus any remaining from other providers

## Related Work

- **T01**: Foundation (convention document, Cursor rules, Forge integration)
- **T02**: AWS presets (25 components, 49 presets)
- **T03**: GCP presets (19 components, 36 presets)
- **T04**: Azure presets (29 components, 55 presets)
- **T05**: Kubernetes presets (51 components, 83 presets)
- **T06**: OpenStack presets (27 components, 44 presets)
- **T07**: Scaleway presets (18 components, 35 presets)

---

**Status**: Production Ready
**Timeline**: Session 9 (T08a)
