---
title: "Presets"
description: "Ready-to-deploy configuration presets for SSH Key"
type: "preset-list"
componentSlug: "ssh-key"
componentTitle: "SSH Key"
provider: "hetznercloud"
icon: "package"
order: 200
presets:
  - slug: "01-import-public-key"
    rank: "01"
    title: "Import Existing Public Key"
    excerpt: "This preset registers an existing SSH public key in Hetzner Cloud so it can be injected into servers at creation time. Generate a keypair locally with `ssh-keygen` and import only the public half --..."
---

# SSH Key Presets

Ready-to-deploy configuration presets for SSH Key. Each preset is a complete manifest you can copy, customize, and deploy.
