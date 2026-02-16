---
title: "Development Standalone MongoDB"
description: "This preset creates a single-node Scaleway MongoDB instance using the smallest available node type. It is the most affordable path to a working MongoDB database for development, testing, and..."
type: "preset"
rank: "01"
presetSlug: "01-dev-standalone"
componentSlug: "mongodb-instance"
componentTitle: "MongoDB Instance"
provider: "scaleway"
icon: "package"
order: 1
---

# Development Standalone MongoDB

This preset creates a single-node Scaleway MongoDB instance using the smallest available node type. It is the most affordable path to a working MongoDB database for development, testing, and prototyping.

## When to Use

- Development and testing environments
- Small applications with light document database needs
- Learning MongoDB on Scaleway

## Key Configuration Choices

- **MongoDB 7.0** (`version: 7.0.12`) -- latest stable version with native query improvements and Atlas-compatible features
- **MGDB-PLAY2-NANO node** (`nodeType: MGDB-PLAY2-NANO`) -- the smallest and most affordable MongoDB node
- **Single node** (`nodeNumber: 1`) -- no replica set; acceptable for non-critical environments
- **SBS 5k storage** (`volumeType: sbs_5k`) -- network-attached block storage with 5,000 IOPS baseline
- **10 GB volume** (`volumeSizeInGb: 10`) -- starting size; must be a multiple of 5 GB
- **No Private Network** -- accessible via public endpoint; add `privateNetworkId` for production

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-admin-user>` | Database admin username (max 63 characters) | Choose a username |
| `<your-admin-password>` | Database admin password (min 8 characters) | Generate a strong password |

## Related Presets

- **02-production-replica-set** -- Use instead for production with a 3-node replica set and Private Network connectivity
