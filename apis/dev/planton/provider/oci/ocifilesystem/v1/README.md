# OciFileSystem

## Overview

OciFileSystem is an Planton component that deploys an OCI File Storage file system together with a dedicated mount target and one or more NFS exports. It provides a single declarative manifest to create a fully functional NFS-accessible file system with per-client access control.

## Purpose

OCI File Storage provides NFS-compatible managed file systems for workloads that need shared, POSIX-compliant storage — application servers, container workloads, CI/CD pipelines, and analytics jobs. This component bundles the file system, mount target, and exports into one resource so that the entire NFS stack is deployed and torn down as a unit.

## Key Features

- **Single-resource deployment** — one manifest creates the file system, mount target, export set configuration, and all NFS exports.
- **Dedicated mount target** — each file system gets its own NFS endpoint with configurable subnet, IP address, hostname, and throughput.
- **Multiple NFS exports** — expose the file system at multiple paths, each with independent access control rules.
- **Per-source access control** — export options define NFS permissions (read/write, identity squashing, privileged ports) per source CIDR.
- **Customer-managed encryption** — optional KMS key for server-side encryption at rest.
- **Snapshot policy attachment** — optional reference to an existing filesystem snapshot policy for automated snapshots.
- **Export set tuning** — configure `maxFsStatBytes` and `maxFsStatFiles` to control NFS capacity reporting via statfs.
- **Network security groups** — associate NSGs with the mount target for NFS traffic control.
- **Foreign key references** — `compartmentId`, `subnetId`, `nsgIds`, `kmsKeyId`, and `filesystemSnapshotPolicyId` support `valueFrom` to reference other Planton-managed resources.

## Constraints

- File system and mount target must be in the same availability domain.
- OCI default service limit is 2 mount targets per availability domain — request a limit increase for more.
- Changing `availabilityDomain`, `mountTarget.subnetId`, `mountTarget.hostnameLabel`, `mountTarget.ipAddress`, or `exports[].path` forces recreation of the affected resources.
- At least one export is required.
- Export paths must start with `/` and be unique within the mount target's export set.

## Use Cases

| Scenario | Configuration |
|----------|---------------|
| Development shared storage | Minimal file system with one export and default access |
| Multi-team shared file system | Multiple exports with per-CIDR access control |
| Application server NFS mount | Fixed IP + DNS hostname for predictable mount point |
| Encrypted shared storage | KMS key for compliance requirements |
| High-throughput workloads | `requestedThroughput` on mount target |

## Production Features

- **Freeform tags** — automatically populated from `metadata.labels`, including `resource_kind`, `resource_id`, `organization`, and `environment`.
- **KMS encryption** — customer-managed keys for encryption at rest when regulatory requirements exceed Oracle-managed defaults.
- **NSG integration** — associate network security groups with the mount target to control NFS traffic at port level (2049/TCP, 111/TCP).
- **Identity squashing** — `root_squash` and `all_squash` modes remap client UIDs/GIDs to limit NFS client privileges on the server.
- **Statfs tuning** — `maxFsStatBytes` and `maxFsStatFiles` control what NFS clients see for available capacity, useful for quota enforcement at the NFS protocol level.
