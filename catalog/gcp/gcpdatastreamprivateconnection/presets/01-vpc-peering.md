# VPC Peering

## Use Case

Let Datastream reach databases with no public address by peering its network with your VPC. Every connection profile in the region that needs the VPC shares this one link.

## When to Use

- Database VMs or on-premises databases reached over VPN or Interconnect
- Sources a proxy VM in your VPC fronts (Google's pattern for a private Cloud SQL or AlloyDB instance)

## What This Creates

- A private connection in `us-central1` peered with the `data-vpc` network through `10.200.0.0/29`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `vpcPeeringConfig.subnet` | `10.200.0.0/29` | Pick a free /29 no subnet, peering, or route in the VPC uses. |
| `vpcPeeringConfig.vpc` | `GcpVpcNetwork` reference | Point at the network your databases are reachable from. |
