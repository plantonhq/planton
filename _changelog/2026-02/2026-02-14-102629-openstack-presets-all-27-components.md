# OpenStack Presets: All 27 Components

**Date**: February 14, 2026
**Type**: Feature
**Components**: Presets System, OpenStack Provider

## Summary

Created 44 production-quality presets across all 27 OpenStack deployment components (88 files total: YAML + MD pairs). This completes the OpenStack provider in the presets initiative, bringing the total to 151 of 213 components covered (268 presets across 5 providers).

## Problem Statement / Motivation

OpenStack's 27 components span networking, security, identity, load balancing, compute, storage, DNS, and container orchestration. Users need opinionated starting points for each, especially given OpenStack's modular architecture where a single use case (e.g., a load-balanced web app) requires coordinating 5+ standalone resources.

### Pain Points

- OpenStack's Octavia load balancer is decomposed into 5 separate resources (LB, Listener, Pool, Member, Monitor) -- users need consistent cross-reference placeholders across the set
- 6 components (entire Security/Identity group) have no hack manifests or examples.md, making configuration knowledge harder to discover
- Networking resources form a foundational dependency chain (Network -> Subnet -> Router -> RouterInterface) with correct placeholder naming critical for downstream consumers

## Solution / What's New

44 presets organized in 4 functional batches:

### Batch 1: Networking Foundation (7 components, 11 presets)

Network, Subnet (standard-dhcp, isolated-no-gateway), Router (edge-with-snat, internal-only), RouterInterface, FloatingIp (allocation-only, with-port-association), FloatingIpAssociate, NetworkPort (standard-fixed-ip, no-security-groups).

### Batch 2: Security & Identity (6 components, 9 presets)

SecurityGroup (web-server, restrictive), SecurityGroupRule (allow-ssh, allow-http-https), Keypair (import-public-key), ApplicationCredential (restricted-readonly, compute-scoped), RoleAssignment (project-user-member), Project (standard).

### Batch 3: Load Balancer Stack (5 components, 9 presets)

LoadBalancer (standard), Listener (http, https-terminated, tcp-passthrough), Pool (round-robin, sticky-session), Member (standard), Monitor (http-health-check, tcp-health-check).

### Batch 4: Compute, Storage, DNS & Containers (9 components, 15 presets)

Instance (standard-vm, boot-from-volume), Image (cloud-image-from-url), ServerGroup (anti-affinity, affinity), Volume (blank-data, bootable-from-image), VolumeAttach (standard), DnsZone (primary-zone), DnsRecord (a-record, cname-record), ContainerCluster (dev-single-master, ha-multi-master), ContainerClusterTemplate (standard-kubernetes, production-kubernetes).

## Implementation Details

- All presets use camelCase field naming (proto3 JSON canonical form), consistent with AWS/GCP/Azure/Kubernetes presets
- StringValueOrRef fields use `value:` wrapper per convention
- Cross-reference placeholders are consistent across the LB stack (e.g., `<loadbalancer-id>` in Listener matches LB output naming)
- Security/Identity presets crafted entirely from spec.proto analysis (no hack manifests or examples.md existed for reference)
- Join resources (RouterInterface, FloatingIpAssociate, VolumeAttach) get exactly 1 preset each -- they have no meaningful configuration variants

## Benefits

- **Quick deployment**: Users can deploy any OpenStack resource by copying a preset and replacing placeholders
- **Consistent cross-references**: LB stack presets use matching placeholder names across all 5 resources
- **Documentation where none existed**: Security/Identity components now have concrete configuration examples for the first time
- **Production patterns**: Presets encode real-world best practices (e.g., 3-master HA clusters, boot-from-volume for persistent roots, anti-affinity server groups for fault isolation)

## Impact

- **27 OpenStack components** now have presets (100% coverage)
- **88 new files** added (44 YAML + 44 MD)
- **Cumulative progress**: 151/213 components covered, 268 total presets across AWS (49), GCP (36), Azure (55), Kubernetes (83), OpenStack (44) -- with 1 from the pilot

## Related Work

- Presets system foundation: `2026-02-14-075740-presets-system-foundation.md`
- AWS presets: `2026-02-14-083641-aws-presets-all-25-components.md`
- GCP presets: `2026-02-14-085825-gcp-presets-all-19-components.md`
- Azure presets: `2026-02-14-093405-azure-presets-all-29-components.md`
- Kubernetes presets: `2026-02-14-100325-kubernetes-presets-all-51-components.md`
- Next: Scaleway (18), DigitalOcean (1), Cloudflare (2), Civo (1), Snowflake (1), OpenFGA (3) -- 26 remaining

---

**Status**: Production Ready
