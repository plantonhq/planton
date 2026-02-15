---
title: "Presets"
description: "Ready-to-deploy configuration presets for Volume"
type: "preset-list"
componentSlug: "volume"
componentTitle: "Volume"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-blank-data"
    rank: "01"
    title: "Blank Data Volume"
    excerpt: "This preset creates an empty Cinder volume for application data storage. The volume is unformatted -- attach it to an instance via `OpenStackVolumeAttach`, then partition and format it from within..."
  - slug: "02-bootable-from-image"
    rank: "02"
    title: "Bootable Volume from Image"
    excerpt: "This preset creates a Cinder volume pre-populated with a Glance image. The resulting volume is bootable and can be used as a root disk for instances. Create the volume first, then reference it in an..."
---

# Volume Presets

Ready-to-deploy configuration presets for Volume. Each preset is a complete manifest you can copy, customize, and deploy.
