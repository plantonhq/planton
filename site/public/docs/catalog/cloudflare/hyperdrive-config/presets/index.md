---
title: "Presets"
description: "Ready-to-deploy configuration presets for Hyperdrive Config"
type: "preset-list"
componentSlug: "hyperdrive-config"
componentTitle: "Hyperdrive Config"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-postgres-basic"
    rank: "01"
    title: "Preset: Basic PostgreSQL Hyperdrive"
    excerpt: "A Hyperdrive config that pools and caches connections to a regional PostgreSQL database, ready to bind to a Worker."
  - slug: "02-postgres-mtls"
    rank: "02"
    title: "Preset: PostgreSQL Hyperdrive with mTLS"
    excerpt: "A Hyperdrive config that connects to a PostgreSQL origin requiring mutual TLS, verifying both the CA and the server hostname."
  - slug: "03-postgres-vpc"
    rank: "03"
    title: "Preset: PostgreSQL Hyperdrive over a Workers VPC Service"
    excerpt: "A Hyperdrive config that reaches a private PostgreSQL origin by egressing through a Workers VPC Service, rather than dialing a publicly reachable host."
---

# Hyperdrive Config Presets

Ready-to-deploy configuration presets for Hyperdrive Config. Each preset is a complete manifest you can copy, customize, and deploy.
