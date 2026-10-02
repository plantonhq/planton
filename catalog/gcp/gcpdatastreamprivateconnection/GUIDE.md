# GcpDatastreamPrivateConnection Guide

The judgment this guide protects: a private connection is network plumbing shared by every source in a VPC, so it belongs to whoever owns the network, and it is chosen once -- nearly every field is immutable.

## Peering or PSC interface

- **`vpcPeeringConfig`** -- Datastream's network peers with your VPC through a /29 you set aside. Datastream then reaches anything your VPC routes to: VMs, and on-premises ranges over VPN or Interconnect. Pick a range no subnet, peering, or route uses today or will tomorrow.
- **`pscInterfaceConfig`** -- Datastream plugs into a network attachment you own. There is no range to reserve and no peering to manage. Use it when every range is taken, when policy forbids peering, or when your network already standardizes on PSC.

## Private Cloud SQL and AlloyDB

Peering is not transitive. A private Cloud SQL or AlloyDB instance lives in a network Google peers with your VPC, so Datastream's peering cannot reach it directly. Google's pattern is a small proxy VM in your VPC (the profile's hostname is the VM) or a PSC interface. For Cloud SQL, the simpler path is often its public IP with Datastream's regional IPs allowlisted, and no private connection at all.

## Lifecycle

Delete the profiles that use a connection before the connection. The default `deletionPolicy` of `FORCE` also removes the routes Datastream created on it.
