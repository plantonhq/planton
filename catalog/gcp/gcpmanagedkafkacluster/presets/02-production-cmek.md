# Production with CMEK

## Use Case

A production event backbone: 12 vCPU and 48 GiB, a 1 TB broker disk, reachable from two VPC networks (its own and a Shared VPC host), encrypted with your key, rebalanced on scale-up, and protected from destroy.

## When to Use

- Production streams with customer-managed encryption requirements
- Clients in more than one VPC network
- Clusters a destroy must never remove by accident

## What This Creates

- A 12-vCPU, 48 GiB cluster with 1000 GiB per broker, attached to two networks, CMEK-encrypted, `PREVENT` on destroy

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `kmsKey` | `events` key | Must be in the cluster's region; grant the Managed Kafka service agent encrypt/decrypt first. Fixed at creation. |
| `capacityConfig` | 12 vCPU, 48 GiB | Size to peak throughput; scales in place. |
| `networkConfigs` | two subnets | One subnet per network your clients use, up to ten. |
