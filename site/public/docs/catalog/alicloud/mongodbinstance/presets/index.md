---
title: "Presets"
description: "Ready-to-deploy configuration presets for MongodbInstance"
type: "preset-list"
componentSlug: "mongodbinstance"
componentTitle: "MongodbInstance"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-development"
    rank: "01"
    title: "Preset: Development MongoDB Instance"
    excerpt: "A minimal MongoDB 7.0 replica-set instance for development and testing."
  - slug: "02-production-ha"
    rank: "02"
    title: "Preset: Production HA MongoDB Instance"
    excerpt: "A production MongoDB replica set deployed across three availability zones with read replicas, backup policies, and operational safeguards."
  - slug: "03-encrypted-compliance"
    rank: "03"
    title: "Preset: Encrypted Compliance MongoDB Instance"
    excerpt: "A security-hardened MongoDB instance with TDE encryption, SSL, subscription billing, and daily backups for compliance-sensitive workloads."
---

# MongodbInstance Presets

Ready-to-deploy configuration presets for MongodbInstance. Each preset is a complete manifest you can copy, customize, and deploy.
