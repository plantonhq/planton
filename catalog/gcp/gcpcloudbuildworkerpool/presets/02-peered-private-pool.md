# Peered Private Pool

## Use Case

Run builds inside your VPC so they reach private services -- a package index, an artifact store, a database to migrate, a private GKE control plane -- with no public IPs on the workers.

## When to Use

- Builds that must reach resources with no public endpoint
- Organizations that forbid public IPs on build machines

## What This Creates

- The Cloud Build API on the project
- A private pool peered into the `ci-vpc` network, taking a /26 for its workers, with no external IPs

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `networkConfig.peeredNetwork` | `ci-vpc` reference | The network to peer into; it needs private services access (a `GcpServiceNetworkingConnection`) first. |
| `networkConfig.peeredNetworkIpRange` | `/26` | Size the block for the number of concurrent builds, or pin it with an address (`192.168.0.0/26`). |
| `workerConfig.noExternalIp` | `true` | Set false only if builds need direct internet egress and the network has no Cloud NAT. |
