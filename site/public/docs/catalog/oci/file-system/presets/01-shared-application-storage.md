---
title: "Shared Application Storage"
description: "This preset creates an NFS file system with a single export path for shared application data. Access is restricted to the subnet CIDR with root squash for basic security. This is the standard pattern..."
type: "preset"
rank: "01"
presetSlug: "01-shared-application-storage"
componentSlug: "file-system"
componentTitle: "File System"
provider: "oci"
icon: "package"
order: 1
---

# Shared Application Storage

This preset creates an NFS file system with a single export path for shared application data. Access is restricted to the subnet CIDR with root squash for basic security. This is the standard pattern for shared storage across compute instances, containers, or OKE pods that need a POSIX-compatible filesystem.

## When to Use

- Shared storage for application deployments where multiple compute instances or containers need access to the same files
- OKE persistent volumes using the OCI File Storage CSI driver
- Content management systems, media processing pipelines, or any workload requiring shared file access
- Development environments where multiple team members need shared file storage

## Key Configuration Choices

- **Single export path** (`/shared`) -- one NFS export serving as the root of shared application data. Clients mount it as `mount -t nfs <mount-target-ip>:/shared /mnt/shared`.
- **Subnet CIDR access** (`source: <subnet-cidr>`) -- restricts NFS access to clients within the specified subnet. Replace with the actual CIDR (e.g., `10.0.1.0/24`) of the subnet where your compute instances reside.
- **Root squash** (`identitySquash: root_squash`) -- remaps root (UID 0) on NFS clients to the nobody user (UID 65534) on the server. This prevents root on a client from having root privileges on the file system, providing a basic security boundary.
- **Privileged source port required** (`requirePrivilegedSourcePort: true`) -- only connections from ports below 1024 are accepted, which is the standard NFS security practice on UNIX systems.
- **DNS hostname label** (`hostnameLabel: sharedfs`) -- enables DNS-based mount commands (e.g., `mount -t nfs sharedfs.subnet.vcn.oraclevcn.com:/shared /mnt/shared`) instead of IP-based mounts.
- **Default throughput** -- OCI provides a baseline throughput tier. Add `requestedThroughput` if the workload requires higher NFS throughput.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment for the file system | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<availability-domain>` | AD for the file system and mount target (e.g., `Uocm:US-ASHBURN-AD-1`) | OCI Console > Compute > Availability Domains |
| `<private-subnet-ocid>` | OCID of the private subnet for the mount target | OCI Console > Networking > Subnets, or `OciSubnet` outputs |
| `<subnet-cidr>` | CIDR block of the subnet where NFS clients reside (e.g., `10.0.1.0/24`) | OCI Console > Networking > Subnets > CIDR Block |

## Related Presets

- **02-restricted-multi-export** -- Use instead when multiple export paths with differentiated access controls are needed
