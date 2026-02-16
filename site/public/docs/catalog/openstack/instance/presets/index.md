---
title: "Presets"
description: "Ready-to-deploy configuration presets for Instance"
type: "preset-list"
componentSlug: "instance"
componentTitle: "Instance"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-standard-vm"
    rank: "01"
    title: "Standard VM Instance"
    excerpt: "This preset creates a compute instance booting from an image with a specified flavor, keypair, and network. The root disk is ephemeral (lives on the hypervisor). This is the simplest and most common..."
  - slug: "02-boot-from-volume"
    rank: "02"
    title: "Boot-from-Volume Instance"
    excerpt: "This preset creates a compute instance that boots from a Cinder volume instead of an ephemeral disk. The root volume is created from a Glance image and persists independently of the instance (unless..."
---

# Instance Presets

Ready-to-deploy configuration presets for Instance. Each preset is a complete manifest you can copy, customize, and deploy.
