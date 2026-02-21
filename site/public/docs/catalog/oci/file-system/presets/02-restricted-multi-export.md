---
title: "Restricted Multi-Export"
description: "This preset creates an NFS file system with two export paths serving different purposes: `/app-data` for read-write application data and `/logs` with split access -- read-write for the application..."
type: "preset"
rank: "02"
presetSlug: "02-restricted-multi-export"
componentSlug: "file-system"
componentTitle: "File System"
provider: "oci"
icon: "package"
order: 2
---

# Restricted Multi-Export

This preset creates an NFS file system with two export paths serving different purposes: `/app-data` for read-write application data and `/logs` with split access -- read-write for the application subnet and read-only for a monitoring subnet. The mount target is protected by an NSG and the file system uses KMS encryption, making this suitable for production workloads with strict access separation.

## When to Use

- Production environments where application data and logs need separate NFS paths with different access policies
- Architectures with monitoring or log aggregation systems that need read-only access to application logs
- Security-conscious deployments requiring NSG-protected NFS endpoints and customer-managed encryption
- Multi-tier applications where different subnets need different levels of file system access

## Key Configuration Choices

- **Two export paths** -- `/app-data` and `/logs` provide logical separation. Application servers mount both; monitoring systems mount only `/logs` with read-only access.
- **Split access on /logs** -- the application subnet gets read-write access (applications write logs), while the monitoring subnet gets read-only access (log collectors read logs). This enforces least-privilege access at the NFS level.
- **All squash for monitoring clients** (`identitySquash: all_squash` on `/logs` monitoring rule) -- all UIDs from monitoring clients are mapped to nobody, ensuring monitoring systems cannot escalate privileges even if compromised.
- **Root squash for application clients** (`identitySquash: root_squash`) -- root on application servers is remapped to nobody, preventing accidental or malicious root-level changes to file ownership.
- **NSG-protected mount target** (`nsgIds`) -- restricts NFS traffic (ports 2049/TCP and 111/TCP) to authorized security group members only.
- **KMS encryption** (`kmsKeyId`) -- customer-managed encryption key for data at rest. Provides key rotation control and meets compliance requirements.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment for the file system | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<availability-domain>` | AD for the file system and mount target | OCI Console > Compute > Availability Domains |
| `<kms-key-ocid>` | OCID of the KMS encryption key | OCI Console > Security > Vault > Keys, or `OciKmsKey` outputs |
| `<private-subnet-ocid>` | OCID of the private subnet for the mount target | OCI Console > Networking > Subnets, or `OciSubnet` outputs |
| `<nfs-nsg-ocid>` | OCID of the NSG allowing NFS traffic (ports 2049, 111) | OCI Console > Networking > NSGs, or `OciSecurityGroup` outputs |
| `<app-subnet-cidr>` | CIDR of the application subnet (e.g., `10.0.1.0/24`) | OCI Console > Networking > Subnets |
| `<monitoring-subnet-cidr>` | CIDR of the monitoring/log-collector subnet (e.g., `10.0.2.0/24`) | OCI Console > Networking > Subnets |

## Related Presets

- **01-shared-application-storage** -- Use instead for simpler setups with a single export and no access separation
