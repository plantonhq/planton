---
title: "Presets"
description: "Ready-to-deploy configuration presets for Hetzner Cloud Certificate"
type: "preset-list"
componentSlug: "hetzner-cloud-certificate"
componentTitle: "Hetzner Cloud Certificate"
provider: "hetznercloud"
icon: "package"
order: 200
presets:
  - slug: "01-managed-lets-encrypt"
    rank: "01"
    title: "Managed Let's Encrypt Certificate"
    excerpt: "This preset creates a Hetzner Cloud managed certificate that automatically obtains and renews a TLS certificate from Let's Encrypt. You specify one or more domain names, and Hetzner Cloud handles..."
  - slug: "02-uploaded-certificate"
    rank: "02"
    title: "Uploaded Certificate"
    excerpt: "This preset uploads a user-provided TLS certificate and private key to Hetzner Cloud. You supply PEM-encoded files, and Hetzner Cloud stores them for use by load balancer HTTPS services. Unlike the..."
---

# Hetzner Cloud Certificate Presets

Ready-to-deploy configuration presets for Hetzner Cloud Certificate. Each preset is a complete manifest you can copy, customize, and deploy.
