---
title: "Presets"
description: "Ready-to-deploy configuration presets for File System"
type: "preset-list"
componentSlug: "file-system"
componentTitle: "File System"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-shared-application-storage"
    rank: "01"
    title: "Shared Application Storage"
    excerpt: "This preset creates an NFS file system with a single export path for shared application data. Access is restricted to the subnet CIDR with root squash for basic security. This is the standard pattern..."
  - slug: "02-restricted-multi-export"
    rank: "02"
    title: "Restricted Multi-Export"
    excerpt: "This preset creates an NFS file system with two export paths serving different purposes: `/app-data` for read-write application data and `/logs` with split access -- read-write for the application..."
---

# File System Presets

Ready-to-deploy configuration presets for File System. Each preset is a complete manifest you can copy, customize, and deploy.
