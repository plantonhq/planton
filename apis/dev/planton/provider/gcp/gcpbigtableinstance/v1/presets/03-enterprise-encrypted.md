# Enterprise Encrypted

Multi-cluster production instance with CMEK encryption, aggressive autoscaling, and storage utilization targets.

## Description

This preset provisions an enterprise-grade Bigtable instance with two CMEK-encrypted clusters, autoscaling from 5 to 50 nodes, and explicit storage utilization targets. All data at rest is encrypted with customer-managed Cloud KMS keys. This configuration is suitable for regulated industries and workloads requiring compliance with data sovereignty or encryption requirements.

## Use Case

- Regulated industries (finance, healthcare) requiring customer-managed encryption
- Large-scale production workloads with predictable growth patterns
- Data platforms requiring explicit storage utilization management
- Enterprise environments with centralized key management

## What This Preset Configures

- **Two clusters** — Automatic replication across two zones
- **CMEK encryption** — Customer-managed Cloud KMS key on both clusters
- **Aggressive autoscaling** — 5 to 50 nodes per cluster, targeting 60% CPU and 4096 GB storage per node
- **SSD storage** — Default, lowest-latency storage type
- **deletion_protection: true** — Instance cannot be destroyed without explicit override

## When to Use It

Use this preset for enterprise workloads requiring CMEK encryption and large-scale capacity. Ensure the Cloud KMS key exists in the same region as the cluster zones and that the Bigtable service account has the `cloudkms.cryptoKeyEncrypterDecrypter` role.
