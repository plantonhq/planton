---
title: "Presets"
description: "Ready-to-deploy configuration presets for Image"
type: "preset-list"
componentSlug: "image"
componentTitle: "Image"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-cloud-image-from-url"
    rank: "01"
    title: "Cloud Image from URL"
    excerpt: "This preset imports a cloud image into Glance from a URL. The image is downloaded by the Glance service and stored in its backend (Ceph, Swift, filesystem). Most Linux cloud images are distributed as..."
---

# Image Presets

Ready-to-deploy configuration presets for Image. Each preset is a complete manifest you can copy, customize, and deploy.
