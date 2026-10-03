# Production Enterprise

The production posture: a regional tier that survives zone failures,
deletion protection as the destroy guard, private-services networking,
and locked-down NFS exports.

## What this preset creates

An ENTERPRISE instance in a region (not a zone — regional tiers take a
region for `location`), exporting one share `data` that only RFC1918
clients (`10.0.0.0/8`) can mount, with root squashed to the anonymous
user. `deletionProtectionEnabled: true` means no teardown can destroy
the instance until the flag is deliberately flipped false in a
reviewable change. The VPC arrives as a reference to a `GcpVpcNetwork`
resource, and `PRIVATE_SERVICE_ACCESS` rides the VPC's existing
service-networking connection — the mode Shared VPC consumers require.

## Prerequisites

- A `GcpVpcNetwork` named `prod-vpc` (replace with yours, or set a
  literal `value`).
- An existing service-networking connection on that VPC —
  `PRIVATE_SERVICE_ACCESS` uses it; this kind does not create it.

## Remix ideas

- Add `kmsKeyName` (a `GcpKmsKey` reference) for CMEK-at-rest.
- Add `protocol: NFS_V4_1` — supported on ENTERPRISE — where NFSv4.1
  semantics or Kerberos matter.
- Stand up a DR replica: create a second instance with
  `initialReplication` whose `peerInstances` references this one (its
  `instance_id` output is exactly the path the API wants). The replica
  is the STANDBY; replication is fixed at create time.
- Tighten `nfsExportOptions.ipRanges` to specific subnets, or add a
  second READ_ONLY export for consumers that should never write.
