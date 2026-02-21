# OCI Redis Cluster

Deploys an OCI Cache (Redis) cluster — a fully managed, Redis-compatible in-memory caching service on Oracle Cloud Infrastructure. Supports non-sharded and sharded topologies.

## Overview

This component provisions a single `oci_redis_redis_cluster` resource in a specified compartment and subnet. It wraps the OCI Cache service, handling display-name defaulting, freeform-tag propagation from metadata, and Pulumi stack output export for downstream consumption via `StringValueOrRef`.

## Purpose

Provide a declarative, YAML-driven interface for creating Redis-compatible cache clusters on OCI. The component abstracts away direct OCI API calls and Pulumi boilerplate, exposing only the fields relevant to cluster topology, sizing, and network placement.

## Key Features

- **Non-sharded and sharded modes** — `clusterMode` selects between a single-shard topology (one primary + replicas) and a multi-shard topology (horizontal data partitioning). When unset, OCI defaults to non-sharded.
- **Per-node sizing** — `nodeCount` and `nodeMemoryInGbs` control replica count and memory per node. Both are updatable without recreation.
- **Network placement** — the cluster is placed in a specific subnet. Optional `nsgIds` attach network security groups for fine-grained access control.
- **Custom configuration** — `configSetId` references an OCI Cache Config Set for tuning Redis parameters (maxmemory-policy, timeout, etc.) without managing the config set lifecycle.
- **Foreign key references** — `compartmentId`, `subnetId`, `nsgIds`, and `configSetId` accept both literal OCIDs and `valueFrom` references to other OpenMCF resources.
- **Automatic tagging** — freeform tags are populated from metadata labels, organization, and environment fields. No manual tag management needed.

## Critical Constraints

- **Cluster mode is immutable** — changing `clusterMode` forces cluster recreation. Choose the topology before initial deployment.
- **Subnet change forces recreation** — changing `subnetId` destroys and re-creates the cluster.
- **Shard count requires sharded mode** — the CEL validation rule enforces that `shardCount` must be > 0 when `clusterMode` is `sharded`.
- **Config sets are external** — this component references config sets by OCID but does not create or manage them. Config sets have independent lifecycles.
- **Tags are auto-managed** — `defined_tags`, `system_tags`, and `freeform_tags` are not exposed as spec fields. Freeform tags are derived from metadata.

## Use Cases

| Scenario | Configuration |
|----------|---------------|
| Development cache | Non-sharded, 2 nodes, 2 GB each |
| Session store | Non-sharded, 3 nodes, 8 GB each, NSG-restricted |
| High-throughput production cache | Sharded, 3+ shards, 3 nodes per shard, 16-32 GB each |
| Custom-tuned cache | Any topology + `configSetId` pointing to a Config Set with custom Redis parameters |

## Production Features

- **High availability** — non-sharded clusters with `nodeCount` >= 2 provide automatic failover. Sharded clusters replicate within each shard.
- **Horizontal scaling** — sharded mode distributes keys across shards. Increase `shardCount` to scale capacity and throughput.
- **Network isolation** — subnet placement combined with NSGs restricts access to authorized resources only.
- **Observability** — cluster OCID, primary/replica FQDNs, and primary IP are exported as stack outputs for integration with monitoring and DNS configuration.
