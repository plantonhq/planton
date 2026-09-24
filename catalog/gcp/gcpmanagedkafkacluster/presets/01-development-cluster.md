# Development Cluster

## Use Case

The smallest Kafka cluster Google sells, for development and integration testing: 3 vCPU and 3 GiB on one subnet.

## When to Use

- Building and testing producers and consumers
- Short-lived environments
- Learning the service before sizing production

## What This Creates

- A 3-vCPU, 3 GiB cluster in `us-central1`, reachable from the `default` subnet's network

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us-central1` | Put the cluster where its clients run. Fixed at creation. |
| `capacityConfig` | 3 vCPU, 3 GiB | Raise vCPUs for throughput; memory 1-8 GiB per vCPU. |
| `networkConfigs` | `default` subnet | Point at the subnet whose network your clients live in. |
