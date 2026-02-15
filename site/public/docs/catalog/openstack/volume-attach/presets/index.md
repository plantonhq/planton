---
title: "Presets"
description: "Ready-to-deploy configuration presets for Volume Attach"
type: "preset-list"
componentSlug: "volume-attach"
componentTitle: "Volume Attach"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Volume Attachment"
    excerpt: "This preset attaches a Cinder volume to a compute instance. The volume appears as a block device inside the instance (e.g., `/dev/vdb`). The device path is auto-assigned by Nova -- add `device` to..."
---

# Volume Attach Presets

Ready-to-deploy configuration presets for Volume Attach. Each preset is a complete manifest you can copy, customize, and deploy.
