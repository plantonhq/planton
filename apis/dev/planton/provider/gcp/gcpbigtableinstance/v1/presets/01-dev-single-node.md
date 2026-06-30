# Dev Single Node

Minimal development instance with a single auto-allocated cluster. Deletion protection disabled for easy teardown.

## Description

This preset provisions the smallest possible Bigtable instance suitable for development and testing. With no explicit node count or autoscaling, Bigtable auto-allocates nodes based on the data footprint (typically 1 node for small datasets). SSD storage is applied by default. Deletion protection is disabled so instances can be destroyed without override.

## Use Case

- Local development and integration testing
- CI/CD pipelines that need a temporary Bigtable instance
- Proof-of-concept or prototyping with NoSQL workloads
- Cost-sensitive dev/test workloads

## What This Preset Configures

- **Single cluster** — One cluster in a single zone
- **Auto-allocated nodes** — Bigtable manages node count based on data footprint
- **SSD storage** — Default, lowest-latency storage type
- **No CMEK** — Data encrypted with Google-managed keys
- **deletion_protection: false** — Instance can be destroyed without explicit override

## When to Use It

Use this preset when you need a quick, low-cost Bigtable instance for non-production workloads. Choose a different preset if you need multi-cluster replication, autoscaling controls, or CMEK encryption.
