---
title: "Presets"
description: "Ready-to-deploy configuration presets for Virtual Machine"
type: "preset-list"
componentSlug: "virtual-machine"
componentTitle: "Virtual Machine"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-ubuntu-ssh"
    rank: "01"
    title: "Ubuntu 22.04 LTS with SSH Key Authentication"
    excerpt: "This preset deploys an Ubuntu 22.04 LTS Gen2 VM with SSH key authentication, no public IP, boot diagnostics enabled, and a 30 GB Premium SSD OS disk. This is the standard configuration for secure..."
  - slug: "02-windows-rdp"
    rank: "02"
    title: "Windows Server 2022 with RDP Access"
    excerpt: "This preset deploys a Windows Server 2022 Datacenter Gen2 VM with password authentication, a public IP for RDP access, boot diagnostics enabled, and a 128 GB Premium SSD OS disk. This configuration..."
---

# Virtual Machine Presets

Ready-to-deploy configuration presets for Virtual Machine. Each preset is a complete manifest you can copy, customize, and deploy.
