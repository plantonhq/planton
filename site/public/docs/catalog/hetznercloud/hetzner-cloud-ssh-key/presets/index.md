---
title: "Presets"
description: "Ready-to-deploy configuration presets for Hetzner Cloud SSH Key"
type: "preset-list"
componentSlug: "hetzner-cloud-ssh-key"
componentTitle: "Hetzner Cloud SSH Key"
provider: "hetznercloud"
icon: "package"
order: 200
presets:
  - slug: "01-import-public-key"
    rank: "01"
    title: "Import Existing Public Key"
    excerpt: "This preset registers an existing SSH public key in Hetzner Cloud so it can be injected into servers at creation time. Generate a keypair locally with `ssh-keygen` and import only the public half --..."
---

# Hetzner Cloud SSH Key Presets

Ready-to-deploy configuration presets for Hetzner Cloud SSH Key. Each preset is a complete manifest you can copy, customize, and deploy.
