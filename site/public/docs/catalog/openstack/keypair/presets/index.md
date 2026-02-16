---
title: "Presets"
description: "Ready-to-deploy configuration presets for Keypair"
type: "preset-list"
componentSlug: "keypair"
componentTitle: "Keypair"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-import-public-key"
    rank: "01"
    title: "Import Existing Public Key"
    excerpt: "This preset imports an existing SSH public key into OpenStack. This is the recommended approach for production -- generate a keypair locally with `ssh-keygen` and import only the public key. The..."
---

# Keypair Presets

Ready-to-deploy configuration presets for Keypair. Each preset is a complete manifest you can copy, customize, and deploy.
