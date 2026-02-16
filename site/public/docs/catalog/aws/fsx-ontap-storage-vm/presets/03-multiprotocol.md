---
title: "Multiprotocol SVM (NFS + SMB)"
description: "Dual-protocol Storage Virtual Machine with MIXED security style and Active Directory. Provides both NFS and SMB access to the same volumes, enabling mixed Linux/Windows environments to share data."
type: "preset"
rank: "03"
presetSlug: "03-multiprotocol"
componentSlug: "fsx-ontap-storage-vm"
componentTitle: "FSx ONTAP Storage VM"
provider: "aws"
icon: "package"
order: 3
---

# Multiprotocol SVM (NFS + SMB)

Dual-protocol Storage Virtual Machine with MIXED security style and Active Directory. Provides both NFS and SMB access to the same volumes, enabling mixed Linux/Windows environments to share data.

## When to Use

- Mixed OS environments where both Linux and Windows clients access the same data
- Migration scenarios transitioning from NFS to SMB (or vice versa)
- Engineering teams with Linux build servers and Windows developer workstations
- VMware environments with both Linux and Windows guest VMs
- Any workload requiring simultaneous NFS and SMB access

## What It Configures

- **MIXED security style** — Both UNIX and NTFS permissions supported. The effective security style depends on which protocol last modified permissions on a given file
- **Active Directory** — Full AD configuration with custom administrators group
- **SVM admin password** — vsadmin SSH access enabled
- **All four endpoints** — iSCSI, management, NFS, and SMB all available

## What to Customize

- Replace all `<REPLACE>` placeholders with actual values
- **Critical**: Replace password placeholders with real credentials
- Adjust AD configuration (`dns_ips`, `domain_name`, `organizational_unit_distinguished_name`)
- Change `file_system_administrators_group` from "FSx Admins" to your AD group
- Remove `svm_admin_password` if SVM CLI access is not needed

## Important Note on MIXED Security Style

With MIXED security style, the effective permission model depends on which protocol last modified the file or directory. This can lead to unexpected behavior if not understood:
- Files created via NFS use UNIX permissions
- Files created via SMB use NTFS ACLs
- The last protocol to set permissions "wins"

For predictable behavior, most production deployments use separate SVMs: one UNIX (NFS-only) and one NTFS (SMB-only). Use MIXED only when dual-protocol access to the same data is a hard requirement.
