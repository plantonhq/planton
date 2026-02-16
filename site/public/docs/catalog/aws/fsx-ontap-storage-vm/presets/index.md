---
title: "Presets"
description: "Ready-to-deploy configuration presets for FSx ONTAP Storage VM"
type: "preset-list"
componentSlug: "fsx-ontap-storage-vm"
componentTitle: "FSx ONTAP Storage VM"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-nfs-unix"
    rank: "01"
    title: "NFS-Only UNIX SVM"
    excerpt: "Basic NFS-only Storage Virtual Machine with UNIX security style. No Active Directory, no SMB. The simplest and most common SVM configuration for Linux/NFS workloads."
  - slug: "02-smb-windows"
    rank: "02"
    title: "SMB Windows SVM with Active Directory"
    excerpt: "Windows-focused Storage Virtual Machine with NTFS security style and Active Directory domain join. Enables SMB/CIFS file share access with Windows ACLs and identity-based access control."
  - slug: "03-multiprotocol"
    rank: "03"
    title: "Multiprotocol SVM (NFS + SMB)"
    excerpt: "Dual-protocol Storage Virtual Machine with MIXED security style and Active Directory. Provides both NFS and SMB access to the same volumes, enabling mixed Linux/Windows environments to share data."
---

# FSx ONTAP Storage VM Presets

Ready-to-deploy configuration presets for FSx ONTAP Storage VM. Each preset is a complete manifest you can copy, customize, and deploy.
