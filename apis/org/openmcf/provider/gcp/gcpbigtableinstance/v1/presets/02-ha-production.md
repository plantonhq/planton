# HA Production

Multi-cluster production instance with autoscaling and automatic replication across two zones.

## Description

This preset provisions a production-grade Bigtable instance with two clusters in different zones. Bigtable automatically replicates data between clusters, providing high availability and automatic failover. Autoscaling is configured to handle variable workloads between 3 and 30 nodes per cluster. Deletion protection is enabled to prevent accidental destruction.

## Use Case

- Production time-series, IoT, and analytics workloads
- Applications requiring high availability and automatic failover
- Workloads with variable traffic patterns that benefit from autoscaling
- Multi-zone replication for data durability

## What This Preset Configures

- **Two clusters** — Automatic replication across two zones in the same region
- **Autoscaling** — 3 to 30 nodes per cluster, targeting 65% CPU utilization
- **SSD storage** — Default, lowest-latency storage type
- **No CMEK** — Data encrypted with Google-managed keys (add kmsKeyName for CMEK)
- **deletion_protection: true** — Instance cannot be destroyed without explicit override

## When to Use It

Use this preset for production workloads requiring high availability. Add CMEK encryption by setting `kmsKeyName` on each cluster if compliance requires customer-managed keys.
