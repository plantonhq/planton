---
title: "Development Instance"
description: "This preset creates a small Scaleway Instance with a public IP for quick development and testing. It uses the DEV1-S type (2 vCPU, 2 GB RAM) with Ubuntu 22.04 -- the most affordable and commonly used..."
type: "preset"
rank: "01"
presetSlug: "01-dev-instance"
componentSlug: "instance"
componentTitle: "Instance"
provider: "scaleway"
icon: "package"
order: 1
---

# Development Instance

This preset creates a small Scaleway Instance with a public IP for quick development and testing. It uses the DEV1-S type (2 vCPU, 2 GB RAM) with Ubuntu 22.04 -- the most affordable and commonly used starting point on Scaleway.

## When to Use

- Quick development or prototyping environments
- Learning and experimentation with Scaleway
- Standalone services that need direct public access (e.g., a dev web server)

## Key Configuration Choices

- **DEV1-S instance** (`type: DEV1-S`) -- 2 vCPU, 2 GB RAM, 20 GB local SSD; the smallest general-purpose instance
- **Ubuntu 22.04** (`image: ubuntu_jammy`) -- widely supported LTS release; change to `debian_bookworm`, `centos_stream_9`, or a custom image UUID as needed
- **Public IP assigned** (`publicIp: {}`) -- the instance is directly reachable from the internet; suitable for development but not recommended for production
- **Started state** (`state: started`) -- the instance boots immediately after creation
- **No Private Network** -- omitted for simplicity; add `privateNetworkId` for environments with Private Network connectivity
- **No security group** -- uses Scaleway's default (allow-all); add `securityGroupId` for production

## Placeholders to Replace

No placeholders -- this preset is ready to deploy as-is. Customize `type`, `image`, and `zone` to match your requirements.

## Related Presets

- **02-production-private** -- Use instead for production workloads that should be on a Private Network without a public IP
